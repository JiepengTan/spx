package ui

import (
	"github.com/realdream-ai/mathf"
	. "github.com/realdream-ai/mathf"

	"github.com/goplus/spx/internal/engine"
	. "github.com/goplus/spx/internal/engine"
)

type UiMonitor struct {
	UiNode
	bgAll          *UiNode
	valueOnly      *UiNode
	labelName      *UiNode
	labelBg        *UiNode
	labelValue     *UiNode
	labelValueOnly *UiNode
}
type UpdateFunc func(float64)

func NewUiMonitor() *UiMonitor {
	panel := engine.SyncCreateEngineUiNode[UiMonitor]("")
	return panel
}
func (pself *UiMonitor) OnStart() {
	pself.bgAll = BindUI[UiNode](pself.GetId(), "BG")
	pself.labelName = BindUI[UiNode](pself.GetId(), "BG/H/LabelName")
	pself.labelBg = BindUI[UiNode](pself.GetId(), "BG/H/C")
	pself.labelValue = BindUI[UiNode](pself.GetId(), "BG/H/C/H/LabelValue")

	pself.valueOnly = BindUI[UiNode](pself.GetId(), "ValueOnly")
	pself.labelValueOnly = BindUI[UiNode](pself.GetId(), "ValueOnly/LabelValue")

}
func (pself *UiMonitor) ShowAll(isOn bool) {
	UiMgr.SetVisible(pself.bgAll.GetId(), isOn)
	UiMgr.SetVisible(pself.valueOnly.GetId(), !isOn)
}

func (pself *UiMonitor) SetVisible(isOn bool) {
	UiMgr.SetVisible(pself.GetId(), isOn)
}

func (pself *UiMonitor) UpdateScale(x float64) {
	UiMgr.SetScale(pself.GetId(), mathf.NewVec2(x, x))
}
func (pself *UiMonitor) UpdatePos(x, y float64) {
	pos := WorldToUI(x, y)
	UiMgr.SetGlobalPosition(pself.GetId(), pos)
}

func (pself *UiMonitor) UpdateText(name, value string) {
	UiMgr.SetText(pself.labelName.GetId(), name)
	UiMgr.SetText(pself.labelValue.GetId(), value)
	UiMgr.SetText(pself.labelValueOnly.GetId(), value)
}
func (pself *UiMonitor) UpdateColor(color Color) {
	UiMgr.SetColor(pself.labelBg.GetId(), color)
	UiMgr.SetColor(pself.valueOnly.GetId(), color)
}
