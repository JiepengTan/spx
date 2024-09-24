package engine

import gde "godot-ext/gdspx/pkg/engine"

type ProxySprite struct {
	gde.Sprite
}

func (pself *ProxySprite) OnCustomeChanged(path string) {
	resPath := "res://assets/" + path
	println("OnCustomeChanged", resPath)
	pself.SetTexture(resPath)
}

func NewSpriteProxy() *ProxySprite {
	player := gde.CreateEmptySprite[ProxySprite]()
	return player
}
