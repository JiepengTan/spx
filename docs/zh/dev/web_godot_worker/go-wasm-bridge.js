/**
 * Go WASM Bridge for Godot Workers
 * 
 * 这个文件提供了在Godot Worker中集成Go WASM模块的完整解决方案
 * 包括模块加载、函数调用、错误处理和性能优化
 */

class GoWasmBridge {
    constructor() {
        this.goInstance = null;
        this.goRuntime = null;
        this.isReady = false;
        this.pendingCalls = [];
        this.callCounter = 0;
        this.activeCalls = new Map();
        
        // 配置选项
        this.config = {
            wasmPath: './main.wasm',
            timeout: 10000, // 10秒超时
            enableDebug: true
        };
        
        // 绑定方法
        this.loadGoModule = this.loadGoModule.bind(this);
        this.callGoFunction = this.callGoFunction.bind(this);
        this.handleGoMessage = this.handleGoMessage.bind(this);
    }
    
    /**
     * 初始化Go WASM模块
     * @param {Object} options 配置选项
     * @returns {Promise} 初始化Promise
     */
    async initialize(options = {}) {
        // 合并配置
        Object.assign(this.config, options);
        
        try {
            this.log('开始初始化Go WASM模块...');
            
            // 加载Go运行时
            await this.loadGoRuntime();
            
            // 加载Go WASM模块
            await this.loadGoModule();
            
            this.log('Go WASM模块初始化完成');
            return true;
            
        } catch (error) {
            this.error('Go WASM模块初始化失败:', error);
            throw error;
        }
    }
    
    /**
     * 加载Go运行时
     * @returns {Promise}
     */
    loadGoRuntime() {
        return new Promise((resolve, reject) => {
            try {
                // 导入Go运行时脚本
                if (this.config.runtimePath !== undefined && this.config.runtimePath !== null && this.config.runtimePath !== '') {
                    importScripts(this.config.runtimePath);
                }
                
                // 创建Go实例
                this.goRuntime = new Go();
                this.log('Go运行时加载成功');
                resolve();
                
            } catch (error) {
                reject(new Error(`加载Go运行时失败: ${error.message}`));
            }
        });
    }
    
    /**
     * 加载Go WASM模块
     * @returns {Promise}
     */
    async loadGoModule() {
        try {
            // 获取WASM字节码
            const wasmBytes = await this.fetchWasm(this.config.wasmPath);
            
            // 实例化WASM模块
            const wasmModule = await WebAssembly.instantiate(wasmBytes, this.goRuntime.importObject);
            this.goInstance = wasmModule.instance;
            
            // 设置消息监听
            this.setupMessageHandling();
            
            // 运行Go程序
            this.goRuntime.run(this.goInstance);
            
            this.log('Go WASM模块加载成功');
            
        } catch (error) {
            throw new Error(`加载Go WASM模块失败: ${error.message}`);
        }
    }
    
    /**
     * 获取WASM字节码
     * @param {string} wasmPath WASM文件路径
     * @returns {Promise<ArrayBuffer>}
     */
    async fetchWasm(wasmPath) {
        try {
            const response = await fetch(wasmPath);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            return await response.arrayBuffer();
        } catch (error) {
            throw new Error(`获取WASM文件失败: ${error.message}`);
        }
    }
    
    /**
     * 设置消息处理机制
     */
    setupMessageHandling() {
        // 监听来自Go的消息
        const originalPostMessage = self.postMessage;
        self.postMessage = (data) => {
            if (this.isGoMessage(data)) {
                this.handleGoMessage(data);
            } else {
                originalPostMessage.call(self, data);
            }
        };
    }
    
    /**
     * 判断是否为Go消息
     * @param {*} data 消息数据
     * @returns {boolean}
     */
    isGoMessage(data) {
        return data && typeof data === 'object' && 
               (data.cmd === 'goReady' || data.source === 'go-wasm');
    }
    
    /**
     * 处理来自Go的消息
     * @param {Object} data 消息数据
     */
    handleGoMessage(data) {
        switch (data.cmd) {
            case 'goReady':
                this.handleGoReady(data);
                break;
            case 'goFunction':
                this.handleGoFunctionCall(data);
                break;
            default:
                this.log('收到未知Go消息:', data);
        }
    }
    
    /**
     * 处理Go模块就绪消息
     * @param {Object} data 消息数据
     */
    handleGoReady(data) {
        this.isReady = true;
        this.log('Go模块就绪，可用函数:', data.functions);
        
        // 处理等待的函数调用
        this.processPendingCalls();
        
        // 通知主线程
        self.postMessage({
            cmd: 'goModuleReady',
            availableFunctions: data.functions || [],
            source: 'go-wasm-bridge'
        });
    }
    
    /**
     * 处理等待的函数调用
     */
    processPendingCalls() {
        while (this.pendingCalls.length > 0) {
            const call = this.pendingCalls.shift();
            this.executeGoFunction(call.funcName, call.args, call.resolve, call.reject);
        }
    }
    
    /**
     * 调用Go函数
     * @param {string} funcName 函数名
     * @param {...*} args 参数
     * @returns {Promise} 调用结果
     */
    callGoFunction(funcName, ...args) {
        return new Promise((resolve, reject) => {
            if (!this.isReady) {
                // 模块未就绪，加入待处理队列
                this.pendingCalls.push({ funcName, args, resolve, reject });
                return;
            }
            
            this.executeGoFunction(funcName, args, resolve, reject);
        });
    }
    
    /**
     * 执行Go函数
     * @param {string} funcName 函数名
     * @param {Array} args 参数数组
     * @param {Function} resolve 成功回调
     * @param {Function} reject 失败回调
     */
    executeGoFunction(funcName, args, resolve, reject) {
        try {
            // 验证函数存在
            const goFunc = self[funcName];
            if (typeof goFunc !== 'function') {
                reject(new Error(`Go函数 ${funcName} 不存在`));
                return;
            }
            
            // 设置超时
            const timeoutId = setTimeout(() => {
                reject(new Error(`Go函数 ${funcName} 调用超时`));
            }, this.config.timeout);
            
            // 调用函数
            const result = goFunc(...args);
            
            // 处理返回值
            if (result && typeof result.then === 'function') {
                // Promise返回值
                result
                    .then(value => {
                        clearTimeout(timeoutId);
                        resolve(value);
                    })
                    .catch(error => {
                        clearTimeout(timeoutId);
                        reject(error);
                    });
            } else {
                // 同步返回值
                clearTimeout(timeoutId);
                resolve(result);
            }
            
        } catch (error) {
            reject(new Error(`执行Go函数 ${funcName} 失败: ${error.message}`));
        }
    }
    
    /**
     * 批量调用Go函数
     * @param {Array} calls 调用配置数组 [{funcName, args}, ...]
     * @returns {Promise<Array>} 结果数组
     */
    async callGoFunctions(calls) {
        const promises = calls.map(call => 
            this.callGoFunction(call.funcName, ...(call.args || []))
        );
        return await Promise.all(promises);
    }
    
    /**
     * 获取可用的Go函数列表
     * @returns {Array} 函数名数组
     */
    getAvailableGoFunctions() {
        const functions = [];
        for (const key in self) {
            if (typeof self[key] === 'function' && key.startsWith('go')) {
                functions.push(key);
            }
        }
        return functions;
    }
    
    /**
     * 安全调用Go函数（带完整错误处理）
     * @param {string} funcName 函数名
     * @param {...*} args 参数
     * @returns {Promise} 调用结果
     */
    async callGoFunctionSafe(funcName, ...args) {
        try {
            // 参数验证
            if (!funcName || typeof funcName !== 'string') {
                throw new Error('函数名必须是有效字符串');
            }
            
            if (!this.isReady) {
                throw new Error('Go模块尚未就绪');
            }
            
            // 调用函数
            const result = await this.callGoFunction(funcName, ...args);
            
            // 结果验证
            if (result && typeof result === 'object' && result.error) {
                throw new Error(`Go函数执行错误: ${result.error}`);
            }
            
            return result;
            
        } catch (error) {
            this.error(`安全调用Go函数 ${funcName} 失败:`, error);
            
            // 记录调试信息
            if (this.config.enableDebug) {
                this.log('调试信息:', {
                    funcName,
                    args,
                    isReady: this.isReady,
                    availableFunctions: this.getAvailableGoFunctions()
                });
            }
            
            throw error;
        }
    }
    
    /**
     * 处理大数据传输的优化调用
     * @param {string} funcName 函数名
     * @param {ArrayBuffer} transferableData 可传输数据
     * @param {...*} args 其他参数
     * @returns {Promise} 调用结果
     */
    async callGoFunctionWithTransfer(funcName, transferableData, ...args) {
        // 注意：在Worker内部，Transferable Objects的优化有限
        // 但这个接口为将来的优化预留空间
        return this.callGoFunction(funcName, transferableData, ...args);
    }
    
    /**
     * 销毁Go模块实例
     */
    destroy() {
        this.log('销毁Go WASM模块实例');
        
        // 清理待处理的调用
        this.pendingCalls.forEach(call => {
            call.reject(new Error('Go模块已销毁'));
        });
        this.pendingCalls = [];
        
        // 清理活跃调用
        this.activeCalls.forEach(call => {
            call.reject(new Error('Go模块已销毁'));
        });
        this.activeCalls.clear();
        
        // 重置状态
        this.isReady = false;
        this.goInstance = null;
        this.goRuntime = null;
    }
    
    /**
     * 日志输出
     * @param {...*} args 日志参数
     */
    log(...args) {
        if (this.config.enableDebug) {
            console.log('[GoWasmBridge]', ...args);
        }
    }
    
    /**
     * 错误日志输出
     * @param {...*} args 错误参数
     */
    error(...args) {
        console.error('[GoWasmBridge]', ...args);
    }
}

// 导出供Worker使用
if (typeof self !== 'undefined' && typeof module === 'undefined') {
    // 在Worker环境中直接使用
    self.GoWasmBridge = GoWasmBridge;
} else if (typeof module !== 'undefined' && module.exports) {
    // Node.js环境
    module.exports = GoWasmBridge;
} else if (typeof window !== 'undefined') {
    // 浏览器环境
    window.GoWasmBridge = GoWasmBridge;
}

/**
 * 使用示例：
 * 
 * // 在Worker中使用
 * const bridge = new GoWasmBridge();
 * 
 * // 初始化
 * await bridge.initialize({
 *     wasmPath: './main.wasm',
 *     runtimePath: './wasm_exec.js',
 *     timeout: 5000,
 *     enableDebug: true
 * });
 * 
 * // 调用Go函数
 * const result = await bridge.callGoFunction('goCalculateSum', 10, 20);
 * console.log('计算结果:', result);
 * 
 * // 安全调用
 * try {
 *     const safeResult = await bridge.callGoFunctionSafe('goProcessData', data);
 *     console.log('处理结果:', safeResult);
 * } catch (error) {
 *     console.error('调用失败:', error);
 * }
 * 
 * // 批量调用
 * const batchResults = await bridge.callGoFunctions([
 *     { funcName: 'goFunc1', args: [1, 2] },
 *     { funcName: 'goFunc2', args: ['hello'] }
 * ]);
 */ 