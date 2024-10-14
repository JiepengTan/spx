package engine

import (
	. "godot-ext/gdspx/pkg/engine"
)

type ProxySprite struct {
	Sprite
	x, y    float64
	Name    string
	PicPath string
	Target  interface{}
}

func NewSpriteProxy(obj interface{}) *ProxySprite {
	proxy := CreateEmptySprite[ProxySprite]()
	proxy.Target = obj
	return proxy
}

func (pself *ProxySprite) OnCostumeChange(path string) {
	//resPath := "res://assets/" + path
	//println("OnCostumeChange", resPath)
}

func (pself *ProxySprite) UpdateTexture(path string) {
	if path == "" {
		return
	}
	resPath := ToEnginePath(path)
	pself.PicPath = resPath
	pself.SetTexture(pself.PicPath)
}

func (pself *ProxySprite) UpdatePosRot(x, y float64, rot float64) {
	pself.x = x
	pself.y = y
	pself.SetPosition(Vec2{X: float32(x), Y: float32(y)})
	rad := HeadingToRad(rot)
	pself.SetRotation(rad)
}

func (pself *ProxySprite) OnTriggerEnter(target ISpriter) {
	sprite, ok := target.(*ProxySprite)
	if ok {
		tempTriggerPairs = append(tempTriggerPairs, TriggerPair{Src: pself, Dst: sprite})
	}
}
