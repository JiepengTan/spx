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
	"path"
	"sort"

	spxfs "github.com/goplus/spx/v2/fs"
	"github.com/goplus/spx/v2/internal/engine"
	tm "github.com/goplus/spx/v2/internal/tilemap"

	"github.com/goplus/spbase/mathf"
)

// DecoratorJSON represents the structure of decorator.json file (new format)
type DecoratorJSON struct {
	Version    int                `json:"version"`
	Decorators []tm.DecoratorNode `json:"decorators"`
}

type gameTilemapMgr struct {
	g              *Game
	datas          *tm.TscnMapData
	decoratorDatas *DecoratorJSON
	useNewLoader   bool   // true if using C++ TileMapParser (new format)
	tilemapPath    string // path to tilemap.json
	tilemapDir     string // directory containing tilemap.json
}

func (p *gameTilemapMgr) init(g *Game, fs spxfs.Dir, tilemapPath string) {
	p.g = g
	p.tilemapPath = tilemapPath
	if tilemapPath == "" {
		return
	}

	// Get directory containing tilemap.json
	p.tilemapDir = path.Dir(tilemapPath)

	// Check if JSON is in new format (version >= 1)
	if p.isNewFormat(fs, tilemapPath) {
		// New format: use C++ TileMapParser for loading tilemap
		enginePath := engine.ToAssetPath(tilemapPath)
		fmt.Printf("[TILEMAP] Using C++ TileMapParser for: %s\n", enginePath)
		tilemapparserMgr.LoadTilemap(enginePath)
		p.useNewLoader = true

		// Load decorator.json from the same directory
		decoratorPath := path.Join(p.tilemapDir, "decorator.json")
		p.loadDecoratorJSON(fs, decoratorPath)
		return
	}

	// Old format: use existing Go loader
	var data tm.TscnMapData
	err := loadJson(&data, fs, tilemapPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to load tilemap JSON file %s: %v", tilemapPath, err))
	}
	p.datas = &data
	tm.ConvertData(&data)
}

// loadDecoratorJSON loads decorator data from a separate decorator.json file
func (p *gameTilemapMgr) loadDecoratorJSON(fs spxfs.Dir, decoratorPath string) {
	var data DecoratorJSON
	err := loadJson(&data, fs, decoratorPath)
	if err != nil {
		fmt.Printf("[TILEMAP] No decorator.json found at %s (this is OK if no decorators)\n", decoratorPath)
		return
	}
	p.decoratorDatas = &data
	fmt.Printf("[TILEMAP] Loaded %d decorators from %s\n", len(data.Decorators), decoratorPath)
}

func (p *gameTilemapMgr) hasData() bool {
	return p.datas != nil || p.useNewLoader
}

// isNewFormat checks if the tilemap JSON is in the new format (version >= 1)
// New format uses C++ TileMapParser with Base64 encoded tile_map_data
func (p *gameTilemapMgr) isNewFormat(fs spxfs.Dir, path string) bool {
	var versionCheck struct {
		Version int `json:"version"`
	}
	if err := loadJson(&versionCheck, fs, path); err != nil {
		return false
	}
	return versionCheck.Version >= 1
}

func (p *gameTilemapMgr) loadTilemaps(datas *tm.TscnMapData) {
	tm.LoadTilemaps(datas, p.g.setTileInfo__1, p.g.setTileMapLayerIndex, p.g.PlaceTiles__1)
}
func (p *gameTilemapMgr) loadDecorators(datas *tm.TscnMapData) {
	p.loadDecoratorNodes(datas.Decorators, "tilemaps")
}

func (p *gameTilemapMgr) loadDecoratorNodes(decorators []tm.DecoratorNode, tilemapDir string) {
	const headingOffset = -90.0
	for _, item := range decorators {
		position := item.Position.ToVec2()
		pivot := item.Pivot.ToVec2()
		relativePath := path.Join(tilemapDir, item.Path)
		assetPath := engine.ToAssetPath(relativePath)
		texSize := resMgr.GetImageSize(assetPath)
		colliderPivot := item.ColliderPivot.ToVec2().Add(pivot)
		pivot = pivot.Sub(texSize.Divf(2))
		p.g.createStaticSprite(relativePath, position, item.Ratation+headingOffset,
			item.Scale.ToVec2(), int64(item.ZIndex), pivot, item.ColliderType, colliderPivot, item.ColliderParams)
	}
}

// loadDecoratorsFromJSON loads decorators from the separate decorator.json file (new format)
func (p *gameTilemapMgr) loadDecoratorsFromJSON() {
	if p.decoratorDatas == nil || len(p.decoratorDatas.Decorators) == 0 {
		return
	}
	p.loadDecoratorNodes(p.decoratorDatas.Decorators, p.tilemapDir)
	fmt.Printf("====>[TILEMAP] Created %d decorator sprites\n", len(p.decoratorDatas.Decorators))
}

func (p *gameTilemapMgr) loadSprites(datas *tm.TscnMapData) {

	sort.Slice(datas.Sprites, func(i, j int) bool {
		return datas.Sprites[i].Path < datas.Sprites[j].Path
	})

	for _, item := range datas.Sprites {
		sp, ok := p.g.sprs[item.Path]
		if ok {
			x, y := item.Position.X, item.Position.Y
			doClone(sp, nil, true, func(sprite *SpriteImpl) {
				sprite.SetXYpos(x, y)
				sprite.Show()
			})
		}
	}
}

func (p *gameTilemapMgr) parseTilemap() {
	// Handle new format: load decorators from separate JSON file
	if p.useNewLoader {
		p.loadDecoratorsFromJSON()
		return
	}

	// Old format: load from combined TscnMapData
	if p.datas == nil {
		return
	}
	p.loadTilemaps(p.datas)
	p.loadDecorators(p.datas)
	//p.loadSprites(p.datas)

	// Update world size based on actual tilemap content
	p.calcWorldSize()
}

// calcWorldSize calculates and updates world size based on actual tile distribution in tilemap
func (p *gameTilemapMgr) calcWorldSize() {
	if p.datas == nil || len(p.datas.TileMap.Layers) == 0 {
		fmt.Println("[TILEMAP DEBUG] No tilemap data or layers, skipping world size update")
		return
	}

	tileSizeX := int(p.datas.TileMap.TileSize.Width)
	tileSizeY := int(p.datas.TileMap.TileSize.Height)

	var minX, maxX, minY, maxY int32 = 0, 0, 0, 0
	hasAnyTiles := false
	totalTiles := 0

	for _, layer := range p.datas.TileMap.Layers {
		tiles := p.parseTileDataForBounds(layer.TileData)
		totalTiles += len(tiles)
		for _, tile := range tiles {
			if !hasAnyTiles {
				minX, maxX = tile.X, tile.X
				minY, maxY = tile.Y, tile.Y
				hasAnyTiles = true
			} else {
				if tile.X < minX {
					minX = tile.X
				}
				if tile.X > maxX {
					maxX = tile.X
				}
				if tile.Y < minY {
					minY = tile.Y
				}
				if tile.Y > maxY {
					maxY = tile.Y
				}
			}
		}
	}

	if hasAnyTiles {
		minWorldX := int((minX) * int32(tileSizeX))
		maxWorldX := int((maxX + 1) * int32(tileSizeX)) // +1 to include the full size of the last tile
		minWorldY := int((minY - 1) * int32(tileSizeY)) // -1 to include the full size of the last tile
		maxWorldY := int((maxY) * int32(tileSizeY))

		worldWidth := maxWorldX - minWorldX
		worldHeight := maxWorldY - minWorldY

		p.g.minWorldX_ = minWorldX
		p.g.minWorldY_ = minWorldY
		p.g.worldWidth_ = worldWidth
		p.g.worldHeight_ = worldHeight

	} else {
		fmt.Println("[TILEMAP DEBUG] No tiles found in any layer")
	}
}

// parseTileDataForBounds parses tile data for boundary calculation (copied logic from internal/tilemap package)
func (p *gameTilemapMgr) parseTileDataForBounds(tileData []int32) []mathf.Vec2i {
	tileCount := len(tileData) / 5
	tiles := make([]mathf.Vec2i, 0, tileCount)

	for i := 0; i < len(tileData); i += 5 {
		if i+4 >= len(tileData) {
			break
		}

		tileX := tileData[i+1]
		tileY := tileData[i+2]

		tile := mathf.Vec2i{
			X: tileX,
			Y: tileY,
		}

		tiles = append(tiles, tile)
	}

	return tiles
}
