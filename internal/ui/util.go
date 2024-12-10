package ui

import (
	"github.com/realdream-ai/mathf"
	. "github.com/realdream-ai/mathf"

	. "github.com/goplus/spx/internal/engine"
	gdx "github.com/realdream-ai/gdspx/pkg/engine"
)

type UiNode struct {
	gdx.UiNode
}

// convert world space position to screen space
func WorldToUI(x, y float64) Vec2 {
	viewport := CameraMgr.GetViewportRect()
	winX := float64(viewport.Size.X)
	winY := float64(viewport.Size.Y)
	return mathf.NewVec2(x+winX/2-float64(viewport.Position.X), (-y+winY/2)-float64(viewport.Position.Y))
}
