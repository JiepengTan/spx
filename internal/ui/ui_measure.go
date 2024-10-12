package ui

import (
	. "godot-ext/gdspx/pkg/engine"
	"math"

	"github.com/goplus/spx/internal/engine"
)

type UiMeasure struct {
	UiNode
	container      *UiNode
	imageLine      *UiNode
	labelValue     *UiNode
	labelContainer *UiNode
}

func NewUiMeasure() *UiMeasure {
	panel := CreateEngineUI[UiMeasure]("")
	return panel
}

func (pself *UiMeasure) OnStart() {
	pself.container = BindUI[UiNode](pself.GetId(), "C")
	pself.imageLine = BindUI[UiNode](pself.GetId(), "C/Line")
	pself.labelContainer = BindUI[UiNode](pself.GetId(), "LC")
	pself.labelValue = BindUI[UiNode](pself.GetId(), "LC/Label")
}

func (pself *UiMeasure) UpdateInfo(x, y float64, length, rot float64, name string, color Color) {
	extraLen := 4.0 //hack for engine picture size
	length += extraLen
	rad := DegToRad(float32(rot))
	s, c := math.Sincos(float64(rad))
	halfX, halfY := (c * length / 2), (s * length / 2)
	pos := PosGame2UI(x, y)
	labelPos := pos
	pos.X -= float32(halfX)
	pos.Y -= float32(halfY)
	pself.container.SetGlobalPosition(pos)
	pself.container.SetColor(color)
	pself.container.SetSize(engine.NewVec2(length+extraLen, 26))
	pself.container.SetRotation(rad)

	pself.labelContainer.SetGlobalPosition(labelPos)
	pself.labelContainer.SetColor(color)
	pself.labelValue.SetText(name)
}
