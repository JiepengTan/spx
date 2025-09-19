/*
 * Copyright (c) 2021 The XGo Authors (xgo.dev). All rights reserved.
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

	spxfs "github.com/goplus/spx/v2/fs"
	tm "github.com/goplus/spx/v2/internal/tilemap"
)

type tilemapMgr struct {
	g     *Game
	datas *tm.TscnMapData
}

// loadTilemapJson loads tilemap data from a JSON file
// Usage: mapData := loadTilemapJson("path/to/tm.json")
func (p *tilemapMgr) init(g *Game, fs spxfs.Dir, path string) {
	p.g = g
	if path == "" {
		return
	}
	var data tm.TscnMapData
	err := loadJson(&data, fs, path)
	if err != nil {
		panic(fmt.Sprintf("Failed to load tilemap JSON file %s: %v", path, err))
	}
	p.datas = &data
}

func (p *tilemapMgr) loadTilemaps(datas *tm.TscnMapData) {
	tm.LoadTilemaps(datas, p.g.SetTileMapLayerIndex, p.g.PlaceTiles__1)

}
func (p *tilemapMgr) loadSprite2ds(datas *tm.TscnMapData) {
	for _, item := range datas.Sprite2Ds {
		p.g.CreatePureSprite(item.TexturePath, item.Position.X, -item.Position.Y, int64(item.ZIndex))
	}
}

func (p *tilemapMgr) loadGameObjs(datas *tm.TscnMapData) {
	for _, item := range datas.Prefabs {
		name := item.PrefabPath
		x, y := item.Position.X, -item.Position.Y
		trans := tm.Transform{X: x, Y: y}
		sp, ok := p.g.sprs[name]
		if ok {
			Gopt_SpriteImpl_Clone__1(sp, trans)
		}
	}
}

func (p *tilemapMgr) parseTilemap() {
	if p.datas == nil {
		return
	}
	p.loadTilemaps(p.datas)
	p.loadSprite2ds(p.datas)
	p.loadGameObjs(p.datas)
}
