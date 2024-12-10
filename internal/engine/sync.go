package engine

import (
	gdx "github.com/realdream-ai/gdspx/pkg/engine"
)

// =============== factory ===================

// it's always called in main thread
func BindUI[T any](parentNode gdx.Object, path string) *T {
	return gdx.BindUI[T](parentNode, path)
}
func CreateEmptySprite[T any]() *T {
	return gdx.CreateEmptySprite[T]()
}
func SyncCreateUiNode[T any](path string) *T {
	var _ret1 *T
	WaitMainThread(func() {
		_ret1 = gdx.CreateUI[T](path)
	})
	return _ret1
}
func SyncCreateEngineUiNode[T any](path string) *T {
	var _ret1 *T
	WaitMainThread(func() {
		_ret1 = gdx.CreateEngineUI[T](path)
	})
	return _ret1
}

func SyncCreateSprite[T any]() *T {
	var _ret1 *T
	WaitMainThread(func() {
		_ret1 = gdx.CreateSprite[T]()
	})
	return _ret1
}

func SyncCreateEmptySprite[T any]() *T {
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

// =============== window ===================
func SyncSetRunnableOnUnfocused(flag bool) {
	if !flag {
		println("TODO tanjp SyncSetRunnableOnUnfocused")
	}
}

func SyncReadAllText(path string) string {
	return ResMgr.ReadAllText(path)
}

// =============== setting ===================

func SyncSetDebugMode(isDebug bool) {
	PlatformMgr.SetDebugMode(isDebug)
}

// =============== setting ===================
func ScreenToWorld(x, y float64) (float64, float64) {
	camPos := gdx.CameraMgr.GetCameraPosition()
	posX, posY := float64(camPos.X), -float64(camPos.Y)
	x += posX
	y += posY
	return x, y
}

func WorldToScreen(x, y float64) (float64, float64) {
	camPos := gdx.CameraMgr.GetCameraPosition()
	posX, posY := float64(camPos.X), -float64(camPos.Y)
	x -= posX
	y -= posY
	return x, y
}

func SyncScreenToWorld(x, y float64) (float64, float64) {
	var _ret1, _ret2 float64
	WaitMainThread(func() {
		_ret1, _ret2 = ScreenToWorld(x, y)
	})
	return _ret1, _ret2
}
func SyncWorldToScreen(x, y float64) (float64, float64) {
	var _ret1, _ret2 float64
	WaitMainThread(func() {
		_ret1, _ret2 = WorldToScreen(x, y)
	})
	return _ret1, _ret2
}

func SyncReloadScene() {
	WaitMainThread(func() {
		gdx.ClearAllSprites()
	})
}
func DegToRad(p_y float64) float64 {
	return p_y * (gdx.Math_PI / 180.0)
}
