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
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// -------------------------------------------------------------------------------------

type drawContext struct {
	*ebiten.Image
}

type hitContext struct {
	Pos image.Point
}

type hitResult struct {
	Target interface{}
}

type Shape interface {
}

// -------------------------------------------------------------------------------------

type spriteDrawInfo struct {
	sprite  *Sprite
	geo     ebiten.GeoM
	visible bool
}

func (p *Sprite) getDrawInfo() *spriteDrawInfo {
	return &spriteDrawInfo{
		sprite:  p,
		visible: p.isVisible,
	}
}

func (p *Sprite) touchPoint(x, y float64) bool {
	return true
}

func (p *Sprite) touchingSprite(dst *Sprite) bool {
	// TODO tanjp
	return false
}

func (p *Sprite) applyPivot(c *costume, cx, cy *float64) {
	*cx += p.pivot.X * float64(c.bitmapResolution)
	*cy -= p.pivot.Y * float64(c.bitmapResolution)
}
