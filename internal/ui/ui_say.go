package ui

import (
	"math"
	"strings"

	"github.com/goplus/spx/internal/engine"
	. "github.com/goplus/spx/internal/engine"
	"github.com/goplus/spx/internal/text"
	"github.com/realdream-ai/mathf"
)

const (
	sayMsgSpliteWidth   = 25
	sayMsgLineHeight    = 26
	sayMsgDefaultHeight = 77
)

type UiSay struct {
	UiNode
	vboxL  *UiNode
	labelL *UiNode
	vboxR  *UiNode
	labelR *UiNode
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

func (pself *UiSay) SetText(winX, winY float64, x, y float64, w, h float64, msg string) {
	camPos := CameraMgr.GetLocalPosition(mathf.NewVec2(x, y))
	x, y = camPos.X, camPos.Y
	isLeft := x <= 0
	xPos := x
	yPos := y + h/2
	UiMgr.SetVisible(pself.vboxL.GetId(), isLeft)
	UiMgr.SetVisible(pself.vboxR.GetId(), !isLeft)
	label := pself.labelL.GetId()
	if !isLeft {
		label = pself.labelR.GetId()
	}
	hasNextLine := strings.ContainsRune(msg, '\n')
	finalMsg := msg
	if !hasNextLine {
		finalMsg = text.SplitLines(msg, sayMsgSpliteWidth)
	}
	lineCount := strings.Count(finalMsg, "\n")
	uiHeight := sayMsgDefaultHeight + float64(lineCount)*sayMsgLineHeight
	maxYPos := winY/2 - uiHeight
	yPos = math.Max(-winY/2, math.Min(yPos, maxYPos))
	xPos = math.Max(-winX/2, math.Min(x, winX/2))

	UiMgr.SetPosition(pself.GetId(), WorldToUI(xPos, yPos))
	UiMgr.SetText(label, finalMsg)
}
