package coroutine

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
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
	suspended map[Thread]bool
	current   Thread
	mutex     sync.Mutex
	cond      sync.Cond
	sema      sync.Mutex
	frame     int
	curQueue  *Queue[*WaitJob]
	nextQueue *Queue[*WaitJob]
}

const (
	waitTypeFrame = iota
	waitTypeTime
	waitTypeMainThread
)

type WaitJob struct {
	Type  int
	Call  func()
	Time  float64
	Frame int64
}

// New creates a coroutine manager.
func New() *Coroutines {
	p := &Coroutines{
		suspended: make(map[Thread]bool),
	}
	p.cond.L = &p.mutex
	p.curQueue = NewQueue[*WaitJob]()
	p.nextQueue = NewQueue[*WaitJob]()
	return p
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
			p.mutex.Unlock()
			p.sema.Unlock()
			if e := recover(); e != nil {
				if e != ErrAbortThread {
					panic(e)
				}
			}
		}()
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

// Yield suspends a running coroutine.
func (p *Coroutines) Yield(me Thread) {
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
func (p *Coroutines) Resume(th Thread) {
	for {
		done := false
		p.mutex.Lock()
		if p.suspended[th] {
			p.suspended[th] = false
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

// Sched func.
func (p *Coroutines) Sched(me Thread) {
	go func() {
		p.Resume(me)
	}()
	p.Yield(me)
}

func (p *Coroutines) Wait(t float64) {
	me := p.Current()
	dstTime := time.TimeSinceLevelLoad() + t
	go func() {
		done := make(chan int)
		job := &WaitJob{
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
	me := p.Current()
	frame := time.Frame()
	go func() {
		done := make(chan int)
		job := &WaitJob{
			Type: waitTypeFrame,
			Call: func() {
				done <- 1
			},
			Frame: frame,
		}
		p.curQueue.Enqueue(job)
		<-done
		p.Resume(me)
	}()
	p.Yield(me)
}

func (p *Coroutines) WaitMainThread() {
	me := p.Current()
	go func() {
		done := make(chan int)
		job := &WaitJob{
			Type: waitTypeMainThread,
			Call: func() {
				done <- 1
			},
		}
		p.curQueue.Enqueue(job)
		<-done
		p.Resume(me)
	}()
	p.Yield(me)
}

func (p *Coroutines) HandleJobs() {
	curQueue := p.curQueue
	nextQueue := p.nextQueue
	curFrame := time.Frame()
	curTime := time.TimeSinceLevelLoad()
	startTime := time.RealTimeSinceStart()

	for curQueue.Count() > 0 {
		task, _ := curQueue.Dequeue()
		switch task.Type {
		case waitTypeFrame:
			if task.Frame >= curFrame {
				nextQueue.Enqueue(task)
			} else {
				task.Call()
			}
		case waitTypeTime:
			if task.Time >= curTime {
				nextQueue.Enqueue(task)
			} else {
				task.Call()
			}
		case waitTypeMainThread:
			task.Call()
		}
		if time.RealTimeSinceStart()-startTime > 1 {
			println("Warning: engine update > 1 seconds, please check your code !")
		}
	}
	curQueue.Move(nextQueue)
}

// -------------------------------------------------------------------------------------
