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
	"math"
	"sync/atomic"

	"github.com/goplus/spx/internal/engine"

	gdspx "godot-ext/gdspx/pkg/engine"
)

func (p *Game) OnEngineStart() {
	go p.onStartAsync()
}

func (p *Game) OnEngineDestroy() {
}

func (p *Game) OnEngineUpdate(delta float32) {
	if !p.isRunned {
		return
	}
	// all these functions is called in main thread
	p.updateInput()
	p.updateCamera()
	p.updateUI()
	p.updateLogic()
	p.updateProxy()
	p.updatePhysic()
}

func (p *Game) onStartAsync() {
	gamer := p.gamer_
	if me, ok := gamer.(interface{ MainEntry() }); ok {
		me.MainEntry()
	}
	if !p.isRunned {
		Gopt_Game_Run(gamer, "assets")
	}
}

func (p *Game) updateLogic() error {
	p.startFlag.Do(func() {
		p.fireEvent(&eventStart{})
	})

	p.tickMgr.update()
	return nil
}

func (p *Game) updateCamera() {
	isOn, x, y := p.Camera.getFollowPos()
	if isOn {
		gdspx.CameraMgr.SetCameraPosition(engine.NewVec2(x, -y))
	}
}

func (p *Game) updateInput() {
	pos := gdspx.InputMgr.GetMousePos()
	atomic.StoreInt64(&p.gMouseX, int64(pos.X))
	atomic.StoreInt64(&p.gMouseY, int64(pos.Y))
}

func (p *Game) updateUI() {
	newItems := make([]Shape, len(p.items))
	copy(newItems, p.items)
	for _, item := range newItems {
		if result, ok := item.(interface{ OnUpdate(float32) }); ok {
			result.OnUpdate(0.01)
		}
	}
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
				initSpritePhysicInfo(sprite, sprite.proxy)
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
				proxy.UpdatePosRot(x, y, sprite.Heading()-sprite.initDirection)
				proxy.UpdateTexture(sprite.getCostumePath())
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
}

func (*Game) updatePhysic() {
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

func initSpritePhysicInfo(sprite *Sprite, proxy *engine.ProxySprite) {
	// update collision layers
	proxy.SetTriggerLayer(sprite.triggerLayer)
	proxy.SetTriggerMask(sprite.triggerMask)
	proxy.SetCollisionLayer(sprite.collisionLayer)
	proxy.SetCollisionMask(sprite.collisionMask)

	// set trigger & collider
	switch sprite.colliderType {
	case physicColliderCircle:
		proxy.SetColliderCircle(sprite.colliderCenter.ToVec2(), float32(math.Max(sprite.colliderRadius, 0.01)))
	case physicColliderRect:
		proxy.SetColliderRect(sprite.colliderCenter.ToVec2(), sprite.colliderSize.ToVec2())
	}

	switch sprite.triggerType {
	case physicColliderCircle:
		proxy.SetTriggerCircle(sprite.triggerCenter.ToVec2(), float32(math.Max(sprite.triggerRadius, 0.01)))
	case physicColliderRect:
		proxy.SetTriggerRect(sprite.triggerCenter.ToVec2(), sprite.triggerSize.ToVec2())
	}
}
