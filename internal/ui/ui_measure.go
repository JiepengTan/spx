package ui

import (
	. "godot-ext/gdspx/pkg/engine"

	"github.com/goplus/spx/internal/engine"
)

type UiMeasure struct {
	UiNode
	container  *UiNode
	imageLine  *UiNode
	labelValue *UiNode
}

func NewUiMeasure() *UiMeasure {
	panel := CreateEngineUI[UiMeasure]("")
	return panel
}

func (pself *UiMeasure) OnStart() {
	pself.container = BindUI[UiNode](pself.GetId(), "C")
	pself.labelValue = BindUI[UiNode](pself.GetId(), "C/Label")
	pself.imageLine = BindUI[UiNode](pself.GetId(), "C/Line")
}

func (pself *UiMeasure) UpdateInfo(x, y float64, length, rot float64, name string, color Color) {
	pos := PosGame2UI(x, y)
	println(int(x), int(y), "length:", int(length), "pos", pos.String(), "rot ", int(rot))
	pself.SetGlobalPosition(pos)
	pself.container.SetSize(engine.NewVec2(length, 26))
	pself.container.SetRotation(DegToRad(float32(rot)))
	pself.labelValue.SetText(name)
}
