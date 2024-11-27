package engine

import (
	"github.com/goplus/spx/internal/coroutine"
	"github.com/goplus/spx/internal/time"
)

var (
	gco *coroutine.Coroutines
)

func SetCoroutines(co *coroutine.Coroutines) {
	gco = co
}

func Wait(secs float64) float64 {
	startTime := time.TimeSinceLevelLoad()
	gco.Wait(secs)
	return time.TimeSinceLevelLoad() - startTime
}
func WaitNextFrame() float64 {
	startTime := time.RealTimeSinceStart()
	gco.WaitNextFrame()
	return time.RealTimeSinceStart() - startTime
}
func CallOnMainThread(call func()) {
	gco.CallOnMainThread(call)
}

func updateCoroutines() {
	gco.HandleJobs()
}
