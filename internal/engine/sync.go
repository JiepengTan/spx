package engine

import (
	. "godot-ext/gdspx/pkg/engine"
)

func SyncInputMousePressed() bool {
	return SyncInputGetMouseState(0) || SyncInputGetMouseState(1)
}
func SyncCreateUiNode[T any](path string) *T {
	var __ret *T
	done := make(chan struct{})
	job := func() {
		__ret = CreateUI[T](path)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
func SyncCreateEngineUiNode[T any](path string) *T {
	var __ret *T
	done := make(chan struct{})
	job := func() {
		__ret = CreateEngineUI[T](path)
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return __ret
}
