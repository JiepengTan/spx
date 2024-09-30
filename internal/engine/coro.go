package engine

import (
	. "godot-ext/gdspx/pkg/engine"
	"time"
)

const (
	maxExecTime = 16 * time.Millisecond
)

var (
	jobQueue  = make(chan Job, 1)
	gameFrame = 0
)

type Job func()

func handleEngineCoroutines() {
	startTime := time.Now()
	timer := time.NewTimer(maxExecTime)
	defer timer.Stop()

	for {
		isTimeout := false
		select {
		case job, ok := <-jobQueue:
			if !ok {
				return
			}
			job()
		case <-timer.C:
			isTimeout = true
			break
		}

		if isTimeout {
			break
		}
		if time.Since(startTime) > maxExecTime {
			break
		}
	}
}
