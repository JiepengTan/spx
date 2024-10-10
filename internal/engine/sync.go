package engine

import (
	. "godot-ext/gdspx/pkg/engine"
	"time"

	"github.com/goplus/spx/internal/coroutine"
)

var (
	Gco *coroutine.Coroutines
)

func Wait(secs float64) {
	Gco.Sleep(time.Duration(secs * 1e9))
}

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
