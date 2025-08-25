# Go SpriteImpl 极简物理API设计文档

## 设计理念

采用CodeMonkey风格的极简API设计，只保留最核心的8个方法，覆盖90%的游戏物理需求。简洁、直观、易用。

## 基础数据结构

```go
// Vector2 二维向量
type Vector2 struct {
    X, Y float32
}

// PhysicsMode 物理模式
type PhysicsMode int

const (
    Dynamic    PhysicsMode = 0  // 受物理影响（重力、碰撞）
    Kinematic  PhysicsMode = 1  // 代码控制移动，有碰撞检测
    NoPhysics  PhysicsMode = 2  // 纯视觉，无碰撞，无物理
)
```

## 核心API定义

### 1. 物理控制接口（4个核心方法）

```go
// SetPhysicsMode 设置物理模式
// 
// 参数:
//   mode: 物理模式
//     - Dynamic: 受重力和外力影响，有碰撞检测（类似Unity isKinematic=false）
//     - Kinematic: 只能通过代码控制移动，有碰撞检测（类似Unity isKinematic=true）
//     - NoPhysics: 纯视觉对象，无碰撞，无物理效果（类似Unity无Collider）
// 
// 返回:
//   error: 设置失败时返回错误
//
// 使用场景:
//   - 玩家角色使用Dynamic模式
//   - NPC巡逻使用Kinematic模式  
//   - 背景装饰使用NoPhysics模式
func (p *SpriteImpl) SetPhysicsMode(mode PhysicsMode) error
```

```go
// SetVelocity 设置速度向量
//
// 参数:
//   velocity: 速度向量，单位：像素/秒
//     - X: 水平速度，正值向右，负值向左
//     - Y: 垂直速度，正值向上，负值向下（注意：重力会影响Y轴速度）
//
// 返回:
//   error: 设置失败时返回错误（如物理模式为NoPhysics时可能失败）
//
// 注意事项:
//   - Dynamic模式：设置的速度会受重力和外力影响而改变
//   - Kinematic模式：精确按设定速度移动，不受重力影响  
//   - NoPhysics模式：直接用于transform移动，不调用物理引擎
//
// 常用模式:
//   - 角色水平移动：SetVelocity(Vector2{X: speed * direction, Y: GetVelocity().Y})
//   - 停止移动：SetVelocity(Vector2{X: 0, Y: GetVelocity().Y})
func (p *SpriteImpl) SetVelocity(velocity Vector2) error
```

```go
// SetGravity 设置重力缩放
//
// 参数:
//   gravity: 重力缩放因子
//     - 1.0: 正常重力
//     - 0.0: 无重力（漂浮效果）
//     - 2.0: 双倍重力（快速下落）
//     - -1.0: 反重力（向上飞行）
//     - 0.3: 水中效果（缓慢下沉）
//
// 返回:
//   error: 设置失败时返回错误
//
// 注意事项:
//   - 只对Dynamic模式有效
//   - Kinematic和NoPhysics模式会忽略重力设置
//   - 可以在运行时动态调整，实现各种环境效果
//
// 使用场景:
//   - 正常角色: SetGravity(1.0)
//   - 游泳状态: SetGravity(0.2)  
//   - 飞行道具: SetGravity(0.0)
//   - 快速下落: SetGravity(3.0)
func (p *SpriteImpl) SetGravity(gravity float32) error
```

```go
// AddImpulse 施加瞬间冲量
//
// 参数:
//   impulse: 冲量向量，单位：像素/秒
//     - X: 水平冲量，正值向右推，负值向左推
//     - Y: 垂直冲量，正值向上推，负值向下推
//
// 返回:
//   error: 施加失败时返回错误
//
// 特点:
//   - 瞬间改变速度，不是持续施力
//   - 只对Dynamic模式有效
//   - 会叠加到当前速度上：新速度 = 当前速度 + 冲量
//
// 常用场景:
//   - 跳跃: AddImpulse(Vector2{X: 0, Y: -400})
//   - 击退: AddImpulse(Vector2{X: -300, Y: -100})  
//   - 爆炸推力: AddImpulse(Vector2{X: 200, Y: -200})
//   - 弹跳: AddImpulse(Vector2{X: 0, Y: -300})
//
// 技巧:
//   - 可以用来实现二段跳、冲刺、击飞等效果
//   - 与SetVelocity的区别：AddImpulse是累加，SetVelocity是替换
func (p *SpriteImpl) AddImpulse(impulse Vector2) error
```

### 2. 基础属性接口（2个必需方法）

```go
// SetPosition 设置位置（瞬移）
//
// 参数:
//   pos: 目标位置，单位：像素
//     - X: 水平位置，0为屏幕左侧
//     - Y: 垂直位置，0为屏幕上方
//
// 返回:
//   error: 设置失败时返回错误
//
// 特点:
//   - 瞬间移动，不经过中间位置
//   - 会自动清除当前速度，避免意外移动
//   - 对所有物理模式都有效
//
// 使用场景:
//   - 游戏开始时设置初始位置
//   - 传送门效果
//   - 重生点复活
//   - 场景切换时的位置重置
//   - 调试时快速移动到指定位置
//
// 注意事项:
//   - 不会触发碰撞检测过程
//   - 如果目标位置有障碍物，对象可能卡在墙内
func (p *SpriteImpl) SetPosition(pos Vector2) error
```

```go
// SetColliderSize 设置碰撞器大小
//
// 参数:
//   width: 碰撞器宽度，单位：像素
//   height: 碰撞器高度，单位：像素
//
// 返回:
//   error: 设置失败时返回错误（如尺寸为负值）
//
// 特点:
//   - 碰撞器是矩形，以对象中心为基准
//   - 只影响碰撞检测，不影响视觉显示
//   - 可以在运行时动态调整
//
// 使用场景:
//   - 角色变身时调整碰撞体积
//   - 不同状态下的碰撞范围（蹲下时变矮）
//   - 精确调整碰撞检测的感觉
//   - 区分显示大小和碰撞大小
//
// 建议:
//   - 碰撞器通常比显示图像略小，提供更好的游戏手感
//   - 玩家角色可以设置得稍小一些，增加容错性
//   - 敌人可以设置得稍大一些，增加挑战性
func (p *SpriteImpl) SetColliderSize(width, height float32) error
```

### 3. 查询接口（2个实用方法）

```go
// IsGrounded 检查是否在地面上
//
// 返回:
//   bool: true表示脚下有可站立的表面，false表示在空中
//
// 检测逻辑:
//   - 从对象底部向下发射短距离射线
//   - 检测是否碰到静态物体或平台
//   - 只有Dynamic和Kinematic模式返回有意义的值
//
// 使用场景:
//   - 判断是否可以跳跃：if IsGrounded() { AddImpulse(...) }
//   - 播放落地音效：if IsGrounded() && wasInAir { playLandSound() }
//   - 切换动画状态：if IsGrounded() { playIdleAnim() } else { playFallAnim() }
//   - 限制二段跳：if IsGrounded() { jumpCount = 0 }
//
// 注意事项:
//   - NoPhysics模式始终返回false
//   - 在斜坡上也会返回true
//   - 检测距离很短，轻微离地就会返回false
func (p *SpriteImpl) IsGrounded() bool
```

```go
// Raycast 发射射线检测
//
// 参数:
//   from: 射线起点，单位：像素
//   to: 射线终点，单位：像素
//
// 返回:
//   hit: 是否击中对象
//   point: 击中点的位置，如果没击中则为零值
//   target: 被击中的对象，如果没击中则为nil
//
// 检测特性:
//   - 检测第一个碰到的对象（按距离排序）
//   - 忽略NoPhysics模式的对象
//   - 忽略射线起点所在的对象（避免自己检测自己）
//
// 使用场景:
//   - 武器射击检测：hit, point, enemy := player.Raycast(gunPos, targetPos)
//   - 视线检测：canSeePlayer := enemy.Raycast(enemyPos, playerPos).hit
//   - 地面检测：isGround := player.Raycast(playerPos, playerPos + Vector2{0, 50}).hit
//   - 寻路避障：isBlocked := unit.Raycast(currentPos, nextPos).hit
//   - 激光特效：从起点到击中点画线
//
// 技巧:
//   - 可以用很短的射线实现精确的接触检测
//   - 可以用很长的射线实现远距离视线检测
//   - 结合循环可以实现扇形范围检测
func (p *SpriteImpl) Raycast(from, to Vector2) (hit bool, point Vector2, target *SpriteImpl)
```

## 完整API概览

```go
// 物理控制（4个核心）
func (p *SpriteImpl) SetPhysicsMode(mode PhysicsMode) error
func (p *SpriteImpl) SetVelocity(velocity Vector2) error  
func (p *SpriteImpl) SetGravity(gravity float32) error
func (p *SpriteImpl) AddImpulse(impulse Vector2) error

// 基础属性（2个必需）
func (p *SpriteImpl) SetPosition(pos Vector2) error
func (p *SpriteImpl) SetColliderSize(width, height float32) error

// 查询方法（2个实用）
func (p *SpriteImpl) IsGrounded() bool
func (p *SpriteImpl) Raycast(from, to Vector2) (bool, Vector2, *SpriteImpl)
```

## 用户故事与使用示例

### 用户故事1: 制作平台跳跃游戏

**需求**: *"我要做一个马里奥风格的平台游戏，玩家可以左右移动和跳跃"*

```go
func setupPlayer() *SpriteImpl {
    player := NewSpriteImpl()
    
    // 设置为动态物理模式，受重力影响
    player.SetPhysicsMode(Dynamic)
    player.SetGravity(1.0)  // 正常重力
    player.SetColliderSize(24, 32)  // 玩家碰撞器
    player.SetPosition(Vector2{X: 100, Y: 100})  // 初始位置
    
    return player
}

func updatePlayer(player *SpriteImpl, input GameInput) {
    currentVel := player.GetVelocity()
    
    // 左右移动
    var moveSpeed float32 = 200
    if input.Left {
        player.SetVelocity(Vector2{X: -moveSpeed, Y: currentVel.Y})
    } else if input.Right {
        player.SetVelocity(Vector2{X: moveSpeed, Y: currentVel.Y})
    } else {
        // 停止水平移动
        player.SetVelocity(Vector2{X: 0, Y: currentVel.Y})
    }
    
    // 跳跃（只有在地面才能跳）
    if input.Jump && player.IsGrounded() {
        player.AddImpulse(Vector2{X: 0, Y: -300})  // 向上跳跃
    }
}
```

### 用户故事2: 制作巡逻敌人

**需求**: *"我需要一个在平台间来回巡逻的敌人，不受重力影响"*

```go
type PatrolEnemy struct {
    sprite *SpriteImpl
    patrolPoints []Vector2
    currentTarget int
    speed float32
}

func createPatrolEnemy(points []Vector2) *PatrolEnemy {
    enemy := &PatrolEnemy{
        sprite: NewSpriteImpl(),
        patrolPoints: points,
        currentTarget: 0,
        speed: 50,
    }
    
    // 设置为运动学模式，不受重力影响
    enemy.sprite.SetPhysicsMode(Kinematic)
    enemy.sprite.SetColliderSize(20, 20)
    enemy.sprite.SetPosition(points[0])  // 从第一个点开始
    
    return enemy
}

func (e *PatrolEnemy) Update() {
    currentPos := e.sprite.GetPosition()
    targetPos := e.patrolPoints[e.currentTarget]
    
    // 计算移动方向
    direction := Vector2{
        X: targetPos.X - currentPos.X,
        Y: targetPos.Y - currentPos.Y,
    }
    
    // 检查是否到达目标点
    distance := math.Sqrt(float64(direction.X*direction.X + direction.Y*direction.Y))
    if distance < 10 {
        // 切换到下一个巡逻点
        e.currentTarget = (e.currentTarget + 1) % len(e.patrolPoints)
        return
    }
    
    // 标准化方向向量并设置速度
    length := float32(distance)
    if length > 0 {
        direction.X /= length
        direction.Y /= length
        e.sprite.SetVelocity(Vector2{
            X: direction.X * e.speed,
            Y: direction.Y * e.speed,
        })
    }
}
```

### 用户故事3: 制作射击游戏

**需求**: *"玩家可以射击，子弹要检测碰撞并消除敌人"*

```go
type Bullet struct {
    sprite *SpriteImpl
    damage int
    maxDistance float32
    startPos Vector2
}

func fireBullet(from Vector2, direction Vector2, speed float32) *Bullet {
    bullet := &Bullet{
        sprite: NewSpriteImpl(),
        damage: 10,
        maxDistance: 500,
        startPos: from,
    }
    
    // 子弹使用运动学模式，不受重力影响
    bullet.sprite.SetPhysicsMode(Kinematic)
    bullet.sprite.SetColliderSize(4, 4)  // 小碰撞器
    bullet.sprite.SetPosition(from)
    
    // 设置子弹速度
    bullet.sprite.SetVelocity(Vector2{
        X: direction.X * speed,
        Y: direction.Y * speed,
    })
    
    return bullet
}

func (b *Bullet) Update() {
    currentPos := b.sprite.GetPosition()
    
    // 检查飞行距离
    distance := math.Sqrt(float64(
        (currentPos.X-b.startPos.X)*(currentPos.X-b.startPos.X) +
        (currentPos.Y-b.startPos.Y)*(currentPos.Y-b.startPos.Y),
    ))
    
    if distance > float64(b.maxDistance) {
        b.Destroy()
        return
    }
    
    // 使用射线检测前方是否有敌人
    velocity := b.sprite.GetVelocity()
    nextPos := Vector2{
        X: currentPos.X + velocity.X * 0.016, // 假设16ms一帧
        Y: currentPos.Y + velocity.Y * 0.016,
    }
    
    hit, point, target := b.sprite.Raycast(currentPos, nextPos)
    if hit && target != nil {
        // 击中目标
        if target.HasTag("enemy") {
            target.TakeDamage(b.damage)
            b.Destroy()
        }
    }
}
```

### 用户故事4: 制作可推动箱子

**需求**: *"场景中有箱子，玩家可以推动它们，箱子会受重力影响"*

```go
func createPushableBox(pos Vector2) *SpriteImpl {
    box := NewSpriteImpl()
    
    // 箱子使用动态物理，可以被推动
    box.SetPhysicsMode(Dynamic)
    box.SetGravity(1.0)  // 正常重力
    box.SetColliderSize(32, 32)
    box.SetPosition(pos)
    
    return box
}

func handlePlayerPushBox(player, box *SpriteImpl) {
    playerPos := player.GetPosition()
    boxPos := box.GetPosition()
    
    // 检查玩家是否贴着箱子
    distance := math.Abs(float64(playerPos.X - boxPos.X))
    if distance < 30 { // 在推动范围内
        // 计算推动方向
        pushDirection := float32(1)
        if playerPos.X > boxPos.X {
            pushDirection = -1  // 向左推
        }
        
        // 给箱子施加推力
        box.AddImpulse(Vector2{X: pushDirection * 100, Y: 0})
    }
}
```

### 用户故事5: 制作飞行道具效果

**需求**: *"玩家吃到飞行道具后可以飞行一段时间，不受重力影响"*

```go
type Player struct {
    sprite *SpriteImpl
    isFlying bool
    flyTimeLeft float32
}

func (p *Player) CollectFlyPowerup() {
    p.isFlying = true
    p.flyTimeLeft = 5.0  // 5秒飞行时间
    
    // 关闭重力
    p.sprite.SetGravity(0.0)
    
    // 给一个向上的推力
    p.sprite.AddImpulse(Vector2{X: 0, Y: -200})
}

func (p *Player) UpdateFlying(deltaTime float32, input GameInput) {
    if !p.isFlying {
        return
    }
    
    // 飞行时间倒计时
    p.flyTimeLeft -= deltaTime
    if p.flyTimeLeft <= 0 {
        p.StopFlying()
        return
    }
    
    // 飞行控制
    var flySpeed float32 = 150
    velocity := Vector2{X: 0, Y: 0}
    
    if input.Up {
        velocity.Y = -flySpeed
    } else if input.Down {
        velocity.Y = flySpeed
    }
    
    if input.Left {
        velocity.X = -flySpeed
    } else if input.Right {
        velocity.X = flySpeed
    }
    
    p.sprite.SetVelocity(velocity)
}

func (p *Player) StopFlying() {
    p.isFlying = false
    p.flyTimeLeft = 0
    
    // 恢复重力
    p.sprite.SetGravity(1.0)
}
```

### 用户故事6: 制作背景装饰元素

**需求**: *"背景中有飘动的云朵和飞鸟，它们不参与游戏碰撞"*

```go
type BackgroundCloud struct {
    sprite *SpriteImpl
    floatAmplitude float32
    floatSpeed float32
    startY float32
}

func createCloud(pos Vector2) *BackgroundCloud {
    cloud := &BackgroundCloud{
        sprite: NewSpriteImpl(),
        floatAmplitude: 20,
        floatSpeed: 0.5,
        startY: pos.Y,
    }
    
    // 云朵使用无物理模式，纯装饰
    cloud.sprite.SetPhysicsMode(NoPhysics)
    cloud.sprite.SetPosition(pos)
    // 不需要设置碰撞器，因为不参与碰撞
    
    return cloud
}

func (c *BackgroundCloud) Update(gameTime float32) {
    // 左右飘动
    currentPos := c.sprite.GetPosition()
    
    // 计算上下浮动
    floatY := c.startY + float32(math.Sin(float64(gameTime * c.floatSpeed))) * c.floatAmplitude
    
    // 水平移动
    newPos := Vector2{
        X: currentPos.X - 10, // 向左慢慢飘动
        Y: floatY,
    }
    
    // 直接设置位置（NoPhysics模式下推荐用SetPosition而不是SetVelocity）
    c.sprite.SetPosition(newPos)
    
    // 屏幕外循环
    if newPos.X < -100 {
        c.sprite.SetPosition(Vector2{X: 900, Y: floatY}) // 从右边重新出现
    }
}
```

### 用户故事7: 制作水下关卡

**需求**: *"某些关卡在水下，角色移动变慢，重力减小"*

```go
type WaterZone struct {
    bounds Rectangle  // 水域范围
    player *SpriteImpl
    inWater bool
    normalGravity float32
    normalSpeed float32
}

func (w *WaterZone) CheckPlayerInWater() {
    playerPos := w.player.GetPosition()
    
    wasInWater := w.inWater
    w.inWater = w.bounds.Contains(playerPos)
    
    // 状态改变时调整物理参数
    if w.inWater && !wasInWater {
        // 进入水中
        w.normalGravity = 1.0  // 记录正常重力
        w.normalSpeed = 200    // 记录正常速度
        
        w.player.SetGravity(0.3)  // 水中重力减小
    } else if !w.inWater && wasInWater {
        // 离开水中
        w.player.SetGravity(w.normalGravity)  // 恢复正常重力
    }
}

func (w *WaterZone) UpdatePlayerMovement(input GameInput) {
    speed := w.normalSpeed
    if w.inWater {
        speed *= 0.6  // 水中速度减慢
    }
    
    currentVel := w.player.GetVelocity()
    var newVelX float32 = 0
    
    if input.Left {
        newVelX = -speed
    } else if input.Right {
        newVelX = speed
    }
    
    // 水中可以向上游
    var newVelY float32 = currentVel.Y
    if w.inWater && input.Up {
        newVelY = -speed * 0.8  // 向上游的速度
    }
    
    w.player.SetVelocity(Vector2{X: newVelX, Y: newVelY})
}
```

### 用户故事8: 制作简单AI敌人

**需求**: *"敌人会追踪玩家，但被墙壁阻挡时会寻找绕路"*

```go
type ChaseEnemy struct {
    sprite *SpriteImpl
    target *SpriteImpl
    speed float32
    detectionRange float32
}

func createChaseEnemy(pos Vector2, target *SpriteImpl) *ChaseEnemy {
    enemy := &ChaseEnemy{
        sprite: NewSpriteImpl(),
        target: target,
        speed: 80,
        detectionRange: 200,
    }
    
    // 敌人使用运动学模式，可以精确控制移动
    enemy.sprite.SetPhysicsMode(Kinematic)
    enemy.sprite.SetGravity(0.0)  // 不受重力影响，可以飞行追击
    enemy.sprite.SetColliderSize(24, 24)
    enemy.sprite.SetPosition(pos)
    
    return enemy
}

func (e *ChaseEnemy) Update() {
    enemyPos := e.sprite.GetPosition()
    targetPos := e.target.GetPosition()
    
    // 检查距离
    dx := targetPos.X - enemyPos.X
    dy := targetPos.Y - enemyPos.Y
    distance := math.Sqrt(float64(dx*dx + dy*dy))
    
    // 超出检测范围则停止追击
    if distance > float64(e.detectionRange) {
        e.sprite.SetVelocity(Vector2{X: 0, Y: 0})
        return
    }
    
    // 使用射线检测是否有障碍物
    hit, _, _ := e.sprite.Raycast(enemyPos, targetPos)
    
    if !hit {
        // 没有障碍物，直接追击
        direction := Vector2{
            X: float32(dx) / float32(distance),
            Y: float32(dy) / float32(distance),
        }
        e.sprite.SetVelocity(Vector2{
            X: direction.X * e.speed,
            Y: direction.Y * e.speed,
        })
    } else {
        // 有障碍物，尝试绕路（简单的左右尝试）
        // 尝试向左绕行
        leftPos := Vector2{X: enemyPos.X - 50, Y: targetPos.Y}
        leftHit, _, _ := e.sprite.Raycast(enemyPos, leftPos)
        
        if !leftHit {
            e.sprite.SetVelocity(Vector2{X: -e.speed, Y: 0})
        } else {
            // 尝试向右绕行
            rightPos := Vector2{X: enemyPos.X + 50, Y: targetPos.Y}
            rightHit, _, _ := e.sprite.Raycast(enemyPos, rightPos)
            
            if !rightHit {
                e.sprite.SetVelocity(Vector2{X: e.speed, Y: 0})
            } else {
                // 都被挡住了，停止移动
                e.sprite.SetVelocity(Vector2{X: 0, Y: 0})
            }
        }
    }
}
```

## API使用建议

### 1. 物理模式选择指南

```go
// Dynamic - 适用于:
// - 玩家角色（需要重力和真实物理）
// - 可推动的物品（箱子、球等）
// - 受重力影响的敌人
player.SetPhysicsMode(Dynamic)

// Kinematic - 适用于:
// - 巡逻敌人（精确控制路径）
// - 移动平台（按预定轨迹移动）
// - 子弹和抛射物（不受重力影响）
// - 飞行敌人（不需要重力）
npc.SetPhysicsMode(Kinematic)

// NoPhysics - 适用于:
// - 背景装饰（云朵、远山等）
// - 粒子效果
// - UI跟随元素
// - 不需要碰撞的视觉效果
decoration.SetPhysicsMode(NoPhysics)
```

### 2. 性能优化建议

```go
// 1. 合理使用NoPhysics模式减少计算开销
backgroundElements.SetPhysicsMode(NoPhysics)

// 2. 子弹等短生命周期对象使用Kinematic
bullet.SetPhysicsMode(Kinematic)  // 而不是Dynamic

// 3. 远离玩家的对象可以降低更新频率或暂时禁用物理
if distanceToPlayer > 500 {
    distantEnemy.SetPhysicsMode(NoPhysics)  // 临时关闭物理
}
```

### 3. 常见问题解决方案

```go
// 问题1: 角色卡在墙里
// 解决: 调整碰撞器大小，确保比显示图像略小
player.SetColliderSize(playerWidth * 0.8, playerHeight * 0.9)

// 问题2: 跳跃感觉不够responsive
// 解决: 使用AddImpulse而不是SetVelocity，并调整重力
if input.Jump && player.IsGrounded() {
    player.AddImpulse(Vector2{X: 0, Y: -350})  // 增大跳跃力
    player.SetGravity(1.2)  // 稍微增大重力，让跳跃更紧凑
}

// 问题3: 移动太滑
// 解决: 及时停止水平移动
if !input.Left && !input.Right {
    currentVel := player.GetVelocity()
    player.SetVelocity(Vector2{X: 0, Y: currentVel.Y})  // 立即停止水平移动
}
```

## 总结

这套8个API的极简设计具有以下优势：

1. **学习成本低** - 只需掌握8个方法
2. **覆盖面广** - 能实现90%的游戏物理需求
3. **性能优秀** - 三种物理模式针对不同需求优化
4. **易于调试** - API简单，问题容易定位
5. **扩展性好** - 基础API稳定，后续可以在上层封装更多功能

通过这些用户故事和示例，你可以看到这套简洁的API能够支撑各种复杂的游戏场景，同时保持代码的清晰和可维护性。