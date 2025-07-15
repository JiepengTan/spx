## Godot MovieWriter 实时录制技术方案

### 0. 调研
https://zhuanlan.zhihu.com/p/343790643


### 1. 当前系统的实时录制限制

#### 核心问题分析

**1. 固定帧率机制的限制**：

```cpp
// 在 main_timer_sync.cpp 中
MainFrameTime MainTimerSync::advance_checked(double p_physics_step, int p_physics_ticks_per_second, double p_process_step) {
    if (fixed_fps != -1) {
        p_process_step = 1.0 / fixed_fps;  // 强制固定时间步长
    }
    // ...
}
```

这种设计完全忽略了实际的帧时间，导致：
- 无法与显示器刷新率同步
- 无法适应性能波动
- 录制速度固定，无法实时预览

**2. AudioDriverDummy 的限制**：

```cpp
// 在录制模式下强制使用 Dummy 音频驱动
if (Engine::get_singleton()->get_write_movie_path() != String()) {
    audio_driver_idx = AudioDriverManager::get_driver_count() - 1;  // Dummy driver
    AudioDriverDummy::get_dummy_singleton()->set_use_threads(false);
}
```

这导致：
- 用户无法听到声音
- 无法进行实时录制反馈
- 音频与视频分离处理

**3. 同步I/O性能瓶颈**：

```cpp
void MovieWriter::add_frame() {
    // 同步获取图像数据
    RID main_vp_texture = RenderingServer::get_singleton()->viewport_get_texture(main_vp_rid);
    Ref<Image> vp_tex = RenderingServer::get_singleton()->texture_2d_get(main_vp_texture);
    
    // 同步音频处理
    AudioDriverDummy::get_dummy_singleton()->mix_audio(mix_rate / fps, audio_mix_buffer.ptr());
    
    // 同步文件写入
    write_frame(vp_tex, audio_mix_buffer.ptr());
}
```

### 2. 实时录制解决方案

####  .双模式录制系统

**设计思路**：
- 保留现有的非实时录制模式
- 新增实时录制模式
- 通过配置参数切换

**实现方案**：

```cpp
enum RecordingMode {
    RECORDING_MODE_OFFLINE,    // 当前的非实时模式
    RECORDING_MODE_REALTIME    // 新的实时模式
};

class RealtimeMovieWriter : public MovieWriter {
private:
    RecordingMode recording_mode;
    std::thread encoding_thread;
    std::queue<FrameData> frame_queue;
    std::mutex queue_mutex;
    
public:
    void set_recording_mode(RecordingMode mode) {
        recording_mode = mode;
    }
    
    virtual void add_frame() override {
        if (recording_mode == RECORDING_MODE_REALTIME) {
            add_frame_realtime();
        } else {
            add_frame_offline();
        }
    }
    
private:
    void add_frame_realtime() {
        // 异步获取帧数据
        FrameData frame;
        capture_frame_async(frame);
        
        // 加入编码队列
        enqueue_frame(frame);
    }
    
    void add_frame_offline() {
        // 原有的同步实现
        MovieWriter::add_frame();
    }
};
```

#### 2. 异步编码管道

**多线程架构**：

```cpp
class AsyncEncoder {
private:
    // 三个独立的线程
    std::thread capture_thread;    // 图像捕获线程
    std::thread encode_thread;     // 编码线程
    std::thread write_thread;      // 文件写入线程
    
    // 线程间通信队列
    ThreadSafeQueue<RawFrame> capture_queue;
    ThreadSafeQueue<EncodedFrame> encode_queue;
    
public:
    void start_realtime_recording() {
        capture_thread = std::thread(&AsyncEncoder::capture_loop, this);
        encode_thread = std::thread(&AsyncEncoder::encode_loop, this);
        write_thread = std::thread(&AsyncEncoder::write_loop, this);
    }
    
private:
    void capture_loop() {
        while (is_recording) {
            RawFrame frame;
            capture_frame(frame);
            capture_queue.push(frame);
        }
    }
    
    void encode_loop() {
        while (is_recording) {
            RawFrame raw_frame;
            if (capture_queue.try_pop(raw_frame)) {
                EncodedFrame encoded_frame;
                encode_frame(raw_frame, encoded_frame);
                encode_queue.push(encoded_frame);
            }
        }
    }
    
    void write_loop() {
        while (is_recording) {
            EncodedFrame encoded_frame;
            if (encode_queue.try_pop(encoded_frame)) {
                write_frame_to_file(encoded_frame);
            }
        }
    }
};
```

#### 3. 智能帧率适配

**自适应帧率系统**：

```cpp
class AdaptiveFrameRate {
private:
    double target_fps;
    double current_fps;
    double performance_factor;
    
public:
    void update_performance_metrics() {
        // 监控编码性能
        double encode_time = measure_encode_time();
        double available_time = 1.0 / target_fps;
        
        performance_factor = available_time / encode_time;
        
        // 动态调整策略
        if (performance_factor < 0.8) {
            // 性能不足，降低质量或跳帧
            reduce_quality();
        } else if (performance_factor > 1.2) {
            // 性能充足，提高质量
            increase_quality();
        }
    }
    
    bool should_skip_frame() {
        return performance_factor < 0.5;
    }
};
```

#### 4. 混合音频驱动

**实时音频处理**：

```cpp
class HybridAudioDriver : public AudioDriver {
private:
    AudioDriver* real_driver;      // 实际音频驱动
    AudioDriver* dummy_driver;     // 录制用驱动
    bool recording_enabled;
    
public:
    void start_recording() {
        recording_enabled = true;
        dummy_driver->start();
    }
    
    void stop_recording() {
        recording_enabled = false;
        dummy_driver->stop();
    }
    
    virtual void audio_server_process(int p_frames, int32_t *p_buffer, bool p_update_mix_time = true) override {
        // 正常音频播放
        real_driver->audio_server_process(p_frames, p_buffer, p_update_mix_time);
        
        // 如果正在录制，同时发送到录制驱动
        if (recording_enabled) {
            dummy_driver->audio_server_process(p_frames, p_buffer, false);
        }
    }
};
```

#### 5. 硬件加速编码

**GPU 编码集成**：

```cpp
class HardwareEncoder {
private:
    enum EncoderType {
        ENCODER_SOFTWARE,
        ENCODER_NVENC,
        ENCODER_QUICKSYNC,
        ENCODER_AMF
    };
    
    EncoderType encoder_type;
    
public:
    void initialize_hardware_encoder() {
        // 检测可用的硬件编码器
        if (is_nvenc_available()) {
            encoder_type = ENCODER_NVENC;
            init_nvenc();
        } else if (is_quicksync_available()) {
            encoder_type = ENCODER_QUICKSYNC;
            init_quicksync();
        } else if (is_amf_available()) {
            encoder_type = ENCODER_AMF;
            init_amf();
        } else {
            encoder_type = ENCODER_SOFTWARE;
            init_software_encoder();
        }
    }
    
    void encode_frame_gpu(const Ref<Image>& frame, EncodedFrame& output) {
        switch (encoder_type) {
            case ENCODER_NVENC:
                encode_nvenc(frame, output);
                break;
            case ENCODER_QUICKSYNC:
                encode_quicksync(frame, output);
                break;
            // ... 其他编码器
        }
    }
};
```

### 3. 用户界面改进

#### 实时录制控制面板

```cpp
class RealtimeRecordingPanel : public Control {
private:
    Button* start_button;
    Button* pause_button;
    Button* stop_button;
    ProgressBar* performance_bar;
    Label* status_label;
    
public:
    void _ready() override {
        // 创建用户界面
        create_recording_controls();
        
        // 连接信号
        start_button->connect("pressed", callable_mp(this, &RealtimeRecordingPanel::start_recording));
        pause_button->connect("pressed", callable_mp(this, &RealtimeRecordingPanel::pause_recording));
        stop_button->connect("pressed", callable_mp(this, &RealtimeRecordingPanel::stop_recording));
    }
    
    void update_performance_indicator(double performance_factor) {
        performance_bar->set_value(performance_factor * 100);
        
        if (performance_factor < 0.8) {
            status_label->set_text("Performance Warning: Consider lowering quality");
            status_label->set_modulate(Color::YELLOW);
        } else {
            status_label->set_text("Recording: Good Performance");
            status_label->set_modulate(Color::GREEN);
        }
    }
};
```

### 4. 配置选项

#### 实时录制配置

```cpp
// 项目设置
GLOBAL_DEF("movie_writer/realtime/enabled", false);
GLOBAL_DEF("movie_writer/realtime/target_fps", 60);
GLOBAL_DEF("movie_writer/realtime/adaptive_quality", true);
GLOBAL_DEF("movie_writer/realtime/hardware_acceleration", true);
GLOBAL_DEF("movie_writer/realtime/audio_preview", true);
GLOBAL_DEF("movie_writer/realtime/buffer_size", 10);

// 编码质量设置
GLOBAL_DEF("movie_writer/encoding/video_codec", "h264");
GLOBAL_DEF("movie_writer/encoding/audio_codec", "aac");
GLOBAL_DEF("movie_writer/encoding/bitrate", 8000);
GLOBAL_DEF("movie_writer/encoding/quality_preset", "medium");
```

### 5. 实现优先级

#### 短期目标（3-6个月）

1. **双模式架构**：
   - 实现 RealtimeMovieWriter 基类
   - 添加录制模式切换
   - 保持向后兼容性

2. **异步编码管道**：
   - 实现多线程编码框架
   - 添加帧缓冲队列
   - 优化内存管理

#### 中期目标（6-12个月）

1. **硬件加速支持**：
   - 集成 NVENC 编码器
   - 添加 QuickSync 支持
   - 实现 AMF 编码器

2. **现代编码格式**：
   - 支持 H.264/H.265
   - 添加 VP8/VP9 支持
   - 实现音频压缩

#### 长期目标（12个月以上）

1. **Web 平台支持**：
   - 使用 MediaRecorder API
   - 实现 IndexedDB 存储
   - 添加流媒体推送

2. **云录制服务**：
   - 实现录制服务器
   - 添加远程存储
   - 支持协作录制

### 6. 性能基准测试

#### 测试场景

```cpp
class RecordingBenchmark {
public:
    struct BenchmarkResult {
        double average_fps;
        double min_fps;
        double max_fps;
        double cpu_usage;
        double memory_usage;
        double encode_time;
        double file_size;
    };
    
    BenchmarkResult run_benchmark(RecordingMode mode, int duration_seconds) {
        BenchmarkResult result;
        
        // 设置测试环境
        setup_test_scene();
        
        // 开始录制
        start_recording(mode);
        
        // 运行测试
        auto start_time = std::chrono::steady_clock::now();
        while (get_elapsed_time(start_time) < duration_seconds) {
            // 收集性能数据
            collect_performance_metrics(result);
            
            // 更新场景
            update_test_scene();
        }
        
        // 停止录制
        stop_recording();
        
        return result;
    }
};
```

这个实时录制方案提供了：
- **兼容性**：保持现有非实时录制功能
- **性能**：通过异步处理和硬件加速提高效率
- **用户体验**：实时预览和性能反馈
- **扩展性**：支持多种编码格式和平台

通过这种渐进式的实现方法，可以在不破坏现有功能的基础上，为 Godot 添加强大的实时录制能力。 