package engine

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"

	stime "time"

	"github.com/goplus/spx/internal/enginewrap"
	"github.com/goplus/spx/internal/time"
	gdx "github.com/realdream-ai/gdspx/pkg/engine"
	gde "github.com/realdream-ai/gdspx/pkg/gdspx"
)

// copy these variable to any namespace you want
var (
	audioMgr    enginewrap.AudioMgrImpl
	cameraMgr   enginewrap.CameraMgrImpl
	inputMgr    enginewrap.InputMgrImpl
	physicMgr   enginewrap.PhysicMgrImpl
	platformMgr enginewrap.PlatformMgrImpl
	resMgr      enginewrap.ResMgrImpl
	sceneMgr    enginewrap.SceneMgrImpl
	spriteMgr   enginewrap.SpriteMgrImpl
	uiMgr       enginewrap.UiMgrImpl
)

type TriggerEvent struct {
	Src *Sprite
	Dst *Sprite
}
type KeyEvent struct {
	Id        int64
	IsPressed bool
}

var (
	game              IGame
	triggerEventsTemp []TriggerEvent
	triggerEvents     []TriggerEvent
	triggerMutex      sync.Mutex

	keyEventsTemp []KeyEvent
	keyEvents     []KeyEvent
	keyMutex      sync.Mutex

	// time
	startTimestamp     stime.Time
	lastTimestamp      stime.Time
	timeSinceLevelLoad float64

	// statistic info
	fps float64
)
var (
	debugLastTime  float64 = 0
	debugLastFrame int64   = 0
)

type IGame interface {
	OnEngineStart()
	OnEngineUpdate(delta float64)
	OnEngineRender(delta float64)
	OnEngineDestroy()
}

func Main(g IGame) {
	enginewrap.Init(WaitMainThread)
	game = g
	gde.LinkEngine(gdx.EngineCallbackInfo{
		OnEngineStart:   onStart,
		OnEngineUpdate:  onUpdate,
		OnEngineDestroy: onDestroy,
		OnKeyPressed:    onKeyPressed,
		OnKeyReleased:   onKeyReleased,
	})
}

func OnGameStarted() {
	gco.OnInited()
}

// callbacks
func onStart() {
	runtime.GOMAXPROCS(1)
	triggerEventsTemp = make([]TriggerEvent, 0)
	triggerEvents = make([]TriggerEvent, 0)
	keyEventsTemp = make([]KeyEvent, 0)
	keyEvents = make([]KeyEvent, 0)

	time.Start(func(scale float64) {
		platformMgr.SetTimeScale(scale)
	})

	startTimestamp = stime.Now()
	lastTimestamp = stime.Now()
	game.OnEngineStart()
}

// TimingInfo 记录各个模块的执行时间和相关信息
type TimingInfo struct {
	PreCall    float64       // 调用前的准备时间
	ActualCall float64       // 实际函数执行时间
	PostCall   float64       // 调用后的清理时间
	GCStats    debug.GCStats // GC统计信息
}

// 全局变量，用于存储性能分析数据
var (
	prevGCStats debug.GCStats
	timingData  = make(map[string]TimingInfo)

	// 控制是否打印详细性能统计信息
	printDetailedStats bool
)

// 获取GC统计信息的差异
func getGCDiff() debug.GCStats {
	var currentStats debug.GCStats
	debug.ReadGCStats(&currentStats)

	diff := debug.GCStats{
		NumGC:          currentStats.NumGC - prevGCStats.NumGC,
		PauseTotal:     currentStats.PauseTotal - prevGCStats.PauseTotal,
		PauseQuantiles: currentStats.PauseQuantiles,
		PauseEnd:       currentStats.PauseEnd,
	}

	// 只复制最近的暂停时间
	if len(currentStats.Pause) > 0 && len(prevGCStats.Pause) > 0 {
		diff.Pause = make([]stime.Duration, len(currentStats.Pause))
		copy(diff.Pause, currentStats.Pause)
	}

	prevGCStats = currentStats
	return diff
}

// 测量函数执行时间，包括前后的准备和清理时间
func measureFunctionTime(name string, fn func()) {
	// 记录调用前的时间
	preCallStart := stime.Now()

	// 获取调用前的GC统计信息
	debug.ReadGCStats(&prevGCStats)

	// 准备阶段结束
	preCallEnd := stime.Now()

	// 执行实际函数
	actualCallStart := stime.Now()
	fn()
	actualCallEnd := stime.Now()

	// 执行后处理开始
	postCallStart := stime.Now()

	// 获取GC差异
	gcDiff := getGCDiff()

	// 后处理结束
	postCallEnd := stime.Now()

	// 计算各阶段时间（毫秒）
	preCallTime := preCallEnd.Sub(preCallStart).Seconds() * 1000
	actualCallTime := actualCallEnd.Sub(actualCallStart).Seconds() * 1000
	postCallTime := postCallEnd.Sub(postCallStart).Seconds() * 1000

	// 存储计时信息
	timingData[name] = TimingInfo{
		PreCall:    preCallTime,
		ActualCall: actualCallTime,
		PostCall:   postCallTime,
		GCStats:    gcDiff,
	}
}

// PrintTimingInfo 打印计时信息
func PrintTimingInfo() {
	fmt.Println("========== 引擎模块详细计时信息 ==========")
	for name, info := range timingData {
		total := info.PreCall + info.ActualCall + info.PostCall
		fmt.Printf("%s: 总计 %.3fms (准备: %.3fms, 执行: %.3fms, 清理: %.3fms)\n",
			name, total, info.PreCall, info.ActualCall, info.PostCall)

		if info.GCStats.NumGC > 0 {
			fmt.Printf("  GC: %d次, 总暂停: %.3fms\n",
				info.GCStats.NumGC,
				float64(info.GCStats.PauseTotal)/float64(stime.Millisecond))
		}
	}

	// 如果是协程更新，还要打印协程内部的详细统计
	if info, ok := timingData["CoroUpdateJobs"]; ok {
		PrintCoroStats(info)
	}

	fmt.Println("====================================")
}

// PrintCoroStats 打印协程模块的详细统计信息
func PrintCoroStats(coroInfo TimingInfo) {
	// 如果gco为空，则返回
	if gco == nil {
		fmt.Println("协程系统未初始化")
		return
	}

	// 导入协程模块的统计信息
	stats := gco.GetLastUpdateStats()

	// 计算引擎测量的总时间与协程内部测量的时间差异
	coroInfo, ok := timingData["CoroUpdateJobs"]
	if !ok {
		return
	}

	// 计算引擎测量的总时间
	engineMeasuredTotal := coroInfo.PreCall + coroInfo.ActualCall + coroInfo.PostCall

	// 协程内部测量的总时间可能包含两种数据：
	// 1. 如果有TotalTime字段，使用它作为协程内部测量的总时间
	// 2. 否则，使用各部分之和
	var coroInternalTotal float64
	var coroInternalParts float64

	if stats.TotalTime > 0 {
		coroInternalTotal = stats.TotalTime
		coroInternalParts = stats.InitTime + stats.LoopTime + stats.MoveTime
	} else {
		coroInternalTotal = stats.InitTime + stats.LoopTime + stats.MoveTime
		coroInternalParts = coroInternalTotal
	}

	// 计算差异
	difference := engineMeasuredTotal - coroInternalTotal

	// 计算协程内部差异
	coroInternalDifference := 0.0
	if stats.TotalTime > 0 {
		coroInternalDifference = stats.TotalTime - coroInternalParts
	}

	fmt.Println("\n========== 协程模块详细统计 ==========")
	fmt.Printf("引擎测量总时间: %.3fms (准备: %.3fms, 执行: %.3fms, 清理: %.3fms 上次总时间 %.3fms)\n",
		engineMeasuredTotal, coroInfo.PreCall, coroInfo.ActualCall, coroInfo.PostCall, lastUpdateDuration)
	fmt.Printf("协程内部测量总时间: %.3fms\n", coroInternalTotal)
	fmt.Printf("时间差异: %.3fms (%.2f%%)\n",
		difference, (difference/engineMeasuredTotal)*100)

	// 如果有内部差异，显示它
	if stats.TotalTime > 0 && coroInternalDifference > 0.1 {
		fmt.Printf("协程内部差异: %.3fms (%.2f%%)\n",
			coroInternalDifference, (coroInternalDifference/stats.TotalTime)*100)
	}

	fmt.Println("\n协程内部详细时间分布:")
	fmt.Printf("  初始化: %.3fms (%.2f%%)\n",
		stats.InitTime, (stats.InitTime/coroInternalTotal)*100)
	fmt.Printf("  主循环: %.3fms (%.2f%%)\n",
		stats.LoopTime, (stats.LoopTime/coroInternalTotal)*100)

	if stats.LoopIterations > 0 {
		fmt.Printf("    - 循环迭代次数: %d\n", stats.LoopIterations)
	}

	fmt.Printf("    - 任务处理: %.3fms (任务数: %d)\n",
		stats.TaskProcessing, stats.TaskCounts)
	fmt.Printf("    - 等待时间: %.3fms\n", stats.WaitTime)
	fmt.Printf("  队列移动: %.3fms (%.2f%%, 下一帧任务: %d)\n",
		stats.MoveTime, (stats.MoveTime/coroInternalTotal)*100, stats.NextCount)

	if stats.ExternalTime > 0 {
		fmt.Printf("  外部时间: %.3fms (%.2f%%)\n",
			stats.ExternalTime, (stats.ExternalTime/coroInternalTotal)*100)
	}

	if stats.GCCount > 0 {
		fmt.Printf("  GC: %d次, 暂停: %.3fms\n", stats.GCCount, stats.GCPauses)
	}

	// 分析可能的原因
	if difference > 5 { // 如果差异大于5毫秒
		fmt.Println("\n可能的性能差异原因:")

		// 检查GC
		if coroInfo.GCStats.NumGC > 0 {
			fmt.Printf("  - 垃圾回收: 引擎测量期间发生了%d次GC，总暂停时间%.3fms\n",
				coroInfo.GCStats.NumGC,
				float64(coroInfo.GCStats.PauseTotal)/float64(stime.Millisecond))
		}

		// 检查协程内部GC
		if stats.GCCount > 0 {
			fmt.Printf("  - 协程内部GC: 协程测量期间发生了%d次GC，总暂停时间%.3fms\n",
				stats.GCCount, stats.GCPauses)
		}

		// 检查函数调用开销
		if coroInfo.PreCall > 1 || coroInfo.PostCall > 1 {
			fmt.Printf("  - 函数调用开销: 准备阶段%.3fms，清理阶段%.3fms\n",
				coroInfo.PreCall, coroInfo.PostCall)
		}

		// 检查循环迭代次数
		if stats.LoopIterations > 100 {
			fmt.Printf("  - 循环迭代次数过多: %d次\n", stats.LoopIterations)
		}

		// 检查协程内部是否有未计时的部分
		unaccountedTime := stats.LoopTime - (stats.TaskProcessing + stats.WaitTime)
		if unaccountedTime > 5 {
			fmt.Printf("  - 协程内部有未计时部分: 约%.3fms\n", unaccountedTime)
		}

		// 检查外部时间
		if stats.ExternalTime > 5 {
			fmt.Printf("  - 协程外部时间过长: %.3fms\n", stats.ExternalTime)
		}

		// 检查Go运行时调度开销
		fmt.Println("  - Go运行时调度: 可能存在goroutine调度延迟")
		fmt.Println("  - 外部因素: 可能存在其他程序或系统资源竞争")
	}
}

var lastUpdateDuration float64

func onUpdate(delta float64) {
	totalStart := stime.Now()

	// 清除上一帧的计时数据
	timingData = make(map[string]TimingInfo)

	updateTime(float64(delta))
	cacheTriggerEvents()
	cacheKeyEvents()

	// 测量游戏更新时间
	measureFunctionTime("GameUpdate", func() {
		game.OnEngineUpdate(delta)
	})

	// 测量协程更新时间
	measureFunctionTime("CoroUpdateJobs", func() {
		gco.UpdateJobs()
	})

	// 测量游戏渲染时间
	measureFunctionTime("GameRender", func() {
		game.OnEngineRender(delta)
	})

	// 计算总时间
	total := stime.Since(totalStart).Seconds() * 1000

	// 打印简要信息
	if total > 50 {
		fmt.Printf("总时间: %.3fms (GameUpdate: %.3fms, CoroUpdateJobs: %.3fms, GameRender: %.3fms)\n",
			total, timingData["GameUpdate"].ActualCall, timingData["CoroUpdateJobs"].ActualCall, timingData["GameRender"].ActualCall)
	}

	// 如果协程更新时间超过阈值，或者前后处理时间过长，打印详细信息
	if printDetailedStats ||
		timingData["CoroUpdateJobs"].ActualCall > 10 ||
		(timingData["CoroUpdateJobs"].PreCall+timingData["CoroUpdateJobs"].PostCall) > 5 ||
		(gco != nil && timingData["CoroUpdateJobs"].ActualCall > 10 &&
			timingData["CoroUpdateJobs"].ActualCall > 2*getCoroStatsTotal()) {
		PrintTimingInfo()

		// 如果启用了详细统计信息打印，则打印协程统计信息
		if printDetailedStats && gco != nil {
			coroInfo := TimingInfo{
				PreCall:    timingData["CoroUpdateJobs"].PreCall,
				ActualCall: timingData["CoroUpdateJobs"].ActualCall,
				PostCall:   timingData["CoroUpdateJobs"].PostCall,
				GCStats:    timingData["CoroUpdateJobs"].GCStats,
			}
			PrintCoroStats(coroInfo)
		}
	}

	lastUpdateDuration = total
}

// TogglePrintDetailedStats 切换是否打印详细性能统计信息
func TogglePrintDetailedStats() {
	printDetailedStats = !printDetailedStats
	fmt.Printf("\n详细性能统计信息打印已%s\n", map[bool]string{true: "开启", false: "关闭"}[printDetailedStats])
}

// getCoroStatsTotal 返回协程内部统计的总时间
func getCoroStatsTotal() float64 {
	if gco == nil {
		return 0
	}

	stats := gco.GetLastUpdateStats()
	return stats.InitTime + stats.LoopTime + stats.MoveTime
}

func onDestroy() {
	game.OnEngineDestroy()
}

func onKeyPressed(id int64) {
	// F12键用于切换详细性能统计信息的显示
	if id == gdx.KeyCode.F12 { // F12键的键值
		TogglePrintDetailedStats()
	}

	keyEventsTemp = append(keyEventsTemp, KeyEvent{Id: id, IsPressed: true})
}
func onKeyReleased(id int64) {
	keyEventsTemp = append(keyEventsTemp, KeyEvent{Id: id, IsPressed: false})
}

func calcfps() {
	curTime := time.RealTimeSinceStart()
	timeDiff := curTime - debugLastTime
	frameDiff := time.Frame() - debugLastFrame
	if timeDiff > 0.25 {
		fps = float64(frameDiff) / timeDiff
		debugLastFrame = time.Frame()
		debugLastTime = curTime
	}
}

func updateTime(delta float64) {
	deltaTime := delta
	timeSinceLevelLoad += deltaTime

	curTime := stime.Now()
	unscaledTimeSinceLevelLoad := curTime.Sub(startTimestamp).Seconds()
	unscaledDeltaTime := curTime.Sub(lastTimestamp).Seconds()
	lastTimestamp = curTime
	timeScale := SyncGetTimeScale()
	calcfps()
	time.Update(float64(timeScale), unscaledTimeSinceLevelLoad, timeSinceLevelLoad, deltaTime, unscaledDeltaTime, fps)
}

func cacheTriggerEvents() {
	triggerMutex.Lock()
	triggerEvents = append(triggerEvents, triggerEventsTemp...)
	triggerMutex.Unlock()
	triggerEventsTemp = triggerEventsTemp[:0]
}
func GetTriggerEvents(lst []TriggerEvent) []TriggerEvent {
	triggerMutex.Lock()
	lst = append(lst, triggerEvents...)
	triggerEvents = triggerEvents[:0]
	triggerMutex.Unlock()
	return lst
}
func cacheKeyEvents() {
	keyMutex.Lock()
	keyEvents = append(keyEvents, keyEventsTemp...)
	keyMutex.Unlock()
	keyEventsTemp = keyEventsTemp[:0]
}

func GetKeyEvents(lst []KeyEvent) []KeyEvent {
	keyMutex.Lock()
	lst = append(lst, keyEvents...)
	keyEvents = keyEvents[:0]
	keyMutex.Unlock()
	return lst
}
