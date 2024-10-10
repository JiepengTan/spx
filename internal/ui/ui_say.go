package ui

import (
	. "godot-ext/gdspx/pkg/engine"

	"github.com/goplus/spx/internal/engine"
)

type UiSay struct {
	UiNode
	vboxL      *UiNode
	labelL     *UiNode
	vboxR      *UiNode
	labelR     *UiNode
	WinX, WinY float64
}

func NewUiSay() *UiSay {
	panel := engine.SyncCreateEngineUiNode[UiSay]("")
	return panel
}

func (pself *UiSay) OnStart() {
	pself.vboxL = BindUI[UiNode](pself.GetId(), "VL")
	pself.labelL = BindUI[UiNode](pself.GetId(), "VL/BG/Label")
	pself.vboxR = BindUI[UiNode](pself.GetId(), "VR")
	pself.labelR = BindUI[UiNode](pself.GetId(), "VR/BG/Label")
}

func (pself *UiSay) SetText(x, y float64, msg string) {
	isLeft := x <= 0
	engine.SyncUiSetVisible(pself.vboxL.GetId(), isLeft)
	engine.SyncUiSetVisible(pself.vboxR.GetId(), !isLeft)
	label := pself.labelL.GetId()
	if !isLeft {
		label = pself.labelR.GetId()
	}
	engine.SyncUiSetPosition(pself.GetId(), PosGame2UI(x, y+65))
	engine.SyncUiSetText(label, msg)
}
