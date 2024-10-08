package engine

import (
	. "godot-ext/gdspx/pkg/engine"
)

func NewVec2(x, y float64) Vec2 {
	return Vec2{X: float32(x), Y: float32(y)}
}
