package ui

import (
	. "godot-ext/gdspx/pkg/engine"
)

type UiMeasure struct {
	UiNode
	frame *UiNode
	label *UiNode
	image *UiNode
}

func (pself *UiMeasure) OnStart() {
	pself.frame = BindUI[UiNode](pself.GetId(), "BG")
	pself.label = BindUI[UiNode](pself.GetId(), "BG/Label")
	pself.image = BindUI[UiNode](pself.GetId(), "ImageDot")
}

func (pself *UiMeasure) SetText(txt string) {
	pself.label.SetText(txt)
}

func (pself *UiMeasure) refresh() {
}
