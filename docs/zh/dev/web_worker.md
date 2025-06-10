# Godot Web导出的Worker实现分析

## 概述

本文档详细分析了Godot引擎在Web平台上如何实现独立的Worker运行机制。通过分析`godot.editor.html`、`godot.editor.js`、`godot.editor.worker.js`和`godot.editor.audio.worklet.js`等核心文件，深入了解其多线程架构设计。

## 1. Worker类型架构

Godot使用了两种类型的Worker来实现不同的功能：

### 1.1 pthread Worker (`godot.editor.worker.js`)
- **用途**: 处理多线程计算任务
- **基础**: 基于Emscripten的pthread实现
- **特点**: 每个Worker运行完整的WebAssembly模块实例
- **适用场景**: CPU密集型计算、游戏逻辑处理

### 1.2 Audio Worklet (`godot.editor.audio.worklet.js`)
- **用途**: 专门用于音频处理
- **基础**: 运行在AudioWorklet上下文中
- **特点**: 提供低延迟的音频处理能力
- **适用场景**: 实时音频处理、音效播放

## 2. Worker管理系统

### 2.1 PThread对象结构

在主线程的`godot.editor.js`中，通过`PThread`对象管理Worker池：

```javascript
PThread = {
    unusedWorkers: [],    // 未使用的worker池
    runningWorkers: [],   // 正在运行的worker
    pthreads: {},        // pthread指针到worker的映射
    nextWorkerID: 1,     // worker ID计数器
    
    // Worker分配
    allocateUnusedWorker: function() {
        var worker = new Worker("godot.editor.worker.js");
        PThread.unusedWorkers.push(worker);
    },
    
    // 获取新Worker
    getNewWorker: function() {
        if (PThread.unusedWorkers.length == 0) {
            PThread.allocateUnusedWorker();
            PThread.loadWasmModuleToWorker(PThread.unusedWorkers[0]);
        }
        return PThread.unusedWorkers.pop();
    }
}
```

### 2.2 Worker生命周期管理

- **创建**: 动态创建Worker实例
- **初始化**: 加载WebAssembly模块和共享内存
- **运行**: 处理具体任务
- **回收**: 任务完成后回到Worker池
- **销毁**: 在需要时终止Worker

## 3. Worker初始化流程

### 3.1 步骤1: 主线程创建Worker

```javascript
// 在PThread.allocateUnusedWorker()中
var worker = new Worker("godot.editor.worker.js");
```

### 3.2 步骤2: 向Worker发送加载命令

```javascript
// 在PThread.loadWasmModuleToWorker()中
worker.postMessage({
    "cmd": "load",
    "handlers": handlers,                    // 事件处理器
    "urlOrBlob": Module["mainScriptUrlOrBlob"], // 主脚本URL
    "wasmMemory": wasmMemory,               // 共享内存
    "wasmModule": wasmModule,               // WebAssembly模块
    "workerID": worker.workerID             // Worker唯一ID
});
```

### 3.3 步骤3: Worker接收并处理加载命令

```javascript
// 在godot.editor.worker.js中
if (e.data.cmd === 'load') {
    // 设置模块配置
    Module['wasmModule'] = e.data.wasmModule;
    Module['wasmMemory'] = e.data.wasmMemory;
    Module['workerID'] = e.data.workerID;
    Module['ENVIRONMENT_IS_PTHREAD'] = true;
    
    // 设置事件处理器
    for (const handler of e.data.handlers) {
        Module[handler] = function() {
            postMessage({ cmd: 'callHandler', handler, args: [...arguments] });
        }
    }
    
    // 导入主脚本并初始化Godot
    importScripts(e.data.urlOrBlob);
    Godot(Module);
}
```

## 4. 关键技术实现

### 4.1 共享内存机制

#### SharedArrayBuffer的使用
```javascript
// 内存视图更新函数
function GROWABLE_HEAP_I8() {
    if (wasmMemory.buffer != HEAP8.buffer) {
        updateMemoryViews();
    }
    return HEAP8;
}
```

- 使用`SharedArrayBuffer`实现主线程和Worker之间的内存共享
- WebAssembly模块在所有线程间共享同一个内存实例
- 支持动态内存增长，自动更新内存视图

### 4.2 消息传递系统

#### 支持的命令类型
```javascript
// Worker消息处理
function handleMessage(e) {
    if (e.data.cmd === 'load') {
        // 加载WebAssembly模块
    } else if (e.data.cmd === 'run') {
        // 运行线程入口点
        Module['invokeEntryPoint'](e.data.start_routine, e.data.arg);
    } else if (e.data.cmd === 'cancel') {
        // 取消线程执行
        Module['__emscripten_thread_exit'](-1);
    } else if (e.data.cmd === 'checkMailbox') {
        // 检查消息邮箱
        Module['checkMailbox']();
    }
}
```

### 4.3 原子操作和同步

#### 环形缓冲区实现
```javascript
// 在audio.worklet.js中
class RingBuffer {
    constructor(p_buffer, p_state, p_threads) {
        this.buffer = p_buffer;
        this.avail = p_state;
        this.threads = p_threads;
    }
    
    read(output) {
        // 使用Atomics进行线程安全操作
        if (this.threads) {
            Atomics.add(this.avail, 0, -output.length);
            Atomics.notify(this.avail, 0);
        }
    }
    
    write(p_buffer) {
        if (this.threads) {
            Atomics.add(this.avail, 0, to_write);
            Atomics.notify(this.avail, 0);
        }
    }
}
```

## 5. 音频处理Worker详解

### 5.1 GodotProcessor类

```javascript
class GodotProcessor extends AudioWorkletProcessor {
    constructor() {
        super();
        this.threads = false;
        this.running = true;
        this.output = null;
        this.input = null;
        // 初始化环形缓冲区
    }
    
    process(inputs, outputs, parameters) {
        // 处理音频输入输出
        // 使用环形缓冲区进行数据传输
        return this.running;
    }
}
```

### 5.2 音频数据流处理

- **输入处理**: 从Web Audio API接收音频数据
- **环形缓冲**: 使用线程安全的环形缓冲区
- **输出处理**: 将处理后的音频数据发送回主线程

## 6. 实际运行机制

### 6.1 线程创建流程

1. **请求线程**: Godot调用`pthread_create`
2. **获取Worker**: `PThread.getNewWorker()`从池中获取Worker
3. **初始化线程**: 发送`run`命令到Worker
4. **执行任务**: Worker调用`invokeEntryPoint`执行具体任务
5. **清理回收**: 任务完成后回收Worker到池中

### 6.2 内存管理

```javascript
// WebAssembly实例化
Module['instantiateWasm'] = (info, receiveInstance) => {
    var module = Module['wasmModule'];
    var instance = new WebAssembly.Instance(module, info);
    return receiveInstance(instance);
}
```

### 6.3 线程同步

- 使用`Atomics.wait`和`Atomics.notify`进行线程同步
- 邮箱机制(`checkMailbox`)处理线程间通信
- 线程本地存储(TLS)管理线程特定数据

## 7. HTML文件的作用

### 7.1 主要功能

`godot.editor.html`文件主要负责：

```javascript
// 引擎配置
const GODOT_CONFIG = {
    "args": [],
    "canvasResizePolicy": 2,
    "executable": "godot.editor",
    "experimentalVK": false,
    "fileSizes": {
        "godot.editor.pck": 80384,
        "godot.editor.wasm": 39020050
    }
};

// 创建引擎实例
const engine = new Engine(GODOT_CONFIG);
```

### 7.2 启动流程

1. **加载资源**: 显示进度条，加载WASM和PCK文件
2. **初始化引擎**: 创建Engine实例
3. **启动游戏**: 调用`engine.startGame()`
4. **用户界面**: 提供Canvas和状态显示

## 8. 性能优化策略

### 8.1 Worker池管理

- **预分配**: 提前创建Worker避免动态创建开销
- **复用**: Worker完成任务后回到池中重复使用
- **负载均衡**: 动态分配任务到可用Worker

### 8.2 内存优化

- **共享内存**: 减少数据复制开销
- **视图更新**: 动态更新内存视图适应增长
- **垃圾回收**: 及时清理不需要的数据

### 8.3 通信优化

- **批量操作**: 合并多个小操作减少消息传递
- **传输对象**: 使用Transferable Objects避免数据复制
- **优先级队列**: 重要消息优先处理

## 9. 技术限制和挑战

### 9.1 浏览器兼容性

- **SharedArrayBuffer**: 需要HTTPS和特定安全策略
- **AudioWorklet**: 现代浏览器支持
- **Worker**: 基本所有浏览器支持

### 9.2 性能限制

- **上下文切换**: Worker间切换有一定开销
- **内存带宽**: 共享内存访问可能成为瓶颈
- **调试困难**: 多线程程序调试复杂

## 10. 总结

Godot的Web Worker实现是一个完整的多线程系统，通过以下关键技术实现：

1. **Worker池管理** - 动态创建和管理Worker实例
2. **模块共享** - 所有Worker共享同一个WebAssembly模块  
3. **内存共享** - 使用SharedArrayBuffer实现高效的内存共享
4. **消息通信** - 完善的消息传递机制处理线程间通信
5. **专门优化** - 针对音频处理使用专门的AudioWorklet

这种设计让Godot能够在Web平台上充分利用多核CPU，提供接近原生应用的性能表现，是现代Web应用多线程架构的优秀范例。

## 11. 在Worker中集成Go WASM模块

### 11.1 架构设计

在Godot Worker中集成Go WASM模块需要设计一个三层架构：

```
┌─────────────────┐
│   Godot WASM    │ ← 游戏引擎核心
├─────────────────┤
│  JavaScript胶水  │ ← 函数桥接和数据转换
├─────────────────┤
│    Go WASM      │ ← 业务逻辑处理
└─────────────────┘
```

### 11.2 Go WASM模块准备

#### 11.2.1 Go代码示例

```go
// main.go
package main

import (
    "encoding/json"
    "syscall/js"
)

// 定义要暴露给JavaScript的函数
func calculateSum(this js.Value, args []js.Value) interface{} {
    if len(args) < 2 {
        return map[string]interface{}{
            "error": "需要至少两个参数",
        }
    }
    
    a := args[0].Float()
    b := args[1].Float()
    result := a + b
    
    return map[string]interface{}{
        "result": result,
        "status": "success",
    }
}

// 处理复杂数据结构
func processGameData(this js.Value, args []js.Value) interface{} {
    if len(args) < 1 {
        return map[string]interface{}{
            "error": "需要传入游戏数据",
        }
    }
    
    // 从JavaScript接收JSON数据
    jsonStr := args[0].String()
    var gameData map[string]interface{}
    
    if err := json.Unmarshal([]byte(jsonStr), &gameData); err != nil {
        return map[string]interface{}{
            "error": "JSON解析失败: " + err.Error(),
        }
    }
    
    // 处理游戏数据逻辑
    result := processLogic(gameData)
    
    // 返回处理结果
    resultJson, _ := json.Marshal(result)
    return string(resultJson)
}

func processLogic(data map[string]interface{}) map[string]interface{} {
    // 具体的业务逻辑处理
    return map[string]interface{}{
        "processed": true,
        "timestamp": js.Global().Get("Date").New().Call("getTime").Int(),
        "originalData": data,
    }
}

// 异步处理函数
func asyncProcess(this js.Value, args []js.Value) interface{} {
    // 创建Promise
    promise := js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
        resolve := args[0]
        reject := args[1]
        
        // 模拟异步处理
        js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
            result := map[string]interface{}{
                "message": "异步处理完成",
                "data": "processed_data",
            }
            resolve.Invoke(result)
            return nil
        }), 1000)
        
        return nil
    }))
    
    return promise
}

func main() {
    // 防止程序退出
    c := make(chan struct{}, 0)
    
    // 向JavaScript全局对象注册函数
    js.Global().Set("goCalculateSum", js.FuncOf(calculateSum))
    js.Global().Set("goProcessGameData", js.FuncOf(processGameData))
    js.Global().Set("goAsyncProcess", js.FuncOf(asyncProcess))
    
    // 通知JavaScript Go模块已准备就绪
    js.Global().Call("postMessage", map[string]interface{}{
        "cmd": "goReady",
        "functions": []string{
            "goCalculateSum",
            "goProcessGameData", 
            "goAsyncProcess",
        },
    })
    
    <-c
}
```

#### 11.2.2 编译Go模块

```bash
# 编译Go代码为WASM
GOOS=js GOARCH=wasm go build -o main.wasm main.go

# 复制Go运行时支持文件
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
```

### 11.3 JavaScript胶水层实现

#### 11.3.1 修改Worker文件

```javascript
// godot.editor.worker.js - 增强版本
'use strict';

var Module = {};
var GoModule = {}; // Go模块相关
var initializedJS = false;
var goReady = false;
var pendingGoCalls = []; // 等待Go模块就绪的调用

// Go WASM支持
var goWasmSupport = {
    // 加载Go运行时
    loadGoRuntime: function() {
        return new Promise((resolve, reject) => {
            try {
                // 导入Go运行时
                importScripts('./wasm_exec.js');
                
                // 创建Go实例
                const go = new Go();
                GoModule.go = go;
                
                // 加载Go WASM模块
                WebAssembly.instantiateStreaming(fetch('./main.wasm'), go.importObject)
                    .then((result) => {
                        GoModule.instance = result.instance;
                        
                        // 运行Go程序
                        go.run(result.instance);
                        
                        resolve();
                    })
                    .catch(reject);
            } catch (error) {
                reject(error);
            }
        });
    },
    
    // 调用Go函数的包装器
    callGoFunction: function(funcName, ...args) {
        return new Promise((resolve, reject) => {
            if (!goReady) {
                // 如果Go模块未就绪，加入待处理队列
                pendingGoCalls.push({ funcName, args, resolve, reject });
                return;
            }
            
            try {
                const goFunc = self[funcName];
                if (typeof goFunc === 'function') {
                    const result = goFunc(...args);
                    
                    // 处理Promise返回值
                    if (result && typeof result.then === 'function') {
                        result.then(resolve).catch(reject);
                    } else {
                        resolve(result);
                    }
                } else {
                    reject(new Error(`Go函数 ${funcName} 不存在`));
                }
            } catch (error) {
                reject(error);
            }
        });
    }
};

// 原有的消息处理函数增强
function handleMessage(e) {
    try {
        if (e.data.cmd === 'load') {
            // 原有Godot加载逻辑
            let messageQueue = [];
            self.onmessage = (e) => messageQueue.push(e);

            self.startWorker = (instance) => {
                Module = instance;
                
                // 同时初始化Go模块
                goWasmSupport.loadGoRuntime()
                    .then(() => {
                        console.log('Go WASM模块加载成功');
                        postMessage({ 'cmd': 'loaded' });
                    })
                    .catch((error) => {
                        console.error('Go WASM模块加载失败:', error);
                        postMessage({ 'cmd': 'loaded' }); // 即使Go模块失败也继续
                    });

                for (let msg of messageQueue) {
                    handleMessage(msg);
                }
                self.onmessage = handleMessage;
            };

            // 原有Godot模块加载逻辑
            Module['wasmModule'] = e.data.wasmModule;
            Module['wasmMemory'] = e.data.wasmMemory;
            Module['buffer'] = Module['wasmMemory'].buffer;
            Module['workerID'] = e.data.workerID;
            Module['ENVIRONMENT_IS_PTHREAD'] = true;

            for (const handler of e.data.handlers) {
                Module[handler] = function() {
                    postMessage({ cmd: 'callHandler', handler, args: [...arguments] });
                }
            }

            if (typeof e.data.urlOrBlob == 'string') {
                importScripts(e.data.urlOrBlob);
            } else {
                var objectUrl = URL.createObjectURL(e.data.urlOrBlob);
                importScripts(objectUrl);
                URL.revokeObjectURL(objectUrl);
            }
            Godot(Module);
            
        } else if (e.data.cmd === 'callGoFunction') {
            // 新增：调用Go函数
            const { funcName, args, callId } = e.data;
            
            goWasmSupport.callGoFunction(funcName, ...args)
                .then(result => {
                    postMessage({
                        cmd: 'goFunctionResult',
                        callId: callId,
                        success: true,
                        result: result
                    });
                })
                .catch(error => {
                    postMessage({
                        cmd: 'goFunctionResult',
                        callId: callId,
                        success: false,
                        error: error.message
                    });
                });
                
        } else if (e.data.cmd === 'run') {
            // 原有运行逻辑
            Module['__emscripten_thread_init'](e.data.pthread_ptr, 0, 0, 1);
            Module['__emscripten_thread_mailbox_await'](e.data.pthread_ptr);

            assert(e.data.pthread_ptr);
            Module['establishStackSpace']();
            Module['PThread'].receiveObjectTransfer(e.data);
            Module['PThread'].threadInitTLS();

            if (!initializedJS) {
                initializedJS = true;
            }

            try {
                Module['invokeEntryPoint'](e.data.start_routine, e.data.arg);
            } catch(ex) {
                if (ex != 'unwind') {
                    throw ex;
                }
            }
        } else if (e.data.cmd === 'cancel') {
            if (Module['_pthread_self']()) {
                Module['__emscripten_thread_exit'](-1);
            }
        } else if (e.data.target === 'setimmediate') {
            // no-op
        } else if (e.data.cmd === 'checkMailbox') {
            if (initializedJS) {
                Module['checkMailbox']();
            }
        } else if (e.data.cmd) {
            err('worker.js received unknown command ' + e.data.cmd);
            err(e.data);
        }
    } catch(ex) {
        err('worker.js onmessage() captured an uncaught exception: ' + ex);
        if (ex && ex.stack) err(ex.stack);
        if (Module['__emscripten_thread_crashed']) {
            Module['__emscripten_thread_crashed']();
        }
        throw ex;
    }
}

// 监听来自Go的消息
self.addEventListener('message', function(e) {
    if (e.data && e.data.cmd === 'goReady') {
        goReady = true;
        console.log('Go模块就绪，可用函数:', e.data.functions);
        
        // 处理等待的Go函数调用
        pendingGoCalls.forEach(({ funcName, args, resolve, reject }) => {
            goWasmSupport.callGoFunction(funcName, ...args)
                .then(resolve)
                .catch(reject);
        });
        pendingGoCalls = [];
        
        // 通知主线程Go模块已就绪
        postMessage({
            cmd: 'goModuleReady',
            availableFunctions: e.data.functions
        });
    }
});

self.onmessage = handleMessage;
```

### 11.4 主线程集成

#### 11.4.1 扩展PThread管理

```javascript
// 在godot.editor.js中扩展PThread对象
PThread.goWasmBridge = {
    callCounter: 0,
    pendingCalls: new Map(),
    
    // 调用Worker中的Go函数
    callGoFunction: function(workerID, funcName, ...args) {
        return new Promise((resolve, reject) => {
            const callId = ++this.callCounter;
            this.pendingCalls.set(callId, { resolve, reject });
            
            // 找到对应的Worker
            const worker = PThread.runningWorkers.find(w => w.workerID === workerID) ||
                          PThread.unusedWorkers.find(w => w.workerID === workerID);
            
            if (!worker) {
                reject(new Error(`Worker ${workerID} 不存在`));
                return;
            }
            
            worker.postMessage({
                cmd: 'callGoFunction',
                funcName: funcName,
                args: args,
                callId: callId
            });
            
            // 设置超时
            setTimeout(() => {
                if (this.pendingCalls.has(callId)) {
                    this.pendingCalls.delete(callId);
                    reject(new Error('Go函数调用超时'));
                }
            }, 10000); // 10秒超时
        });
    },
    
    // 处理Go函数调用结果
    handleGoFunctionResult: function(data) {
        const { callId, success, result, error } = data;
        const pending = this.pendingCalls.get(callId);
        
        if (pending) {
            this.pendingCalls.delete(callId);
            if (success) {
                pending.resolve(result);
            } else {
                pending.reject(new Error(error));
            }
        }
    }
};

// 扩展原有的Worker消息处理
const originalLoadWasmModuleToWorker = PThread.loadWasmModuleToWorker;
PThread.loadWasmModuleToWorker = function(worker) {
    return originalLoadWasmModuleToWorker.call(this, worker).then(() => {
        // 增强消息处理以支持Go函数调用结果
        const originalOnMessage = worker.onmessage;
        worker.onmessage = function(e) {
            const data = e.data;
            
            if (data.cmd === 'goFunctionResult') {
                PThread.goWasmBridge.handleGoFunctionResult(data);
            } else if (data.cmd === 'goModuleReady') {
                console.log(`Worker ${worker.workerID} 的Go模块已就绪:`, data.availableFunctions);
                worker.goReady = true;
                worker.availableGoFunctions = data.availableFunctions;
            } else {
                originalOnMessage.call(this, e);
            }
        };
        
        return worker;
    });
};
```

### 11.5 Godot中的JavaScript接口

#### 11.5.1 创建全局接口函数

```javascript
// 在godot.editor.html或主JavaScript文件中添加
window.GodotGoWasmBridge = {
    // 简单的同步调用包装
    callGoFunction: async function(funcName, ...args) {
        // 获取第一个可用的Worker
        const worker = PThread.runningWorkers.find(w => w.goReady) ||
                      PThread.unusedWorkers.find(w => w.goReady);
        
        if (!worker) {
            throw new Error('没有可用的Go WASM Worker');
        }
        
        return await PThread.goWasmBridge.callGoFunction(worker.workerID, funcName, ...args);
    },
    
    // 批量调用多个Go函数
    callGoFunctions: async function(calls) {
        const promises = calls.map(call => 
            this.callGoFunction(call.funcName, ...call.args)
        );
        return await Promise.all(promises);
    },
    
    // 获取可用的Go函数列表
    getAvailableGoFunctions: function() {
        const worker = PThread.runningWorkers.find(w => w.goReady) ||
                      PThread.unusedWorkers.find(w => w.goReady);
        return worker ? worker.availableGoFunctions : [];
    }
};

// 暴露给Godot的FFI接口
if (typeof GodotFFI !== 'undefined') {
    GodotFFI.goWasm = {
        // 简单数据类型调用
        callSimple: function(funcName, arg1, arg2) {
            return window.GodotGoWasmBridge.callGoFunction(funcName, arg1, arg2);
        },
        
        // 复杂数据调用（JSON）
        callWithJson: function(funcName, jsonString) {
            return window.GodotGoWasmBridge.callGoFunction(funcName, jsonString);
        },
        
        // 异步调用
        callAsync: async function(funcName, ...args) {
            try {
                const result = await window.GodotGoWasmBridge.callGoFunction(funcName, ...args);
                return JSON.stringify({ success: true, data: result });
            } catch (error) {
                return JSON.stringify({ success: false, error: error.message });
            }
        }
    };
}
```

### 11.6 使用示例

#### 11.6.1 在Godot中调用Go函数

```javascript
// 在Godot的JavaScript代码中
async function callGoFromGodot() {
    try {
        // 调用简单的计算函数
        const sumResult = await GodotGoWasmBridge.callGoFunction('goCalculateSum', 10, 20);
        console.log('求和结果:', sumResult);
        
        // 调用复杂数据处理函数
        const gameData = {
            player: { x: 100, y: 200, health: 80 },
            enemies: [
                { x: 150, y: 250, type: 'orc' },
                { x: 300, y: 400, type: 'skeleton' }
            ]
        };
        
        const processResult = await GodotGoWasmBridge.callGoFunction(
            'goProcessGameData', 
            JSON.stringify(gameData)
        );
        console.log('处理结果:', JSON.parse(processResult));
        
        // 异步处理
        const asyncResult = await GodotGoWasmBridge.callGoFunction('goAsyncProcess');
        console.log('异步处理结果:', asyncResult);
        
    } catch (error) {
        console.error('调用Go函数失败:', error);
    }
}

// 批量调用示例
async function batchCallGo() {
    const calls = [
        { funcName: 'goCalculateSum', args: [1, 2] },
        { funcName: 'goCalculateSum', args: [3, 4] },
        { funcName: 'goCalculateSum', args: [5, 6] }
    ];
    
    try {
        const results = await GodotGoWasmBridge.callGoFunctions(calls);
        console.log('批量调用结果:', results);
    } catch (error) {
        console.error('批量调用失败:', error);
    }
}
```

### 11.7 错误处理和调试

#### 11.7.1 错误处理策略

```javascript
// 增强的错误处理
PThread.goWasmBridge.callGoFunctionSafe = async function(workerID, funcName, ...args) {
    try {
        // 参数验证
        if (!funcName || typeof funcName !== 'string') {
            throw new Error('函数名必须是字符串');
        }
        
        // Worker验证
        const worker = PThread.runningWorkers.find(w => w.workerID === workerID) ||
                      PThread.unusedWorkers.find(w => w.workerID === workerID);
        
        if (!worker) {
            throw new Error(`Worker ${workerID} 不存在`);
        }
        
        if (!worker.goReady) {
            throw new Error(`Worker ${workerID} 的Go模块未就绪`);
        }
        
        // 调用函数
        const result = await this.callGoFunction(workerID, funcName, ...args);
        
        // 结果验证
        if (result && typeof result === 'object' && result.error) {
            throw new Error(`Go函数执行错误: ${result.error}`);
        }
        
        return result;
        
    } catch (error) {
        console.error(`调用Go函数 ${funcName} 失败:`, error);
        
        // 记录调试信息
        console.log('调试信息:', {
            workerID,
            funcName,
            args,
            availableWorkers: PThread.runningWorkers.length + PThread.unusedWorkers.length,
            goReadyWorkers: [...PThread.runningWorkers, ...PThread.unusedWorkers]
                           .filter(w => w.goReady).length
        });
        
        throw error;
    }
};
```

### 11.8 性能优化建议

#### 11.8.1 数据传输优化

```javascript
// 使用Transferable Objects优化大数据传输
PThread.goWasmBridge.callGoFunctionWithTransfer = function(workerID, funcName, transferableData, ...args) {
    return new Promise((resolve, reject) => {
        const callId = ++this.callCounter;
        this.pendingCalls.set(callId, { resolve, reject });
        
        const worker = PThread.runningWorkers.find(w => w.workerID === workerID) ||
                      PThread.unusedWorkers.find(w => w.workerID === workerID);
        
        if (!worker) {
            reject(new Error(`Worker ${workerID} 不存在`));
            return;
        }
        
        // 使用Transferable Objects
        const transferList = [];
        if (transferableData instanceof ArrayBuffer) {
            transferList.push(transferableData);
        }
        
        worker.postMessage({
            cmd: 'callGoFunction',
            funcName: funcName,
            args: [transferableData, ...args],
            callId: callId
        }, transferList);
    });
};
```

### 11.9 部署和配置

#### 11.9.1 文件结构

```
project/
├── godot.editor.html          # 主HTML文件
├── godot.editor.js           # Godot主JS文件（增强版）
├── godot.editor.worker.js    # Worker文件（增强版）
├── godot.editor.wasm         # Godot WASM模块
├── godot.editor.pck          # Godot资源包
├── main.wasm                 # Go WASM模块
├── wasm_exec.js             # Go运行时支持
└── godot.editor.audio.worklet.js # 音频处理
```

#### 11.9.2 HTTP头配置

```
# 需要的HTTP响应头（用于SharedArrayBuffer支持）
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin

# WASM MIME类型
application/wasm: .wasm
```

这个集成方案让你可以在Godot的Worker环境中同时运行Go WASM模块，通过JavaScript胶水层实现两个WASM模块间的无缝集成。Go模块可以处理复杂的业务逻辑，而Godot专注于游戏引擎功能，两者通过高效的消息传递机制协同工作。

## 参考文件

- `godot.editor.html` - 主HTML文件和启动逻辑
- `godot.editor.js` - 主引擎代码和Worker管理
- `godot.editor.worker.js` - Worker线程实现
- `godot.editor.audio.worklet.js` - 音频处理Worker
- `godot.editor.wasm` - WebAssembly模块
- `godot.editor.pck` - 游戏资源包
- `main.wasm` - Go WASM模块
- `wasm_exec.js` - Go运行时支持文件 