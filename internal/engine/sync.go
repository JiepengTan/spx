package engine

import (
	. "github.com/realdream-ai/gdspx/pkg/engine"
)

// =============== factory ===================
func SyncCreateUiNode[T any](path string) *T {
	WaitMainThread()
	return CreateUI[T](path)
}
func SyncCreateEngineUiNode[T any](path string) *T {
	WaitMainThread()
	return CreateEngineUI[T](path)
}

func SyncCreateSprite[T any]() *T {
	WaitMainThread()
	return CreateSprite[T]()
}

func SyncCreateEmptySprite[T any]() *T {
	WaitMainThread()
	return CreateEmptySprite[T]()
}

func SyncNewBackdropProxy(obj interface{}, path string, renderScale float64) *ProxySprite {
	WaitMainThread()
	return newBackdropProxy(obj, path, renderScale)
}

func newBackdropProxy(obj interface{}, path string, renderScale float64) *ProxySprite {
	__ret := CreateEmptySprite[ProxySprite]()
	__ret.Target = obj
	__ret.SetZIndex(-1)
	__ret.DisablePhysic()
	__ret.UpdateTexture(path, renderScale)
	return __ret
}

// =============== input ===================
func SyncInputMousePressed() bool {
	return SyncInputGetMouseState(0) || SyncInputGetMouseState(1)
}

// =============== time ===================
func SyncGetCurrentTPS() float64 {
	return 30 // TODO(tanjp) use engine api
}

// =============== window ===================
func SyncSetRunnableOnUnfocused(flag bool) {
	if !flag {
		println("TODO tanjp SyncSetRunnableOnUnfocused")
	}
}

func SyncReadAllText(path string) string {
	return SyncResReadAllText(path)
}

// =============== setting ===================

func SyncSetDebugMode(isDebug bool) {
	SyncPlatformSetDebugMode(isDebug)
}

// =============== setting ===================
func ScreenToWorld(x, y float64) (float64, float64) {
	camPos := CameraMgr.GetCameraPosition()
	posX, posY := float64(camPos.X), -float64(camPos.Y)
	x += posX
	y += posY
	return x, y
}

func WorldToScreen(x, y float64) (float64, float64) {
	camPos := CameraMgr.GetCameraPosition()
	posX, posY := float64(camPos.X), -float64(camPos.Y)
	x -= posX
	y -= posY
	return x, y
}

func SyncScreenToWorld(x, y float64) (float64, float64) {
	WaitMainThread()
	return ScreenToWorld(x, y)
}
func SyncWorldToScreen(x, y float64) (float64, float64) {
	WaitMainThread()
	return WorldToScreen(x, y)
}

func SyncGetCameraLocalPosition(x, y float64) (float64, float64) {
	posX, posY := SyncGetCameraPosition()
	x -= posX
	y -= posY
	return x, y
}
func SyncGetCameraPosition() (float64, float64) {
	pos := SyncCameraGetCameraPosition()
	return float64(pos.X), -float64(pos.Y)
}
func SyncSetCameraPosition(x, y float64) {
	SyncCameraSetCameraPosition(NewVec2(float64(x), -float64(y)))
}
