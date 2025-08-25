# Godot 2D 物理系统控制指南

## Unity vs Godot 对应关系

| Unity概念 | Godot对应 | 说明 |
|-----------|-----------|------|
| `Rigidbody2D.isKinematic = false` | `RigidBody2D` 默认模式 | 受物理影响 |
| `Rigidbody2D.isKinematic = true` | `RigidBody2D.freeze = true` 或 `CharacterBody2D` | 不受物理影响 |
| `Rigidbody2D.velocity` | `RigidBody2D.linear_velocity` | 速度控制 |
| `Rigidbody2D.AddForce` | `RigidBody2D.apply_impulse()` | 施加力/冲量 |
| 角色控制器 | `CharacterBody2D` | 专门的角色控制节点 |

## 基础节点结构

在Godot中，物理对象通常使用：

```
CharacterBody2D (角色控制，推荐用于玩家)
├── CollisionShape2D
└── Sprite2D

RigidBody2D (物理模拟，推荐用于物品/敌人)
├── CollisionShape2D  
└── Sprite2D

StaticBody2D (静态碰撞体，推荐用于地面/墙壁)
├── CollisionShape2D
└── Sprite2D

AnimatableBody2D (可动画的静态体，推荐用于移动平台)
├── CollisionShape2D
└── Sprite2D
```

## 1. 实现 isKinematic = false（受物理影响）

### 方案1：使用 RigidBody2D（默认行为）

```gdscript
extends RigidBody2D

func _ready():
    # RigidBody2D 默认就是受物理影响的
    freeze = false  # 默认值，可以省略
    
    # 可以调整物理属性
    mass = 1.0
    gravity_scale = 1.0
    linear_damp = 0.0
    angular_damp = 0.0
```

### 方案2：使用 CharacterBody2D + 手动物理

```gdscript
extends CharacterBody2D

var gravity = ProjectSettings.get_setting("physics/2d/default_gravity")
var friction = 500.0

func _physics_process(delta):
    # 手动应用重力（受物理影响）
    if not is_on_floor():
        velocity.y += gravity * delta
    
    # 应用摩擦力
    velocity.x = move_toward(velocity.x, 0, friction * delta)
    
    move_and_slide()
```

## 2. 实现 isKinematic = true（不受物理影响）

### 方案1：冻结 RigidBody2D

```gdscript
extends RigidBody2D

func _ready():
    # 类似Unity的 isKinematic = true
    freeze = true
    freeze_mode = RigidBody2D.FREEZE_MODE_KINEMATIC

func move_kinematic(direction: Vector2, delta: float):
    # 通过代码精确控制移动，不受物理影响
    position += direction * delta
```

### 方案2：使用 CharacterBody2D（推荐）

```gdscript
extends CharacterBody2D

# CharacterBody2D 本身就类似 isKinematic = true
# 不自动受重力和外力影响，完全由代码控制

func _physics_process(delta):
    # 只执行你想要的移动，忽略重力等物理效果
    var direction = get_movement_direction()
    velocity = direction * speed
    move_and_slide()
```

### 方案3：使用 AnimatableBody2D

```gdscript
extends AnimatableBody2D

# 专门用于移动平台等需要动画但不受物理影响的对象

func _ready():
    # 使用Tween或AnimationPlayer控制移动
    var tween = create_tween()
    tween.set_loops()
    tween.tween_property(self, "position", Vector2(200, 0), 2.0)
    tween.tween_property(self, "position", Vector2(0, 0), 2.0)
```

## 3. 实现 AddForce 效果

### RigidBody2D 的力控制

```gdscript
extends RigidBody2D

func _ready():
    # 设置为受物理影响模式
    freeze = false

# 施加冲量（瞬间力，类似Unity的AddForce with ForceMode.Impulse）
func apply_force_impulse(force: Vector2):
    apply_central_impulse(force)

# 施加持续力（类似Unity的AddForce）  
func apply_continuous_force(force: Vector2):
    apply_central_force(force)

# 在特定点施加冲量
func apply_force_at_point(force: Vector2, point: Vector2):
    apply_impulse(force, point)

# 使用示例
func _input(event):
    if event.is_action_pressed("jump"):
        apply_central_impulse(Vector2(0, -500))  # 向上跳跃
    
    if event.is_action_pressed("push"):
        apply_central_force(Vector2(100, 0))  # 持续向右推
```

### CharacterBody2D 的模拟力效果

```gdscript
extends CharacterBody2D

var external_forces = Vector2.ZERO
var force_decay = 0.9

func _physics_process(delta):
    # 应用外部力
    velocity += external_forces * delta
    
    # 力的衰减
    external_forces *= force_decay
    
    move_and_slide()

# 模拟施加冲量
func apply_impulse(impulse: Vector2):
    # 直接设置速度
    velocity += impulse

# 模拟施加持续力
func add_force(force: Vector2):
    external_forces += force

# 使用示例
func take_knockback(knockback_force: Vector2):
    apply_impulse(knockback_force)
```

## 4. 重力实现

### 方案1：使用 RigidBody2D 自动重力

```gdscript
extends RigidBody2D

func _ready():
    # 使用项目设置的重力
    gravity_scale = 1.0  # 默认重力
    
    # 自定义重力强度
    gravity_scale = 2.0  # 双倍重力
    gravity_scale = 0.5  # 一半重力
    gravity_scale = 0.0  # 无重力
```

### 方案2：CharacterBody2D 手动重力

```gdscript
extends CharacterBody2D

# 获取项目默认重力
var gravity = ProjectSettings.get_setting("physics/2d/default_gravity")

func _physics_process(delta):
    # 基础重力实现
    if not is_on_floor():
        velocity.y += gravity * delta
    
    move_and_slide()
```

### 方案3：自定义重力系统

```gdscript
extends CharacterBody2D

var custom_gravity = 980.0
var max_fall_speed = 1000.0
var gravity_scale = 1.0

func _physics_process(delta):
    # 可变重力实现
    if not is_on_floor():
        var current_gravity = custom_gravity * gravity_scale
        velocity.y += current_gravity * delta
        
        # 限制最大下落速度
        velocity.y = min(velocity.y, max_fall_speed)
    
    move_and_slide()

# 动态调整重力
func set_gravity_scale(scale: float):
    gravity_scale = scale

# 使用示例
func enter_water():
    set_gravity_scale(0.3)  # 水中重力减小

func exit_water():
    set_gravity_scale(1.0)  # 恢复正常重力
```

### 方案4：高级重力控制

```gdscript
extends CharacterBody2D

var gravity = ProjectSettings.get_setting("physics/2d/default_gravity")
var jump_gravity_scale = 0.6  # 上升时重力较小
var fall_gravity_scale = 1.4  # 下落时重力较大

func _physics_process(delta):
    if not is_on_floor():
        if velocity.y < 0:  # 向上运动（跳跃）
            velocity.y += gravity * jump_gravity_scale * delta
        else:  # 向下运动（下落）
            velocity.y += gravity * fall_gravity_scale * delta
    
    move_and_slide()
```

## 5. 摩擦力实现

### 方案1：RigidBody2D 物理摩擦

```gdscript
extends RigidBody2D

func _ready():
    # 使用物理材质设置摩擦
    var physics_material = PhysicsMaterial.new()
    physics_material.friction = 0.8  # 摩擦系数
    physics_material.bounce = 0.2    # 弹性系数
    
    # 应用到碰撞形状
    var collision_shape = $CollisionShape2D
    collision_shape.physics_material_override = physics_material
    
    # 或者直接设置阻力
    linear_damp = 1.0   # 线性阻力
    angular_damp = 1.0  # 角阻力
```

### 方案2：CharacterBody2D 手动摩擦

```gdscript
extends CharacterBody2D

var friction = 500.0        # 地面摩擦力
var air_resistance = 50.0   # 空气阻力

func _physics_process(delta):
    if is_on_floor():
        # 地面摩擦
        velocity.x = move_toward(velocity.x, 0, friction * delta)
    else:
        # 空气阻力
        velocity.x = move_toward(velocity.x, 0, air_resistance * delta)
    
    move_and_slide()
```

### 方案3：基于材质的摩擦系统

```gdscript
extends CharacterBody2D

var base_friction = 300.0
var current_friction = base_friction

func _physics_process(delta):
    # 应用摩擦力
    if is_on_floor() and velocity.x != 0:
        var friction_force = current_friction * delta
        if abs(velocity.x) <= friction_force:
            velocity.x = 0  # 完全停止
        else:
            velocity.x -= sign(velocity.x) * friction_force
    
    move_and_slide()

# 检测地面材质
func _on_area_2d_body_entered(body):
    if body.has_meta("surface_type"):
        match body.get_meta("surface_type"):
            "ice":
                current_friction = base_friction * 0.1
            "sand":
                current_friction = base_friction * 2.0
            _:
                current_friction = base_friction
```

## 用户故事实现示例

### 1. 制作横版闯关游戏 - 玩家角色

**需求**：*"我要制作类似《超级马里奥》的游戏，玩家可以跑跳，会受重力影响，但移动要很精确"*

```gdscript
extends CharacterBody2D

@export var speed = 300.0
@export var jump_velocity = -400.0
@export var acceleration = 1500.0
@export var friction = 1200.0

# 获取重力设置
var gravity = ProjectSettings.get_setting("physics/2d/default_gravity")

func _physics_process(delta):
    # 添加重力
    if not is_on_floor():
        velocity.y += gravity * delta

    # 处理跳跃
    if Input.is_action_just_pressed("ui_accept") and is_on_floor():
        velocity.y = jump_velocity

    # 处理水平移动
    var direction = Input.get_axis("ui_left", "ui_right")
    
    if direction != 0:
        # 加速移动
        velocity.x = move_toward(velocity.x, direction * speed, acceleration * delta)
    else:
        # 应用摩擦力停止
        velocity.x = move_toward(velocity.x, 0, friction * delta)

    move_and_slide()
    
    # 翻转精灵
    if direction != 0:
        $Sprite2D.flip_h = direction < 0
```

### 2. 制作RPG游戏 - NPC商人

**需求**：*"商人NPC按固定路线巡逻，玩家撞到时能对话，但不能被推走"*

```gdscript
extends CharacterBody2D

@export var patrol_speed = 50.0
@export var patrol_points: Array[Vector2]

var current_target = 0
var dialogue_active = false

func _ready():
    # CharacterBody2D 默认不受外力影响（类似isKinematic = true）
    if patrol_points.is_empty():
        patrol_points = [global_position, global_position + Vector2(100, 0)]

func _physics_process(delta):
    if not dialogue_active:
        patrol_movement(delta)

func patrol_movement(delta):
    if patrol_points.size() < 2:
        return
        
    var target_pos = patrol_points[current_target]
    var direction = (target_pos - global_position).normalized()
    
    # 精确控制移动，不受物理推动
    velocity = direction * patrol_speed
    move_and_slide()
    
    # 检查是否到达目标点
    if global_position.distance_to(target_pos) < 10.0:
        current_target = (current_target + 1) % patrol_points.size()
        # 翻转精灵方向
        $Sprite2D.flip_h = direction.x < 0

# 碰撞检测
func _on_dialogue_area_body_entered(body):
    if body.is_in_group("player"):
        dialogue_active = true
        velocity = Vector2.ZERO  # 停止移动
        start_dialogue()

func _on_dialogue_area_body_exited(body):
    if body.is_in_group("player"):
        dialogue_active = false

func start_dialogue():
    print("欢迎光临我的商店！")
```

### 3. 制作平台游戏 - 移动平台

**需求**：*"上下移动的平台，玩家站上去被带动，平台不被任何东西影响"*

```gdscript
extends AnimatableBody2D  # 专门用于移动平台

@export var move_distance = 200.0
@export var move_duration = 3.0

var start_position: Vector2
var players_on_platform: Array = []

func _ready():
    start_position = global_position
    start_moving()

func start_moving():
    var tween = create_tween()
    tween.set_loops()
    
    # 向上移动
    tween.tween_property(self, "global_position", 
        start_position + Vector2(0, -move_distance), move_duration)
    tween.tween_delay(0.5)
    
    # 向下移动
    tween.tween_property(self, "global_position", 
        start_position, move_duration) 
    tween.tween_delay(0.5)

# 检测玩家踩上平台
func _on_player_detector_body_entered(body):
    if body.is_in_group("player"):
        players_on_platform.append(body)

func _on_player_detector_body_exited(body):
    if body.is_in_group("player"):
        players_on_platform.erase(body)

# 如果使用CharacterBody2D实现移动平台
extends CharacterBody2D

@export var amplitude = 100.0
@export var frequency = 1.0

var start_position: Vector2

func _ready():
    start_position = global_position

func _physics_process(delta):
    # 正弦波移动
    var time = Time.get_time()
    global_position.y = start_position.y + sin(time * frequency) * amplitude
    
    # 计算移动速度，用于推动玩家
    velocity.y = cos(time * frequency) * frequency * amplitude
    
    move_and_slide()
```

### 4. 制作动作游戏 - 受击系统

**需求**：*"玩家被攻击时被击飞，击飞中不能控制，结束后恢复控制"*

```gdscript
extends CharacterBody2D

@export var speed = 200.0
@export var jump_velocity = -400.0

var gravity = ProjectSettings.get_setting("physics/2d/default_gravity")
var is_knocked_back = false
var knockback_timer = 0.0
var knockback_duration = 0.8

func _physics_process(delta):
    # 添加重力
    if not is_on_floor():
        velocity.y += gravity * delta
    
    if is_knocked_back:
        handle_knockback(delta)
    else:
        handle_normal_input(delta)
    
    move_and_slide()

func handle_normal_input(delta):
    # 正常输入处理
    if Input.is_action_just_pressed("ui_accept") and is_on_floor():
        velocity.y = jump_velocity

    var direction = Input.get_axis("ui_left", "ui_right")
    if direction != 0:
        velocity.x = direction * speed
    else:
        velocity.x = move_toward(velocity.x, 0, speed)

func handle_knockback(delta):
    knockback_timer -= delta
    
    # 击飞期间应用空气阻力
    velocity.x = move_toward(velocity.x, 0, 100 * delta)
    
    # 击飞结束
    if knockback_timer <= 0:
        is_knocked_back = false

func take_knockback(knockback_force: Vector2):
    # 开始击飞
    is_knocked_back = true
    knockback_timer = knockback_duration
    
    # 直接设置击飞速度
    velocity = knockback_force
    
    # 视觉反馈
    flash_damage_effect()

func flash_damage_effect():
    var tween = create_tween()
    tween.tween_property($Sprite2D, "modulate", Color.RED, 0.1)
    tween.tween_property($Sprite2D, "modulate", Color.WHITE, 0.1)

# 外部调用示例
func _on_enemy_attack():
    var knockback = Vector2(300, -200)  # 向右上方击飞
    take_knockback(knockback)
```

### 5. 制作解谜游戏 - 推箱子机关

**需求**：*"有些箱子可推动解谜，有些是装饰，推动箱子要有真实物理感"*

```gdscript
extends RigidBody2D

@export var can_be_pushed = true
@export var push_force_threshold = 100.0

func _ready():
    if can_be_pushed:
        # 可推动箱子 - 使用物理
        freeze = false
        mass = 2.0
        linear_damp = 5.0  # 阻力，防止滑太远
        
        # 设置物理材质
        var physics_material = PhysicsMaterial.new()
        physics_material.friction = 0.8
        physics_material.bounce = 0.1
        $CollisionShape2D.physics_material_override = physics_material
        
    else:
        # 装饰箱子 - 不可移动（类似isKinematic = true）
        freeze = true
        freeze_mode = RigidBody2D.FREEZE_MODE_STATIC

func _on_body_entered(body):
    if body.is_in_group("player") and can_be_pushed:
        # 可以添加推动音效
        play_push_sound()

# 检测是否在正确位置（解谜用）
func _on_puzzle_area_body_entered(body):
    if body == self:
        emit_signal("box_in_correct_position")
        # 可以添加视觉反馈
        $Sprite2D.modulate = Color.GREEN

func _on_puzzle_area_body_exited(body):
    if body == self:
        emit_signal("box_left_position")
        $Sprite2D.modulate = Color.WHITE

signal box_in_correct_position
signal box_left_position

func play_push_sound():
    # 播放推箱子音效
    $AudioStreamPlayer2D.play()
```

### 6. 制作竞速游戏 - 载具控制

**需求**：*"玩家可开车，上车时角色跟着车动，下车时恢复步行"*

```gdscript
# 载具脚本
extends RigidBody2D

@export var engine_power = 800.0
@export var steering_power = 30.0

var player_inside = false
var player_node = null

func _ready():
    # 载具使用物理引擎
    freeze = false
    mass = 5.0

func _physics_process(delta):
    if player_inside and player_node:
        handle_vehicle_input()

func handle_vehicle_input():
    var acceleration = Input.get_axis("ui_down", "ui_up")
    var steering = Input.get_axis("ui_left", "ui_right")
    
    # 引擎力
    if acceleration != 0:
        apply_central_force(transform.x * acceleration * engine_power)
    
    # 转向力矩
    if steering != 0 and linear_velocity.length() > 10:
        apply_torque(steering * steering_power)

func enter_vehicle(player):
    player_inside = true
    player_node = player
    
    # 让玩家成为载具的子节点
    player.get_parent().remove_child(player)
    add_child(player)
    player.position = Vector2(0, -20)  # 坐在载具上的位置
    
    # 禁用玩家的物理处理
    player.set_physics_process(false)

func exit_vehicle():
    if player_node:
        player_inside = false
        
        # 恢复玩家到场景
        remove_child(player_node)
        get_tree().current_scene.add_child(player_node)
        player_node.global_position = global_position + Vector2(50, 0)  # 载具旁边
        
        # 重新启用玩家物理
        player_node.set_physics_process(true)
        player_node = null

func _input(event):
    if event.is_action_pressed("exit_vehicle") and player_inside:
        exit_vehicle()

# 玩家脚本中的对应代码
extends CharacterBody2D

var in_vehicle = false

func _on_vehicle_area_body_entered(body):
    if body.has_method("enter_vehicle"):
        body.enter_vehicle(self)
        in_vehicle = true

func _physics_process(delta):
    # 只有不在载具中时才处理正常移动
    if not in_vehicle:
        handle_normal_movement(delta)
```

## 推荐的Godot 2D游戏物理架构

```gdscript
# 玩家角色 - CharacterBody2D
extends CharacterBody2D
# 优势：精确控制、内置碰撞检测、性能好

# 敌人/怪物 - 根据需要选择
extends RigidBody2D     # 需要真实物理交互时
extends CharacterBody2D # 需要精确AI控制时

# 物品/可推动对象 - RigidBody2D  
extends RigidBody2D
# 优势：真实物理交互、自动重力和碰撞

# 移动平台 - AnimatableBody2D
extends AnimatableBody2D
# 优势：可以移动但不受物理影响，专门为此设计

# 静态地面/墙壁 - StaticBody2D
extends StaticBody2D
# 优势：完全静态、性能最好
```

## 总结

Godot的2D物理系统提供了比Unity更细粒度的控制：

- **完全物理控制**：使用 `RigidBody2D`（类似Unity `isKinematic = false`）
- **完全代码控制**：使用 `CharacterBody2D`（类似Unity `isKinematic = true`）  
- **混合控制**：可以在 `CharacterBody2D` 中选择性应用物理效果
- **专门化节点**：`AnimatableBody2D` 专为移动平台设计

关键优势是每种节点类型都针对特定用例进行了优化，让开发者能够选择最适合的解决方案。