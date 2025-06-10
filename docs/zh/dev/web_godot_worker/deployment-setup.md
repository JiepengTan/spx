# Godot + Go WASM 集成部署指南

## 项目结构

```
project/
├── assets/
│   ├── godot.editor.html              # 主HTML文件
│   ├── godot.editor.js               # Godot主JS文件
│   ├── godot.editor.worker.js        # 原始Worker文件
│   ├── godot.editor.worker.enhanced.js # 增强Worker文件
│   ├── godot.editor.wasm             # Godot WASM模块
│   ├── godot.editor.pck              # Godot资源包
│   └── godot.editor.audio.worklet.js # 音频处理
├── go-wasm/
│   ├── main.go                       # Go源代码
│   ├── main.wasm                     # 编译后的Go WASM
│   └── wasm_exec.js                  # Go运行时支持
├── js/
│   ├── go-wasm-bridge.js            # Go WASM桥接器
│   └── integration-example.js       # 集成示例代码
└── server/
    ├── .htaccess                     # Apache配置
    ├── nginx.conf                    # Nginx配置
    └── http-headers.txt              # HTTP头配置
```

## 1. 准备Go WASM模块

### 1.1 编写Go代码

```go
// main.go
package main

import (
    "encoding/json"
    "fmt"
    "syscall/js"
    "time"
)

// 游戏数据结构
type GameData struct {
    PlayerName string                 `json:"playerName"`
    Level      int                    `json:"level"`
    Score      int                    `json:"score"`
    Inventory  map[string]interface{} `json:"inventory"`
}

// 玩家行动结构
type PlayerAction struct {
    Type   string                 `json:"type"`
    Target string                 `json:"target"`
    Data   map[string]interface{} `json:"data"`
}

// 基础计算函数
func calculateSum(this js.Value, args []js.Value) interface{} {
    if len(args) < 2 {
        return map[string]interface{}{"error": "需要两个参数"}
    }
    
    a := args[0].Float()
    b := args[1].Float()
    
    return map[string]interface{}{
        "result": a + b,
        "status": "success",
        "timestamp": time.Now().Unix(),
    }
}

// 游戏初始化
func initializeGame(this js.Value, args []js.Value) interface{} {
    if len(args) < 1 {
        return map[string]interface{}{"error": "需要初始化参数"}
    }
    
    configStr := args[0].String()
    var config map[string]interface{}
    
    if err := json.Unmarshal([]byte(configStr), &config); err != nil {
        return map[string]interface{}{"error": "配置解析失败: " + err.Error()}
    }
    
    // 创建初始游戏状态
    gameState := GameData{
        PlayerName: config["playerName"].(string),
        Level:      1,
        Score:      0,
        Inventory:  make(map[string]interface{}),
    }
    
    gameState.Inventory["health_potions"] = 3
    gameState.Inventory["coins"] = 100
    
    result, _ := json.Marshal(gameState)
    return string(result)
}

// 处理玩家行动
func processPlayerAction(this js.Value, args []js.Value) interface{} {
    if len(args) < 1 {
        return map[string]interface{}{"error": "需要行动数据"}
    }
    
    actionStr := args[0].String()
    var action PlayerAction
    
    if err := json.Unmarshal([]byte(actionStr), &action); err != nil {
        return map[string]interface{}{"error": "行动数据解析失败: " + err.Error()}
    }
    
    // 根据行动类型处理
    switch action.Type {
    case "move":
        return processMovement(action)
    case "attack":
        return processAttack(action)
    case "use_item":
        return processItemUsage(action)
    default:
        return map[string]interface{}{"error": "未知行动类型: " + action.Type}
    }
}

func processMovement(action PlayerAction) interface{} {
    return map[string]interface{}{
        "success": true,
        "message": fmt.Sprintf("移动到 %s", action.Target),
        "energyCost": 1,
        "newPosition": action.Target,
    }
}

func processAttack(action PlayerAction) interface{} {
    damage := 10 + (time.Now().Unix() % 5) // 随机伤害
    return map[string]interface{}{
        "success": true,
        "message": fmt.Sprintf("攻击 %s", action.Target),
        "damage": damage,
        "critical": damage > 12,
    }
}

func processItemUsage(action PlayerAction) interface{} {
    return map[string]interface{}{
        "success": true,
        "message": fmt.Sprintf("使用了 %s", action.Target),
        "effect": "restored_health",
        "value": 25,
    }
}

// 计算游戏统计
func calculateGameStats(this js.Value, args []js.Value) interface{} {
    stats := map[string]interface{}{
        "totalPlayers": 1,
        "uptime": time.Now().Unix(),
        "performance": map[string]interface{}{
            "fps": 60,
            "memory": "24MB",
            "cpu": "15%",
        },
    }
    
    result, _ := json.Marshal(stats)
    return string(result)
}

// 异步处理示例
func asyncProcess(this js.Value, args []js.Value) interface{} {
    promise := js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
        resolve := args[0]
        
        // 模拟异步处理
        js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
            result := map[string]interface{}{
                "message": "异步处理完成",
                "processedAt": time.Now().Unix(),
                "data": "complex_calculation_result",
            }
            resolve.Invoke(result)
            return nil
        }), 1000)
        
        return nil
    }))
    
    return promise
}

// 复杂数据处理
func processComplexData(this js.Value, args []js.Value) interface{} {
    if len(args) < 1 {
        return map[string]interface{}{"error": "需要数据参数"}
    }
    
    dataStr := args[0].String()
    var data map[string]interface{}
    
    if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
        return map[string]interface{}{"error": "数据解析失败: " + err.Error()}
    }
    
    // 处理复杂逻辑
    processedData := make(map[string]interface{})
    processedData["original"] = data
    processedData["processed"] = true
    processedData["timestamp"] = time.Now().Unix()
    processedData["hash"] = fmt.Sprintf("%x", time.Now().UnixNano())
    
    // 模拟复杂计算
    if items, ok := data["items"].([]interface{}); ok {
        processedData["itemCount"] = len(items)
        processedData["totalValue"] = calculateTotalValue(items)
    }
    
    result, _ := json.Marshal(processedData)
    return string(result)
}

func calculateTotalValue(items []interface{}) float64 {
    total := 0.0
    for _, item := range items {
        if itemMap, ok := item.(map[string]interface{}); ok {
            if value, ok := itemMap["value"].(float64); ok {
                total += value
            }
        }
    }
    return total
}

func main() {
    // 防止程序退出
    c := make(chan struct{}, 0)
    
    // 注册所有函数到JavaScript全局对象
    js.Global().Set("goCalculateSum", js.FuncOf(calculateSum))
    js.Global().Set("goInitializeGame", js.FuncOf(initializeGame))
    js.Global().Set("goProcessPlayerAction", js.FuncOf(processPlayerAction))
    js.Global().Set("goCalculateGameStats", js.FuncOf(calculateGameStats))
    js.Global().Set("goAsyncProcess", js.FuncOf(asyncProcess))
    js.Global().Set("goProcessComplexData", js.FuncOf(processComplexData))
    
    // 通知JavaScript Go模块已准备就绪
    js.Global().Call("postMessage", map[string]interface{}{
        "cmd": "goReady",
        "functions": []string{
            "goCalculateSum",
            "goInitializeGame",
            "goProcessPlayerAction",
            "goCalculateGameStats",
            "goAsyncProcess",
            "goProcessComplexData",
        },
        "version": "1.0.0",
        "timestamp": time.Now().Unix(),
    })
    
    fmt.Println("Go WASM模块已启动")
    <-c
}
```

### 1.2 编译Go代码

```bash
#!/bin/bash
# build.sh

echo "编译Go WASM模块..."

# 设置环境变量
export GOOS=js
export GOARCH=wasm

# 编译
go build -o main.wasm main.go

# 复制Go运行时
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .

echo "编译完成！"
echo "生成文件: main.wasm, wasm_exec.js"
```

## 2. 修改Godot Worker

### 2.1 替换原始Worker

将原始的 `godot.editor.worker.js` 替换为增强版本，或者创建新的Worker文件。

```javascript
// godot.editor.worker.enhanced.js
// 使用 integration-example.js 中的 EnhancedGodotWorker 类
```

### 2.2 更新HTML文件

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, user-scalable=no">
    <title>Godot + Go WASM Game</title>
    <!-- 原有样式保持不变 -->
    <style>
    /* ... 原有CSS ... */
    </style>
    <link id='-gd-engine-icon' rel='icon' type='image/png' href='godot.editor.icon.png' />
</head>
<body>
    <canvas id="canvas">
        HTML5 canvas appears to be unsupported in the current browser.
    </canvas>
    
    <!-- 状态显示 -->
    <div id="status">
        <!-- ... 原有状态显示 ... -->
    </div>
    
    <!-- 加载Go WASM桥接器 -->
    <script src="go-wasm-bridge.js"></script>
    <script src="integration-example.js"></script>
    
    <!-- 原有Godot脚本 -->
    <script src="godot.editor.js"></script>
    
    <script>
        // Godot配置（使用增强Worker）
        const GODOT_CONFIG = {
            "args": [],
            "canvasResizePolicy": 2,
            "executable": "godot.editor",
            "experimentalVK": false,
            "fileSizes": {
                "godot.editor.pck": 80384,
                "godot.editor.wasm": 39020050,
                "main.wasm": 1024000  // Go WASM文件大小
            },
            "focusCanvas": true,
            "gdextensionLibs": []
        };
        
        // 修改Worker路径为增强版本
        const originalAllocateUnusedWorker = PThread.allocateUnusedWorker;
        PThread.allocateUnusedWorker = function() {
            var worker = new Worker("godot.editor.worker.enhanced.js");
            PThread.unusedWorkers.push(worker);
        };
        
        const engine = new Engine(GODOT_CONFIG);
        
        // 原有引擎启动逻辑
        // ... 保持不变 ...
        
        // 添加Go WASM就绪监听
        window.addEventListener('goModuleReady', (event) => {
            console.log('Go模块就绪事件:', event.detail);
            
            // 可以在这里添加游戏特定的初始化逻辑
            initializeGameLogic();
        });
        
        async function initializeGameLogic() {
            try {
                // 等待Go模块完全就绪
                await GoWasm.waitForReady();
                
                // 初始化游戏
                const gameState = await GoWasm.call('goInitializeGame', JSON.stringify({
                    playerName: 'WebPlayer',
                    difficulty: 'normal'
                }));
                
                console.log('游戏初始化成功:', gameState);
                
                // 可以触发Godot中的相应事件
                // 例如通过GodotFFI调用Godot函数
                
            } catch (error) {
                console.error('游戏逻辑初始化失败:', error);
            }
        }
    </script>
</body>
</html>
```

## 3. 服务器配置

### 3.1 Apache配置 (.htaccess)

```apache
# .htaccess
# 启用CORS和必要的HTTP头

# 设置WASM文件的MIME类型
AddType application/wasm .wasm

# 启用CORS
Header always set Access-Control-Allow-Origin "*"
Header always set Access-Control-Allow-Methods "GET, POST, OPTIONS"
Header always set Access-Control-Allow-Headers "Content-Type"

# SharedArrayBuffer支持所需的安全头
Header always set Cross-Origin-Embedder-Policy "require-corp"
Header always set Cross-Origin-Opener-Policy "same-origin"

# 缓存控制
<FilesMatch "\.(wasm|js)$">
    Header set Cache-Control "public, max-age=31536000"
</FilesMatch>

<FilesMatch "\.(html|pck)$">
    Header set Cache-Control "public, max-age=3600"
</FilesMatch>

# Gzip压缩
<IfModule mod_deflate.c>
    AddOutputFilterByType DEFLATE text/plain
    AddOutputFilterByType DEFLATE text/html
    AddOutputFilterByType DEFLATE text/xml
    AddOutputFilterByType DEFLATE text/css
    AddOutputFilterByType DEFLATE application/xml
    AddOutputFilterByType DEFLATE application/xhtml+xml
    AddOutputFilterByType DEFLATE application/rss+xml
    AddOutputFilterByType DEFLATE application/javascript
    AddOutputFilterByType DEFLATE application/x-javascript
    AddOutputFilterByType DEFLATE application/wasm
</IfModule>
```

### 3.2 Nginx配置

```nginx
# nginx.conf片段
server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    # SSL证书配置
    ssl_certificate /path/to/your/cert.pem;
    ssl_certificate_key /path/to/your/key.pem;
    
    # 文档根目录
    root /path/to/your/project;
    index index.html;
    
    # WASM MIME类型
    location ~* \.wasm$ {
        add_header Content-Type application/wasm;
        add_header Cross-Origin-Embedder-Policy require-corp;
        add_header Cross-Origin-Opener-Policy same-origin;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
    
    # JavaScript文件
    location ~* \.js$ {
        add_header Cross-Origin-Embedder-Policy require-corp;
        add_header Cross-Origin-Opener-Policy same-origin;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
    
    # HTML文件
    location ~* \.html$ {
        add_header Cross-Origin-Embedder-Policy require-corp;
        add_header Cross-Origin-Opener-Policy same-origin;
        expires 1h;
        add_header Cache-Control "public";
    }
    
    # 其他静态资源
    location ~* \.(pck|png|ico)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
    
    # Gzip压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/javascript
        application/xml+rss
        application/wasm;
}
```

## 4. 开发和测试

### 4.1 本地开发服务器

```python
# dev-server.py
#!/usr/bin/env python3
"""
简单的HTTPS开发服务器，支持SharedArrayBuffer所需的安全头
"""

import http.server
import ssl
import socketserver
from urllib.parse import urlparse
import mimetypes

# 添加WASM MIME类型
mimetypes.add_type('application/wasm', '.wasm')

class CORSHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    def end_headers(self):
        # 添加必要的安全头
        self.send_header('Cross-Origin-Embedder-Policy', 'require-corp')
        self.send_header('Cross-Origin-Opener-Policy', 'same-origin')
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type')
        super().end_headers()

if __name__ == "__main__":
    PORT = 8443
    
    with socketserver.TCPServer(("", PORT), CORSHTTPRequestHandler) as httpd:
        # 创建SSL上下文
        context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
        context.load_cert_chain('localhost.pem', 'localhost-key.pem')  # 需要自签名证书
        
        httpd.socket = context.wrap_socket(httpd.socket, server_side=True)
        
        print(f"HTTPS服务器运行在端口 {PORT}")
        print(f"访问: https://localhost:{PORT}")
        httpd.serve_forever()
```

### 4.2 生成自签名证书

```bash
#!/bin/bash
# generate-cert.sh

# 生成私钥
openssl genrsa -out localhost-key.pem 2048

# 生成证书请求
openssl req -new -key localhost-key.pem -out localhost.csr -subj "/C=US/ST=CA/L=San Francisco/O=Dev/CN=localhost"

# 生成自签名证书
openssl x509 -req -days 365 -in localhost.csr -signkey localhost-key.pem -out localhost.pem

echo "证书生成完成: localhost.pem, localhost-key.pem"
```

## 5. 部署检查清单

### 5.1 文件检查

- [ ] `main.wasm` - Go编译的WASM文件
- [ ] `wasm_exec.js` - Go运行时支持
- [ ] `go-wasm-bridge.js` - 桥接器代码
- [ ] `integration-example.js` - 集成示例
- [ ] `godot.editor.worker.enhanced.js` - 增强Worker
- [ ] 原有Godot文件 (html, js, wasm, pck等)

### 5.2 服务器检查

- [ ] HTTPS配置正确
- [ ] 正确的MIME类型设置
- [ ] 必要的安全头配置
- [ ] CORS配置（如果需要）
- [ ] Gzip压缩启用

### 5.3 功能测试

```javascript
// test-integration.js
// 在浏览器控制台中运行的测试脚本

async function testGoWasmIntegration() {
    console.log('开始测试Go WASM集成...');
    
    try {
        // 1. 检查Go模块状态
        const status = GoWasm.getStatus();
        console.log('Go模块状态:', status);
        
        if (!status.ready) {
            console.log('等待Go模块就绪...');
            await GoWasm.waitForReady();
        }
        
        // 2. 测试基本函数调用
        console.log('测试基本计算...');
        const sumResult = await GoWasm.call('goCalculateSum', 10, 20);
        console.log('求和结果:', sumResult);
        
        // 3. 测试游戏初始化
        console.log('测试游戏初始化...');
        const gameState = await GoWasm.call('goInitializeGame', JSON.stringify({
            playerName: 'TestPlayer',
            difficulty: 'normal'
        }));
        console.log('游戏状态:', JSON.parse(gameState));
        
        // 4. 测试玩家行动处理
        console.log('测试玩家行动...');
        const actionResult = await GoWasm.call('goProcessPlayerAction', JSON.stringify({
            type: 'move',
            target: 'forest',
            data: {}
        }));
        console.log('行动结果:', actionResult);
        
        // 5. 测试异步处理
        console.log('测试异步处理...');
        const asyncResult = await GoWasm.call('goAsyncProcess');
        console.log('异步结果:', asyncResult);
        
        // 6. 测试批量调用
        console.log('测试批量调用...');
        const batchResults = await GoWasm.callBatch([
            { funcName: 'goCalculateSum', args: [1, 2] },
            { funcName: 'goCalculateSum', args: [3, 4] },
            { funcName: 'goCalculateGameStats', args: [] }
        ]);
        console.log('批量结果:', batchResults);
        
        console.log('✅ 所有测试通过！');
        return true;
        
    } catch (error) {
        console.error('❌ 测试失败:', error);
        return false;
    }
}

// 运行测试
testGoWasmIntegration();
```

## 6. 性能优化建议

### 6.1 文件压缩

```bash
# 压缩WASM文件
wasm-opt -O3 -o main.optimized.wasm main.wasm

# 压缩JavaScript文件
terser integration-example.js -o integration-example.min.js -c -m
```

### 6.2 预加载配置

```html
<!-- 在HTML头部添加预加载 -->
<link rel="preload" href="main.wasm" as="fetch" type="application/wasm" crossorigin>
<link rel="preload" href="wasm_exec.js" as="script">
<link rel="preload" href="go-wasm-bridge.js" as="script">
```

### 6.3 Service Worker缓存

```javascript
// sw.js
const CACHE_NAME = 'godot-go-wasm-v1';
const urlsToCache = [
    '/',
    '/godot.editor.html',
    '/godot.editor.js',
    '/godot.editor.wasm',
    '/godot.editor.pck',
    '/main.wasm',
    '/wasm_exec.js',
    '/go-wasm-bridge.js',
    '/integration-example.js'
];

self.addEventListener('install', (event) => {
    event.waitUntil(
        caches.open(CACHE_NAME)
            .then((cache) => cache.addAll(urlsToCache))
    );
});

self.addEventListener('fetch', (event) => {
    event.respondWith(
        caches.match(event.request)
            .then((response) => {
                return response || fetch(event.request);
            })
    );
});
```

这个部署指南提供了完整的设置流程，从Go代码编写到服务器配置，再到测试和优化。按照这个指南，你应该能够成功部署一个集成了Go WASM的Godot Web应用程序。 