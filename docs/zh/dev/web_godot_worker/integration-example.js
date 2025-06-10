/**
 * Godot + Go WASM 集成示例
 * 
 * 这个文件展示了如何在实际项目中集成Go WASM模块到Godot Worker中
 * 包括完整的设置、错误处理和最佳实践
 */

// ================================
// 1. 修改后的 Godot Worker 文件
// ================================

/**
 * godot.editor.worker.enhanced.js
 * 增强版本的Godot Worker，集成Go WASM支持
 */
class EnhancedGodotWorker {
    constructor() {
        this.Module = {};
        this.GoModule = {};
        this.initializedJS = false;
        this.goWasmBridge = null;
        
        // 初始化Go WASM桥接器
        this.initializeGoWasmBridge();
        
        // 绑定消息处理器
        self.onmessage = this.handleMessage.bind(this);
    }
    
    /**
     * 初始化Go WASM桥接器
     */
    async initializeGoWasmBridge() {
        try {
            // 导入GoWasmBridge
            importScripts('./go-wasm-bridge.js');
            
            // 创建桥接器实例
            this.goWasmBridge = new GoWasmBridge();
            
            console.log('Go WASM桥接器创建成功');
        } catch (error) {
            console.error('初始化Go WASM桥接器失败:', error);
        }
    }
    
    /**
     * 处理来自主线程的消息
     */
    async handleMessage(e) {
        try {
            const data = e.data;
            
            switch (data.cmd) {
                case 'load':
                    await this.handleLoadCommand(data);
                    break;
                    
                case 'callGoFunction':
                    await this.handleGoFunctionCall(data);
                    break;
                    
                case 'run':
                    this.handleRunCommand(data);
                    break;
                    
                case 'cancel':
                    this.handleCancelCommand();
                    break;
                    
                case 'checkMailbox':
                    this.handleCheckMailbox();
                    break;
                    
                default:
                    console.warn('未知命令:', data.cmd);
            }
            
        } catch (error) {
            console.error('Worker消息处理错误:', error);
            this.reportError(error);
        }
    }
    
    /**
     * 处理加载命令
     */
    async handleLoadCommand(data) {
        let messageQueue = [];
        self.onmessage = (e) => messageQueue.push(e);
        
        // 设置Worker启动回调
        self.startWorker = async (instance) => {
            this.Module = instance;
            
            try {
                // 并行初始化Godot和Go模块
                await Promise.all([
                    this.initializeGodotModule(),
                    this.initializeGoModule()
                ]);
                
                console.log('Worker初始化完成');
                postMessage({ 'cmd': 'loaded' });
                
                // 处理队列中的消息
                for (let msg of messageQueue) {
                    await this.handleMessage(msg);
                }
                
                // 恢复正常消息处理
                self.onmessage = this.handleMessage.bind(this);
                
            } catch (error) {
                console.error('Worker初始化失败:', error);
                postMessage({ 'cmd': 'loaded', 'error': error.message });
            }
        };
        
        // 设置Godot模块配置
        this.setupGodotModule(data);
        
        // 加载Godot脚本
        this.loadGodotScript(data);
    }
    
    /**
     * 设置Godot模块
     */
    setupGodotModule(data) {
        this.Module['wasmModule'] = data.wasmModule;
        this.Module['wasmMemory'] = data.wasmMemory;
        this.Module['buffer'] = this.Module['wasmMemory'].buffer;
        this.Module['workerID'] = data.workerID;
        this.Module['ENVIRONMENT_IS_PTHREAD'] = true;
        
        // 设置事件处理器
        for (const handler of data.handlers) {
            this.Module[handler] = (...args) => {
                postMessage({ 
                    cmd: 'callHandler', 
                    handler, 
                    args: [...args] 
                });
            };
        }
    }
    
    /**
     * 加载Godot脚本
     */
    loadGodotScript(data) {
        if (typeof data.urlOrBlob === 'string') {
            importScripts(data.urlOrBlob);
        } else {
            const objectUrl = URL.createObjectURL(data.urlOrBlob);
            importScripts(objectUrl);
            URL.revokeObjectURL(objectUrl);
        }
        
        // 启动Godot
        Godot(this.Module);
    }
    
    /**
     * 初始化Godot模块
     */
    async initializeGodotModule() {
        return new Promise((resolve) => {
            // Godot模块的初始化通常是同步的
            console.log('Godot模块初始化完成');
            resolve();
        });
    }
    
    /**
     * 初始化Go模块
     */
    async initializeGoModule() {
        if (!this.goWasmBridge) {
            throw new Error('Go WASM桥接器未初始化');
        }
        
        await this.goWasmBridge.initialize({
            wasmPath: './main.wasm',
            runtimePath: './wasm_exec.js',
            timeout: 10000,
            enableDebug: true
        });
        
        console.log('Go模块初始化完成');
    }
    
    /**
     * 处理Go函数调用
     */
    async handleGoFunctionCall(data) {
        const { funcName, args, callId } = data;
        
        try {
            if (!this.goWasmBridge || !this.goWasmBridge.isReady) {
                throw new Error('Go模块未就绪');
            }
            
            const result = await this.goWasmBridge.callGoFunctionSafe(funcName, ...args);
            
            postMessage({
                cmd: 'goFunctionResult',
                callId: callId,
                success: true,
                result: result
            });
            
        } catch (error) {
            postMessage({
                cmd: 'goFunctionResult',
                callId: callId,
                success: false,
                error: error.message
            });
        }
    }
    
    /**
     * 处理运行命令
     */
    handleRunCommand(data) {
        this.Module['__emscripten_thread_init'](data.pthread_ptr, 0, 0, 1);
        this.Module['__emscripten_thread_mailbox_await'](data.pthread_ptr);
        
        this.Module['establishStackSpace']();
        this.Module['PThread'].receiveObjectTransfer(data);
        this.Module['PThread'].threadInitTLS();
        
        if (!this.initializedJS) {
            this.initializedJS = true;
        }
        
        try {
            this.Module['invokeEntryPoint'](data.start_routine, data.arg);
        } catch (ex) {
            if (ex !== 'unwind') {
                throw ex;
            }
        }
    }
    
    /**
     * 处理取消命令
     */
    handleCancelCommand() {
        if (this.Module['_pthread_self']()) {
            this.Module['__emscripten_thread_exit'](-1);
        }
    }
    
    /**
     * 处理邮箱检查
     */
    handleCheckMailbox() {
        if (this.initializedJS) {
            this.Module['checkMailbox']();
        }
    }
    
    /**
     * 报告错误
     */
    reportError(error) {
        if (this.Module['__emscripten_thread_crashed']) {
            this.Module['__emscripten_thread_crashed']();
        }
        
        postMessage({
            cmd: 'workerError',
            error: error.message,
            stack: error.stack
        });
    }
}

// 创建Worker实例
const enhancedWorker = new EnhancedGodotWorker();

// ================================
// 2. 主线程集成代码
// ================================

/**
 * 主线程的Go WASM集成管理器
 */
class MainThreadGoWasmManager {
    constructor() {
        this.callCounter = 0;
        this.pendingCalls = new Map();
        this.workerGoStatus = new Map(); // 跟踪每个Worker的Go模块状态
        
        // 扩展现有的PThread对象
        this.extendPThreadObject();
    }
    
    /**
     * 扩展PThread对象以支持Go WASM
     */
    extendPThreadObject() {
        // 保存原始方法
        const originalLoadWasmModuleToWorker = PThread.loadWasmModuleToWorker;
        
        // 增强Worker加载方法
        PThread.loadWasmModuleToWorker = (worker) => {
            return originalLoadWasmModuleToWorker.call(PThread, worker).then(() => {
                // 增强消息处理
                const originalOnMessage = worker.onmessage;
                worker.onmessage = (e) => {
                    this.handleWorkerMessage(worker, e);
                    if (originalOnMessage) {
                        originalOnMessage.call(worker, e);
                    }
                };
                
                return worker;
            });
        };
        
        // 添加Go WASM相关方法到PThread
        PThread.goWasm = {
            callFunction: this.callGoFunction.bind(this),
            callFunctionSafe: this.callGoFunctionSafe.bind(this),
            callFunctions: this.callGoFunctions.bind(this),
            getAvailableWorkers: this.getGoReadyWorkers.bind(this),
            getWorkerStatus: this.getWorkerGoStatus.bind(this)
        };
    }
    
    /**
     * 处理Worker消息
     */
    handleWorkerMessage(worker, e) {
        const data = e.data;
        
        switch (data.cmd) {
            case 'goFunctionResult':
                this.handleGoFunctionResult(data);
                break;
                
            case 'goModuleReady':
                this.handleGoModuleReady(worker, data);
                break;
                
            case 'workerError':
                this.handleWorkerError(worker, data);
                break;
        }
    }
    
    /**
     * 处理Go模块就绪
     */
    handleGoModuleReady(worker, data) {
        console.log(`Worker ${worker.workerID} 的Go模块已就绪:`, data.availableFunctions);
        
        // 更新Worker状态
        this.workerGoStatus.set(worker.workerID, {
            ready: true,
            functions: data.availableFunctions,
            lastUpdate: Date.now()
        });
        
        worker.goReady = true;
        worker.availableGoFunctions = data.availableFunctions;
        
        // 触发就绪事件
        this.dispatchGoModuleReadyEvent(worker);
    }
    
    /**
     * 处理Go函数调用结果
     */
    handleGoFunctionResult(data) {
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
    
    /**
     * 处理Worker错误
     */
    handleWorkerError(worker, data) {
        console.error(`Worker ${worker.workerID} 错误:`, data.error);
        
        // 更新状态
        this.workerGoStatus.set(worker.workerID, {
            ready: false,
            error: data.error,
            lastUpdate: Date.now()
        });
    }
    
    /**
     * 调用Go函数
     */
    async callGoFunction(funcName, ...args) {
        const worker = this.getAvailableGoWorker();
        if (!worker) {
            throw new Error('没有可用的Go WASM Worker');
        }
        
        return this.callGoFunctionOnWorker(worker.workerID, funcName, ...args);
    }
    
    /**
     * 在指定Worker上调用Go函数
     */
    callGoFunctionOnWorker(workerID, funcName, ...args) {
        return new Promise((resolve, reject) => {
            const callId = ++this.callCounter;
            this.pendingCalls.set(callId, { resolve, reject });
            
            const worker = this.getWorkerById(workerID);
            if (!worker) {
                reject(new Error(`Worker ${workerID} 不存在`));
                return;
            }
            
            // 设置超时
            const timeout = setTimeout(() => {
                if (this.pendingCalls.has(callId)) {
                    this.pendingCalls.delete(callId);
                    reject(new Error('Go函数调用超时'));
                }
            }, 10000);
            
            // 发送调用请求
            worker.postMessage({
                cmd: 'callGoFunction',
                funcName: funcName,
                args: args,
                callId: callId
            });
        });
    }
    
    /**
     * 安全调用Go函数
     */
    async callGoFunctionSafe(funcName, ...args) {
        try {
            return await this.callGoFunction(funcName, ...args);
        } catch (error) {
            console.error(`安全调用Go函数 ${funcName} 失败:`, error);
            throw error;
        }
    }
    
    /**
     * 批量调用Go函数
     */
    async callGoFunctions(calls) {
        const promises = calls.map(call => 
            this.callGoFunction(call.funcName, ...(call.args || []))
        );
        return await Promise.all(promises);
    }
    
    /**
     * 获取可用的Go Worker
     */
    getAvailableGoWorker() {
        const workers = [...PThread.runningWorkers, ...PThread.unusedWorkers];
        return workers.find(w => w.goReady && this.workerGoStatus.get(w.workerID)?.ready);
    }
    
    /**
     * 获取所有Go就绪的Worker
     */
    getGoReadyWorkers() {
        const workers = [...PThread.runningWorkers, ...PThread.unusedWorkers];
        return workers.filter(w => w.goReady && this.workerGoStatus.get(w.workerID)?.ready);
    }
    
    /**
     * 根据ID获取Worker
     */
    getWorkerById(workerID) {
        const workers = [...PThread.runningWorkers, ...PThread.unusedWorkers];
        return workers.find(w => w.workerID === workerID);
    }
    
    /**
     * 获取Worker的Go模块状态
     */
    getWorkerGoStatus(workerID) {
        return this.workerGoStatus.get(workerID) || { ready: false };
    }
    
    /**
     * 触发Go模块就绪事件
     */
    dispatchGoModuleReadyEvent(worker) {
        const event = new CustomEvent('goModuleReady', {
            detail: {
                workerID: worker.workerID,
                functions: worker.availableGoFunctions
            }
        });
        
        if (typeof window !== 'undefined') {
            window.dispatchEvent(event);
        }
    }
}

// ================================
// 3. 全局接口和便利函数
// ================================

/**
 * 全局Go WASM接口
 */
class GlobalGoWasmInterface {
    constructor() {
        this.manager = new MainThreadGoWasmManager();
        this.isReady = false;
        
        // 监听Go模块就绪事件
        if (typeof window !== 'undefined') {
            window.addEventListener('goModuleReady', this.handleGoModuleReady.bind(this));
        }
    }
    
    /**
     * 处理Go模块就绪
     */
    handleGoModuleReady(event) {
        console.log('Go模块就绪:', event.detail);
        this.isReady = true;
    }
    
    /**
     * 等待Go模块就绪
     */
    async waitForReady(timeout = 30000) {
        if (this.isReady) return true;
        
        return new Promise((resolve, reject) => {
            const timeoutId = setTimeout(() => {
                reject(new Error('等待Go模块就绪超时'));
            }, timeout);
            
            const handler = () => {
                clearTimeout(timeoutId);
                window.removeEventListener('goModuleReady', handler);
                resolve(true);
            };
            
            window.addEventListener('goModuleReady', handler);
        });
    }
    
    /**
     * 调用Go函数的便利方法
     */
    async call(funcName, ...args) {
        await this.waitForReady();
        return this.manager.callGoFunction(funcName, ...args);
    }
    
    /**
     * 安全调用Go函数
     */
    async callSafe(funcName, ...args) {
        try {
            return await this.call(funcName, ...args);
        } catch (error) {
            console.error(`调用Go函数 ${funcName} 失败:`, error);
            return { error: error.message };
        }
    }
    
    /**
     * 批量调用Go函数
     */
    async callBatch(calls) {
        await this.waitForReady();
        return this.manager.callGoFunctions(calls);
    }
    
    /**
     * 获取状态信息
     */
    getStatus() {
        const readyWorkers = this.manager.getGoReadyWorkers();
        return {
            ready: this.isReady,
            workerCount: readyWorkers.length,
            availableFunctions: readyWorkers[0]?.availableGoFunctions || []
        };
    }
}

// 创建全局实例
const GoWasm = new GlobalGoWasmInterface();

// ================================
// 4. 使用示例
// ================================

/**
 * 在Godot项目中的使用示例
 */
class GameLogicWithGoWasm {
    constructor() {
        this.playerData = null;
        this.gameState = null;
    }
    
    /**
     * 初始化游戏逻辑
     */
    async initialize() {
        try {
            // 等待Go模块就绪
            await GoWasm.waitForReady();
            console.log('Go WASM模块就绪，状态:', GoWasm.getStatus());
            
            // 初始化游戏数据
            await this.initializeGameData();
            
        } catch (error) {
            console.error('游戏逻辑初始化失败:', error);
        }
    }
    
    /**
     * 初始化游戏数据
     */
    async initializeGameData() {
        // 调用Go函数处理初始化
        const initResult = await GoWasm.call('goInitializeGame', {
            playerName: 'Player1',
            difficulty: 'normal',
            gameMode: 'adventure'
        });
        
        console.log('游戏初始化结果:', initResult);
        this.gameState = initResult;
    }
    
    /**
     * 处理玩家行动
     */
    async handlePlayerAction(action) {
        try {
            const result = await GoWasm.call('goProcessPlayerAction', {
                action: action,
                playerState: this.playerData,
                gameState: this.gameState
            });
            
            // 更新游戏状态
            this.updateGameState(result);
            return result;
            
        } catch (error) {
            console.error('处理玩家行动失败:', error);
            return { error: error.message };
        }
    }
    
    /**
     * 计算游戏数据
     */
    async calculateGameStats() {
        const batchCalls = [
            { funcName: 'goCalculatePlayerStats', args: [this.playerData] },
            { funcName: 'goCalculateWorldStats', args: [this.gameState] },
            { funcName: 'goCalculatePerformance', args: [] }
        ];
        
        try {
            const results = await GoWasm.callBatch(batchCalls);
            console.log('批量计算结果:', results);
            return results;
            
        } catch (error) {
            console.error('批量计算失败:', error);
            return [];
        }
    }
    
    /**
     * 更新游戏状态
     */
    updateGameState(newData) {
        if (newData.playerData) {
            this.playerData = newData.playerData;
        }
        if (newData.gameState) {
            this.gameState = newData.gameState;
        }
    }
    
    /**
     * 安全的异步处理
     */
    async safeAsyncProcess(data) {
        const result = await GoWasm.callSafe('goAsyncProcess', JSON.stringify(data));
        
        if (result.error) {
            console.warn('异步处理出现错误:', result.error);
            return null;
        }
        
        return JSON.parse(result);
    }
}

// ================================
// 5. 导出和全局访问
// ================================

// 将接口暴露给全局
if (typeof window !== 'undefined') {
    window.GoWasm = GoWasm;
    window.GameLogicWithGoWasm = GameLogicWithGoWasm;
}

// 将接口暴露给Godot FFI
if (typeof GodotFFI !== 'undefined') {
    GodotFFI.goWasm = {
        call: GoWasm.call.bind(GoWasm),
        callSafe: GoWasm.callSafe.bind(GoWasm),
        callBatch: GoWasm.callBatch.bind(GoWasm),
        getStatus: GoWasm.getStatus.bind(GoWasm),
        waitForReady: GoWasm.waitForReady.bind(GoWasm)
    };
}

/**
 * 使用示例代码：
 * 
 * // 1. 基本使用
 * const result = await GoWasm.call('goCalculateSum', 10, 20);
 * console.log('计算结果:', result);
 * 
 * // 2. 安全调用
 * const safeResult = await GoWasm.callSafe('goRiskyFunction', data);
 * if (safeResult.error) {
 *     console.error('调用失败:', safeResult.error);
 * }
 * 
 * // 3. 批量调用
 * const batchResults = await GoWasm.callBatch([
 *     { funcName: 'goFunc1', args: [1, 2] },
 *     { funcName: 'goFunc2', args: ['hello'] }
 * ]);
 * 
 * // 4. 在游戏逻辑中使用
 * const gameLogic = new GameLogicWithGoWasm();
 * await gameLogic.initialize();
 * const actionResult = await gameLogic.handlePlayerAction('move_forward');
 */

console.log('Godot + Go WASM 集成模块加载完成'); 