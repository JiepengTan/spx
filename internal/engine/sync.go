package engine

import (
	. "godot-ext/gdspx/pkg/engine"
)

type ProxySprite struct {
	Sprite
	x, y    float64
	Name    string
	PicPath string
	target  interface{}
}

func NewSpriteProxy(obj interface{}) *ProxySprite {
	player := CreateEmptySprite[ProxySprite]()
	player.target = obj
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
	//println(pself.Name, "SyncPos", int(x), int(y))
	pself.x = x
	pself.y = y
	pself.SetPosition(Vec2{X: float32(x), Y: float32(y)})
}

func (pself *ProxySprite) OnTriggerEnter(target ISpriter) {
	sprite, ok := target.(*ProxySprite)
	if ok {
		println(pself.Name, " OnTriggerEnter ", sprite.Name)
	}
}
