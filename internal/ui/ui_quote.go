package ui

import (
	"github.com/realdream-ai/mathf"

	"github.com/goplus/spx/internal/engine"
	. "github.com/goplus/spx/internal/engine"
)

type UiQuote struct {
	UiNode
	container *UiNode
	imageL    *UiNode
	imageR    *UiNode
	labelDes  *UiNode
	labelMsg  *UiNode
}

func NewUiQuote() *UiQuote {
	panel := engine.SyncCreateEngineUiNode[UiQuote]("")
	return panel
}

func (pself *UiQuote) OnStart() {
	pself.container = BindUI[UiNode](pself.GetId(), "C")
	pself.imageL = BindUI[UiNode](pself.GetId(), "C/ImageL")
	pself.imageR = BindUI[UiNode](pself.GetId(), "C/ImageR")
	pself.labelDes = BindUI[UiNode](pself.GetId(), "C/LabelDes")
	pself.labelMsg = BindUI[UiNode](pself.GetId(), "C/LabelMsg")
}

func (pself *UiQuote) SetText(pos mathf.Vec2, size mathf.Vec2, msg, description string) {
	pos = CameraMgr.GetLocalPosition(pos)
	UiMgr.SetGlobalPosition(pself.container.GetId(), WorldToUI(pos.Sub(mathf.NewVec2(size.X, -size.Y))))
	UiMgr.SetSize(pself.container.GetId(), size.Mulf(2))
	UiMgr.SetText(pself.labelMsg.GetId(), msg)
	UiMgr.SetText(pself.labelDes.GetId(), description)
}
