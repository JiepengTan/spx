package engine

import (
	gde "godot-ext/gdspx/pkg/engine"
	"math"
)

type ProxySprite struct {
	gde.Sprite
	x, y    float64
	Name    string
	PicPath string
}

func NewSpriteProxy() *ProxySprite {
	player := gde.CreateEmptySprite[ProxySprite]()
	return player
}
func (pself *ProxySprite) OnCostumeChange(path string) {
	//resPath := "res://assets/" + path
	//println("OnCostumeChange", resPath)
}

func (pself *ProxySprite) SyncTexture(path string) {
	if path == "" {
		return
	}
	resPath := "res://assets/" + path
	pself.PicPath = resPath
	pself.SetTexture(pself.PicPath)
}

func (pself *ProxySprite) SyncPos(x, y float64) {
	if math.Abs(pself.x-x) < 0.1 && math.Abs(pself.y-y) < 0.1 {
		return
	}
	//println(pself.Name, "SyncPos", int(x), int(y))
	pself.x = x
	pself.y = y
	pself.SetPosition(gde.Vec2{X: float32(x), Y: float32(y)})
}
