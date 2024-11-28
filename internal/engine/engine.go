package engine

import (
	"sync"

	stime "time"

	"github.com/goplus/spx/internal/time"
	. "github.com/realdream-ai/gdspx/pkg/engine"
	"github.com/realdream-ai/gdspx/pkg/gdspx"
)

type TriggerEvent struct {
	Src *ProxySprite
	Dst *ProxySprite
}
type KeyEvent struct {
	Id        int64
	IsPressed bool
}

var (
	game              Gamer
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
	debugTimer     float64 = 0
	debugLastFrame int64   = 0
)

type Gamer interface {
	OnEngineStart()
	OnEngineUpdate(delta float32)
	OnEngineRender(delta float32)
	OnEngineDestroy()
}

func GdspxMain(g Gamer) {
	game = g
	gdspx.LinkEngine(EngineCallbackInfo{
		OnEngineStart:   onStart,
		OnEngineUpdate:  onUpdate,
		OnEngineDestroy: onDestroy,
		OnKeyPressed:    onKeyPressed,
		OnKeyReleased:   onKeyReleased,
	})
}

// callbacks
func onStart() {
	triggerEventsTemp = make([]TriggerEvent, 0)
	triggerEvents = make([]TriggerEvent, 0)
	keyEventsTemp = make([]KeyEvent, 0)
	keyEvents = make([]KeyEvent, 0)

	time.Start(onSetTimeScale)
	startTimestamp = stime.Now()
	lastTimestamp = stime.Now()
	game.OnEngineStart()
}

func onUpdate(delta float32) {
	updateTime(float64(delta))
	cacheTriggerEvents()
	cacheKeyEvents()
	game.OnEngineUpdate(delta)
	gco.HandleJobs()
	game.OnEngineRender(delta)
}

func calcfps() {
	timer := time.RealTimeSinceStart()
	timeDiff := timer - debugTimer
	frameDiff := float64(time.Frame() - debugLastFrame)
	if timeDiff > 0.25 {
		fps = frameDiff / timeDiff
		debugLastFrame = time.Frame()
		debugTimer = timer
	}
}
func onSetTimeScale(scale float64) {
	SyncPlatformSetTimeScale(float32(scale))
}

func updateTime(delta float64) {
	deltaTime := delta
	timeSinceLevelLoad += deltaTime

	curTime := stime.Now()
	unscaledTimeSinceLevelLoad := curTime.Sub(startTimestamp).Seconds()
	unscaledDeltaTime := curTime.Sub(lastTimestamp).Seconds()
	lastTimestamp = curTime
	timeScale := PlatformMgr.GetTimeScale()
	calcfps()
	time.Update(float64(timeScale), unscaledTimeSinceLevelLoad, timeSinceLevelLoad, deltaTime, unscaledDeltaTime, fps)
}

func onDestroy() {
	game.OnEngineDestroy()
}

func onKeyPressed(id int64) {
	keyEventsTemp = append(keyEventsTemp, KeyEvent{Id: id, IsPressed: true})
}
func onKeyReleased(id int64) {
	keyEventsTemp = append(keyEventsTemp, KeyEvent{Id: id, IsPressed: false})
}

func cacheTriggerEvents() {
	triggerMutex.Lock()
	triggerEvents = append(triggerEvents, triggerEventsTemp...)
	triggerMutex.Unlock()
	triggerEventsTemp = triggerEventsTemp[:0]
}
func IsWebIntepreterMode() bool {
	return gdspx.IsWebIntepreterMode()
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

func GetFPS() float64 {
	return fps
}
func GetTPS() float64 {
	return 30
}
