package tilemap

import (
	"fmt"
	"sort"
)

// Type definitions for tilemap data structures (prefixed with tscn to avoid naming conflicts)

// tscnPoint represents a 2D coordinate
type tscnPoint struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
}

// tscnWorldPoint represents a 2D coordinate in world space (pixels)
type tscnWorldPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// tscnTileSize represents the dimensions of a tile
type tscnTileSize struct {
	Width  int32 `json:"width"`
	Height int32 `json:"height"`
}

// tscnPhysicsData represents physics properties of a tile
type tscnPhysicsData struct {
	CollisionPoints []tscnWorldPoint `json:"collision_points,omitempty"`
}

// tscnTileInfo represents information about a single tile in the tileset
type tscnTileInfo struct {
	AtlasCoords tscnPoint       `json:"atlas_coords"`
	Physics     tscnPhysicsData `json:"physics,omitempty"`
}

// tscnTileSource represents a tileset source
type tscnTileSource struct {
	ID          int32          `json:"id"`
	TexturePath string         `json:"texture_path"`
	Tiles       []tscnTileInfo `json:"tiles"`
}

// tscnTileSet represents the complete tileset information
type tscnTileSet struct {
	Sources []tscnTileSource `json:"sources"`
}

// tscnTileInstance represents a placed tile in the map
type tscnTileInstance struct {
	TileCoords  tscnPoint      `json:"tile_coords"`
	WorldCoords tscnWorldPoint `json:"world_coords"`
	SourceID    int32          `json:"source_id"`
	AtlasCoords tscnPoint      `json:"atlas_coords"`
}

// tscnLayer represents a tilemap layer with compact tile data format
type tscnLayer struct {
	ID       int32   `json:"id"`
	Name     string  `json:"name"`
	TileData []int32 `json:"tile_data"`
}

// tscnTileMapData represents the complete tilemap data
type tscnTileMapData struct {
	Format   int32        `json:"format"`
	TileSize tscnTileSize `json:"tile_size"`
	TileSet  tscnTileSet  `json:"tileset"`
	Layers   []tscnLayer  `json:"layers"`
}

// tscnSprite2DNode represents a Sprite2D node in the scene
type tscnSprite2DNode struct {
	Name        string         `json:"name"`
	Parent      string         `json:"parent"`
	Position    tscnWorldPoint `json:"position"`
	TexturePath string         `json:"texture_path"`
	ZIndex      int32          `json:"z_index,omitempty"`
}

// tscnPrefabNode represents an instantiated prefab node in the scene
type tscnPrefabNode struct {
	Name       string                 `json:"name"`
	Parent     string                 `json:"parent"`
	Position   tscnWorldPoint         `json:"position"`
	PrefabPath string                 `json:"prefab_path"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// TscnMapData represents the root structure for JSON output
type TscnMapData struct {
	TileMap   tscnTileMapData    `json:"tilemap"`
	Sprite2Ds []tscnSprite2DNode `json:"sprite2ds"`
	Prefabs   []tscnPrefabNode   `json:"prefabs"`
}

type Transform struct {
	x   float64
	y   float64
	dir int64
}

func NewTransform(x, y float64, dir int64) *Transform {
	return &Transform{x: x, y: y, dir: dir}
}
func (p *Transform) X() float64 {
	return p.x
}
func (p *Transform) Y() float64 {
	return p.y
}
func (p *Transform) Dir() int64 {
	return p.dir
}
func (p *Transform) ToString() string {
	return fmt.Sprintf("TransformData %.1f %.1f %d", p.x, p.y, p.dir)
}

// Runtime utilities for parsing tile data
func LoadTilemaps(datas *TscnMapData, funcSetTile func(texturePath string, isCollision bool), funcSetLayer func(layerIndex int64),
	funcPlaceTiles func(positions []float64, texturePath string, layerIndex int64)) {
	paths := make(map[int32]string)
	for _, item := range datas.TileMap.TileSet.Sources {
		paths[item.ID] = item.TexturePath
		hasCollision := false
		for _, tile := range item.Tiles {
			hasCollision = hasCollision || tile.Physics.CollisionPoints != nil
		}
		funcSetTile(item.TexturePath, false)
	}
	for idx, layer := range datas.TileMap.Layers {
		layerId := int64(idx)
		funcSetLayer(layerId)
		tileData := layer.TileData
		tiles := parseTileData(tileData, datas.TileMap.TileSize)
		sort.Slice(tiles, func(i, j int) bool {
			return tiles[i].SourceID < tiles[j].SourceID
		})
		lastId := int32(-1)
		path := ""
		positions := make([]float64, 0, len(tiles)*2)
		positions = positions[:0]
		for _, tile := range tiles {
			if lastId != tile.SourceID {
				if len(positions) > 0 {
					funcPlaceTiles(positions, path, layerId)
				}
				positions = positions[:0]
				lastId = tile.SourceID
				path = paths[tile.SourceID]
			}
			y, x := tile.WorldCoords.Y, tile.WorldCoords.X
			positions = append(positions, x, -y)
		}
		if len(positions) > 0 {
			funcPlaceTiles(positions, path, layerId)
		}
	}
}

// TileMapParser provides utilities for parsing compact tile data
// ParseTileData converts compact tile data array to tile instances
// The tile_data format is: [x, source_id, atlas_coords_encoded, x2, source_id2, atlas_coords_encoded2, ...]
// Where atlas_coords_encoded combines atlas X and Y coordinates
func parseTileData(tileData []int32, tileSize tscnTileSize) []tscnTileInstance {
	var tiles []tscnTileInstance

	for i := 0; i < len(tileData); i += 3 {
		if i+2 >= len(tileData) {
			break
		}

		tilePos := tileData[i]
		sourceID := tileData[i+1]
		atlasEncoded := tileData[i+2]

		// Decode tile position (Godot uses a specific encoding)
		tileX := tilePos & 0xFFFF
		if tileX >= 0x8000 {
			tileX -= 0x10000 // Handle negative coordinates
		}
		tileY := (tilePos >> 16) & 0xFFFF
		if tileY >= 0x8000 {
			tileY -= 0x10000 // Handle negative coordinates
		}

		// Decode atlas coordinates (usually just X and Y)
		atlasX := atlasEncoded & 0xFFFF
		atlasY := (atlasEncoded >> 16) & 0xFFFF

		tile := tscnTileInstance{
			TileCoords: tscnPoint{X: tileX, Y: tileY},
			WorldCoords: tscnWorldPoint{
				X: float64(tileX * tileSize.Width),
				Y: float64(tileY * tileSize.Height),
			},
			SourceID:    sourceID,
			AtlasCoords: tscnPoint{X: atlasX, Y: atlasY},
		}

		tiles = append(tiles, tile)
	}

	return tiles
}
