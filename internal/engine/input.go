package engine

import (
	. "godot-ext/gdspx/pkg/engine"
)

//----------------------------------------
func GetMousePos() (x, y int) {
	pos := InputMgr.GetMousePos()
	return int(pos.X), int(pos.Y)
}
