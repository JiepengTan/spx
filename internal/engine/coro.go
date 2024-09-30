package engine

import (
	"time"
)

const (
	maxExecTime = 16 * time.Millisecond
)

var (
	updateJobQueue = make(chan Job, 1)
)

type Job func()

func handleEngineCoroutines() {
	startTime := time.Now()
	timer := time.NewTimer(maxExecTime)
	defer timer.Stop()

	for {
		isTimeout := false
		select {
		case job, ok := <-updateJobQueue:
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
