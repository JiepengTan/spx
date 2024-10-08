/*
 * Copyright (c) 2021 The GoPlus Authors (goplus.org). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package spx

import (
	"fmt"

	"golang.org/x/image/font"

	"github.com/goplus/spx/internal/gdi"
	xfont "github.com/goplus/spx/internal/gdi/font"
)

var (
	defaultFont   gdi.Font
	defaultFont2  gdi.Font
	defaultFontSm gdi.Font
)

func init() {
	const dpi = 72
	defaultFont = xfont.NewDefault(&xfont.Options{
		Size:    15,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	defaultFont2 = xfont.NewDefault(&xfont.Options{ // for stageMonitor
		Size:    12,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	defaultFontSm = xfont.NewDefault(&xfont.Options{
		Size:    11,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}

// -------------------------------------------------------------------------------------

const (
	styleSay   = 1
	styleThink = 2
)

type sayOrThinker struct {
	sp    *Sprite
	msg   string
	style int // styleSay, styleThink
}

const (
	sayCornerSize = 8
	thinkRadius   = 5
	screenGap     = 4
	leadingWidth  = 15
	gapWidth      = 40
	trackDx       = 5
	trackCx       = gapWidth + trackDx
	trackCy       = 17
	minWidth      = leadingWidth + leadingWidth + gapWidth
)

func (p *sayOrThinker) hit(hc hitContext) (hr hitResult, ok bool) {
	return
}

// -------------------------------------------------------------------------------------

func (p *Sprite) sayOrThink(msgv interface{}, style int) {
	msg, ok := msgv.(string)
	if !ok {
		msg = fmt.Sprint(msgv)
	}

	if msg == "" {
		p.doStopSay()
		return
	}

	old := p.sayObj
	if old == nil {
		p.sayObj = &sayOrThinker{sp: p, msg: msg, style: style}
		p.g.addShape(p.sayObj)
	} else {
		old.msg, old.style = msg, style
		p.g.activateShape(old)
	}
}

func (p *Sprite) waitStopSay(secs float64) {
	p.g.Wait(secs)
	p.doStopSay()
}

func (p *Sprite) doStopSay() {
	if p.sayObj != nil {
		p.g.removeShape(p.sayObj)
		p.sayObj = nil
	}
}

// -------------------------------------------------------------------------------------
