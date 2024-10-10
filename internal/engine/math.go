package engine

import (
	. "godot-ext/gdspx/pkg/engine"
)

func NewVec2(x, y float64) Vec2 {
	return Vec2{X: float32(x), Y: float32(y)}
}
func NewRect2(x, y, width, hegiht float64) Rect2 {
	return Rect2{NewVec2(x, y), NewVec2(width, hegiht)}
}
func Clamp01d(val float64) float64 {
	if val < 0 {
		val = 0
	}
	if val > 1 {
		val = 1
	}
	return val
}
func Clamp01(val float32) float32 {
	if val < 0 {
		val = 0
	}
	if val > 1 {
		val = 1
	}
	return val
}
func Clamp01i(val int64) int64 {
	if val < 0 {
		val = 0
	}
	if val > 1 {
		val = 1
	}
	return val
}
