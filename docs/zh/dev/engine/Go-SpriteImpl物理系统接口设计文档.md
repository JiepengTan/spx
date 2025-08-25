# Go SpriteImpl 物理系统接口设计文档

## 概述

本文档定义了Go语言中`SpriteImpl`类型的物理系统接口设计，该接口基于Unity风格的物理控制系统，支持三种物理模式：Dynamic（动态物理）、Kinematic（运动学控制）、NoPhysics（无物理效果）。

## 设计原则

1. **类型安全**：使用强类型枚举和结构体
2. **错误处理**：所有可能失败的操作都返回error
3. **命名清晰**：遵循Go命名约定，方法名清楚表达功能
4. **性能考虑**：提供批量操作和性能控制选项
5. **易用性**：提供便捷方法和常用场景的预设
6. **可扩展**：接口设计支持未来功能扩展

## 基础数据结构

### 向量和枚举类型

```go
// Vector2 二维向量
type Vector2 struct {
    X, Y float32
}

// PhysicsMode 物理模式枚举
type PhysicsMode int

const (
    PhysicsDynamic    PhysicsMode = 0  // 受物理影响（类似Unity isKinematic = false）
    PhysicsKinematic  PhysicsMode = 1  // 代码控制（类似Unity isKinematic = true）
    PhysicsNoPhysics  PhysicsMode = 2  // 无物理效果（类似Unity无Collider/Rigidbody）
)

// ForceMode 力的模式
type ForceMode int

const (
    ForceNormal        ForceMode = 0  // 持续力
    ForceImpulse       ForceMode = 1  // 瞬间冲量
    ForceVelocityChange ForceMode = 2  // 直接改变速度
    ForceAcceleration  ForceMode = 3  // 加速度
)

// ConstraintFlags 约束标志
type ConstraintFlags int

const (
    ConstraintNone       ConstraintFlags = 0
    ConstraintFreezeX    ConstraintFlags = 1 << 0  // 冻结X轴位置
    ConstraintFreezeY    ConstraintFlags = 1 << 1  // 冻结Y轴位置
    ConstraintFreezeRot  ConstraintFlags = 1 << 2  // 冻结旋转
    ConstraintFreezeAll  ConstraintFlags = ConstraintFreezeX | ConstraintFreezeY | ConstraintFreezeRot
)
```

### 复合数据结构

```go
// PhysicsProperties 物理属性
type PhysicsProperties struct {
    Mass         float32  // 质量
    Drag         float32  // 线性阻力
    AngularDrag  float32  // 角阻力
    GravityScale float32  // 重力缩放
    UseGravity   bool     // 是否使用重力
}

// CollisionInfo 碰撞信息
type CollisionInfo struct {
    Other     *SpriteImpl  // 碰撞的另一个对象
    Point     Vector2      // 碰撞点
    Normal    Vector2      // 碰撞法线
    Impulse   Vector2      // 碰撞冲量
    Timestamp int64        // 碰撞时间戳
}

// PhysicsState 物理状态（用于保存/恢复）
type PhysicsState struct {
    Mode         PhysicsMode       // 物理模式
    Velocity     Vector2           // 速度
    Position     Vector2           // 位置
    Properties   PhysicsProperties // 物理属性
    Constraints  ConstraintFlags   // 约束设置
}

// PhysicsDebugInfo 调试信息
type PhysicsDebugInfo struct {
    Mode            string   // 当前物理模式
    CurrentVelocity Vector2  // 当前速度
    CurrentForces   Vector2  // 当前受力
    IsGrounded      bool     // 是否在地面
    CollisionCount  int      // 碰撞数量
    UpdateRate      float32  // 更新频率
}
```

## 核心接口定义

### 1. 物理模式控制接口

```go
// SetPhysicsMode 设置物理模式
// mode: 要设置的物理模式
// 返回: 设置失败时返回错误
func (p *SpriteImpl) SetPhysicsMode(mode PhysicsMode) error

// GetPhysicsMode 获取当前物理模式
// 返回: 当前的物理模式
func (p *SpriteImpl) GetPhysicsMode() PhysicsMode

// IsPhysicsEnabled 检查物理系统是否启用
// 返回: true表示物理系统启用（Dynamic或Kinematic模式）
func (p *SpriteImpl) IsPhysicsEnabled() bool

// 兼容Unity风格的接口（可选）
// SetKinematic 设置是否为Kinematic模式
// kinematic: true设置为Kinematic，false设置为Dynamic
func (p *SpriteImpl) SetKinematic(kinematic bool) error

// IsKinematic 检查是否为Kinematic模式
// 返回: true表示当前为Kinematic模式
func (p *SpriteImpl) IsKinematic() bool
```

### 2. 速度和位置控制接口

```go
// SetVelocity 设置速度向量
// velocity: 要设置的速度向量
func (p *SpriteImpl) SetVelocity(velocity Vector2) error

// GetVelocity 获取当前速度向量
// 返回: 当前速度向量
func (p *SpriteImpl) GetVelocity() Vector2

// SetVelocityX 仅设置X轴速度
// x: X轴速度值
func (p *SpriteImpl) SetVelocityX(x float32) error

// SetVelocityY 仅设置Y轴速度
// y: Y轴速度值
func (p *SpriteImpl) SetVelocityY(y float32) error

// SetPosition 设置位置
// pos: 目标位置
func (p *SpriteImpl) SetPosition(pos Vector2) error

// GetPosition 获取当前位置
// 返回: 当前位置
func (p *SpriteImpl) GetPosition() Vector2

// TeleportTo 瞬移到指定位置（清除速度）
// pos: 目标位置
func (p *SpriteImpl) TeleportTo(pos Vector2) error

// MoveWithoutPhysics 无物理模式下的移动
// direction: 移动方向（会被标准化）
// speed: 移动速度
func (p *SpriteImpl) MoveWithoutPhysics(direction Vector2, speed float32) error

// SetMovementVelocity 设置移动速度（NoPhysics模式专用）
// velocity: 移动速度向量
func (p *SpriteImpl) SetMovementVelocity(velocity Vector2) error
```

### 3. 力控制接口

```go
// AddForce 施加力
// force: 力向量
// mode: 力的模式（持续力/冲量/速度变化/加速度）
func (p *SpriteImpl) AddForce(force Vector2, mode ForceMode) error

// AddForceAtPosition 在指定位置施加力
// force: 力向量
// position: 施力位置
// mode: 力的模式
func (p *SpriteImpl) AddForceAtPosition(force Vector2, position Vector2, mode ForceMode) error

// ClearForces 清除所有外部力
func (p *SpriteImpl) ClearForces() error

// ApplyImpulse 施加冲量（便捷方法）
// impulse: 冲量向量
func (p *SpriteImpl) ApplyImpulse(impulse Vector2) error

// ApplyKnockback 施加击退力（便捷方法）
// knockback: 击退力向量
func (p *SpriteImpl) ApplyKnockback(knockback Vector2) error
```

### 4. 物理属性设置接口

```go
// SetPhysicsProperties 批量设置物理属性
// props: 物理属性结构体
func (p *SpriteImpl) SetPhysicsProperties(props PhysicsProperties) error

// GetPhysicsProperties 获取当前物理属性
// 返回: 当前物理属性
func (p *SpriteImpl) GetPhysicsProperties() PhysicsProperties

// SetMass 设置质量
// mass: 质量值（必须大于0）
func (p *SpriteImpl) SetMass(mass float32) error

// SetDrag 设置线性阻力
// drag: 阻力系数（0表示无阻力）
func (p *SpriteImpl) SetDrag(drag float32) error

// SetGravityScale 设置重力缩放
// scale: 重力缩放因子（1.0为正常重力，0.0为无重力）
func (p *SpriteImpl) SetGravityScale(scale float32) error

// SetUseGravity 设置是否使用重力
// enable: true启用重力，false禁用重力
func (p *SpriteImpl) SetUseGravity(enable bool) error

// GetMass 获取质量
func (p *SpriteImpl) GetMass() float32

// GetDrag 获取阻力系数
func (p *SpriteImpl) GetDrag() float32

// GetGravityScale 获取重力缩放
func (p *SpriteImpl) GetGravityScale() float32

// IsUsingGravity 检查是否使用重力
func (p *SpriteImpl) IsUsingGravity() bool
```

### 5. 碰撞和检测接口

```go
// IsGrounded 检查是否在地面上
// 返回: true表示在地面上
func (p *SpriteImpl) IsGrounded() bool

// IsOnWall 检查是否贴着墙壁
// 返回: true表示贴着墙壁
func (p *SpriteImpl) IsOnWall() bool

// IsOnCeiling 检查是否贴着天花板
// 返回: true表示贴着天花板
func (p *SpriteImpl) IsOnCeiling() bool

// IsCollisionEnabled 检查碰撞是否启用
// 返回: true表示碰撞启用
func (p *SpriteImpl) IsCollisionEnabled() bool

// CheckCollisionAtPoint 检查指定点是否有碰撞（NoPhysics模式用）
// point: 检查的点位置
// 返回: (是否有碰撞, 错误信息)
func (p *SpriteImpl) CheckCollisionAtPoint(point Vector2) (bool, error)

// CheckCollisionInRadius 检查指定半径内的碰撞
// radius: 检查半径
// 返回: (碰撞信息列表, 错误信息)
func (p *SpriteImpl) CheckCollisionInRadius(radius float32) ([]CollisionInfo, error)

// CanMoveTo 检查是否可以移动到指定位置
// targetPos: 目标位置
// 返回: (是否可以移动, 错误信息)
func (p *SpriteImpl) CanMoveTo(targetPos Vector2) (bool, error)

// SetCollisionEnabled 启用/禁用碰撞
// enabled: true启用碰撞，false禁用碰撞
func (p *SpriteImpl) SetCollisionEnabled(enabled bool) error
```

### 6. 约束控制接口

```go
// SetConstraints 设置位置/旋转约束
// constraints: 约束标志组合
func (p *SpriteImpl) SetConstraints(constraints ConstraintFlags) error

// GetConstraints 获取当前约束设置
// 返回: 当前约束标志
func (p *SpriteImpl) GetConstraints() ConstraintFlags

// FreezePosition 冻结位置（X和Y轴）
func (p *SpriteImpl) FreezePosition() error

// FreezeRotation 冻结旋转
func (p *SpriteImpl) FreezeRotation() error

// UnfreezeAll 解除所有约束
func (p *SpriteImpl) UnfreezeAll() error
```

### 7. 事件和回调接口

```go
// PhysicsEvent 物理事件接口
type PhysicsEvent interface {
    GetEventType() string
    GetTimestamp() int64
}

// 事件回调函数类型
type PhysicsEventCallback func(event PhysicsEvent)
type CollisionCallback func(collision CollisionInfo)
type VelocityChangeCallback func(oldVel, newVel Vector2)

// OnPhysicsModeChanged 注册物理模式变化回调
// callback: 回调函数
// 返回: 错误信息
func (p *SpriteImpl) OnPhysicsModeChanged(callback PhysicsEventCallback) error

// OnCollision 注册碰撞事件回调
// callback: 碰撞回调函数
func (p *SpriteImpl) OnCollision(callback CollisionCallback) error

// OnVelocityChanged 注册速度变化回调
// callback: 速度变化回调函数
func (p *SpriteImpl) OnVelocityChanged(callback VelocityChangeCallback) error

// RemovePhysicsCallback 移除指定回调
// callbackID: 回调标识符
func (p *SpriteImpl) RemovePhysicsCallback(callbackID string) error

// ClearAllCallbacks 清除所有回调
func (p *SpriteImpl) ClearAllCallbacks() error
```

### 8. 实用工具接口

```go
// SavePhysicsState 保存当前物理状态
// 返回: (物理状态, 错误信息)
func (p *SpriteImpl) SavePhysicsState() (PhysicsState, error)

// RestorePhysicsState 恢复物理状态
// state: 要恢复的物理状态
func (p *SpriteImpl) RestorePhysicsState(state PhysicsState) error

// TemporarilyDisablePhysics 临时禁用物理（保存当前状态）
func (p *SpriteImpl) TemporarilyDisablePhysics() error

// RestorePreviousPhysics 恢复之前的物理状态
func (p *SpriteImpl) RestorePreviousPhysics() error

// SetUpdateEnabled 启用/禁用物理更新
// enabled: true启用更新，false禁用更新
func (p *SpriteImpl) SetUpdateEnabled(enabled bool) error

// SetUpdateInterval 设置更新间隔（NoPhysics模式性能优化用）
// interval: 更新间隔（秒）
func (p *SpriteImpl) SetUpdateInterval(interval float32) error

// IsPhysicsActive 检查物理系统是否活跃
// 返回: true表示物理系统正在更新
func (p *SpriteImpl) IsPhysicsActive() bool

// EnableDebugDraw 启用/禁用调试绘制
// enabled: true启用调试绘制
func (p *SpriteImpl) EnableDebugDraw(enabled bool) error

// GetDebugInfo 获取调试信息
// 返回: 调试信息结构体
func (p *SpriteImpl) GetDebugInfo() PhysicsDebugInfo
```

## 便捷组合接口

### 9. 高级操作接口

```go
// SetupAsPlayer 设置为玩家角色
// speed: 移动速度
// jumpForce: 跳跃力度
func (p *SpriteImpl) SetupAsPlayer(speed, jumpForce float32) error

// SetupAsNPC 设置为NPC
// patrolSpeed: 巡逻速度
func (p *SpriteImpl) SetupAsNPC(patrolSpeed float32) error

// SetupAsPhysicsObject 设置为物理对象
// mass: 质量
// drag: 阻力
func (p *SpriteImpl) SetupAsPhysicsObject(mass, drag float32) error

// StartKnockback 开始击飞状态
// force: 击飞力量
// duration: 击飞持续时间（秒）
func (p *SpriteImpl) StartKnockback(force Vector2, duration float32) error

// IsInKnockback 检查是否在击飞状态
// 返回: true表示正在被击飞
func (p *SpriteImpl) IsInKnockback() bool

// StopKnockback 停止击飞状态
func (p *SpriteImpl) StopKnockback() error

// EnterVehicle 进入载具
// vehicle: 载具对象
func (p *SpriteImpl) EnterVehicle(vehicle *SpriteImpl) error

// ExitVehicle 退出载具
func (p *SpriteImpl) ExitVehicle() error

// IsInVehicle 检查是否在载具中
// 返回: true表示在载具中
func (p *SpriteImpl) IsInVehicle() bool

// SetupAsMovingPlatform 设置为移动平台
// path: 移动路径点数组
// duration: 完成一个循环的时间（秒）
func (p *SpriteImpl) SetupAsMovingPlatform(path []Vector2, duration float32) error

// AttachToParent 附加到父对象
// parent: 父对象
func (p *SpriteImpl) AttachToParent(parent *SpriteImpl) error

// DetachFromParent 从父对象分离
func (p *SpriteImpl) DetachFromParent() error
```

## 使用示例

### 基础使用

```go
package main

import "fmt"

func main() {
    // 创建玩家角色
    player := NewSpriteImpl()
    
    // 设置为动态物理模式
    err := player.SetPhysicsMode(PhysicsDynamic)
    if err != nil {
        fmt.Printf("设置物理模式失败: %v\n", err)
        return
    }
    
    // 设置物理属性
    props := PhysicsProperties{
        Mass:         1.0,
        Drag:         0.0,
        GravityScale: 1.0,
        UseGravity:   true,
    }
    player.SetPhysicsProperties(props)
    
    // 水平移动
    currentVel := player.GetVelocity()
    player.SetVelocity(Vector2{X: 300, Y: currentVel.Y})
    
    // 跳跃
    if player.IsGrounded() {
        player.ApplyImpulse(Vector2{X: 0, Y: -400})
    }
}
```

### 装饰性对象

```go
func createDecorationNPC() *SpriteImpl {
    decoration := NewSpriteImpl()
    
    // 设置为无物理模式
    decoration.SetPhysicsMode(PhysicsNoPhysics)
    
    // 设置浮动效果
    decoration.SetMovementVelocity(Vector2{X: 0, Y: 50})
    
    return decoration
}
```

### 击飞系统

```go
func setupFighter() *SpriteImpl {
    fighter := NewSpriteImpl()
    fighter.SetPhysicsMode(PhysicsDynamic)
    
    // 注册碰撞回调
    fighter.OnCollision(func(collision CollisionInfo) {
        if collision.Other.HasTag("enemy") {
            // 受到击飞
            knockback := Vector2{X: 300, Y: -200}
            fighter.StartKnockback(knockback, 0.8)
        }
    })
    
    return fighter
}
```

### 移动平台

```go
func createMovingPlatform() *SpriteImpl {
    platform := NewSpriteImpl()
    
    // 设置移动路径
    path := []Vector2{
        {X: 0, Y: 0},
        {X: 200, Y: 0},
        {X: 200, Y: 100},
        {X: 0, Y: 100},
    }
    
    platform.SetupAsMovingPlatform(path, 5.0) // 5秒完成一个循环
    
    return platform
}
```

### 载具系统

```go
func setupVehicle() *SpriteImpl {
    vehicle := NewSpriteImpl()
    vehicle.SetPhysicsMode(PhysicsDynamic)
    vehicle.SetMass(3.0)
    vehicle.SetDrag(0.5)
    
    return vehicle
}

func playerEnterVehicle(player, vehicle *SpriteImpl) {
    err := player.EnterVehicle(vehicle)
    if err != nil {
        fmt.Printf("进入载具失败: %v\n", err)
    }
}
```

### 调试和监控

```go
func debugPhysicsObject(obj *SpriteImpl) {
    // 启用调试绘制
    obj.EnableDebugDraw(true)
    
    // 获取调试信息
    debug := obj.GetDebugInfo()
    fmt.Printf("物理模式: %s\n", debug.Mode)
    fmt.Printf("当前速度: (%.2f, %.2f)\n", debug.CurrentVelocity.X, debug.CurrentVelocity.Y)
    fmt.Printf("是否在地面: %t\n", debug.IsGrounded)
    fmt.Printf("碰撞数量: %d\n", debug.CollisionCount)
}
```

## 错误处理

```go
// 定义常见错误类型
var (
    ErrInvalidPhysicsMode = errors.New("无效的物理模式")
    ErrInvalidMass       = errors.New("质量必须大于0")
    ErrNullPointer       = errors.New("空指针引用")
    ErrPhysicsDisabled   = errors.New("物理系统未启用")
    ErrInvalidVelocity   = errors.New("无效的速度值")
)

// 错误处理示例
func safeSetVelocity(sprite *SpriteImpl, velocity Vector2) {
    if err := sprite.SetVelocity(velocity); err != nil {
        switch err {
        case ErrPhysicsDisabled:
            fmt.Println("警告: 物理系统未启用，无法设置速度")
        case ErrInvalidVelocity:
            fmt.Printf("错误: 无效的速度值 (%.2f, %.2f)\n", velocity.X, velocity.Y)
        default:
            fmt.Printf("设置速度时发生未知错误: %v\n", err)
        }
    }
}
```

## 性能优化建议

### 1. 批量操作
```go
// 批量设置多个属性
props := PhysicsProperties{
    Mass: 2.0,
    Drag: 1.0,
    GravityScale: 1.5,
    UseGravity: true,
}
sprite.SetPhysicsProperties(props) // 一次设置多个属性
```

### 2. 条件检查
```go
// 避免不必要的操作
if sprite.GetPhysicsMode() != PhysicsNoPhysics {
    sprite.AddForce(force, ForceImpulse)
}
```

### 3. 回调管理
```go
// 及时清理回调，避免内存泄漏
defer sprite.ClearAllCallbacks()
```

## 线程安全说明

所有接口操作都应该在主线程中调用。如果需要在其他goroutine中操作物理对象，应该使用通道或其他同步机制：

```go
// 线程安全的物理操作
type SafePhysicsOperation struct {
    sprite *SpriteImpl
    mutex  sync.Mutex
}

func (s *SafePhysicsOperation) SetVelocity(velocity Vector2) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    return s.sprite.SetVelocity(velocity)
}
```

## 总结

这套接口设计提供了完整的物理系统控制能力，包括：

- **三种物理模式**：Dynamic、Kinematic、NoPhysics
- **完整的力学控制**：速度、位置、力、约束
- **事件系统**：碰撞、状态变化回调
- **便捷功能**：预设配置、击飞系统、载具系统
- **调试支持**：状态查询、可视化调试
- **性能优化**：批量操作、更新控制

接口设计遵循Go语言习惯，提供了类型安全和错误处理，同时保持了Unity风格的易用性。通过这套接口，开发者可以轻松实现各种复杂的物理交互效果。