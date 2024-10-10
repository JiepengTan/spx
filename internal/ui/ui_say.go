package ui

import (
	. "godot-ext/gdspx/pkg/engine"

	"github.com/goplus/spx/internal/engine"
)

type UiSay struct {
	UiNode
	frame      *UiNode
	label      *UiNode
	imageL     *UiNode
	imageR     *UiNode
	WinX, WinY float64
}

func NewUiSay(msg string) *UiSay {
	panel := engine.SyncCreateEngineUiNode[UiSay]("")
	return panel
}

func (pself *UiSay) OnStart() {
	pself.frame = BindUI[UiNode](pself.GetId(), "V/BG")
	pself.label = BindUI[UiNode](pself.GetId(), "V/BG/Label")
	pself.imageL = BindUI[UiNode](pself.GetId(), "V/HL")
	pself.imageR = BindUI[UiNode](pself.GetId(), "V/HR")
}

func (pself *UiSay) SetText(x, y float64, txt string) {
	isLeft := x <= 0
	engine.SyncUiSetVisible(pself.imageL.GetId(), isLeft)
	engine.SyncUiSetVisible(pself.imageR.GetId(), !isLeft)
	rect := engine.NewRect2(x+pself.WinX/2, (-y + pself.WinY/2), 120, 10)
	engine.SyncUiSetText(pself.label.GetId(), txt)
	engine.SyncUiSetRect(pself.GetId(), rect)
}

func (pself *UiSay) refresh() {
}
