package engine

import (
	. "godot-ext/gdspx/pkg/engine"
)

//----------------------------------------
func GetMousePos() (x, y int) {
	pos := InputMgr.GetMousePos()
	return int(pos.X), int(pos.Y)
}

func IsMousePressed() bool {
	return false
}

func IsKeyPressed(key int64) bool {
	return InputMgr.GetKey(key)
}

func IsStateKey(key int64) bool {
	return false
}
