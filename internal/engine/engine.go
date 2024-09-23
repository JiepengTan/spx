package engine

import (
	gde "godot-ext/gdspx/pkg/engine"
	"godot-ext/gdspx/pkg/gdspx"
)

func GdspxMain() {
	registerTypes()
	gdspx.LinkEngine(gde.EngineCallbackInfo{
		OnEngineStart:   onStart,
		OnEngineUpdate:  onUpdate,
		OnEngineDestroy: onDestroy,
	})
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
	print("onStart")
}

func onUpdate(delta float32) {
}

func onDestroy() {
	print("onStart")
}
