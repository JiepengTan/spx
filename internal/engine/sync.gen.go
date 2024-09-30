package engine

import (
	. "godot-ext/gdspx/pkg/engine"
)

func SyncGetMousePos() Vec2 {
	var retValue Vec2
	done := make(chan struct{})
	job := func() {
		retValue = InputMgr.GetMousePos()
		done <- struct{}{}
	}
	updateJobQueue <- job
	<-done
	return retValue
}
