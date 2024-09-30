package engine

import (
	. "godot-ext/gdspx/pkg/engine"
	"godot-ext/gdspx/pkg/gdspx"
	"sync"
)

var (
	game             Gamer
	tempTriggerPairs []TriggerPair
	TriggerPairs     []TriggerPair
	mu               sync.Mutex
)

type Gamer interface {
	OnEngineStart()
	OnEngineUpdate(delta float32)
	OnEngineDestroy()
}

func GdspxMain(g Gamer) {
	game = g
	gdspx.LinkEngine(EngineCallbackInfo{
		OnEngineStart:   onStart,
		OnEngineUpdate:  onUpdate,
		OnEngineDestroy: onDestroy,
	})
}

// callbacks
func onStart() {

	println("OnEngineStart")
	PlatformMgr.SetDebugMode(true)
	tempTriggerPairs = make([]TriggerPair, 0)
	TriggerPairs = make([]TriggerPair, 0)
	game.OnEngineStart()
}

func onUpdate(delta float32) {
	// cache trigger pairs
	cacheTriggerPairs()
	game.OnEngineUpdate(delta)
	handleEngineCoroutines()

}

func cacheTriggerPairs() {
	mu.Lock()
	TriggerPairs = append(TriggerPairs, tempTriggerPairs...)
	mu.Unlock()
	tempTriggerPairs = tempTriggerPairs[:0]
}

func GetTriggerPairs(lst []TriggerPair) []TriggerPair {
	mu.Lock()
	lst = append(lst, TriggerPairs...)
	if len(TriggerPairs) > 0 {
		println("TriggerPairs", len(TriggerPairs))
	}
	TriggerPairs = TriggerPairs[:0]
	mu.Unlock()
	return lst
}

func onDestroy() {
	game.OnEngineDestroy()
}

func NewVec2(x, y float64) Vec2 {
	return Vec2{X: float32(x), Y: float32(y)}
}
