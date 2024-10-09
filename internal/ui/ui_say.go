package ui

import (
	. "godot-ext/gdspx/pkg/engine"

	"github.com/goplus/spx/internal/engine"
)

type UiSay struct {
	UiNode
	frame *UiNode
	label *UiNode
	image *UiNode
}

func NewUiSay(msg string) *UiSay {
	println("NewUiSay", msg)
	panel := engine.SyncCreateEngineUiNode[UiSay]("")
	return panel
}

func (pself *UiSay) OnStart() {
	pself.frame = BindUI[UiNode](pself.GetId(), "BG")
	pself.label = BindUI[UiNode](pself.GetId(), "BG/Label")
	pself.image = BindUI[UiNode](pself.GetId(), "ImageDot")
}

func (pself *UiSay) SetText(txt string) {
	engine.SyncUiSetText(pself.label.GetId(), txt)
}

func (pself *UiSay) refresh() {
}
