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
	"github.com/goplus/spx/internal/engine"
)

var (
	gGamer Gamer
	gGame  *Game
)

func (p *Game) OnEngineStart() {
	if me, ok := gGamer.(interface{ MainEntry() }); ok {
		me.MainEntry()
	}
	if !gGame.isRunned {
		Gopt_Game_Run(gGamer, "assets")
	}
}

func (p *Game) OnEngineDestroy() {
}

func (p *Game) OnEngineUpdate(delta float32) {
	p.Update()
	p.updateProxy()
}

func (p *Game) updateProxy() {
	count := 0
	items := p.getItems()
	for _, item := range items {
		sprite, ok := item.(*Sprite)
		if ok {
			var proxy *engine.ProxySprite
			// bind proxy
			if sprite.proxy == nil && !sprite.HasDestroyed {
				sprite.proxy = engine.NewSpriteProxy(sprite)
				sprite.onBindProxy()
				initSpritePhysic(sprite, sprite.proxy)
				sprite.proxy.SetScale(engine.NewVec2(0.5, 0.5)) // TODO(tanjp) remove this hack
			}
			proxy = sprite.proxy
			if sprite.HasDestroyed {
				continue
			}
			proxy.Name = sprite.name
			// sync position
			if sprite.isVisible {
				x, y := sprite.getXY()
				proxy.SyncPos(x, y)
				proxy.SyncTexture(sprite.getCostumePath())
				count++
			}
			proxy.SetVisible(sprite.isVisible)
		}
	}

	// unbind proxy
	for _, item := range p.destroyItems {
		sprite, ok := item.(*Sprite)
		if ok && sprite.proxy != nil {
			sprite.proxy.Destroy()
			sprite.proxy = nil
		}
	}
	p.destroyItems = nil

	// update physic
	triggers := make([]engine.TriggerPair, 0)
	triggers = engine.GetTriggerPairs(triggers)
	for _, pair := range triggers {
		src := pair.Src.Target
		dst := pair.Dst.Target
		srcSprite, ok1 := src.(*Sprite)
		dstSrpite, ok2 := dst.(*Sprite)
		if ok1 && ok2 {
			if srcSprite.isVisible && !srcSprite.isDying && dstSrpite.isVisible && !dstSrpite.isDying {
				srcSprite.hasOnTouched = true
				srcSprite.fireTouched(dstSrpite)
			}

		} else {
			panic("unexpected trigger pair ")
		}
	}
}
