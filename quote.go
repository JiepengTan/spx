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

const (
	quotePadding     = 5.0
	quoteLineWidth   = 8.0
	quoteHeadLen     = 16.0
	quoteTextPadding = 3.0
	quoteBorderRadis = 10.0
)

func init() {
	const dpi = 72
}

type quoter struct {
	sprite      *Sprite
	message     string
	description string
}

func (p *Sprite) quote_(message, description string) {
	old := p.quoteObj
	if old == nil {
		p.quoteObj = &quoter{sprite: p, message: message, description: description}
		p.g.addShape(p.quoteObj)
	} else {
		old.message, old.description = message, description
		p.g.activateShape(old)
	}
}

func (p *Sprite) waitStopQuote(secs float64) {
	p.g.Wait(secs)
	p.doStopQuote()
}

func (p *Sprite) doStopQuote() {
	if p.quoteObj != nil {
		p.g.removeShape(p.quoteObj)
		p.quoteObj = nil
	}
}
