package engine

import (
	"math"

	gdx "github.com/realdream-ai/gdspx/pkg/engine"
	"github.com/realdream-ai/mathf"
	. "github.com/realdream-ai/mathf"
)

// =============== factory ===================

// it's always called in main thread
func BindUI[T any](parentNode gdx.Object, path string) *T {
	return gdx.BindUI[T](parentNode, path)
}
func NewSpriteProxy(obj interface{}) *SpriteProxy {
	proxy := gdx.CreateEmptySprite[SpriteProxy]()
	proxy.Target = obj
	return proxy
}

func SyncNewUiNode[T any]() *T {
	var _ret1 *T
	WaitMainThread(func() {
		_ret1 = gdx.CreateEngineUI[T]("")
	})
	return _ret1
}

func SyncNewSprite[T any]() *T {
	var _ret1 *T
	WaitMainThread(func() {
		_ret1 = gdx.CreateSprite[T]()
	})
	return _ret1
}

func SyncNewEmptySprite[T any]() *T {
	var _ret1 *T
	WaitMainThread(func() {
		_ret1 = gdx.CreateEmptySprite[T]()
	})
	return _ret1
}

func SyncNewBackdropProxy(obj interface{}, path string, renderScale float64) *SpriteProxy {
	var _ret1 *SpriteProxy
	WaitMainThread(func() {
		_ret1 = gdx.CreateEmptySprite[SpriteProxy]()
		_ret1.Target = obj
		_ret1.SetZIndex(-1)
		_ret1.DisablePhysic()
		_ret1.UpdateTexture(path, renderScale)
	})
	return _ret1
}

// =============== input ===================
func (pself *inputMgr) MousePressed() bool {
	return InputMgr.GetMouseState(0) || InputMgr.GetMouseState(1)
}

func GetMousePos() Vec2 {
	return gdx.InputMgr.GetMousePos()
}

// =============== window ===================

func (pself *platformMgr) SetRunnableOnUnfocused(flag bool) {
	if !flag {
		println("TODO tanjp SetRunnableOnUnfocused")
	}
}

func SyncReadAllText(path string) string {
	return ResMgr.ReadAllText(path)
}

// =============== setting ===================

func SyncSetDebugMode(isDebug bool) {
	PlatformMgr.SetDebugMode(isDebug)
}

func SetCameraPosition(pos Vec2) {
	gdx.CameraMgr.SetCameraPosition(mathf.NewVec2(pos.X, -pos.Y))
}

func GetBoundFromAlpha(assetPath string) Rect2 {
	return gdx.ResMgr.GetBoundFromAlpha(assetPath)
}

// =============== setting ===================
func ScreenToWorld(pos Vec2) Vec2 {
	camPos := gdx.CameraMgr.GetCameraPosition()
	camPos.Y *= -1
	return pos.Add(camPos)
}

func WorldToScreen(pos Vec2) Vec2 {
	camPos := gdx.CameraMgr.GetCameraPosition()
	camPos.Y *= -1
	return pos.Sub(camPos)
}

func SyncScreenToWorld(pos Vec2) Vec2 {
	var _ret1 Vec2
	WaitMainThread(func() {
		_ret1 = ScreenToWorld(pos)
	})
	return _ret1
}
func SyncWorldToScreen(pos Vec2) Vec2 {
	var _ret1 Vec2
	WaitMainThread(func() {
		_ret1 = WorldToScreen(pos)
	})
	return _ret1
}

func SyncReloadScene() {
	WaitMainThread(func() {
		gdx.ClearAllSprites()
	})
}
func DegToRad(p_y float64) float64 {
	return p_y * (gdx.Math_PI / 180.0)
}
func Sincos(rad float64) Vec2 {
	return NewVec2(math.Sincos(rad))
}
