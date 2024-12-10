package ui

import (
	. "github.com/realdream-ai/mathf"

	. "github.com/goplus/spx/internal/engine"
	gdx "github.com/realdream-ai/gdspx/pkg/engine"
)

type UiNode struct {
	gdx.UiNode
}

// convert world space position to screen space
func WorldToUI(pos Vec2) Vec2 {
	pos.Y *= -1
	viewport := CameraMgr.GetViewportRect()
	return pos.Add(viewport.Size.Mulf(0.5)).Sub(viewport.Position)
}
