# Godot MovieWriter 视频录制原理分析

## 概述

MovieWriter 是 Godot 引擎的视频录制系统，支持非实时高质量视频录制。它可以录制游戏运行过程中的视频和音频，并输出为多种格式的文件。

## 1. 整体架构

### 1.1 类层次结构

MovieWriter 采用抽象工厂模式设计：

```cpp
// 抽象基类
class MovieWriter : public Object {
    // 核心接口方法
    virtual Error write_begin(const Size2i &p_movie_size, uint32_t p_fps, const String &p_base_path);
    virtual Error write_frame(const Ref<Image> &p_image, const int32_t *p_audio_data);
    virtual void write_end();
    virtual bool handles_file(const String &p_path) const;
};

// 具体实现类
class MovieWriterMJPEG : public MovieWriter;  // AVI 格式录制器
class MovieWriterPNGWAV : public MovieWriter;  // PNG 序列录制器
```

### 1.2 工厂注册机制

```cpp
// 在 servers/register_server_types.cpp 中注册
writer_mjpeg = memnew(MovieWriterMJPEG);
MovieWriter::add_writer(writer_mjpeg);

writer_pngwav = memnew(MovieWriterPNGWAV);
MovieWriter::add_writer(writer_pngwav);
```

### 1.3 录制器选择

通过文件扩展名自动选择合适的录制器：

```cpp
MovieWriter *MovieWriter::find_writer_for_file(const String &p_file) {
    for (int32_t i = writer_count - 1; i >= 0; i--) {
        if (writers[i]->handles_file(p_file)) {
            return writers[i];
        }
    }
    return nullptr;
}
```

## 2. 录制流程

### 2.1 初始化阶段

在 `main.cpp` 的 `Main::setup2()` 中检测录制参数：

```cpp
if (Engine::get_singleton()->get_write_movie_path() != String()) {
    movie_writer = MovieWriter::find_writer_for_file(Engine::get_singleton()->get_write_movie_path());
    if (movie_writer == nullptr) {
        ERR_PRINT("Can't find movie writer for file type, aborting: " + Engine::get_singleton()->get_write_movie_path());
        Engine::get_singleton()->set_write_movie_path(String());
    }
}
```

### 2.2 录制循环

在主循环 `Main::iteration()` 的每一帧末尾：

```cpp
if (movie_writer) {
    movie_writer->add_frame();
}
```

### 2.3 关键方法：add_frame()

```cpp
void MovieWriter::add_frame() {
    // 1. 获取当前帧图像
    RID main_vp_rid = RenderingServer::get_singleton()->viewport_find_from_screen_attachment(DisplayServer::MAIN_WINDOW_ID);
    RID main_vp_texture = RenderingServer::get_singleton()->viewport_get_texture(main_vp_rid);
    Ref<Image> vp_tex = RenderingServer::get_singleton()->texture_2d_get(main_vp_texture);
    
    // 2. HDR 格式转换
    if (RenderingServer::get_singleton()->viewport_is_using_hdr_2d(main_vp_rid)) {
        vp_tex->convert(Image::FORMAT_RGBA8);
        vp_tex->linear_to_srgb();
    }
    
    // 3. 收集性能数据
    RenderingServer::get_singleton()->viewport_set_measure_render_time(main_vp_rid, true);
    cpu_time += RenderingServer::get_singleton()->viewport_get_measured_render_time_cpu(main_vp_rid);
    gpu_time += RenderingServer::get_singleton()->viewport_get_measured_render_time_gpu(main_vp_rid);
    
    // 4. 获取音频数据
    AudioDriverDummy::get_dummy_singleton()->mix_audio(mix_rate / fps, audio_mix_buffer.ptr());
    
    // 5. 写入帧数据
    write_frame(vp_tex, audio_mix_buffer.ptr());
}
```

## 3. 数据采集机制

### 3.1 图像采集

**采集流程**：
1. 通过 `RenderingServer` 获取主窗口的 viewport
2. 获取 viewport 的纹理数据
3. 将纹理转换为 `Image` 对象
4. 处理 HDR 格式转换（如果需要）

**关键代码**：
```cpp
RID main_vp_rid = RenderingServer::get_singleton()->viewport_find_from_screen_attachment(DisplayServer::MAIN_WINDOW_ID);
RID main_vp_texture = RenderingServer::get_singleton()->viewport_get_texture(main_vp_rid);
Ref<Image> vp_tex = RenderingServer::get_singleton()->texture_2d_get(main_vp_texture);
```

### 3.2 音频采集

**AudioDriverDummy 的作用**：
- 替代真实的音频驱动，用于录制模式
- 不播放声音，只收集音频数据
- 支持多声道配置

**音频处理流程**：
```cpp
// 设置音频参数
mix_rate = get_audio_mix_rate();
AudioDriverDummy::get_dummy_singleton()->set_mix_rate(mix_rate);
AudioDriverDummy::get_dummy_singleton()->set_speaker_mode(AudioDriver::SpeakerMode(get_audio_speaker_mode()));

// 计算音频块大小
audio_channels = AudioDriverDummy::get_dummy_singleton()->get_channels();
audio_mix_buffer.resize(mix_rate * audio_channels / fps);

// 每帧获取音频数据
AudioDriverDummy::get_dummy_singleton()->mix_audio(mix_rate / fps, audio_mix_buffer.ptr());
```

## 4. 非实时录制特性

### 4.1 固定帧率机制

MovieWriter 的核心特色是**非实时录制**：

- 通过 `--fixed-fps` 参数强制设置固定帧率
- 每帧的 delta 时间保持一致，不受硬件性能影响
- 录制过程可能比实时播放慢，但输出视频质量稳定

### 4.2 时间控制

```cpp
// 在 Main::begin() 中设置
if (fixed_fps != -1) {
    Engine::get_singleton()->set_frame_delay(1.0 / fixed_fps * 1000000); // 微秒
}
```

### 4.3 音频同步

```cpp
// 确保音频采样率能被帧率整除
if ((mix_rate % fps) != 0) {
    WARN_PRINT("MovieWriter's audio mix rate (" + itos(mix_rate) + ") can not be divided by the recording FPS (" + itos(fps) + "). Audio may go out of sync over time.");
}
```

## 5. 具体实现详解

### 5.1 MJPEG 录制器 (MovieWriterMJPEG)

**特点**：
- 输出 AVI 格式文件
- 使用 MJPEG 压缩视频
- 音频为未压缩的 PCM 格式
- 文件大小中等，编码速度快

**实现细节**：
```cpp
Error MovieWriterMJPEG::write_frame(const Ref<Image> &p_image, const int32_t *p_audio_data) {
    // 将图像压缩为 JPEG
    Vector<uint8_t> jpg_buffer = p_image->save_jpg_to_buffer(quality);
    
    // 写入 AVI 文件结构
    f->store_buffer((const uint8_t *)"00db", 4); // 视频流标识
    f->store_32(jpg_buffer.size());
    f->store_buffer(jpg_buffer.ptr(), jpg_buffer.size());
    
    // 处理字节对齐
    if (jpg_buffer.size() & 1) {
        f->store_8(0);
    }
    
    // 写入音频数据
    f->store_buffer((const uint8_t *)"01wb", 4); // 音频流标识
    f->store_32(audio_block_size);
    f->store_buffer((const uint8_t *)p_audio_data, audio_block_size);
}
```

**AVI 文件结构**：
- RIFF 头
- AVI 头信息
- 视频流信息
- 音频流信息
- 数据块（交替存储视频和音频）
- 索引表

### 5.2 PNG+WAV 录制器 (MovieWriterPNGWAV)

**特点**：
- 输出 PNG 图像序列 + WAV 音频文件
- 无损压缩，质量最高
- 文件大小较大，编码速度慢
- 适合后期处理

**实现细节**：
```cpp
Error MovieWriterPNGWAV::write_frame(const Ref<Image> &p_image, const int32_t *p_audio_data) {
    // 保存 PNG 图像文件
    Vector<uint8_t> png_buffer = p_image->save_png_to_buffer();
    Ref<FileAccess> fi = FileAccess::open(base_path + zeros_str(frame_count) + ".png", FileAccess::WRITE);
    fi->store_buffer(png_buffer.ptr(), png_buffer.size());
    
    // 将音频数据写入 WAV 文件
    f_wav->store_buffer((const uint8_t *)p_audio_data, audio_block_size);
    
    frame_count++;
}
```

**文件命名规则**：
- PNG 文件：`output_00000.png`, `output_00001.png`, ...
- WAV 文件：`output.wav`

## 6. 使用方式

### 6.1 命令行参数

```bash
# 录制为 AVI 格式（MJPEG）
godot --write-movie output.avi --fixed-fps 60

# 录制为 PNG 序列
godot --write-movie output.png --fixed-fps 30

# 结合其他参数
godot --write-movie video.avi --fixed-fps 30 --headless
```

### 6.2 项目设置

在 `project.godot` 中可以配置：

```ini
[editor/movie_writer]
mix_rate = 48000            # 音频采样率 (Hz)
speaker_mode = 0            # 音频声道模式 (0=立体声)
mjpeg_quality = 0.75        # MJPEG 压缩质量 (0.01-1.0)
movie_file = ""             # 默认输出文件
disable_vsync = false       # 是否禁用垂直同步
fps = 60                    # 录制帧率
```

### 6.3 音频声道配置

```cpp
enum SpeakerMode {
    SPEAKER_MODE_STEREO = 0,        // 2 声道
    SPEAKER_SURROUND_31 = 1,        // 4 声道
    SPEAKER_SURROUND_51 = 2,        // 6 声道
    SPEAKER_SURROUND_71 = 3,        // 8 声道
};
```

## 7. 性能监控

### 7.1 录制统计

MovieWriter 会收集并报告详细的性能统计：

```cpp
void MovieWriter::end() {
    // 输出录制统计信息
    print_line(vformat("%d frames at %d FPS (movie length: %s), recorded in %s (%d%% of real-time speed).", 
        Engine::get_singleton()->get_frames_drawn(), fps, movie_time, real_time, 
        (float(movie_time_seconds) / real_time_seconds) * 100));
    
    print_line(vformat("CPU time: %.2f seconds (average: %.2f ms/frame)", 
        cpu_time / 1000, cpu_time / Engine::get_singleton()->get_frames_drawn()));
    
    print_line(vformat("GPU time: %.2f seconds (average: %.2f ms/frame)", 
        gpu_time / 1000, gpu_time / Engine::get_singleton()->get_frames_drawn()));
}
```

### 7.2 窗口标题显示

录制过程中会实时更新窗口标题：

```cpp
DisplayServer::get_singleton()->window_set_title(
    vformat("MovieWriter: Frame %d (time: %s) - %s", 
        Engine::get_singleton()->get_frames_drawn(), movie_time, project_name));
```

## 8. 扩展性

### 8.1 自定义录制器

可以通过继承 `MovieWriter` 类来实现自定义录制格式：

```cpp
class CustomMovieWriter : public MovieWriter {
    GDCLASS(CustomMovieWriter, MovieWriter)
    
protected:
    virtual bool handles_file(const String &p_path) const override;
    virtual Error write_begin(const Size2i &p_movie_size, uint32_t p_fps, const String &p_base_path) override;
    virtual Error write_frame(const Ref<Image> &p_image, const int32_t *p_audio_data) override;
    virtual void write_end() override;
};
```

### 8.2 GDExtension 支持

文档提到可以通过 GDExtension 实现自定义录制器，以获得更好的性能。

## 9. 技术优势

### 9.1 稳定的输出质量
- 非实时录制确保每帧都有足够时间渲染
- 固定帧率保证输出视频的时间一致性
- 不受硬件性能波动影响

### 9.2 灵活的格式支持
- 支持有损压缩（MJPEG）和无损压缩（PNG）
- 音频格式可配置（立体声到7.1环绕声）
- 可扩展的架构支持添加新格式

### 9.3 完整的音视频同步
- 音频和视频采用相同的时间基准
- 确保音视频完美同步
- 支持多声道音频录制

### 9.4 便于后期处理
- PNG 序列格式便于视频编辑
- 详细的性能统计数据
- 灵活的参数配置

## 10. 注意事项

### 10.1 磁盘空间
- 录制前会检查可用磁盘空间
- PNG 序列文件可能非常大
- 建议预留足够的存储空间

### 10.2 性能影响
- 录制过程会显著影响运行速度
- 建议在专门的录制环境中使用
- 可能需要调整项目设置以优化录制效果

### 10.3 格式选择
- AVI 格式适合直接播放和分享
- PNG 序列适合后期编辑和处理
- 根据用途选择合适的格式

## 11. 总结

Godot 的 MovieWriter 系统是一个设计精良的视频录制解决方案，通过非实时录制机制保证了高质量的输出。其模块化的架构使得系统具有良好的扩展性，同时提供了丰富的配置选项和详细的性能监控。

这个系统特别适合：
- 制作游戏宣传视频
- 录制教程和演示
- 游戏开发调试
- 自动化测试录制

通过深入理解其工作原理，开发者可以更好地利用这个工具来创建高质量的游戏视频内容。 