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

func (pself *UiSay) SetText(winSize mathf.Vec2, pos mathf.Vec2, size mathf.Vec2, msg string) {
	camPos := CameraMgr.GetLocalPosition(pos)
	x, y := camPos.X, camPos.Y
	isLeft := x <= 0
	xPos := x
	yPos := y + size.Y/2
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
	maxYPos := winSize.Y/2 - uiHeight
	yPos = math.Max(-winSize.Y/2, math.Min(yPos, maxYPos))
	xPos = math.Max(-winSize.X/2, math.Min(x, winSize.X/2))

	UiMgr.SetVisible(pself.vboxL.GetId(), isLeft)
	UiMgr.SetVisible(pself.vboxR.GetId(), !isLeft)
	UiMgr.SetPosition(pself.GetId(), WorldToUI(mathf.NewVec2(xPos, yPos)))
	UiMgr.SetText(label, finalMsg)
}
