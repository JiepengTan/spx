package ui

import (
	"math"

	"github.com/realdream-ai/mathf"
	. "github.com/realdream-ai/mathf"

	"github.com/goplus/spx/internal/engine"
	. "github.com/goplus/spx/internal/engine"
)

type UiMeasure struct {
	UiNode
	container      *UiNode
	imageLine      *UiNode
	labelValue     *UiNode
	labelContainer *UiNode
}

func NewUiMeasure() *UiMeasure {
	panel := engine.SyncCreateEngineUiNode[UiMeasure]("")
	return panel
}

func (pself *UiMeasure) OnStart() {
	pself.container = BindUI[UiNode](pself.GetId(), "C")
	pself.imageLine = BindUI[UiNode](pself.GetId(), "C/Line")
	pself.labelContainer = BindUI[UiNode](pself.GetId(), "LC")
	pself.labelValue = BindUI[UiNode](pself.GetId(), "LC/Label")
}

func (pself *UiMeasure) UpdateInfo(x, y float64, length, heading float64, name string, color Color) {
	extraLen := 4.0 //hack for engine picture size
	length += extraLen
	rad := DegToRad(heading - 90)
	s, c := math.Sincos(float64(rad))
	halfX, halfY := (c * length / 2), (s * length / 2)
	pos := WorldToUI(x, y)
	labelPos := pos
	pos.X -= float64(halfX)
	pos.Y -= float64(halfY)
	UiMgr.SetGlobalPosition(pself.container.GetId(), pos)
	UiMgr.SetColor(pself.container.GetId(), color)
	UiMgr.SetSize(pself.container.GetId(), mathf.NewVec2(length+extraLen, 26))
	UiMgr.SetRotation(pself.container.GetId(), rad)

	UiMgr.SetGlobalPosition(pself.labelContainer.GetId(), labelPos)
	UiMgr.SetColor(pself.labelContainer.GetId(), color)
	UiMgr.SetText(pself.labelValue.GetId(), name)
}
