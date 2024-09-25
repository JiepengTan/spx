package engine

import (
	. "godot-ext/gdspx/pkg/engine"
)

//----------------------------------------

func SetWindowSize(width, height int) {
	println("SetWindowSize", width, height)
	PlatformMgr.SetWindowSize(int64(width), int64(height))
}

func SetWindowResizingMode(mode int) {
}

func SetRunnableOnUnfocused(flag bool) {

}
func SetFullscreen(flag bool) {
	PlatformMgr.SetWindowFullscreen(flag)
}
func SetWindowTitle(title string) {
	PlatformMgr.SetWindowTitle(title)
}
