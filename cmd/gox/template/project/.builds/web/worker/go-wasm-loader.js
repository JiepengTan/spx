
function handleGameAppMessage(data) {
  const workerId = (typeof Module !== 'undefined' && Module['workerID']) || 'unknown';
  const threadInfo = typeof importScripts !== 'undefined' ? 'Worker' : 'MainThread';
  try {
    switch (data.cmd) {
      case 'projectDataUpdate':
        handleProjectDataUpdate(data);
        break;
      default:
        console.warn(`[Thread ${threadInfo}-${workerId}] Unknown GameApp command:`, data.cmd);
        break;
    }
  } catch (error) {
    console.error(`[Thread ${threadInfo}-${workerId}] Error handling GameApp message:`, error);
  }
}

async function handleProjectDataUpdate(data) {
  Module["gameProjectData"] = data.data;
  self.tryRunGoWasm()
}

function tryRunGoWasm() {
  const workerId = (typeof Module !== 'undefined' && Module['workerID']) || 'unknown';
  if (!Module["FFI"]) {
    return;
  }
  if (!Module["gameProjectData"]) {
    return;
  }

  if (self.goBridge && self.goBridge.isReady) {
    try {
      // If Go WASM is ready, can call related functions to process data
      self.goBridge.callGoFunctionSafe('goLoadData', Module["gameProjectData"]);
    } catch (error) {
      console.error(`[Worker ${workerId}] Error calling Go function to process project data:`, error);
    }
  }
}

/**
 * Initializes Go WASM on demand (callable from any thread)
 * This function will be called on godot_js_spx_on_engine_start callback
 */
async function initializeGoWasmOnDemand() {
  const workerId = (typeof Module !== 'undefined' && Module['workerID']) || 'unknown';
  const threadInfo = typeof importScripts !== 'undefined' ? 'Worker' : 'MainThread';

  // If already initialized, return immediately
  if (self.goBridge && self.goBridge.isReady && typeof Module !== 'undefined' && Module['FFI']) {
    console.log(`[Thread ${threadInfo}-${workerId}] Go WASM is ready, no need to reinitialize`);
    return true;
  }
  try {
    // Load Go WASM module
    await loadGoWasmModule();
    return true;
  } catch (error) {
    console.error(`[Thread ${threadInfo}-${workerId}] Go WASM initialization failed:`, error);
    return false;
  }
}

// Expose functions to global scope for godot.editor.js to call
if (typeof self !== 'undefined') {
  self.initializeGoWasmOnDemand = initializeGoWasmOnDemand;
}


/**
 * Simplified Go WASM module loading logic
 */
async function loadGoWasmModule() {
  // If already loaded, return immediately
  if (self.goBridge && self.goBridge.isReady) {
    console.log(`[Godot Worker ${Module['workerID']}] Go WASM is already loaded, using directly`);
    return;
  }

  try {
    // Create Go WASM Bridge instance
    const goBridge = new GoWasmBridge();

    // Initialize Go WASM module
    await goBridge.initialize({
      wasmPath: './gdspx.wasm',
      timeout: 15000,
      enableDebug: false
    });

    // Try to call Go initialization function (optional)
    try {
      const initResult = await goBridge.callGoFunctionSafe('goWasmInit');
      Module['FFI'] = BindFFI(goBridge);
    } catch (goInitError) {
      console.warn(`[Godot Worker ${Module['workerID']}] Go initialization function call failed, but continuing execution:`, goInitError);
    }

    // Expose Go Bridge instance to global scope of current worker
    self.goBridge = goBridge;
  } catch (error) {
    console.error(`[Godot Worker ${Module['workerID']}] Go WASM module loading failed:`, error);
    throw error;
  }
}
