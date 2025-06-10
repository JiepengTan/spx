/**
 * @license
 * Copyright 2015 The Emscripten Authors
 * SPDX-License-Identifier: MIT
 */

// Pthread Web Worker startup routine:
// This is the entry point file that is loaded first by each Web Worker
// that executes pthreads on the Emscripten application.

'use strict';

var Module = {};

// Thread-local guard variable for one-time init of the JS state
var initializedJS = false;

// Go WASM 模块全局状态控制 - 简化版本
var goWasmState = {
  isLoading: false,
  isLoaded: false,
  bridge: null,
  loadingPromise: null,
  error: null
};

function assert(condition, text) {
  if (!condition) abort('Assertion failed: ' + text);
}

function threadPrintErr() {
  var text = Array.prototype.slice.call(arguments).join(' ');
  console.error(text);
}
function threadAlert() {
  var text = Array.prototype.slice.call(arguments).join(' ');
  postMessage({ cmd: 'alert', text: text, threadId: Module['_pthread_self']() });
}
// We don't need out() for now, but may need to add it if we want to use it
// here. Or, if this code all moves into the main JS, that problem will go
// away. (For now, adding it here increases code size for no benefit.)
var out = () => { throw 'out() is not defined in worker.js.'; }
var err = threadPrintErr;
self.alert = threadAlert;

Module['instantiateWasm'] = (info, receiveInstance) => {
  // Instantiate from the module posted from the main thread.
  // We can just use sync instantiation in the worker.
  var module = Module['wasmModule'];
  // We don't need the module anymore; new threads will be spawned from the main thread.
  Module['wasmModule'] = null;
  var instance = new WebAssembly.Instance(module, info);
  // TODO: Due to Closure regression https://github.com/google/closure-compiler/issues/3193,
  // the above line no longer optimizes out down to the following line.
  // When the regression is fixed, we can remove this if/else.
  return receiveInstance(instance);
}

// Turn unhandled rejected promises into errors so that the main thread will be
// notified about them.
self.onunhandledrejection = (e) => {
  throw e.reason ?? e;
};

/**
 * 按需初始化 Go WASM（可在任何线程调用）
 * 这个函数会在 godot_js_spx_on_engine_start 回调时被调用
 */
async function initializeGoWasmOnDemand() {
  const workerId = (typeof Module !== 'undefined' && Module['workerID']) || 'unknown';
  const threadInfo = typeof importScripts !== 'undefined' ? 'Worker' : 'MainThread';

  console.log(`[线程 ${threadInfo}-${workerId}] 开始按需初始化 Go WASM`);

  // 如果已经初始化过了，直接返回
  if (self.goBridge && self.goBridge.isReady && typeof Module !== 'undefined' && Module['FFI']) {
    console.log(`[线程 ${threadInfo}-${workerId}] Go WASM 已就绪，无需重复初始化`);
    return true;
  }

  try {
    // 加载 Go WASM 模块
    await loadGoWasmModule();
    console.log(`[线程 ${threadInfo}-${workerId}] Go WASM 按需初始化成功`);
    return true;
  } catch (error) {
    console.error(`[线程 ${threadInfo}-${workerId}] Go WASM 按需初始化失败:`, error);
    return false;
  }
}

// 将函数暴露到全局作用域，以便在 godot.editor.js 中调用
if (typeof self !== 'undefined') {
  self.initializeGoWasmOnDemand = initializeGoWasmOnDemand;
}


/**
 * 简化的 Go WASM 模块加载逻辑
 */
async function loadGoWasmModule() {
  // 如果已经加载过了，直接返回
  if (self.goBridge && self.goBridge.isReady) {
    console.log(`[Godot Worker ${Module['workerID']}] Go WASM 已加载，直接使用`);
    return;
  }

  try {
    console.log(`[Godot Worker ${Module['workerID']}] 开始加载 Go WASM 模块...`);

    // 导入 Go WASM Bridge (只导入一次)
    if (typeof GoWasmBridge === 'undefined') {
      importScripts('./go-wasm-bridge.js');
    }
    if (typeof BindFFI === 'undefined') {
      importScripts('./wrap.gen.js');
    }

    // 创建 Go WASM Bridge 实例
    const goBridge = new GoWasmBridge();

    // 初始化 Go WASM 模块
    await goBridge.initialize({
      wasmPath: './gdspx.wasm',
      timeout: 15000,
      enableDebug: true
    });

    console.log(`[Godot Worker ${Module['workerID']}] Go WASM 模块加载成1111功`);

    // 尝试调用 Go 的初始化函数（可选）
    try {
      const initResult = await goBridge.callGoFunctionSafe('goWasmInit');
      console.log(`[Godot Worker ${Module['workerID']}] Go 初始化函数执行成功:`, initResult);
      Module['FFI'] = BindFFI(goBridge);
    } catch (goInitError) {
      console.warn(`[Godot Worker ${Module['workerID']}] Go 初始化函数调用失败，但继续执行:`, goInitError);
    }

    // 将 Go Bridge 实例暴露给当前 worker 的全局作用域
    self.goBridge = goBridge;

    console.log(`[Godot Worker ${Module['workerID']}] Go WASM 模块初始化完成`);

  } catch (error) {
    console.error(`[Godot Worker ${Module['workerID']}] Go WASM 模块加载失败:`, error);
    throw error;
  }
}

function handleMessage(e) {
  try {
    if (e.data.cmd === 'load') { // Preload command that is called once per worker to parse and load the Emscripten code.

      // Until we initialize the runtime, queue up any further incoming messages.
      let messageQueue = [];
      self.onmessage = (e) => messageQueue.push(e);

      // And add a callback for when the runtime is initialized.
      self.startWorker = async (instance) => {
        Module = instance;

        // Notify the main thread that this thread has loaded.
        postMessage({ 'cmd': 'loaded' });
        // Process any messages that were queued before the thread was ready.
        for (let msg of messageQueue) {
          handleMessage(msg);
        }
        // Restore the real message handler.
        self.onmessage = handleMessage;
      };

      // Module and memory were sent from main thread
      Module['wasmModule'] = e.data.wasmModule;

      // Use `const` here to ensure that the variable is scoped only to
      // that iteration, allowing safe reference from a closure.
      for (const handler of e.data.handlers) {
        Module[handler] = function () {
          postMessage({ cmd: 'callHandler', handler, args: [...arguments] });
        }
      }

      Module['wasmMemory'] = e.data.wasmMemory;

      Module['buffer'] = Module['wasmMemory'].buffer;

      Module['workerID'] = e.data.workerID;

      Module['ENVIRONMENT_IS_PTHREAD'] = true;

      if (typeof e.data.urlOrBlob == 'string') {
        importScripts(e.data.urlOrBlob);
      } else {
        var objectUrl = URL.createObjectURL(e.data.urlOrBlob);
        importScripts(objectUrl);
        URL.revokeObjectURL(objectUrl);
      }
      Godot(Module);
    } else if (e.data.cmd === 'run') {
      // Pass the thread address to wasm to store it for fast access.
      Module['__emscripten_thread_init'](e.data.pthread_ptr, /*isMainBrowserThread=*/0, /*isMainRuntimeThread=*/0, /*canBlock=*/1);

      // Await mailbox notifications with `Atomics.waitAsync` so we can start
      // using the fast `Atomics.notify` notification path.
      Module['__emscripten_thread_mailbox_await'](e.data.pthread_ptr);

      assert(e.data.pthread_ptr);
      // Also call inside JS module to set up the stack frame for this pthread in JS module scope
      Module['establishStackSpace']();
      Module['PThread'].receiveObjectTransfer(e.data);
      Module['PThread'].threadInitTLS();

      if (!initializedJS) {
        initializedJS = true;
      }

      try {
        Module['invokeEntryPoint'](e.data.start_routine, e.data.arg);
      } catch (ex) {
        if (ex != 'unwind') {
          // The pthread "crashed".  Do not call `_emscripten_thread_exit` (which
          // would make this thread joinable).  Instead, re-throw the exception
          // and let the top level handler propagate it back to the main thread.
          throw ex;
        }
      }
    } else if (e.data.cmd === 'cancel') { // Main thread is asking for a pthread_cancel() on this thread.
      if (Module['_pthread_self']()) {
        Module['__emscripten_thread_exit'](-1);
      }
    } else if (e.data.target === 'setimmediate') {
      // no-op
    } else if (e.data.cmd === 'checkMailbox') {
      if (initializedJS) {
        Module['checkMailbox']();
      }
    } else if (e.data._gameAppMessageId) {
      // 这是来自 GameApp 的消息，带有特殊标识
      handleGameAppMessage(e.data);
    } else if (e.data.cmd) {
      // The received message looks like something that should be handled by this message
      // handler, (since there is a e.data.cmd field present), but is not one of the
      // recognized commands:
      err('worker.js received unknown command ' + e.data.cmd);
      err(e.data);
    }
  } catch (ex) {
    err('worker.js onmessage() captured an uncaught exception: ' + ex);
    if (ex && ex.stack) err(ex.stack);
    if (Module['__emscripten_thread_crashed']) {
      Module['__emscripten_thread_crashed']();
    }
    throw ex;
  }
};

// === 新增：处理来自 GameApp 的消息 ===
function handleGameAppMessage(data) {
  const workerId = (typeof Module !== 'undefined' && Module['workerID']) || 'unknown';
  const threadInfo = typeof importScripts !== 'undefined' ? 'Worker' : 'MainThread';
  
  console.log(`[线程 ${threadInfo}-${workerId}] 收到 GameApp 消息:`, data.cmd || 'unknown', data);
  
  try {
    switch (data.cmd) {
      case 'projectDataUpdate':
        handleProjectDataUpdate(data);
        break;
        
      case 'customCommand':
        handleCustomCommand(data);
        break;
        
      case 'ping':
        // 响应 ping 消息
        postMessage({
          cmd: 'pong',
          workerID: workerId,
          originalMessageId: data._gameAppMessageId,
          timestamp: Date.now()
        });
        break;
        
      case 'getWorkerStatus':
        // 返回 worker 状态信息
        postMessage({
          cmd: 'workerStatus',
          workerID: workerId,
          pthread_ptr: (typeof Module !== 'undefined' && Module['_pthread_self']) ? Module['_pthread_self']() : 0,
          initializedJS: initializedJS,
          originalMessageId: data._gameAppMessageId,
          timestamp: Date.now()
        });
        break;
        
      case 'initGoWasm':
        // 按需初始化 Go WASM
        initializeGoWasmOnDemand().then(success => {
          postMessage({
            cmd: 'goWasmInitResult',
            success: success,
            workerID: workerId,
            originalMessageId: data._gameAppMessageId,
            timestamp: Date.now()
          });
        }).catch(error => {
          postMessage({
            cmd: 'goWasmInitResult',
            success: false,
            error: error.message,
            workerID: workerId,
            originalMessageId: data._gameAppMessageId,
            timestamp: Date.now()
          });
        });
        break;
        
      default:
        console.warn(`[线程 ${threadInfo}-${workerId}] 未知的 GameApp 命令:`, data.cmd);
        // 发送未知命令响应
        postMessage({
          cmd: 'unknownCommand',
          originalCmd: data.cmd,
          workerID: workerId,
          originalMessageId: data._gameAppMessageId,
          timestamp: Date.now()
        });
        break;
    }
  } catch (error) {
    console.error(`[线程 ${threadInfo}-${workerId}] 处理 GameApp 消息时出错:`, error);
    // 发送错误响应
    postMessage({
      cmd: 'error',
      error: error.message,
      originalCmd: data.cmd,
      workerID: workerId,
      originalMessageId: data._gameAppMessageId,
      timestamp: Date.now()
    });
  }
}

async function handleProjectDataUpdate(data) {
  const workerId = (typeof Module !== 'undefined' && Module['workerID']) || 'unknown';
  console.log(`[Worker ${workerId}] 处理项目数据更新，数据大小:`, data.data ? data.data.byteLength : 0);
  
  //await unpackGameData("", data.data, data.packName, data.packUrl)
  // 这里可以处理项目数据更新逻辑
  // 例如：更新本地缓存、通知 Go WASM 等
  
  if (! Module["FFI"]){
    console.log("==> Module[\"FFI\"] is not ready, wait for goWasmInitResult")
    return;
  }
  console.log("==> Module[\"FFI\"] is ready, call onProjectDataUpdate")

  if (self.goBridge && self.goBridge.isReady) {
    try {
      // 如果 Go WASM 已就绪，可以调用相关函数处理数据
      self.goBridge.callGoFunctionSafe('goLoadData', data.data);
    } catch (error) {
      console.error(`[Worker ${workerId}] 调用 Go 函数处理项目数据失败:`, error);
    }
  }

}

async function unpackGameData(dir, projectData, packName, packUrl) {
  const zip1 = new JSZip();
  const zip1Content = await zip1.loadAsync(projectData);
  let datas = []
  for (const [filePath, file] of Object.entries(zip1Content.files)) {
      const content = await file.async('arraybuffer');
      if (!file.dir) {
          console.log("unpackGameData ", filePath, content);
          datas.push({ "path": filePath, "data": content })
      }
  }
  // write project data to file
  datas.push({ "path": "spx_project_data.zip", "data": projectData.buffer })
  console.log("unpackGameData ", "spx_project_data",projectData, projectData.buffer);
  if (packUrl != ""){
      let pckBuffer = await (await fetch(packUrl)).arrayBuffer();
      datas.push({ "path": packName, "data": pckBuffer })
  }
  Module.unpackGameData(dir, datas)
}

function handleCustomCommand(data) {
  const workerId = (typeof Module !== 'undefined' && Module['workerID']) || 'unknown';
  console.log(`[Worker ${workerId}] 处理自定义命令:`, data);
  
  // 这里可以添加自定义命令的处理逻辑
  // 根据需要调用相应的函数或模块
  
  // 发送处理完成响应
  postMessage({
    cmd: 'customCommandComplete',
    workerID: workerId,
    originalMessageId: data._gameAppMessageId,
    result: `自定义命令 ${data.customCmd || 'unknown'} 处理完成`,
    timestamp: Date.now()
  });
}

self.onmessage = handleMessage;


