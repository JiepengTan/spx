# Godot 无物理模式扩展指南

## 概述

本指南展示如何扩展 `UnityLikeBody2D` 统一物理控制器，添加 `NoPhysics` 模式，实现类似Unity中没有Collider和Rigidbody的GameObject行为。

## 设计理念

### 三种物理模式对比

| 模式 | Unity等效 | 特点 | 适用场景 |
|------|-----------|------|----------|
| `DYNAMIC` | `isKinematic = false` | 受物理影响，自动重力和碰撞 | 玩家角色、可推动物品 |
| `KINEMATIC` | `isKinematic = true` | 代码控制移动，有碰撞检测 | NPC、移动平台 |
| `NO_PHYSICS` | 无Collider/Rigidbody | 纯视觉，无碰撞，性能最优 | 装饰元素、特效、UI |

## 扩展核心父类

### UnityLikeBody2D 完整扩展版本

```gdscript
# UnityLikeBody2D.gd - 扩展版本
extends CharacterBody2D
class_name UnityLikeBody2D

# 物理模式枚举
enum PhysicsMode {
    KINEMATIC,      # 类似Unity isKinematic = true
    DYNAMIC,        # 类似Unity isKinematic = false  
    NO_PHYSICS      # 类似Unity没有Collider/Rigidbody
}

# Unity风格的属性
@export var physicsMode: PhysicsMode = PhysicsMode.DYNAMIC
@export var useGravity: bool = true
@export var gravityScale: float = 1.0
@export var mass: float = 1.0
@export var drag: float = 0.0

# 内部状态
var _gravity = ProjectSettings.get_setting("physics/2d/default_gravity")
var _external_forces = Vector2.ZERO
var _applied_forces = Vector2.ZERO
var _friction = 300.0
var _collision_enabled = true

# 兼容性属性（保持向后兼容）
var isKinematic: bool:
    get:
        return physicsMode == PhysicsMode.KINEMATIC
    set(value):
        physicsMode = PhysicsMode.KINEMATIC if value else PhysicsMode.DYNAMIC

func _ready():
    _update_physics_mode()
    if has_method("_on_ready_override"):
        _on_ready_override()

func _physics_process(delta):
    match physicsMode:
        PhysicsMode.DYNAMIC:
            _handle_dynamic_physics(delta)
        PhysicsMode.KINEMATIC:
            _handle_kinematic_physics(delta)
        PhysicsMode.NO_PHYSICS:
            _handle_no_physics(delta)
    
    # 只有在有物理模式时才调用move_and_slide
    if physicsMode != PhysicsMode.NO_PHYSICS:
        move_and_slide()
    
    if has_method("_on_physics_process_override"):
        _on_physics_process_override(delta)

func _handle_dynamic_physics(delta):
    # 原有的物理处理逻辑
    if useGravity and not is_on_floor():
        velocity.y += _gravity * gravityScale * delta
    
    velocity += _applied_forces * delta / mass
    velocity += _external_forces * delta
    
    if drag > 0:
        velocity = velocity.move_toward(Vector2.ZERO, drag * velocity.length() * delta)
    
    if is_on_floor() and abs(velocity.x) > 0:
        velocity.x = move_toward(velocity.x, 0, _friction * delta)
    
    _applied_forces = Vector2.ZERO

func _handle_kinematic_physics(delta):
    # Kinematic模式：只执行用户设置的移动
    pass

func _handle_no_physics(delta):
    # 无物理模式：直接通过transform移动，不使用velocity
    # velocity在这个模式下用作移动速度参考
    if velocity != Vector2.ZERO:
        global_position += velocity * delta
    
    # 清除所有物理相关的力
    _external_forces = Vector2.ZERO
    _applied_forces = Vector2.ZERO

# 设置物理模式
func set_physics_mode(mode: PhysicsMode):
    physicsMode = mode
    _update_physics_mode()

func _update_physics_mode():
    match physicsMode:
        PhysicsMode.DYNAMIC:
            _enable_collision()
            set_physics_process(true)
        PhysicsMode.KINEMATIC:
            _enable_collision()
            set_physics_process(true)
        PhysicsMode.NO_PHYSICS:
            _disable_collision()
            set_physics_process(true)  # 仍需要处理transform移动

func _enable_collision():
    _collision_enabled = true
    # 启用所有碰撞形状
    for child in get_children():
        if child is CollisionShape2D or child is CollisionPolygon2D:
            child.disabled = false

func _disable_collision():
    _collision_enabled = false
    velocity = Vector2.ZERO  # 清除velocity，避免move_and_slide的影响
    # 禁用所有碰撞形状
    for child in get_children():
        if child is CollisionShape2D or child is CollisionPolygon2D:
            child.disabled = true

# Unity风格的API（扩展）
func AddForce(force: Vector2, forceMode: String = "Force"):
    if physicsMode == PhysicsMode.NO_PHYSICS:
        return  # 无物理模式不受力影响
    if physicsMode == PhysicsMode.KINEMATIC:
        return  # Kinematic模式不受力影响
    
    match forceMode:
        "Force":
            _external_forces += force
        "Impulse":
            _applied_forces += force
        "VelocityChange":
            velocity += force
        "Acceleration":
            _external_forces += force * mass

# 无物理模式专用的移动方法
func move_without_physics(direction: Vector2, speed: float):
    if physicsMode != PhysicsMode.NO_PHYSICS:
        return
    
    # 直接移动，不考虑碰撞
    global_position += direction.normalized() * speed * get_physics_process_delta_time()

# 设置移动速度（无物理模式下用于transform移动）
func set_movement_velocity(vel: Vector2):
    velocity = vel

# 检测碰撞状态
func is_collision_enabled() -> bool:
    return _collision_enabled

# 检查是否在地面（无物理模式下总是返回false）
func is_grounded() -> bool:
    if physicsMode == PhysicsMode.NO_PHYSICS:
        return false
    return is_on_floor()
```

## 使用示例

### 1. 装饰性NPC（无物理交互）

```gdscript
# DecorationNPC.gd - 纯装饰，不参与物理
extends UnityLikeBody2D

@export var float_amplitude = 20.0
@export var float_speed = 2.0
@export var patrol_points: Array[Vector2]

var current_patrol_target = 0
var start_position: Vector2

func _on_ready_override():
    # 设置为无物理模式
    set_physics_mode(PhysicsMode.NO_PHYSICS)
    start_position = global_position

func _on_physics_process_override(delta):
    # 方案1: 简单的上下浮动动画
    simple_float_animation()
    
    # 方案2: 巡逻移动（取消注释使用）
    # patrol_movement()

func simple_float_animation():
    # 使用velocity进行浮动移动
    var float_offset = cos(Time.get_time() * float_speed) * float_amplitude * float_speed
    velocity = Vector2(0, float_offset)

func patrol_movement():
    if patrol_points.is_empty():
        return
    
    var target = patrol_points[current_patrol_target]
    var direction = (target - global_position).normalized()
    var distance = global_position.distance_to(target)
    
    if distance > 10:
        set_movement_velocity(direction * 50)
    else:
        current_patrol_target = (current_patrol_target + 1) % patrol_points.size()
        set_movement_velocity(Vector2.ZERO)
```

### 2. UI跟随元素（无碰撞）

```gdscript
# UIFollower.gd - 跟随玩家的UI元素，不产生碰撞
extends UnityLikeBody2D

@export var target: Node2D
@export var follow_speed = 200.0
@export var offset = Vector2(0, -50)
@export var smoothing = 5.0

var target_position: Vector2

func _on_ready_override():
    # 无物理模式，不会与其他对象碰撞
    set_physics_mode(PhysicsMode.NO_PHYSICS)

func _on_physics_process_override(delta):
    if not target:
        return
    
    target_position = target.global_position + offset
    var direction = (target_position - global_position)
    var distance = direction.length()
    
    if distance > 5:
        # 使用平滑跟随
        var smooth_velocity = direction * smoothing
        set_movement_velocity(smooth_velocity)
    else:
        set_movement_velocity(Vector2.ZERO)

# 可以添加额外的UI行为
func show_damage_number(damage: int):
    var label = Label.new()
    label.text = str(damage)
    label.add_theme_color_override("font_color", Color.RED)
    add_child(label)
    
    # 损伤数字上浮动画
    var tween = create_tween()
    tween.parallel().tween_property(label, "position", Vector2(0, -50), 1.0)
    tween.parallel().tween_property(label, "modulate", Color.TRANSPARENT, 1.0)
    tween.tween_callback(label.queue_free)
```

### 3. 背景移动元素

```gdscript
# BackgroundElement.gd - 背景中移动的装饰元素
extends UnityLikeBody2D

@export var scroll_speed = Vector2(-50, 0)
@export var screen_wrap = true
@export var parallax_factor = 0.5

var camera: Camera2D
var viewport_size: Vector2

func _on_ready_override():
    # 背景元素不参与物理
    set_physics_mode(PhysicsMode.NO_PHYSICS)
    
    # 获取相机和视口信息
    camera = get_viewport().get_camera_2d()
    viewport_size = get_viewport().get_visible_rect().size
    
    # 设置基础移动速度
    set_movement_velocity(scroll_speed)

func _on_physics_process_override(delta):
    # 视差滚动效果
    if camera:
        apply_parallax_effect()
    
    # 屏幕循环
    if screen_wrap:
        handle_screen_wrapping()

func apply_parallax_effect():
    # 根据相机位置调整移动速度
    var camera_velocity = Vector2.ZERO
    if camera.has_method("get_velocity"):
        camera_velocity = camera.get_velocity() if camera.has_method("get_velocity") else Vector2.ZERO
    
    var parallax_velocity = scroll_speed + (camera_velocity * parallax_factor)
    set_movement_velocity(parallax_velocity)

func handle_screen_wrapping():
    var camera_pos = camera.global_position if camera else Vector2.ZERO
    var left_bound = camera_pos.x - viewport_size.x * 0.6
    var right_bound = camera_pos.x + viewport_size.x * 0.6
    
    if global_position.x < left_bound:
        global_position.x = right_bound + 100
    elif global_position.x > right_bound + 200:
        global_position.x = left_bound - 100
```

### 4. 特效对象（纯视觉）

```gdscript
# VisualEffect.gd - 纯视觉特效，不参与物理
extends UnityLikeBody2D

@export var effect_duration = 2.0
@export var move_speed = Vector2(0, -100)
@export var fade_out = true
@export var scale_animation = true

func _on_ready_override():
    # 无物理模式
    set_physics_mode(PhysicsMode.NO_PHYSICS)
    
    # 设置移动速度
    set_movement_velocity(move_speed)
    
    # 启动效果动画
    start_effect_animations()
    
    # 自动销毁
    var timer = get_tree().create_timer(effect_duration)
    timer.timeout.connect(_on_effect_finished)

func start_effect_animations():
    var tween = create_tween()
    tween.set_parallel(true)
    
    # 淡出动画
    if fade_out:
        tween.tween_property(self, "modulate", Color.TRANSPARENT, effect_duration)
    
    # 缩放动画
    if scale_animation:
        tween.tween_property(self, "scale", Vector2(2.0, 2.0), effect_duration * 0.5)
        tween.tween_property(self, "scale", Vector2(0.1, 0.1), effect_duration * 0.5).set_delay(effect_duration * 0.5)

func _on_effect_finished():
    queue_free()

# 可以在效果进行中切换到物理模式
func enable_physics_drop():
    set_physics_mode(PhysicsMode.DYNAMIC)
    useGravity = true
    velocity.y = 0  # 重置垂直速度，让重力接管
```

### 5. 动态模式切换对象

```gdscript
# ModeSwitcher.gd - 可以在不同模式间切换的对象
extends UnityLikeBody2D

enum State {
    DECORATION,    # 装饰状态（无物理）
    INTERACTIVE,   # 可交互状态（Kinematic）
    PHYSICAL       # 物理状态（Dynamic）
}

@export var current_state: State = State.DECORATION
@export var transition_duration = 0.5

var state_colors = {
    State.DECORATION: Color.GRAY,
    State.INTERACTIVE: Color.BLUE, 
    State.PHYSICAL: Color.WHITE
}

func _on_ready_override():
    set_state(current_state)

func set_state(new_state: State):
    if current_state == new_state:
        return
    
    var old_state = current_state
    current_state = new_state
    
    # 切换物理模式
    match new_state:
        State.DECORATION:
            set_physics_mode(PhysicsMode.NO_PHYSICS)
            # 简单的漂浮动画
            start_decoration_animation()
            
        State.INTERACTIVE:
            set_physics_mode(PhysicsMode.KINEMATIC)
            useGravity = false
            velocity = Vector2.ZERO
            
        State.PHYSICAL:
            set_physics_mode(PhysicsMode.DYNAMIC)
            useGravity = true
    
    # 视觉反馈
    animate_state_transition(new_state)
    
    # 发出状态变化信号
    emit_signal("state_changed", old_state, new_state)

func start_decoration_animation():
    # 装饰状态下的浮动效果
    var float_amplitude = 30.0
    var float_speed = 1.5
    
    func float_update():
        if current_state == State.DECORATION:
            var float_y = sin(Time.get_time() * float_speed) * float_amplitude * float_speed
            velocity = Vector2(0, float_y)

func animate_state_transition(new_state: State):
    var target_color = state_colors.get(new_state, Color.WHITE)
    
    var tween = create_tween()
    tween.set_parallel(true)
    
    # 颜色过渡
    tween.tween_property($Sprite2D, "modulate", target_color, transition_duration)
    
    # 缩放动画表示状态切换
    tween.tween_property(self, "scale", Vector2(1.2, 1.2), transition_duration * 0.5)
    tween.tween_property(self, "scale", Vector2(1.0, 1.0), transition_duration * 0.5).set_delay(transition_duration * 0.5)

# 外部触发状态切换
func activate():
    match current_state:
        State.DECORATION:
            set_state(State.INTERACTIVE)
        State.INTERACTIVE:
            set_state(State.PHYSICAL)
        State.PHYSICAL:
            set_state(State.DECORATION)

func deactivate():
    set_state(State.DECORATION)

# 响应外部事件
func _on_player_interaction():
    if current_state == State.INTERACTIVE:
        set_state(State.PHYSICAL)

func _on_timer_timeout():
    if current_state == State.PHYSICAL:
        set_state(State.DECORATION)

signal state_changed(old_state: State, new_state: State)
```

## 实用工具方法扩展

### 临时模式控制

```gdscript
# UnityLikeBody2D.gd 中的额外工具方法

# 临时禁用物理（保存当前状态）
var _saved_physics_mode: PhysicsMode

func temporarily_disable_physics():
    _saved_physics_mode = physicsMode
    set_physics_mode(PhysicsMode.NO_PHYSICS)

func restore_physics():
    if _saved_physics_mode != null:
        set_physics_mode(_saved_physics_mode)

# 检查当前物理模式
func is_dynamic() -> bool:
    return physicsMode == PhysicsMode.DYNAMIC

func is_kinematic() -> bool:
    return physicsMode == PhysicsMode.KINEMATIC

func is_no_physics() -> bool:
    return physicsMode == PhysicsMode.NO_PHYSICS

# 获取模式信息
func get_physics_mode_string() -> String:
    match physicsMode:
        PhysicsMode.DYNAMIC:
            return "Dynamic"
        PhysicsMode.KINEMATIC:
            return "Kinematic" 
        PhysicsMode.NO_PHYSICS:
            return "NoPhysics"
        _:
            return "Unknown"
```

### 手动碰撞检测

```gdscript
# 无物理模式下的手动碰撞检测
func manual_collision_check(target_position: Vector2) -> bool:
    if physicsMode != PhysicsMode.NO_PHYSICS:
        return false
    
    # 使用Space2D进行手动碰撞检测
    var space_state = get_world_2d().direct_space_state
    var query = PhysicsRayQueryParameters2D.create(global_position, target_position)
    var result = space_state.intersect_ray(query)
    
    return result.size() > 0

# 区域检测
func manual_area_check(radius: float) -> Array:
    if physicsMode != PhysicsMode.NO_PHYSICS:
        return []
    
    var space_state = get_world_2d().direct_space_state
    var query = PhysicsPointQueryParameters2D.new()
    query.position = global_position
    query.collision_mask = collision_mask
    
    var results = space_state.intersect_point(query, 32)
    return results

# 检查是否可以移动到指定位置
func can_move_to(target_pos: Vector2) -> bool:
    if physicsMode != PhysicsMode.NO_PHYSICS:
        return true  # 其他模式由物理引擎处理
    
    return not manual_collision_check(target_pos)
```

### 性能优化

```gdscript
# 无物理模式的性能优化
var _no_physics_update_interval = 0.1  # 降低更新频率
var _no_physics_timer = 0.0

func _handle_no_physics_optimized(delta):
    _no_physics_timer += delta
    
    # 只在需要时更新
    if _no_physics_timer >= _no_physics_update_interval:
        _no_physics_timer = 0.0
        
        if velocity != Vector2.ZERO:
            global_position += velocity * _no_physics_update_interval
        
        # 清除力
        _external_forces = Vector2.ZERO
        _applied_forces = Vector2.ZERO

# 距离优化：远离相机的对象降低更新频率
func should_update_this_frame() -> bool:
    if physicsMode != PhysicsMode.NO_PHYSICS:
        return true
    
    var camera = get_viewport().get_camera_2d()
    if not camera:
        return true
    
    var distance = global_position.distance_to(camera.global_position)
    var viewport_size = get_viewport().get_visible_rect().size
    var max_distance = viewport_size.length()
    
    # 距离越远，更新频率越低
    if distance > max_distance * 2:
        return Engine.get_process_frames() % 4 == 0  # 每4帧更新一次
    elif distance > max_distance:
        return Engine.get_process_frames() % 2 == 0  # 每2帧更新一次
    
    return true  # 近距离每帧更新
```

## 调试和监控

### 可视化调试

```gdscript
# 调试信息显示
func _draw():
    if not Engine.is_editor_hint() and OS.is_debug_build():
        draw_physics_mode_indicator()
        draw_velocity_indicator()

func draw_physics_mode_indicator():
    var color = Color.WHITE
    var radius = 8.0
    
    match physicsMode:
        PhysicsMode.DYNAMIC:
            color = Color.GREEN
        PhysicsMode.KINEMATIC:
            color = Color.BLUE
        PhysicsMode.NO_PHYSICS:
            color = Color.RED
            radius = 6.0  # 稍小一点表示无物理
    
    draw_circle(Vector2.ZERO, radius, color)

func draw_velocity_indicator():
    if velocity != Vector2.ZERO:
        var color = Color.YELLOW
        if physicsMode == PhysicsMode.NO_PHYSICS:
            color = Color.ORANGE
        
        draw_line(Vector2.ZERO, velocity.normalized() * 30, color, 2)
        
        # 显示速度数值
        var speed_text = "%.1f" % velocity.length()
        # 注意：这里需要CanvasItem的draw_string方法，实际使用时需要准备字体资源
```

### 事件系统

```gdscript
# 信号事件
signal physics_mode_changed(old_mode: PhysicsMode, new_mode: PhysicsMode)
signal collision_state_changed(enabled: bool)
signal no_physics_movement(old_pos: Vector2, new_pos: Vector2)

func set_physics_mode(mode: PhysicsMode):
    var old_mode = physicsMode
    physicsMode = mode
    _update_physics_mode()
    
    if old_mode != mode:
        emit_signal("physics_mode_changed", old_mode, mode)

func _handle_no_physics(delta):
    var old_pos = global_position
    
    if velocity != Vector2.ZERO:
        global_position += velocity * delta
        emit_signal("no_physics_movement", old_pos, global_position)
    
    _external_forces = Vector2.ZERO
    _applied_forces = Vector2.ZERO
```

## 使用场景总结

### NoPhysics 模式的优势

1. **性能最优**：
   - 无碰撞检测开销
   - 可自定义更新频率
   - 直接transform操作

2. **完全控制**：
   - 不受物理引擎限制
   - 可实现任意移动模式
   - 易于实现特效动画

3. **灵活切换**：
   - 可随时切换到其他模式
   - 状态保存和恢复
   - 动态行为调整

### 推荐使用场景

| 场景类型 | 示例 | 推荐模式 |
|----------|------|----------|
| 背景装饰 | 云朵、树叶、背景NPC | NoPhysics |
| UI元素 | 血条跟随、伤害数字 | NoPhysics |  
| 视觉特效 | 粒子效果、魔法光圈 | NoPhysics |
| 过场动画 | 镜头移动期间的角色 | NoPhysics |
| 远距离对象 | 远景中的移动元素 | NoPhysics |

## 最佳实践

1. **合理选择模式**：根据对象的实际需求选择合适的物理模式
2. **性能优化**：对于大量的装饰性对象，使用NoPhysics模式
3. **动态切换**：在需要时动态切换模式，如过场动画期间
4. **调试支持**：在开发期间启用可视化调试
5. **事件驱动**：使用信号系统实现模式切换的响应逻辑

通过这个扩展，你现在拥有了完整的三模式物理系统：Dynamic（真实物理）、Kinematic（代码控制）、NoPhysics（纯视觉），完美覆盖了所有Unity物理状态，同时获得了更好的性能和更灵活的控制！