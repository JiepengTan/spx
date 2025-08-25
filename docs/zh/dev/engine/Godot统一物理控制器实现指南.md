# Godot 统一物理控制器实现指南

## 概述

本指南展示如何创建一个基于 `CharacterBody2D` 的统一物理控制器父类，完美模拟Unity的 `isKinematic`、`velocity`、`AddForce` 等概念。只需继承一个父类，就能获得Unity风格的物理控制API。

## 核心设计理念

- **单一继承**：所有物理对象都继承自同一个父类
- **Unity兼容**：提供与Unity相似的API接口
- **灵活控制**：支持Kinematic和非Kinematic两种模式
- **性能优化**：基于CharacterBody2D，避免不必要的物理计算

## 统一物理控制器实现

### 核心父类 - UnityLikeBody2D

```gdscript
# UnityLikeBody2D.gd - 统一的物理控制器父类
extends CharacterBody2D
class_name UnityLikeBody2D

# Unity风格的属性
@export var isKinematic: bool = false
@export var useGravity: bool = true
@export var gravityScale: float = 1.0
@export var mass: float = 1.0
@export var drag: float = 0.0
@export var angularDrag: float = 0.05

# 内部物理状态
var _gravity = ProjectSettings.get_setting("physics/2d/default_gravity")
var _external_forces = Vector2.ZERO
var _applied_forces = Vector2.ZERO
var _friction = 300.0

# Unity风格的velocity属性（与CharacterBody2D.velocity同步）
var _velocity: Vector2:
    get:
        return velocity
    set(value):
        velocity = value

func _ready():
    # 确保子类可以重写
    if has_method("_on_ready_override"):
        _on_ready_override()

func _physics_process(delta):
    if not isKinematic:
        # 非Kinematic模式：自动处理物理
        _handle_physics(delta)
    else:
        # Kinematic模式：只执行用户设置的velocity
        pass
    
    # 执行移动
    move_and_slide()
    
    # 子类重写点
    if has_method("_on_physics_process_override"):
        _on_physics_process_override(delta)

func _handle_physics(delta):
    # 应用重力
    if useGravity and not is_on_floor():
        velocity.y += _gravity * gravityScale * delta
    
    # 应用外部力
    velocity += _applied_forces * delta / mass
    
    # 应用持续力
    velocity += _external_forces * delta
    
    # 应用阻力
    if drag > 0:
        velocity = velocity.move_toward(Vector2.ZERO, drag * velocity.length() * delta)
    
    # 地面摩擦
    if is_on_floor() and abs(velocity.x) > 0:
        velocity.x = move_toward(velocity.x, 0, _friction * delta)
    
    # 清除单帧力
    _applied_forces = Vector2.ZERO

# Unity风格的API接口

# 设置isKinematic状态
func set_kinematic(kinematic: bool):
    isKinematic = kinematic
    if kinematic:
        # Kinematic模式下，清除所有物理效果
        _external_forces = Vector2.ZERO
        _applied_forces = Vector2.ZERO

# 施加瞬时冲量（类似Unity的AddForce with ForceMode.Impulse）
func AddForce(force: Vector2, forceMode: String = "Force"):
    if isKinematic:
        return  # Kinematic对象不受力影响
    
    match forceMode:
        "Force":
            _external_forces += force
        "Impulse":
            _applied_forces += force
        "VelocityChange":
            velocity += force
        "Acceleration":
            _external_forces += force * mass

# 直接设置velocity（Unity风格）
func set_velocity_unity(new_velocity: Vector2):
    velocity = new_velocity

# 获取velocity（Unity风格）
func get_velocity_unity() -> Vector2:
    return velocity

# 检查是否在地面（Unity风格）
func is_grounded() -> bool:
    return is_on_floor()

# 传送到指定位置
func teleport_to(pos: Vector2):
    global_position = pos
    velocity = Vector2.ZERO
```

## 使用示例

### 1. 玩家角色实现

```gdscript
# Player.gd
extends UnityLikeBody2D

@export var moveSpeed = 300.0
@export var jumpForce = -400.0

func _on_ready_override():
    # 玩家设置为非Kinematic，受重力影响
    isKinematic = false
    useGravity = true
    gravityScale = 1.0

func _on_physics_process_override(delta):
    handle_input()

func handle_input():
    # 水平移动
    var direction = Input.get_axis("ui_left", "ui_right")
    velocity.x = direction * moveSpeed
    
    # 跳跃
    if Input.is_action_just_pressed("ui_accept") and is_grounded():
        velocity.y = jumpForce

# 受击效果
func take_damage(knockback_force: Vector2):
    AddForce(knockback_force, "Impulse")
```

### 2. NPC商人实现

```gdscript
# Merchant.gd  
extends UnityLikeBody2D

@export var patrolSpeed = 50.0
@export var patrolPoints: Array[Vector2]
var currentTarget = 0

func _on_ready_override():
    # NPC设置为Kinematic，不受物理影响
    isKinematic = true
    useGravity = false

func _on_physics_process_override(delta):
    if patrolPoints.size() >= 2:
        patrol_movement()

func patrol_movement():
    var target = patrolPoints[currentTarget]
    var direction = (target - global_position).normalized()
    
    # 直接设置移动速度（Kinematic模式）
    velocity = direction * patrolSpeed
    
    if global_position.distance_to(target) < 10:
        currentTarget = (currentTarget + 1) % patrolPoints.size()

# 对话系统
func _on_dialogue_area_body_entered(body):
    if body.is_in_group("player"):
        velocity = Vector2.ZERO  # 停止移动
        start_dialogue()

func start_dialogue():
    print("欢迎光临我的商店！")
```

### 3. 可推动箱子实现

```gdscript
# PushableBox.gd
extends UnityLikeBody2D

@export var canBePushed = true

func _on_ready_override():
    if canBePushed:
        # 可推动 - 非Kinematic
        isKinematic = false
        useGravity = true
        mass = 2.0
        drag = 2.0  # 有阻力，不会滑太远
    else:
        # 装饰箱子 - Kinematic
        isKinematic = true
        useGravity = false

# 被推动时的反应
func _on_player_collision(player_velocity: Vector2):
    if canBePushed and player_velocity.length() > 50:
        AddForce(player_velocity * 0.5, "Force")

# 解谜机关检测
signal box_in_correct_position
signal box_left_position

func _on_puzzle_area_body_entered(body):
    if body == self:
        emit_signal("box_in_correct_position")
        $Sprite2D.modulate = Color.GREEN

func _on_puzzle_area_body_exited(body):
    if body == self:
        emit_signal("box_left_position")
        $Sprite2D.modulate = Color.WHITE
```

### 4. 移动平台实现

```gdscript
# MovingPlatform.gd
extends UnityLikeBody2D

@export var moveDistance = Vector2(200, 0)
@export var moveDuration = 3.0

var players_on_platform: Array = []

func _on_ready_override():
    # 平台为Kinematic，不受外力影响
    isKinematic = true
    useGravity = false
    start_moving()

func start_moving():
    var tween = create_tween()
    tween.set_loops()
    
    var start_pos = global_position
    var end_pos = start_pos + moveDistance
    
    # 使用Tween控制移动，同时更新velocity用于推动玩家
    tween.tween_method(move_to_position, start_pos, end_pos, moveDuration)
    tween.tween_method(move_to_position, end_pos, start_pos, moveDuration)

func move_to_position(pos: Vector2):
    var old_pos = global_position
    global_position = pos
    # 计算移动速度，用于推动站在平台上的对象
    if get_physics_process_delta_time() > 0:
        velocity = (pos - old_pos) / get_physics_process_delta_time()

# 检测玩家站在平台上
func _on_player_detector_body_entered(body):
    if body.is_in_group("player"):
        players_on_platform.append(body)

func _on_player_detector_body_exited(body):
    if body.is_in_group("player"):
        players_on_platform.erase(body)
```

### 5. 受击系统实现

```gdscript
# Fighter.gd
extends UnityLikeBody2D

@export var health = 100
@export var knockbackResistance = 1.0

var isKnockedBack = false
var knockbackTimer = 0.0
var originalKinematic = false

func _on_ready_override():
    # 战斗角色通常是非Kinematic
    isKinematic = false
    useGravity = true

func _on_physics_process_override(delta):
    if isKnockedBack:
        handle_knockback(delta)
    else:
        handle_normal_behavior()

func handle_knockback(delta):
    knockbackTimer -= delta
    
    # 击飞期间的空气阻力
    velocity.x = move_toward(velocity.x, 0, 200 * delta)
    
    if knockbackTimer <= 0:
        isKnockedBack = false
        # 恢复原始Kinematic状态
        isKinematic = originalKinematic

func handle_normal_behavior():
    # 正常行为逻辑（AI、玩家输入等）
    pass

func take_hit(damage: int, knockback_force: Vector2):
    health -= damage
    
    if health <= 0:
        die()
        return
    
    # 开始击飞
    originalKinematic = isKinematic
    isKinematic = false  # 暂时变为非Kinematic接受击飞
    
    # 施加击飞力（考虑抗击退性）
    var final_knockback = knockback_force / knockbackResistance
    AddForce(final_knockback, "VelocityChange")
    
    # 设置击飞状态
    isKnockedBack = true
    knockbackTimer = 0.8
    
    # 视觉反馈
    flash_damage_effect()

func flash_damage_effect():
    var tween = create_tween()
    tween.tween_property($Sprite2D, "modulate", Color.RED, 0.1)
    tween.tween_property($Sprite2D, "modulate", Color.WHITE, 0.1)

func die():
    # 死亡处理
    set_physics_process(false)
    $Sprite2D.modulate = Color.GRAY
```

### 6. 载具系统实现

```gdscript
# Vehicle.gd
extends UnityLikeBody2D

@export var enginePower = 800.0
@export var steeringPower = 30.0
@export var maxSpeed = 500.0

var playerInside = false
var playerNode = null

func _on_ready_override():
    # 载具使用非Kinematic物理
    isKinematic = false
    useGravity = true
    mass = 3.0
    drag = 1.0

func _on_physics_process_override(delta):
    if playerInside and playerNode:
        handle_vehicle_input()

func handle_vehicle_input():
    var acceleration = Input.get_axis("ui_down", "ui_up")
    var steering = Input.get_axis("ui_left", "ui_right")
    
    # 引擎力
    if acceleration != 0:
        var engine_force = Vector2(acceleration * enginePower, 0)
        AddForce(engine_force, "Force")
    
    # 简单的转向（通过修改velocity方向）
    if steering != 0 and velocity.length() > 10:
        var turn_amount = steering * steeringPower * deg_to_rad(1)
        velocity = velocity.rotated(turn_amount * get_physics_process_delta_time())
    
    # 限制最大速度
    if velocity.length() > maxSpeed:
        velocity = velocity.normalized() * maxSpeed

func enter_vehicle(player):
    playerInside = true
    playerNode = player
    
    # 让玩家成为载具的子节点
    player.get_parent().remove_child(player)
    add_child(player)
    player.position = Vector2(0, -20)  # 坐在载具上的位置
    
    # 禁用玩家的物理处理（如果玩家也是UnityLikeBody2D）
    if player.has_method("set_physics_process"):
        player.set_physics_process(false)

func exit_vehicle():
    if playerNode:
        playerInside = false
        
        # 恢复玩家到场景
        remove_child(playerNode)
        get_tree().current_scene.add_child(playerNode)
        playerNode.global_position = global_position + Vector2(50, 0)  # 载具旁边
        
        # 重新启用玩家物理
        if playerNode.has_method("set_physics_process"):
            playerNode.set_physics_process(true)
        
        playerNode = null

func _input(event):
    if event.is_action_pressed("exit_vehicle") and playerInside:
        exit_vehicle()

# 载具进入检测区域
func _on_enter_area_body_entered(body):
    if body.is_in_group("player") and not playerInside:
        # 显示进入提示
        show_enter_prompt()

func show_enter_prompt():
    print("按 E 键进入载具")
```

## 高级功能扩展

### 约束系统

```gdscript
# UnityLikeBody2D.gd 的约束系统扩展

# 约束标志（类似Unity的RigidbodyConstraints）
enum ConstraintFlags {
    NONE = 0,
    FREEZE_POSITION_X = 1,
    FREEZE_POSITION_Y = 2,
    FREEZE_ROTATION = 4,
    FREEZE_ALL = 7
}

var constraints: int = ConstraintFlags.NONE

func _handle_constraints():
    if constraints & ConstraintFlags.FREEZE_POSITION_X:
        velocity.x = 0
    if constraints & ConstraintFlags.FREEZE_POSITION_Y:
        velocity.y = 0
    # 旋转约束可以通过禁用rotation属性实现

# 设置约束的便捷方法
func freeze_position_x():
    constraints |= ConstraintFlags.FREEZE_POSITION_X

func freeze_position_y():
    constraints |= ConstraintFlags.FREEZE_POSITION_Y

func freeze_rotation():
    constraints |= ConstraintFlags.FREEZE_ROTATION

func unfreeze_all():
    constraints = ConstraintFlags.NONE
```

### 材质系统

```gdscript
# 物理材质系统
class PhysicsMaterial2DUnity:
    var friction: float = 0.4
    var bounce: float = 0.0
    var friction_combine: String = "Average"  # Average, Minimum, Maximum, Multiply
    var bounce_combine: String = "Average"

var physicsMaterial: PhysicsMaterial2DUnity

func set_physics_material(material: PhysicsMaterial2DUnity):
    physicsMaterial = material
    _friction = material.friction * 1000  # 转换为Godot的摩擦力单位

func create_physics_material(friction: float, bounce: float = 0.0) -> PhysicsMaterial2DUnity:
    var material = PhysicsMaterial2DUnity.new()
    material.friction = friction
    material.bounce = bounce
    return material

# 处理弹性碰撞
func _handle_bounce_on_collision(collision_normal: Vector2):
    if physicsMaterial and physicsMaterial.bounce > 0:
        var bounce_velocity = velocity.bounce(collision_normal) * physicsMaterial.bounce
        velocity = bounce_velocity
```

### 性能优化系统

```gdscript
# 睡眠系统（优化性能）
var _sleep_threshold = 10.0
var _sleep_timer = 0.0
var _is_sleeping = false

func _check_sleep(delta):
    if isKinematic:
        return  # Kinematic对象不需要睡眠
    
    if velocity.length() < _sleep_threshold and is_on_floor():
        _sleep_timer += delta
        if _sleep_timer > 1.0 and not _is_sleeping:
            enter_sleep()
    else:
        if _is_sleeping:
            exit_sleep()
        _sleep_timer = 0.0

func enter_sleep():
    _is_sleeping = true
    set_physics_process(false)  # 停止物理更新
    # 可以添加睡眠视觉效果

func exit_sleep():
    _is_sleeping = false
    set_physics_process(true)   # 恢复物理更新

# 外部唤醒接口
func wake_up():
    if _is_sleeping:
        exit_sleep()
```

### 事件系统

```gdscript
# 物理事件信号
signal velocity_changed(old_velocity: Vector2, new_velocity: Vector2)
signal kinematic_state_changed(is_kinematic: bool)
signal collision_detected(collision_info: Dictionary)
signal force_applied(force: Vector2, force_mode: String)

# 在相应位置发出信号
func AddForce(force: Vector2, forceMode: String = "Force"):
    if isKinematic:
        return
    
    # ... 原有逻辑 ...
    
    emit_signal("force_applied", force, forceMode)

func set_kinematic(kinematic: bool):
    var old_kinematic = isKinematic
    isKinematic = kinematic
    
    # ... 原有逻辑 ...
    
    if old_kinematic != kinematic:
        emit_signal("kinematic_state_changed", kinematic)
```

## 使用指南

### 推荐设置

根据游戏对象类型，推荐以下配置：

```gdscript
# 玩家角色
isKinematic = false
useGravity = true
gravityScale = 1.0
mass = 1.0
drag = 0.0

# NPC/敌人（AI控制）
isKinematic = true
useGravity = false
mass = 1.0

# 可推动物品
isKinematic = false
useGravity = true
mass = 2.0
drag = 1.0

# 移动平台
isKinematic = true
useGravity = false

# 载具
isKinematic = false
useGravity = true
mass = 3.0
drag = 0.5
```

### 最佳实践

1. **统一继承**：所有需要物理交互的对象都继承自 `UnityLikeBody2D`
2. **合理设置**：根据对象用途合理设置 `isKinematic` 和其他物理参数
3. **性能考虑**：对于简单的装饰对象，考虑使用 `StaticBody2D` 而非此系统
4. **事件驱动**：利用信号系统实现松耦合的游戏逻辑
5. **渐进增强**：从基本功能开始，根据需要逐步添加高级功能

## 总结

这个统一物理控制器提供了：

- **Unity兼容的API**：`isKinematic`、`AddForce`、`velocity` 等
- **灵活的模式切换**：可以在运行时动态切换物理模式
- **完整的功能集**：重力、摩擦、约束、材质、性能优化等
- **易于扩展**：通过重写方法轻松添加自定义行为
- **高性能**：基于 `CharacterBody2D`，避免不必要的计算

通过继承这一个父类，你就能获得与Unity类似的物理控制体验，同时享受Godot的性能优势！