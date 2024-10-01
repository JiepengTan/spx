package engine

import (
	. "godot-ext/gdspx/pkg/engine"
)

func UpdateCameraPosition(x, y float64) {
	CameraMgr.SetCameraPosition(NewVec2(x, -y)) // TODO revert camera's y pos
}
