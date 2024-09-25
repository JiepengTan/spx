package engine

import (
	. "godot-ext/gdspx/pkg/engine"
	"godot-ext/gdspx/pkg/gdspx"
)

var game Gamer

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
	game.OnEngineStart()
}

func onUpdate(delta float32) {
	//pos := InputMgr.GetMousePos()
	//println("OnEngineUpdate", delta, int(pos.X), int(pos.Y))
	game.OnEngineUpdate(delta)
}

func onDestroy() {
	game.OnEngineDestroy()
}
