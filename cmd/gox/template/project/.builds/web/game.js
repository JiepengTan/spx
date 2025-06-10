class GameApp {
    constructor(config) {
        config = config || {};
        this.config = config;
        this.editor = null;
        this.game = null;
        this.persistentPath = '/home/web_user';
        this.tempZipPath = '/tmp/preload.zip';
        this.packName =  'godot.editor.pck';
        this.projectDataName =  'project.data';
        this.isRuntimeMode = config.isRuntimeMode;
        this.tempGamePath = '/home/spx_game_cache';
        this.projectInstallName = config.projectName || "Game";
        this.logLevel = config.logLevel || 0;
        this.projectData = config.projectData;
        this.oldData = config.projectData;
        this.persistentPaths = [this.persistentPath];
        this.gameCanvas = config.gameCanvas;
        this.editorCanvas = config.editorCanvas || config.gameCanvas;
        this.exitFunc = null;
        this.basePath = 'godot.editor'
        this.isEditor = true;
        this.assetURLs = config.assetURLs;
        this.useAssetCache = config.useAssetCache;
        this.pthreads = null;
        
        // === Worker 管理相关属性 ===
        this.workerMessageHandlers = new Map(); // 存储消息处理器
        this.workerMessageId = 0; // 消息ID计数器
        
        this.editorConfig = {
            "executable": "godot.editor",
            'unloadAfterInit': false,
            'canvas': this.editorCanvas,
            'canvasResizePolicy': 0,
            "logLevel": this.logLevel,
            'persistentPaths': this.persistentPaths,
            'onExecute': (args) => {
                this.logVerbose("onExecute  ", args);
            },
            'onExit': () => {
                if (this.exitFunc) {
                    this.exitFunc();
                }
            }
        };
        this.gameConfig = {
            "executable": "godot.editor",
            'persistentPaths': this.persistentPaths,
            'unloadAfterInit': false,
            'canvas': this.gameCanvas,
            'logLevel': this.logLevel,
            'canvasResizePolicy': 1,
            'onExit': () => {
                this.onGameExit()
            },
            'ffi': config.ffi
        };
        this.logicPromise = Promise.resolve();
        this.curProjectHash = ''
    }
    
    logVerbose(...args) {
        if (this.logLevel == LOG_LEVEL_VERBOSE) {
            console.log(...args);
        }
    }
    startTask(prepareFunc, taskFunc, ...args) {
        if (prepareFunc != null) {
            prepareFunc()
        }
        this.logicPromise = this.logicPromise.then(async () => {
            let promise = new Promise(async (resolve, reject) => {
                await taskFunc.call(this, resolve, reject, ...args);
            })
            await promise
        })
        return this.logicPromise
    }

    async StartProject() {
        return this.startTask(null, this.startProject)
    }

    async UpdateProject(newData, addInfos, deleteInfos, updateInfos) {
        return this.startTask(null, this.updateProject, newData, addInfos, deleteInfos, updateInfos)
    }

    async StopProject() {
        return this.startTask(null, this.stopProject)
    }

    async RunGame() {
        return this.startTask(() => { this.runGameTask++ }, this.runGame)
    }

    async StopGame() {
        return this.startTask(() => { this.stopGameTask++ }, this.stopGame)
    }

    async startProject(resolve, reject) {
        if (this.editor != null) {
            console.error("project already loaded!")
        }
        this.isEditor = true
        if(this.isRuntimeMode){
            await this.checkEngineCache()
            resolve()
            return 
        }

        let url = this.assetURLs["engineres.zip"]
        let engineData = await (await fetch(url)).arrayBuffer();
      
        try {
            this.onProgress(0.1);
            this.clearPersistence(this.tempZipPath);
            let isCacheValid = await this.checkAndUpdateCache(engineData, true);
            await this.checkEngineCache()
            this.editor = new Engine(this.editorConfig);
            if (!isCacheValid) {
                this.exitFunc = () => {
                    this.exitFunc = null
                    this.editor = new Engine(this.editorConfig);
                    this.runEditor(resolve, reject)
                };
                // install project
                this.editor.init().then(async () => {
                    this.writePersistence(this.editor, this.tempZipPath, engineData);
                    const args = ['--project-manager', '--single-window', "--install_project_name", this.projectInstallName];
                    this.editor.start({
                        'args': args, 'persistentDrops': true,
                        "logLevel": this.logLevel
                    }).then(async () => {
                        this.editorCanvas.focus();
                    })
                });
            } else {
                this.logVerbose("cache is valid, skip it")
                resolve()
            }
        } catch (error) {
            console.error("Error checking database existence: ", error);
        }
    }

    async updateProject(resolve, reject, newData, addInfos, deleteInfos, updateInfos) {
        this.projectData = newData
        resolve()
    }

    async stopProject(resolve, reject) {
        if (!this.isRuntimeMode) {
            if (this.editor == null) {
                resolve()
                return
            }
        }
 
        this.stopGameTask++
        await this.stopGame(() => {
            this.isEditor = true
            this.onProgress(1.0);
            if (this.editor != null){
                this.editor.requestQuit()
            }
            this.logVerbose("on editor quit")
            this.editor = null
            this.exitFunc = null
            resolve();
        }, null)
    }

    runEditor(resolve, reject) {
        let args = [
            "--path",
            this.getInstallPath(),
            "--single-window",
            "--editor",
        ];
        this.exitFunc = null;
        this.logVerbose("runEditor ", args);
        this.onProgress(0.2);
        this.editor.init().then(() => {
            this.onProgress(0.4);
            this.editor.start({
                'args': args, 'persistentDrops': false,
                'canvas': this.editorCanvas,
                "logLevel": this.logLevel
            }).then(async () => {
                await this.waitFsSyncDone(this.editorCanvas)
                this.onProgress(0.9);
                await this.mergeProjectWithEngineRes()
                this.onProgress(1.0);
                this.editorCanvas.focus();
                await this.updateProjectHash(this.curProjectHash)
                this.logVerbose("==> editor start done")
                resolve()
            });
        });
    }

    async runGame(resolve, reject) {
        this.runGameTask--
        // if stopGame is called before runing game, then do nothing
        if (this.stopGameTask > 0) {
            this.logVerbose("stopGame is called before runing game")
            resolve()
            return
        }

        this.isEditor = false
        let args = []
        if (!this.isRuntimeMode){
            args = [
                "--path",
                this.getInstallPath(),
                "--editor-pid",
                "0",
                "res://main.tscn",
            ];
        }else{
            args = [ 
                '--main-pack', this.tempGamePath+ "/" + this.packName,
                '--main-project-data', this.tempGamePath+ "/" + this.projectDataName,
            ];
        }
           
        this.logVerbose("RunGame ", args);
        if (this.game) {
            this.logVerbose('A game is already running. Close it first');
            resolve()
            return;
        }
        this.onProgress(0.5);
        this.game = new Engine(this.gameConfig);
        let curGame = this.game
        curGame.init().then(async () => {
            this.onProgress(0.7);
            await this.unpackGameData(curGame, this.tempGamePath, this.projectData,this.packName, this.isRuntimeMode? this.assetURLs["godot.editor.pck"]:"" )
              
            curGame.start({ 'args': args, 'canvas': this.gameCanvas }).then(async () => {
                this.pthreads = curGame.getPThread()
                console.log("==> pthreads =", this.pthreads)
                // 等待2秒钟
                this.logVerbose("==> waited seconds after fs sync 1");
                await new Promise(resolve => setTimeout(resolve, 1000));
                this.broadcastProjectDataUpdate(this.projectData,this.packName, this.isRuntimeMode? this.assetURLs["godot.editor.pck"]:"")
                //await this.waitFsSyncDone(this.gameCanvas)
                //  this.onProgress(0.9);
                this.onProgress(1.0);
                this.gameCanvas.focus();
                this.logVerbose("==> game start done")
                resolve()
            });
        });
    }


    async unpackGameData(curGame,dir, projectData, pckName, packUrl) {
        let pckData = null;
        if (packUrl != ""){
            pckData = await (await fetch(packUrl)).arrayBuffer();
        }
        console.log("unpackGameData ", dir,this.projectDataName,  pckName);
        await curGame.unpackGameData(dir,this.projectDataName, projectData.buffer, pckName, pckData)
    }

    async stopGame(resolve, reject) {
        this.stopGameTask--
        if (this.game == null) {
            // no game is running, do nothing
            resolve()
            this.logVerbose("no game is running")
            return
        }
        this.stopGameResolve = () => {
            this.game = null
            resolve();
            this.stopGameResolve = null
        }
        this.pthreads = null
        this.isEditor = true
        this.onProgress(1.0);
        this.game.requestQuit()
    }

    onGameExit() {
        this.game = null
        this.logVerbose("on game quit")
        if (this.stopGameResolve) {
            this.stopGameResolve()
        }
    }

    //------------------ update project ------------------
    async waitFsSyncDone(canvas) {
        return new Promise((resolve, reject) => {
            this.logVerbose("waitFsSyncDone start")
            const evt = new CustomEvent('spx_wait_fs_sync_done', {
                detail: {
                    "resolve": async () => {
                        this.logVerbose("waitFsSyncDone done")
                        resolve()
                    },
                }
            });
            canvas.dispatchEvent(evt);
        })
    }

    //------------------ install project ------------------
    getInstallPath() {
        return `${this.persistentPath}/${this.projectInstallName}`;
    }

    writePersistence(engine, targetPath, value) {
        if (engine == null) {
            console.error("please init egnine first!")
            return
        }
        engine.copyToFS(targetPath, value);
    }
    clearPersistence(targetPath) {
        const req = indexedDB.deleteDatabase(targetPath);
        req.onerror = (err) => {
            alert('Error deleting local files. Please retry after reloading the page.');
        };
        this.logVerbose("clear persistence cache", targetPath);
    }

    getObjectStore(dbName, storeName, mode, storeKeyPath) {
        return new Promise((resolve, reject) => {
            let request = indexedDB.open(dbName);

            request.onupgradeneeded = function (event) {
                let db = event.target.result;
                if (!db.objectStoreNames.contains(storeName)) {
                    if (storeKeyPath) {
                        db.createObjectStore(storeName, { keyPath: storeKeyPath });
                    } else {
                        db.createObjectStore(storeName);
                    }

                }
            };

            request.onsuccess = function (event) {
                let db = event.target.result;
                if (!db.objectStoreNames.contains(storeName)) {
                    reject(`Object store "${storeName}" not found`);
                    db.close();
                    return;
                }

                let transaction = db.transaction(storeName, mode);
                let objectStore = transaction.objectStore(storeName);
                resolve({ db, objectStore, transaction });
            };

            request.onerror = function (event) {
                reject('Error opening database: ' + dbName + " " + storeName + " " + event.target.error);
            };

            request.onblocked = function (event) {
                reject('Database is blocked. Please close other tabs or windows using this database. ', dbName + " " + storeName + " " + event.target.error);
            }
        });
    }

    queryIndexDB(dbName, storeName, key) {
        return this.getObjectStore(dbName, storeName, 'readonly').then(({ db, objectStore, transaction }) => {
            return new Promise((resolve, reject) => {
                let getRequest = objectStore.get(key);

                getRequest.onsuccess = function () {
                    resolve(getRequest.result);
                };

                getRequest.onerror = function () {
                    reject('Error checking key existence');
                };

                transaction.oncomplete = function () {
                    db.close();
                };
            });
        });
    }

    updateIndexDB(dbName, storeName, key, value) {
        return this.getObjectStore(dbName, storeName, 'readwrite', key).then(({ db, objectStore, transaction }) => {
            return new Promise((resolve, reject) => {
                let putRequest = objectStore.put(value, key);

                putRequest.onsuccess = function () {
                    resolve('Value successfully written to the database');
                };

                putRequest.onerror = function () {
                    reject('Error writing value to the database');
                };

                transaction.oncomplete = function () {
                    db.close();
                };
            });
        });
    }
    async getCache(storeName) {
        try {
            let cacheValue = await this.queryIndexDB(this.persistentPath, 'FILE_DATA', storeName);
            return cacheValue;
        } catch (error) {
            console.error(error);
            return undefined;
        }
    }

    async setCache(storeName, value) {
        try {
            let cacheValue = await this.updateIndexDB(this.persistentPath, 'FILE_DATA', storeName, value);
            return cacheValue;
        } catch (error) {
            console.error(error);
            return undefined;
        }
    }

    async computeHash(data) {
        const hashBuffer = await crypto.subtle.digest('SHA-256', data);
        const hashArray = Array.from(new Uint8Array(hashBuffer));
        return hashArray.map(byte => byte.toString(16).padStart(2, '0')).join('');
    }
    getProjectDataKey() {
        return `${this.persistentPath}/.spx_cache_data/${this.projectInstallName}`
    }
    getProjectHashKey() {
        return `${this.persistentPath}/.spx_cache_hash/${this.projectInstallName}`
    }

    async updateProjectHash(hash) {
        this.logVerbose("updateProjectHash ", hash)
        await this.setCache(this.getProjectHashKey(), hash);
    }
    async checkAndUpdateCache(curData, isClearIfDirty = false) {
        // TODO only cache art resources
        let curHash = await this.computeHash(curData);
        let cachedHash = await this.getCache(this.getProjectHashKey());
        this.curProjectHash = curHash
        this.logVerbose("checkAndUpdateCache ", this.getProjectHashKey(), curHash, " old_hash = ", cachedHash)
        if (cachedHash != undefined && curHash === cachedHash) {
            return true;
        }
        if (isClearIfDirty) {
            await this.updateProjectHash('')
            // clear the dirty cache
            // TOOD only clear the current project's cache data
            this.clearPersistence(this.persistentPath);
            // create a default indexDB
            await this.ensureCacheDB()
        } else {
            await this.updateProjectHash(this.curProjectHash)
        }
        // cache is dirty, update it 
        await this.setCache(this.getProjectDataKey(), curData);
        return false;
    }

    async ensureCacheDB() {
        await this.getObjectStore(this.persistentPath, 'FILE_DATA', 'readonly')
    }

    getEngineHashKey(assetName) {
        return `${this.persistentPath}/.spx_engine_hash/${assetName}`
    }
    getEngineDataKey(assetName) {
        return `${this.persistentPath}/.spx_engine_data/${assetName}`
    }
    async checkEngineCache() {
        let hashes = GetEngineHashes()
        this.logVerbose("curHashes ", hashes)
        this.wasmEngine = await this.checkEngineCacheAsset(hashes, "godot.editor.wasm");
        this.editorConfig.wasmEngine = this.wasmEngine
        this.gameConfig.wasmEngine = this.wasmEngine
    }

    async checkEngineCacheAsset(hashes, assetName) {
        try {
            let url = this.assetURLs[assetName]
            if (!this.useAssetCache) {
                return await (await fetch(url)).arrayBuffer();
            }

            let curHash = hashes[assetName];
            await this.ensureCacheDB();

            const cachedHash = await this.getCache(this.getEngineHashKey(assetName));
            const isCacheValid = cachedHash !== undefined && curHash === cachedHash;

            if (!isCacheValid) {
                this.logVerbose("Download engine asset:", assetName, url);
                const curData = await (await fetch(url)).arrayBuffer();
                await this.setCache(this.getEngineDataKey(assetName), curData);
                await this.setCache(this.getEngineHashKey(assetName), curHash);

                return curData;
            } else {
                this.logVerbose("Load cached engine asset:", assetName);
                const curData = await this.getCache(this.getEngineDataKey(assetName));
                return curData;
            }
        } catch (error) {
            console.error("Error checking engine cache asset:", error);
            throw error;
        }
    }

    //------------------ res merge ------------------
    async mergeZips(zipFile1, zipFile2) {
        const zip1 = new JSZip();
        const zip2 = new JSZip();

        const zip1Content = await zip1.loadAsync(zipFile1);
        const zip2Content = await zip2.loadAsync(zipFile2);

        const newZip = new JSZip();

        for (const [filePath, file] of Object.entries(zip1Content.files)) {
            const content = await file.async('arraybuffer');
            newZip.file(filePath, content);
        }

        for (const [filePath, file] of Object.entries(zip2Content.files)) {
            const content = await file.async('arraybuffer');
            newZip.file(filePath, content);
        }

        return newZip.generateAsync({ type: 'arraybuffer' });
    }

    async mergeProjectWithEngineRes() {
        if (this.hasMergedProject) {
            return
        }
        this.logVerbose("merge zip files");
        const engineDataResp = fetch("engineres.zip");
        let engineData = await (await engineDataResp).arrayBuffer();
        this.projectData = await this.mergeZips(this.projectData, engineData);
        this.hasMergedProject = true
    }

    //------------------ misc ------------------
    onProgress(value) {
        if (this.config.onProgress != null) {
            this.config.onProgress(value);
        }
    }



    // === PThread Worker 消息发送相关方法 ===
    
    /**
     * 获取所有可用的 workers (包括运行中的和未使用的)
     */
    getAllWorkers() {
        const workers = [];
        console.log("==> getAllWorkers pthreads =", this.pthreads)
        // 获取 PThread 的 workers
        if (this.pthreads) {
            workers.push(...this.pthreads.runningWorkers);
            workers.push(...this.pthreads.unusedWorkers);
        }
        
        return workers;
    }
    
    /**
     * 获取正在运行的 workers
     */
    getRunningWorkers() {
        const workers = [];
        if (this.pthreads) {
            workers.push(...this.pthreads.runningWorkers);
        }
        return workers;
    }
    
    /**
     * 向所有 worker 发送消息
     * @param {Object} message - 要发送的消息对象
     * @param {Array} transferList - 可转移对象列表 (可选)
     * @param {boolean} onlyRunning - 是否只发送给正在运行的 worker (默认 true)
     * @param {boolean} cloneForEach - 是否为每个 worker 克隆消息数据 (默认 false)
     */
    postMessageToAllWorkers(message, transferList = null, onlyRunning = true, cloneForEach = false) {
        const workers = onlyRunning ? this.getRunningWorkers() : this.getAllWorkers();
        
        this.logVerbose(`向 ${workers.length} 个 worker 发送消息:`, message);
        
        let successCount = 0;
        let errorCount = 0;
        
        workers.forEach((worker, index) => {
            try {
                if (worker && typeof worker.postMessage === 'function') {
                    // 为每个消息添加唯一标识和目标信息
                    let enhancedMessage = {
                        ...message,
                        _gameAppMessageId: ++this.workerMessageId,
                        _targetWorkerIndex: index,
                        _timestamp: Date.now()
                    };
                    
                    // 如果需要克隆数据或者使用 transferList，需要特殊处理
                    if (transferList && cloneForEach) {
                        // 为每个 worker 创建数据副本
                        if (message.data && message.data.buffer) {
                            const clonedData = new Uint8Array(message.data);
                            enhancedMessage.data = clonedData;
                            worker.postMessage(enhancedMessage, [clonedData.buffer]);
                        } else {
                            worker.postMessage(enhancedMessage);
                        }
                    } else {
                        // 不使用 transferList 进行广播，让每个 worker 都能收到数据
                        worker.postMessage(enhancedMessage);
                    }
                    
                    successCount++;
                    this.logVerbose(`消息已发送到 worker ${index} (ID: ${worker.workerID || 'unknown'})`);
                } else {
                    console.warn(`Worker ${index} 无效或没有 postMessage 方法`);
                    errorCount++;
                }
            } catch (error) {
                console.error(`发送消息到 worker ${index} 失败:`, error);
                errorCount++;
            }
        });
        
        this.logVerbose(`消息发送完成: 成功 ${successCount}, 失败 ${errorCount}`);
        return { successCount, errorCount, totalWorkers: workers.length };
    }
    
    /**
     * 向特定 worker 发送消息
     * @param {number} workerIndex - worker 索引
     * @param {Object} message - 要发送的消息对象
     * @param {Array} transferList - 可转移对象列表 (可选)
     * @param {boolean} onlyRunning - 是否只在正在运行的 worker 中查找
     */
    postMessageToWorker(workerIndex, message, transferList = null, onlyRunning = true) {
        const workers = onlyRunning ? this.getRunningWorkers() : this.getAllWorkers();
        
        if (workerIndex < 0 || workerIndex >= workers.length) {
            console.error(`Worker 索引 ${workerIndex} 超出范围 (0-${workers.length - 1})`);
            return false;
        }
        
        const worker = workers[workerIndex];
        if (!worker || typeof worker.postMessage !== 'function') {
            console.error(`Worker ${workerIndex} 无效或没有 postMessage 方法`);
            return false;
        }
        
        try {
            const enhancedMessage = {
                ...message,
                _gameAppMessageId: ++this.workerMessageId,
                _targetWorkerIndex: workerIndex,
                _timestamp: Date.now()
            };
            
            if (transferList) {
                worker.postMessage(enhancedMessage, transferList);
            } else {
                worker.postMessage(enhancedMessage);
            }
            
            this.logVerbose(`消息已发送到 worker ${workerIndex} (ID: ${worker.workerID || 'unknown'}):`, message);
            return true;
        } catch (error) {
            console.error(`发送消息到 worker ${workerIndex} 失败:`, error);
            return false;
        }
    }
    
    /**
     * 获取 worker 信息列表
     */
    getWorkerInfo() {
        const runningWorkers = this.getRunningWorkers();
        const allWorkers = this.getAllWorkers();
        
        return {
            runningCount: runningWorkers.length,
            totalCount: allWorkers.length,
            runningWorkers: runningWorkers.map((worker, index) => ({
                index,
                workerID: worker.workerID || 'unknown',
                pthread_ptr: worker.pthread_ptr || 0,
                loaded: worker.loaded || false
            })),
            allWorkers: allWorkers.map((worker, index) => ({
                index,
                workerID: worker.workerID || 'unknown',
                pthread_ptr: worker.pthread_ptr || 0,
                loaded: worker.loaded || false,
                isRunning: runningWorkers.includes(worker)
            }))
        };
    }
    
    /**
     * 广播项目数据更新消息到所有 worker
     * @param {ArrayBuffer|Uint8Array} projectData - 项目数据
     */
    broadcastProjectDataUpdate(projectData, packName, packUrl) {
        const message = {
            cmd: 'projectDataUpdate',
            data: projectData,
            timestamp: Date.now()
        };
        
        // 广播到多个 worker 时不使用 transferList，避免 ArrayBuffer 被 detached
        // 数据会被克隆到每个 worker，虽然内存占用更多但确保所有 worker 都能收到
        return this.postMessageToAllWorkers(message, null, true, false);
    }
    
    /**
     * 向所有 worker 发送自定义命令
     * @param {string} cmd - 命令名称
     * @param {Object} data - 命令数据
     * @param {Array} transferList - 可转移对象列表
     */
    broadcastCustomCommand(cmd, data = {}, transferList = null) {
        const message = {
            cmd: cmd,
            ...data,
            timestamp: Date.now()
        };
        
        return this.postMessageToAllWorkers(message, transferList);
    }
    
    // === 便利方法 ===
    
    /**
     * Ping 所有 worker 以检查响应性
     */
    async pingAllWorkers(timeout = 5000) {
        const message = { cmd: 'ping' };
        const result = this.postMessageToAllWorkers(message);
        
        this.logVerbose(`Ping 发送到 ${result.totalWorkers} 个 worker`);
        return result;
    }
    
    /**
     * 获取所有 worker 的状态信息
     */
    async requestAllWorkerStatus(timeout = 5000) {
        const message = { cmd: 'getWorkerStatus' };
        const result = this.postMessageToAllWorkers(message);
        
        this.logVerbose(`状态请求发送到 ${result.totalWorkers} 个 worker`);
        return result;
    }
    
    /**
     * 初始化所有 worker 的 Go WASM 模块
     */
    async initGoWasmInAllWorkers(timeout = 15000) {
        const message = { cmd: 'initGoWasm' };
        const result = this.postMessageToAllWorkers(message);
        
        this.logVerbose(`Go WASM 初始化请求发送到 ${result.totalWorkers} 个 worker`);
        return result;
    }
    
    // === 示例用法方法 ===
    
    /**
     * 示例：发送项目数据到所有 worker
     */
    async syncProjectDataToWorkers() {
        if (!this.projectData) {
            console.warn('没有项目数据可同步');
            return;
        }
        
        this.logVerbose('开始同步项目数据到所有 worker...');
        const result = this.broadcastProjectDataUpdate(this.projectData);
        this.logVerbose('项目数据同步请求已发送:', result);
        return result;
    }
    
    /**
     * 示例：向所有 worker 发送自定义游戏命令
     */
    async sendGameCommand(command, params = {}) {
        const result = this.broadcastCustomCommand('gameCommand', {
            gameCommand: command,
            params: params
        });
        
        this.logVerbose(`游戏命令 "${command}" 已发送到所有 worker:`, result);
        return result;
    }
    
    /**
     * 示例：设置 worker 配置
     */
    async configureWorkers(config) {
        const result = this.broadcastCustomCommand('configure', {
            configuration: config
        });
        
        this.logVerbose('Worker 配置已发送:', result);
        return result;
    }

}

function GetEngineHashes() { 
	return {
"gdspx.wasm":"18f40edf2359eac08b41f2a23d65fc7c905b6c197821463c5f0ee4784f1c9604",
"godot.editor.wasm":"47897986d2212982e1c3ebfca7c8ab1a229cdcb23fe70fe8a75dfc60aea957d1",

	}
}
	