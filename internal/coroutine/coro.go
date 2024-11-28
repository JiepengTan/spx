package coroutine

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	stime "time"
	"unsafe"

	"github.com/goplus/spx/internal/time"
)

var (
	// ErrCannotYieldANonrunningThread represents an "can not yield a non-running thread" error.
	ErrCannotYieldANonrunningThread = errors.New("can not yield a non-running thread")
	ErrAbortThread                  = errors.New("abort thread")
)

// -------------------------------------------------------------------------------------

type ThreadObj interface {
}

type threadImpl struct {
	Obj      ThreadObj
	stopped_ bool
	frame    int
}

func (p *threadImpl) Stopped() bool {
	return p.stopped_
}

// Thread represents a coroutine id.
type Thread = *threadImpl

// Coroutines represents a coroutine manager.
type Coroutines struct {
	hasInited bool
	suspended map[Thread]bool
	current   Thread
	mutex     sync.Mutex
	cond      sync.Cond
	sema      sync.Mutex
	frame     int
	curQueue  *Queue[*WaitJob]
	nextQueue *Queue[*WaitJob]
	curId     int64

	waitingFrame map[Thread]bool
}

const (
	waitTypeFrame = iota
	waitTypeTime
	waitTypeMainThread
)

type WaitJob struct {
	Id    int64
	Type  int
	Call  func()
	Time  float64
	Frame int64
}

// New creates a coroutine manager.
func New() *Coroutines {
	p := &Coroutines{
		suspended:    make(map[Thread]bool),
		waitingFrame: make(map[Thread]bool),
	}
	p.cond.L = &p.mutex
	p.curQueue = NewQueue[*WaitJob]()
	p.nextQueue = NewQueue[*WaitJob]()
	p.hasInited = false
	return p
}

func (p *Coroutines) OnInited() {
	p.hasInited = true
}

// Create creates a new coroutine.
func (p *Coroutines) Create(tobj ThreadObj, fn func(me Thread) int) Thread {
	return p.CreateAndStart(false, tobj, fn)
}

func (p *Coroutines) setCurrent(id Thread) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&p.current)), unsafe.Pointer(id))
}

func (p *Coroutines) Current() Thread {
	return Thread(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&p.current))))
}

// CreateAndStart creates and executes the new coroutine.
func (p *Coroutines) CreateAndStart(start bool, tobj ThreadObj, fn func(me Thread) int) Thread {
	id := &threadImpl{Obj: tobj, frame: p.frame}
	go func() {
		p.sema.Lock()
		p.setCurrent(id)
		defer func() {
			p.mutex.Lock()
			delete(p.suspended, id)
			delete(p.waitingFrame, id)

			p.mutex.Unlock()
			p.sema.Unlock()
			if e := recover(); e != nil {
				if e != ErrAbortThread {
					panic(e)
				}
			}
		}()
		p.mutex.Lock()
		p.waitingFrame[id] = false
		p.mutex.Unlock()
		fn(id)
	}()
	if start {
		runtime.Gosched()
	}
	return id
}

func (p *Coroutines) Abort() {
	panic(ErrAbortThread)
}

func (p *Coroutines) StopIf(filter func(th Thread) bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for th := range p.suspended {
		if filter(th) {
			th.stopped_ = true
		}
	}
}
func (p *Coroutines) GetActiveCount() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	count := 0
	for _, val := range p.waitingFrame {
		if !val {
			count++
		}
	}
	return count
}

// Yield suspends a running coroutine.
func (p *Coroutines) Yield(me Thread) {
	p.yield(me, true)
}

func (p *Coroutines) Resume(me Thread) {
	p.resume(me, true)
}

func (p *Coroutines) yield(me Thread, isWaiting bool) {
	if isWaiting {
		p.mutex.Lock()
		p.waitingFrame[me] = true
		p.mutex.Unlock()
	}
	if p.Current() != me {
		panic(ErrCannotYieldANonrunningThread)
	}
	p.sema.Unlock()
	p.mutex.Lock()
	p.suspended[me] = true
	for p.suspended[me] {
		p.cond.Wait()
	}
	p.mutex.Unlock()
	p.sema.Lock()

	p.setCurrent(me)
	if me.stopped_ { // check stopped
		panic(ErrAbortThread)
	}
}

// Resume resumes a suspended coroutine.
func (p *Coroutines) resume(me Thread, isWaiting bool) {
	if isWaiting {
		p.mutex.Lock()
		p.waitingFrame[me] = false
		p.mutex.Unlock()
	}
	for {
		done := false
		p.mutex.Lock()
		if p.suspended[me] {
			p.suspended[me] = false
			p.cond.Broadcast()
			done = true
		}
		p.mutex.Unlock()
		if done {
			return
		}
		runtime.Gosched()
	}
}

func (p *Coroutines) Wait(t float64) {
	id := atomic.AddInt64(&p.curId, 1)
	//println("Wait ", p.curId, id)
	me := p.Current()
	dstTime := time.TimeSinceLevelLoad() + t
	go func() {
		done := make(chan int)
		job := &WaitJob{
			Id:   id,
			Type: waitTypeTime,
			Call: func() {
				done <- 1
			},
			Time: dstTime,
		}
		p.curQueue.Enqueue(job)
		<-done
		p.Resume(me)
	}()
	p.Yield(me)
}

func (p *Coroutines) WaitNextFrame() {
	id := atomic.AddInt64(&p.curId, 1)
	me := p.Current()
	frame := time.Frame()
	go func() {
		done := make(chan int)
		job := &WaitJob{
			Id:   id,
			Type: waitTypeFrame,
			Call: func() {
				done <- 1
			},
			Frame: frame,
		}
		p.curQueue.Enqueue(job)
		<-done
		if time.Frame()-frame > 1 {
			println("Warning!!!: WaitNextFrame wait too many frames, count=", time.Frame()-frame, "id", id)
		}
		p.Resume(me)
	}()
	p.Yield(me)
}

func (p *Coroutines) WaitMainThread(call func()) {
	//id := atomic.AddInt64(&p.curId, 1)
	me := p.Current()
	coro := func(isResume bool) {
		done := make(chan int)
		job := &WaitJob{
			Id:   0,
			Type: waitTypeMainThread,
			Call: func() {
				call()
				done <- 1
			},
		}
		p.curQueue.Enqueue(job)
		<-done
		if isResume {
			// main thread call does NOT count as blocking
			p.resume(me, false)
		}
	}
	if p.hasInited {
		go coro(true)
		// main thread call does NOT count as blocking
		p.yield(me, false)
	} else {
		coro(false)
	}
}
func (p *Coroutines) WaitToDo(fn func()) {
	me := p.Current()
	go func() {
		fn()
		p.Resume(me)
	}()
	p.Yield(me)
}
func WaitForChan[T any](p *Coroutines, done chan T, data *T) {
	me := p.Current()
	go func() {
		*data = <-done
		p.Resume(me)
	}()
	p.Yield(me)
}

func (p *Coroutines) HandleJobs() {
	timer := time.RealTimeSinceStart()
	msg := p.handleJobs()
	delta := (time.RealTimeSinceStart() - timer) * 1000
	fmt.Printf("HandleJobs use time %f ms %s \n", delta, msg)
}
func (p *Coroutines) handleJobs() string {
	curQueue := p.curQueue
	nextQueue := p.nextQueue
	curFrame := time.Frame()
	curTime := time.TimeSinceLevelLoad()
	debugStartTime := time.RealTimeSinceStart()
	taskCount := 0
	//println("===== HandleJobs ======")
	for !p.hasInited || curQueue.Count() > 0 ||
		p.GetActiveCount() > 0 {
		if curQueue.Count() == 0 {
			stime.Sleep(stime.Microsecond * 100) // sleep 0.05 ms
			continue
		}
		task := curQueue.Dequeue()
		switch task.Type {
		case waitTypeFrame:
			if task.Frame >= curFrame {
				nextQueue.Enqueue(task)
			} else {
				task.Call()
				taskCount++
				//println("frame call: ", task.Id)
			}
		case waitTypeTime:
			//println("wait call:", task.Id)
			if task.Time >= curTime {
				nextQueue.Enqueue(task)
			} else {
				task.Call()
			}
		case waitTypeMainThread:
			//println("main call:", task.Id)
			task.Call()
		}
		if time.RealTimeSinceStart()-debugStartTime > 1 {
			println("Warning: engine update > 1 seconds, please check your code !")
		}
	}
	curQueue.Move(nextQueue)
	return fmt.Sprintf("curFrame%d \t,taskCount %d\t ,curTime %f , moveCount %d\t", curFrame, taskCount, curTime, nextQueue.Count())

}

// -------------------------------------------------------------------------------------
