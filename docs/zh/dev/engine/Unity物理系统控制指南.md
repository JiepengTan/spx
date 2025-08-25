# Unity 物理系统控制指南

## 基础概念

### 运行时控制角色是否受物理控制的方法

在Unity中，可以通过以下几种方式在运行时控制角色是否受到物理控制：

#### 1. 使用 Rigidbody.isKinematic

```csharp
// 获取角色的Rigidbody组件
Rigidbody rb = GetComponent<Rigidbody>();

// 禁用物理控制（角色不受重力和碰撞影响）
rb.isKinematic = true;

// 启用物理控制（角色受重力和碰撞影响）
rb.isKinematic = false;
```

#### 2. 启用/禁用 Rigidbody 组件

```csharp
Rigidbody rb = GetComponent<Rigidbody>();

// 禁用物理
rb.enabled = false;

// 启用物理
rb.enabled = true;
```

#### 3. 控制重力影响

```csharp
Rigidbody rb = GetComponent<Rigidbody>();

// 禁用重力
rb.useGravity = false;

// 启用重力
rb.useGravity = true;
```

#### 4. 冻结特定轴向的物理

```csharp
Rigidbody rb = GetComponent<Rigidbody>();

// 冻结所有旋转和位置
rb.constraints = RigidbodyConstraints.FreezeAll;

// 只冻结Y轴位置（防止角色上下移动）
rb.constraints = RigidbodyConstraints.FreezePositionY;

// 解除所有约束
rb.constraints = RigidbodyConstraints.None;
```

## isKinematic 对其他物体的影响

### Kinematic 物体的特性

**Kinematic 物体 (`isKinematic = true`)：**
- 不受重力和力的影响
- 可以影响其他非Kinematic物体
- 自身不会被其他物体推动

### 对其他物体的影响

#### 1. 碰撞检测仍然有效
```csharp
// Kinematic物体仍然可以触发碰撞事件
void OnCollisionEnter(Collision collision)
{
    if (collision.gameObject.CompareTag("Player"))
    {
        // 这个事件仍然会被触发
    }
}
```

#### 2. 可以推动非Kinematic物体
```csharp
// Kinematic物体移动时会推动其他物体
transform.Translate(Vector3.forward * speed * Time.deltaTime);
// 如果路径上有非Kinematic物体，会被推开
```

#### 3. 触发器功能正常
```csharp
// 如果Collider设置为isTrigger = true
void OnTriggerEnter(Collider other)
{
    // 仍然可以检测到进入触发器的物体
}
```

## 超级玛丽式角色控制

### Rigidbody 设置

```csharp
public class MarioController : MonoBehaviour
{
    [Header("Movement")]
    public float moveSpeed = 8f;
    public float jumpForce = 12f;
    
    [Header("Ground Check")]
    public Transform groundCheck;
    public LayerMask groundLayer;
    public float groundCheckRadius = 0.2f;
    
    private Rigidbody2D rb;
    private bool isGrounded;
    private float horizontalInput;
    
    void Start()
    {
        rb = GetComponent<Rigidbody2D>();
        
        // 关键设置：
        rb.isKinematic = false;          // 受物理影响
        rb.freezeRotation = true;        // 防止角色翻转
        // 或者用约束：
        // rb.constraints = RigidbodyConstraints2D.FreezeRotation;
    }
}
```

### 控制方式

#### 方案1：修改速度（推荐）
```csharp
void Update()
{
    horizontalInput = Input.GetAxisRaw("Horizontal");
    
    // 检查是否在地面
    isGrounded = Physics2D.OverlapCircle(groundCheck.position, groundCheckRadius, groundLayer);
    
    // 跳跃
    if (Input.GetKeyDown(KeyCode.Space) && isGrounded)
    {
        rb.velocity = new Vector2(rb.velocity.x, jumpForce);
    }
}

void FixedUpdate()
{
    // 水平移动：直接设置x轴速度，保持y轴速度
    rb.velocity = new Vector2(horizontalInput * moveSpeed, rb.velocity.y);
}
```

#### 方案2：施加力
```csharp
void FixedUpdate()
{
    // 施加水平力
    rb.AddForce(new Vector2(horizontalInput * moveSpeed, 0));
    
    // 限制最大速度
    if (Mathf.Abs(rb.velocity.x) > maxSpeed)
    {
        rb.velocity = new Vector2(Mathf.Sign(rb.velocity.x) * maxSpeed, rb.velocity.y);
    }
}
```

### 物理材质设置

```csharp
void Start()
{
    // 创建物理材质，避免角色"粘墙"
    PhysicsMaterial2D playerMaterial = new PhysicsMaterial2D("PlayerMaterial");
    playerMaterial.friction = 0f;        // 无摩擦
    playerMaterial.bounciness = 0f;      // 无弹性
    
    GetComponent<Collider2D>().material = playerMaterial;
}
```

## 跳跃机制详解

### velocity.y 的自动衰减

当你设置 `velocity.y = 20` 后：

1. **初始瞬间**：角色以 20 单位/秒的速度向上运动
2. **重力作用**：每帧重力都会减少 `velocity.y` 的值
3. **自动衰减**：速度逐渐从 20 → 15 → 10 → 5 → 0 → -5 → -10...
4. **下落**：当 `velocity.y` 变为负值时，角色开始下落

```csharp
void Update()
{
    if (Input.GetKeyDown(KeyCode.Space) && isGrounded)
    {
        rb.velocity = new Vector2(rb.velocity.x, 20f); // 设置跳跃
    }
    
    // 可以实时查看velocity变化
    Debug.Log($"Current velocity.y: {rb.velocity.y}");
}
```

### velocity.x 的持续性

**不一定会一直保持！** `velocity.x` 是否保持 100 取决于物理环境：

- **摩擦力**：有摩擦力时会逐渐减速
- **空气阻力**：`rb.drag` 值会影响速度衰减
- **碰撞**：碰撞会改变velocity值

```csharp
// 在平台游戏中，通常每帧更新 velocity.x 以确保精确控制
void FixedUpdate()
{
    float horizontalInput = Input.GetAxisRaw("Horizontal");
    rb.velocity = new Vector2(horizontalInput * moveSpeed, rb.velocity.y);
}
```

## 击飞效果实现

### 对于 isKinematic = false 的物体

#### 方案1：使用 AddForce（推荐）

```csharp
public void KnockbackBox(Rigidbody2D boxRb, Vector2 direction, float force)
{
    boxRb.AddForce(direction.normalized * force, ForceMode2D.Impulse);
}

// 使用示例
void OnCollisionEnter2D(Collision2D collision)
{
    if (collision.gameObject.CompareTag("Box"))
    {
        Rigidbody2D boxRb = collision.gameObject.GetComponent<Rigidbody2D>();
        
        // 向右上方击飞
        Vector2 knockbackDirection = new Vector2(1f, 1f);
        KnockbackBox(boxRb, knockbackDirection, 500f);
    }
}
```

#### 方案2：直接设置 velocity

```csharp
public void KnockbackBox(Rigidbody2D boxRb, Vector2 knockbackVelocity)
{
    boxRb.velocity = knockbackVelocity;
}

// 使用示例
void HitBox()
{
    Vector2 knockbackVelocity = new Vector2(15f, 20f); // x=15, y=20
    KnockbackBox(boxRb, knockbackVelocity);
}
```

### 对于 isKinematic = true 的物体

#### 方案1：临时切换为非Kinematic（推荐）

```csharp
public class Monster : MonoBehaviour
{
    private Rigidbody2D rb;
    private bool isKnockedBack = false;
    private bool wasKinematic;
    
    void Start()
    {
        rb = GetComponent<Rigidbody2D>();
        rb.isKinematic = true; // 正常状态下为Kinematic
    }
    
    public void GetKnockedBack(Vector2 force)
    {
        StartCoroutine(KnockbackCoroutine(force));
    }
    
    IEnumerator KnockbackCoroutine(Vector2 force)
    {
        // 暂时切换为非Kinematic
        wasKinematic = rb.isKinematic;
        rb.isKinematic = false;
        isKnockedBack = true;
        
        // 施加击飞力
        rb.AddForce(force, ForceMode2D.Impulse);
        
        // 等待击飞结束
        yield return new WaitForSeconds(1f);
        
        // 恢复Kinematic状态
        rb.isKinematic = wasKinematic;
        rb.velocity = Vector2.zero; // 清除残余速度
        isKnockedBack = false;
    }
    
    void Update()
    {
        // 击飞期间暂停AI逻辑
        if (!isKnockedBack)
        {
            // 怪物的正常AI逻辑
            NormalAIBehavior();
        }
    }
}
```

#### 方案2：手动模拟抛物线运动

```csharp
public class Monster : MonoBehaviour
{
    private bool isKnockedBack = false;
    private Vector2 knockbackVelocity;
    private float gravity = -20f;
    
    public void GetKnockedBack(Vector2 initialVelocity)
    {
        StartCoroutine(SimulateKnockback(initialVelocity));
    }
    
    IEnumerator SimulateKnockback(Vector2 initialVelocity)
    {
        isKnockedBack = true;
        knockbackVelocity = initialVelocity;
        
        while (knockbackVelocity.y > 0 || !IsGrounded())
        {
            // 模拟重力
            knockbackVelocity.y += gravity * Time.deltaTime;
            
            // 模拟空气阻力
            knockbackVelocity.x *= 0.98f;
            
            // 移动怪物
            transform.Translate(knockbackVelocity * Time.deltaTime);
            
            yield return null;
        }
        
        isKnockedBack = false;
    }
}
```

## isKinematic 的使用场景

### 何时使用 isKinematic = true

#### 1. 需要精确控制的移动物体
```csharp
// 移动平台 - 需要按固定轨迹移动，不受其他物理影响
public class MovingPlatform : MonoBehaviour
{
    void Start()
    {
        GetComponent<Rigidbody2D>().isKinematic = true;
        // 平台按预定路线移动，不会被玩家推动或重力影响
    }
}
```

#### 2. 角色在特殊状态下
```csharp
public class Player : MonoBehaviour
{
    void EnterCutscene()
    {
        rb.isKinematic = true; // 过场动画中，角色不受物理影响
    }
    
    void EnterVehicle()
    {
        rb.isKinematic = true; // 进入载具，角色移动由载具控制
    }
    
    void StartClimbing()
    {
        rb.isKinematic = true; // 爬墙时，不受重力影响
    }
}
```

#### 3. NPC和怪物的AI控制
```csharp
public class NPCPatrol : MonoBehaviour
{
    void Start()
    {
        rb.isKinematic = true; // NPC按AI逻辑移动，不被物理推动
    }
    
    void Update()
    {
        // 精确控制NPC移动，不会因为碰撞而偏离路径
        transform.Translate(patrolDirection * speed * Time.deltaTime);
    }
}
```

#### 4. 门、开关等机关
```csharp
public class Door : MonoBehaviour
{
    void Start()
    {
        rb.isKinematic = true; // 门只能按脚本控制开关，不受力影响
    }
    
    public void OpenDoor()
    {
        // 门按动画或脚本精确移动，不会被其他物体干扰
        transform.DOMove(openPosition, 1f);
    }
}
```

## 用户故事示例

### 1. 制作横版闯关游戏 - 玩家角色

**用户需求**：*"我要制作类似《超级玛里奥》的游戏，玩家可以跑跳，会受重力影响，但移动要很精确"*

```csharp
public class PlayerController : MonoBehaviour
{
    void Start()
    {
        Rigidbody2D rb = GetComponent<Rigidbody2D>();
        
        // 设置参数
        rb.isKinematic = false;     // 受物理影响（重力、碰撞）
        rb.freezeRotation = true;   // 不会翻转倒下
        rb.gravityScale = 3f;       // 增强重力感，快速下落
        rb.drag = 0f;              // 无阻力，移动更灵敏
    }
    
    void FixedUpdate()
    {
        // 精确控制水平移动
        rb.velocity = new Vector2(input * moveSpeed, rb.velocity.y);
        
        // 跳跃
        if (jumpPressed && isGrounded)
            rb.velocity = new Vector2(rb.velocity.x, jumpForce);
    }
}
```

### 2. 制作RPG游戏 - NPC商人

**用户需求**：*"我的游戏中有个商人NPC，他要按固定路线巡逻，玩家撞到他时要能对话，但不能把他推走"*

```csharp
public class MerchantNPC : MonoBehaviour
{
    void Start()
    {
        Rigidbody2D rb = GetComponent<Rigidbody2D>();
        
        // 设置参数
        rb.isKinematic = true;      // 不受物理推动
        // 保留Collider用于碰撞检测
    }
    
    void Update()
    {
        // AI控制巡逻，不会被玩家推动
        PatrolMovement();
    }
    
    void OnTriggerEnter2D(Collider2D other)
    {
        if (other.CompareTag("Player"))
        {
            StartDialogue(); // 仍能检测玩家碰撞
        }
    }
}
```

### 3. 制作平台游戏 - 移动平台

**用户需求**：*"我需要一个上下移动的平台，玩家站上去会被带动，但平台不能被任何东西影响"*

```csharp
public class MovingPlatform : MonoBehaviour
{
    void Start()
    {
        Rigidbody2D rb = GetComponent<Rigidbody2D>();
        
        // 设置参数
        rb.isKinematic = true;      // 不受外力影响
        rb.useGravity = false;      // 不受重力
    }
    
    void Update()
    {
        // 按固定路线移动，会推动站在上面的玩家
        transform.Translate(Vector2.up * Mathf.Sin(Time.time) * speed * Time.deltaTime);
    }
    
    void OnCollisionStay2D(Collision2D collision)
    {
        if (collision.gameObject.CompareTag("Player"))
        {
            // 玩家会被平台带动移动
            collision.transform.parent = transform;
        }
    }
}
```

### 4. 制作动作游戏 - 受击系统

**用户需求**：*"玩家被攻击时要被击飞，但击飞过程中不能控制角色，击飞结束后恢复正常控制"*

```csharp
public class PlayerCombat : MonoBehaviour
{
    private bool isKnockedBack = false;
    
    public void TakeHit(Vector2 knockbackForce)
    {
        StartCoroutine(KnockbackSequence(knockbackForce));
    }
    
    IEnumerator KnockbackSequence(Vector2 force)
    {
        Rigidbody2D rb = GetComponent<Rigidbody2D>();
        
        // 击飞阶段：使用物理
        rb.isKinematic = false;
        rb.AddForce(force, ForceMode2D.Impulse);
        isKnockedBack = true;
        
        // 等待击飞结束
        yield return new WaitForSeconds(0.8f);
        
        // 恢复控制
        isKnockedBack = false;
        // rb.isKinematic 保持false，恢复正常物理控制
    }
    
    void Update()
    {
        // 击飞期间禁用输入
        if (!isKnockedBack)
        {
            HandlePlayerInput();
        }
    }
}
```

### 5. 制作解谜游戏 - 推箱子机关

**用户需求**：*"有些箱子可以被推动来解谜，有些箱子是装饰不能动，推动的箱子要有真实的物理感觉"*

```csharp
public class PuzzleBox : MonoBehaviour
{
    [SerializeField] private bool canBePushed = true;
    
    void Start()
    {
        Rigidbody2D rb = GetComponent<Rigidbody2D>();
        
        if (canBePushed)
        {
            // 可推动的箱子
            rb.isKinematic = false;
            rb.drag = 2f;           // 有阻力，不会滑太远
            rb.angularDrag = 1f;    // 限制旋转
            rb.mass = 2f;           // 有重量感
        }
        else
        {
            // 装饰箱子
            rb.isKinematic = true;  // 不能被推动
        }
    }
}
```

### 6. 制作竞速游戏 - 载具控制

**用户需求**：*"玩家可以开车，上车时角色要跟着车动，下车时恢复步行物理"*

```csharp
public class VehicleSystem : MonoBehaviour
{
    public void EnterVehicle(GameObject player)
    {
        Rigidbody2D playerRb = player.GetComponent<Rigidbody2D>();
        
        // 进入载具
        playerRb.isKinematic = true;        // 角色不受物理影响
        player.transform.parent = transform; // 跟随载具移动
        
        // 载具物理
        GetComponent<Rigidbody2D>().isKinematic = false;
    }
    
    public void ExitVehicle(GameObject player)
    {
        Rigidbody2D playerRb = player.GetComponent<Rigidbody2D>();
        
        // 离开载具
        playerRb.isKinematic = false;       // 恢复角色物理
        player.transform.parent = null;     // 取消跟随
    }
}
```

## 总结

这些用户故事展示了在不同游戏场景下，如何根据具体需求灵活使用 `isKinematic`、`velocity`、`AddForce` 等参数来实现理想的游戏体验。关键是要理解：

- **isKinematic = false**：适用于需要真实物理交互的对象
- **isKinematic = true**：适用于需要精确控制且不被干扰的对象
- **velocity**：直接控制物体速度，适合精确移动
- **AddForce**：施加力来产生自然的物理效果，适合击飞等效果

选择合适的方法取决于你想要的游戏体验和控制精度。