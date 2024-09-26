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

	"github.com/goplus/spx/internal/engine"
)

const (
	physicColliderCircle = 0x00
	physicColliderRect   = 0x01
)

func initSpritePhysic(sprite *Sprite, proxy *engine.ProxySprite) {
	// TODO tanjp handle collision events
	proxy.SetScale(engine.NewVec2(0.5, 0.5))
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

func paserColliderType(typeName string) int64 {
	switch typeName {
	case "circle":
		return physicColliderCircle
	case "rect":
		return physicColliderRect
	}
	println("unknown collider type:", typeName)
	return physicColliderCircle
}
