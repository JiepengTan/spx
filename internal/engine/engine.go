package engine

import (
	gde "godot-ext/gdspx/pkg/engine"
	"godot-ext/gdspx/pkg/gdspx"
)

var game Gamer

type Gamer interface {
	RegisterEngineTypes()
	OnEngineStart()
	OnEngineUpdate(delta float32)
	OnEngineDestroy()
}

func GdspxMain(g Gamer) {
	game = g
	game.RegisterEngineTypes()
	gdspx.LinkEngine(gde.EngineCallbackInfo{
		OnEngineStart:   onStart,
		OnEngineUpdate:  onUpdate,
		OnEngineDestroy: onDestroy,
	})
}

func RegisterSpriteType[T any]() {
	gde.RegisterSpriteType[T]()
}

func SetWindowSize(width, height int) {
}

func SetWindowResizingMode(mode int) {
}

func SetRunnableOnUnfocused(flag bool) {

}
func SetFullscreen(flag bool) {

}
func SetWindowTitle(title string) {

}

func RunGame(p interface{}) error {
	return nil
}

func registerTypes() {

}
func onStart() {
	game.OnEngineStart()
}

func onUpdate(delta float32) {
	game.OnEngineUpdate(delta)
}

func onDestroy() {
	game.OnEngineDestroy()
}
