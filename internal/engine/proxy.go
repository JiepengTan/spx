package engine

import (
	gdx "github.com/realdream-ai/gdspx/pkg/engine"
	. "github.com/realdream-ai/mathf"
)

type SpriteProxy struct {
	gdx.Sprite
	x, y    float64
	Name    string
	PicPath string
	Target  interface{}
}

func NewSpriteProxy(obj interface{}) *SpriteProxy {
	proxy := CreateEmptySprite[SpriteProxy]()
	proxy.Target = obj
	return proxy
}

func (pself *SpriteProxy) UpdateTexture(path string, renderScale float64) {
	if path == "" {
		return
	}
	resPath := ToAssetPath(path)
	pself.PicPath = resPath
	pself.SetTexture(pself.PicPath)
	pself.SetRenderScale(NewVec2(renderScale, renderScale))
}
func (pself *SpriteProxy) UpdateTextureAltas(path string, rect2 Rect2, renderScale float64) {
	if path == "" {
		return
	}
	resPath := ToAssetPath(path)
	pself.PicPath = resPath
	pself.SetTextureAltas(pself.PicPath, rect2)
	pself.SetRenderScale(NewVec2(renderScale, renderScale))
}

func (pself *SpriteProxy) UpdateTransform(x, y float64, rot float64, scale64 float64, isSync bool) {
	pself.x = x
	pself.y = y
	rad := DegToRad(rot)
	pos := Vec2{X: float64(x), Y: float64(y)}
	scale := float64(scale64)
	if !isSync {
		pself.SetPosition(pos)
		pself.SetRotation(rad)
		pself.SetScaleX(scale)
	} else {
		WaitMainThread(func() {
			pself.SetPosition(pos)
			pself.SetRotation(rad)
			pself.SetScaleX(scale)
		})
	}
}

func (pself *SpriteProxy) OnTriggerEnter(target gdx.ISpriter) {
	sprite, ok := target.(*SpriteProxy)
	if ok {
		triggerEventsTemp = append(triggerEventsTemp, TriggerEvent{Src: pself, Dst: sprite})
	}
}
