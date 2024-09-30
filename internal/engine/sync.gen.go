package engine

import (
	. "godot-ext/gdspx/pkg/engine"
)

// IAudioMgr
func SyncAudioPlayAudio(path string) {

	done := make(chan struct{})
	job := func() {
		AudioMgr.PlayAudio(path)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncAudioSetAudioVolume(volume float32) {

	done := make(chan struct{})
	job := func() {
		AudioMgr.SetAudioVolume(volume)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncAudioGetAudioVolume() float32 {
	var __ret float32
	done := make(chan struct{})
	job := func() {
		__ret = AudioMgr.GetAudioVolume()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncAudioIsMusicPlaying() bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = AudioMgr.IsMusicPlaying()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncAudioPlayMusic(path string) {

	done := make(chan struct{})
	job := func() {
		AudioMgr.PlayMusic(path)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncAudioSetMusicVolume(volume float32) {

	done := make(chan struct{})
	job := func() {
		AudioMgr.SetMusicVolume(volume)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncAudioGetMusicVolume() float32 {
	var __ret float32
	done := make(chan struct{})
	job := func() {
		__ret = AudioMgr.GetMusicVolume()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncAudioPauseMusic() {

	done := make(chan struct{})
	job := func() {
		AudioMgr.PauseMusic()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncAudioResumeMusic() {

	done := make(chan struct{})
	job := func() {
		AudioMgr.ResumeMusic()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncAudioGetMusicTimer() float32 {
	var __ret float32
	done := make(chan struct{})
	job := func() {
		__ret = AudioMgr.GetMusicTimer()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncAudioSetMusicTimer(time float32) {

	done := make(chan struct{})
	job := func() {
		AudioMgr.SetMusicTimer(time)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}

// ICameraMgr
func SyncCameraGetCameraPosition() Vec2 {
	var __ret Vec2
	done := make(chan struct{})
	job := func() {
		__ret = CameraMgr.GetCameraPosition()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncCameraSetCameraPosition(position Vec2) {

	done := make(chan struct{})
	job := func() {
		CameraMgr.SetCameraPosition(position)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncCameraGetCameraZoom() Vec2 {
	var __ret Vec2
	done := make(chan struct{})
	job := func() {
		__ret = CameraMgr.GetCameraZoom()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncCameraSetCameraZoom(size Vec2) {

	done := make(chan struct{})
	job := func() {
		CameraMgr.SetCameraZoom(size)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncCameraGetViewportRect() Rect2 {
	var __ret Rect2
	done := make(chan struct{})
	job := func() {
		__ret = CameraMgr.GetViewportRect()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}

// IInputMgr
func SyncInputGetMousePos() Vec2 {
	var __ret Vec2
	done := make(chan struct{})
	job := func() {
		__ret = InputMgr.GetMousePos()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncInputGetKey(key int64) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = InputMgr.GetKey(key)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncInputGetMouseState(mouse_id int64) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = InputMgr.GetMouseState(mouse_id)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncInputGetKeyState(key int64) int64 {
	var __ret int64
	done := make(chan struct{})
	job := func() {
		__ret = InputMgr.GetKeyState(key)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncInputGetAxis(neg_action string, pos_action string) float32 {
	var __ret float32
	done := make(chan struct{})
	job := func() {
		__ret = InputMgr.GetAxis(neg_action, pos_action)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncInputIsActionPressed(action string) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = InputMgr.IsActionPressed(action)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncInputIsActionJustPressed(action string) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = InputMgr.IsActionJustPressed(action)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncInputIsActionJustReleased(action string) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = InputMgr.IsActionJustReleased(action)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}

// IPhysicMgr
func SyncPhysicRaycast(from Vec2, to Vec2, collision_mask int64) Object {
	var __ret Object
	done := make(chan struct{})
	job := func() {
		__ret = PhysicMgr.Raycast(from, to, collision_mask)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncPhysicCheckCollision(from Vec2, to Vec2, collision_mask int64, collide_with_areas bool, collide_with_bodies bool) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = PhysicMgr.CheckCollision(from, to, collision_mask, collide_with_areas, collide_with_bodies)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}

// IPlatformMgr
func SyncPlatformSetWindowSize(width int64, height int64) {

	done := make(chan struct{})
	job := func() {
		PlatformMgr.SetWindowSize(width, height)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncPlatformGetWindowSize() Vec2 {
	var __ret Vec2
	done := make(chan struct{})
	job := func() {
		__ret = PlatformMgr.GetWindowSize()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncPlatformSetWindowTitle(title string) {

	done := make(chan struct{})
	job := func() {
		PlatformMgr.SetWindowTitle(title)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncPlatformGetWindowTitle() string {
	var __ret string
	done := make(chan struct{})
	job := func() {
		__ret = PlatformMgr.GetWindowTitle()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncPlatformSetWindowFullscreen(enable bool) {

	done := make(chan struct{})
	job := func() {
		PlatformMgr.SetWindowFullscreen(enable)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncPlatformIsWindowFullscreen() bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = PlatformMgr.IsWindowFullscreen()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncPlatformSetDebugMode(enable bool) {

	done := make(chan struct{})
	job := func() {
		PlatformMgr.SetDebugMode(enable)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncPlatformIsDebugMode() bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = PlatformMgr.IsDebugMode()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}

// ISceneMgr
func SyncSceneChangeSceneToFile(path string) {

	done := make(chan struct{})
	job := func() {
		SceneMgr.ChangeSceneToFile(path)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSceneReloadCurrentScene() int64 {
	var __ret int64
	done := make(chan struct{})
	job := func() {
		__ret = SceneMgr.ReloadCurrentScene()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSceneUnloadCurrentScene() {

	done := make(chan struct{})
	job := func() {
		SceneMgr.UnloadCurrentScene()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}

// ISpriteMgr
func SyncSpriteSetDontDestroyOnLoad(obj Object) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetDontDestroyOnLoad(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteSetProcess(obj Object, is_on bool) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetProcess(obj, is_on)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteSetPhysicProcess(obj Object, is_on bool) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetPhysicProcess(obj, is_on)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteSetChildPosition(obj Object, path string, pos Vec2) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetChildPosition(obj, path, pos)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetChildPosition(obj Object, path string) Vec2 {
	var __ret Vec2
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetChildPosition(obj, path)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetChildRotation(obj Object, path string, rot float32) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetChildRotation(obj, path, rot)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetChildRotation(obj Object, path string) float32 {
	var __ret float32
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetChildRotation(obj, path)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetChildScale(obj Object, path string, scale Vec2) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetChildScale(obj, path, scale)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetChildScale(obj Object, path string) Vec2 {
	var __ret Vec2
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetChildScale(obj, path)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteCheckCollision(obj Object, target Object, is_src_trigger bool, is_dst_trigger bool) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.CheckCollision(obj, target, is_src_trigger, is_dst_trigger)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteCheckCollisionWithPoint(obj Object, point Vec2, is_trigger bool) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.CheckCollisionWithPoint(obj, point, is_trigger)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteCreateSprite(path string) Object {
	var __ret Object
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.CreateSprite(path)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteCloneSprite(obj Object) Object {
	var __ret Object
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.CloneSprite(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteDestroySprite(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.DestroySprite(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteIsSpriteAlive(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.IsSpriteAlive(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetPosition(obj Object, pos Vec2) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetPosition(obj, pos)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetPosition(obj Object) Vec2 {
	var __ret Vec2
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetPosition(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetRotation(obj Object, rot float32) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetRotation(obj, rot)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetRotation(obj Object) float32 {
	var __ret float32
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetRotation(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetScale(obj Object, scale Vec2) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetScale(obj, scale)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetScale(obj Object) Vec2 {
	var __ret Vec2
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetScale(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetColor(obj Object, color Color) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetColor(obj, color)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetColor(obj Object) Color {
	var __ret Color
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetColor(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetTexture(obj Object, path string) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetTexture(obj, path)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetTexture(obj Object) string {
	var __ret string
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetTexture(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetVisible(obj Object, visible bool) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetVisible(obj, visible)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetVisible(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetVisible(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteGetZIndex(obj Object) int64 {
	var __ret int64
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetZIndex(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetZIndex(obj Object, z int64) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetZIndex(obj, z)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpritePlayAnim(obj Object, p_name string, p_custom_scale float32, p_from_end bool) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.PlayAnim(obj, p_name, p_custom_scale, p_from_end)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpritePlayBackwardsAnim(obj Object, p_name string) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.PlayBackwardsAnim(obj, p_name)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpritePauseAnim(obj Object) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.PauseAnim(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteStopAnim(obj Object) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.StopAnim(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteIsPlayingAnim(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.IsPlayingAnim(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetAnim(obj Object, p_name string) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetAnim(obj, p_name)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetAnim(obj Object) string {
	var __ret string
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetAnim(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetAnimFrame(obj Object, p_frame int64) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetAnimFrame(obj, p_frame)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetAnimFrame(obj Object) int64 {
	var __ret int64
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetAnimFrame(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetAnimSpeedScale(obj Object, p_speed_scale float32) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetAnimSpeedScale(obj, p_speed_scale)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetAnimSpeedScale(obj Object) float32 {
	var __ret float32
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetAnimSpeedScale(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteGetAnimPlayingSpeed(obj Object) float32 {
	var __ret float32
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetAnimPlayingSpeed(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetAnimCentered(obj Object, p_center bool) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetAnimCentered(obj, p_center)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteIsAnimCentered(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.IsAnimCentered(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetAnimOffset(obj Object, p_offset Vec2) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetAnimOffset(obj, p_offset)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetAnimOffset(obj Object) Vec2 {
	var __ret Vec2
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetAnimOffset(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetAnimFlipH(obj Object, p_flip bool) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetAnimFlipH(obj, p_flip)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteIsAnimFlippedH(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.IsAnimFlippedH(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetAnimFlipV(obj Object, p_flip bool) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetAnimFlipV(obj, p_flip)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteIsAnimFlippedV(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.IsAnimFlippedV(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetVelocity(obj Object, velocity Vec2) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetVelocity(obj, velocity)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetVelocity(obj Object) Vec2 {
	var __ret Vec2
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetVelocity(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteIsOnFloor(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.IsOnFloor(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteIsOnFloorOnly(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.IsOnFloorOnly(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteIsOnWall(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.IsOnWall(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteIsOnWallOnly(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.IsOnWallOnly(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteIsOnCeiling(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.IsOnCeiling(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteIsOnCeilingOnly(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.IsOnCeilingOnly(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteGetLastMotion(obj Object) Vec2 {
	var __ret Vec2
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetLastMotion(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteGetPositionDelta(obj Object) Vec2 {
	var __ret Vec2
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetPositionDelta(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteGetFloorNormal(obj Object) Vec2 {
	var __ret Vec2
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetFloorNormal(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteGetWallNormal(obj Object) Vec2 {
	var __ret Vec2
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetWallNormal(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteGetRealVelocity(obj Object) Vec2 {
	var __ret Vec2
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetRealVelocity(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteMoveAndSlide(obj Object) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.MoveAndSlide(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteSetGravity(obj Object, gravity float32) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetGravity(obj, gravity)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetGravity(obj Object) float32 {
	var __ret float32
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetGravity(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetMass(obj Object, mass float32) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetMass(obj, mass)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetMass(obj Object) float32 {
	var __ret float32
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetMass(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteAddForce(obj Object, force Vec2) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.AddForce(obj, force)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteAddImpulse(obj Object, impulse Vec2) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.AddImpulse(obj, impulse)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteSetCollisionLayer(obj Object, layer int64) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetCollisionLayer(obj, layer)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetCollisionLayer(obj Object) int64 {
	var __ret int64
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetCollisionLayer(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetCollisionMask(obj Object, mask int64) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetCollisionMask(obj, mask)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetCollisionMask(obj Object) int64 {
	var __ret int64
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetCollisionMask(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetTriggerLayer(obj Object, layer int64) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetTriggerLayer(obj, layer)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetTriggerLayer(obj Object) int64 {
	var __ret int64
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetTriggerLayer(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetTriggerMask(obj Object, mask int64) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetTriggerMask(obj, mask)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteGetTriggerMask(obj Object) int64 {
	var __ret int64
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.GetTriggerMask(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetColliderRect(obj Object, center Vec2, size Vec2) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetColliderRect(obj, center, size)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteSetColliderCircle(obj Object, center Vec2, radius float32) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetColliderCircle(obj, center, radius)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteSetColliderCapsule(obj Object, center Vec2, size Vec2) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetColliderCapsule(obj, center, size)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteSetCollisionEnabled(obj Object, enabled bool) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetCollisionEnabled(obj, enabled)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteIsCollisionEnabled(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.IsCollisionEnabled(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncSpriteSetTriggerRect(obj Object, center Vec2, size Vec2) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetTriggerRect(obj, center, size)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteSetTriggerCircle(obj Object, center Vec2, radius float32) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetTriggerCircle(obj, center, radius)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteSetTriggerCapsule(obj Object, center Vec2, size Vec2) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetTriggerCapsule(obj, center, size)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteSetTriggerEnabled(obj Object, trigger bool) {

	done := make(chan struct{})
	job := func() {
		SpriteMgr.SetTriggerEnabled(obj, trigger)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncSpriteIsTriggerEnabled(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = SpriteMgr.IsTriggerEnabled(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}

// IUiMgr
func SyncUiCreateNode(path string) Object {
	var __ret Object
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.CreateNode(path)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncUiCreateButton(path string, text string) Object {
	var __ret Object
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.CreateButton(path, text)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncUiCreateLabel(path string, text string) Object {
	var __ret Object
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.CreateLabel(path, text)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncUiCreateImage(path string) Object {
	var __ret Object
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.CreateImage(path)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncUiCreateToggle(path string, value bool) Object {
	var __ret Object
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.CreateToggle(path, value)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncUiCreateSlider(path string, value float32) Object {
	var __ret Object
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.CreateSlider(path, value)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncUiCreateInput(path string, text string) Object {
	var __ret Object
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.CreateInput(path, text)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncUiDestroyNode(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.DestroyNode(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncUiGetType(obj Object) int64 {
	var __ret int64
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.GetType(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncUiSetText(obj Object, text string) {

	done := make(chan struct{})
	job := func() {
		UiMgr.SetText(obj, text)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncUiGetText(obj Object) string {
	var __ret string
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.GetText(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncUiSetTexture(obj Object, path string) {

	done := make(chan struct{})
	job := func() {
		UiMgr.SetTexture(obj, path)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncUiGetTexture(obj Object) string {
	var __ret string
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.GetTexture(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncUiSetColor(obj Object, color Color) {

	done := make(chan struct{})
	job := func() {
		UiMgr.SetColor(obj, color)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncUiGetColor(obj Object) Color {
	var __ret Color
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.GetColor(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncUiSetFontSize(obj Object, size int64) {

	done := make(chan struct{})
	job := func() {
		UiMgr.SetFontSize(obj, size)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncUiGetFontSize(obj Object) int64 {
	var __ret int64
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.GetFontSize(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncUiSetVisible(obj Object, visible bool) {

	done := make(chan struct{})
	job := func() {
		UiMgr.SetVisible(obj, visible)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncUiGetVisible(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.GetVisible(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncUiSetInteractable(obj Object, interactable bool) {

	done := make(chan struct{})
	job := func() {
		UiMgr.SetInteractable(obj, interactable)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncUiGetInteractable(obj Object) bool {
	var __ret bool
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.GetInteractable(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncUiSetRect(obj Object, rect Rect2) {

	done := make(chan struct{})
	job := func() {
		UiMgr.SetRect(obj, rect)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
}
func SyncUiGetRect(obj Object) Rect2 {
	var __ret Rect2
	done := make(chan struct{})
	job := func() {
		__ret = UiMgr.GetRect(obj)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
