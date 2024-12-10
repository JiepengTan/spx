package engine

import (
	. "github.com/realdream-ai/mathf"
)

// --------------------------------------------------------------------------
// Override coordinate system-related functions to accommodate the
// difference between SPX and Godot coordinate systems (Y-axis inverted)

func (pself *SpriteProxy) SetTriggerRect(center Vec2, size Vec2) {
	center.Y = -center.Y
	pself.Sprite.SetTriggerRect(center, size)
}

func (pself *SpriteProxy) SetTriggerCapsule(center Vec2, size Vec2) {
	center.Y = -center.Y
	pself.Sprite.SetTriggerCapsule(center, size)
}

func (pself *SpriteProxy) SetTriggerCircle(center Vec2, radius float64) {
	center.Y = -center.Y
	pself.Sprite.SetTriggerCircle(center, radius)
}

func (pself *SpriteProxy) SetColliderRect(center Vec2, size Vec2) {
	center.Y = -center.Y
	pself.Sprite.SetColliderRect(center, size)
}

func (pself *SpriteProxy) SetColliderCapsule(center Vec2, size Vec2) {
	center.Y = -center.Y
	pself.Sprite.SetColliderCapsule(center, size)
}

func (pself *SpriteProxy) SetColliderCircle(center Vec2, radius float64) {
	center.Y = -center.Y
	pself.Sprite.SetColliderCircle(center, radius)
}

// ----------------- camera ---------------------
func (pself *cameraMgr) GetLocalPosition(pos Vec2) Vec2 {
	camPos := pself.GetCameraPosition()
	return pos.Sub(camPos)
}
func (pself *cameraMgr) GetPosition() Vec2 {
	pos := pself.cameraMgrImpl.GetCameraPosition()
	return NewVec2(pos.X, -pos.Y)
}

func (pself *cameraMgr) SetPosition(position Vec2) {
	pself.cameraMgrImpl.SetCameraPosition(NewVec2(position.X, -position.Y))
}
