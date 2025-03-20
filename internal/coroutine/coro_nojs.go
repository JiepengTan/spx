//go:build !js
// +build !js

package coroutine

import (
	"errors"
	"fmt"
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

// UpdateJobsStats 存储协程更新的详细统计信息
type UpdateJobsStats struct {
	InitTime       float64 // 初始化时间
	LoopTime       float64 // 主循环时间
	MoveTime       float64 // 队列移动时间
	WaitTime       float64 // 等待时间
	TaskProcessing float64 // 任务处理时间
	GCPauses       float64 // GC暂停时间
	ExternalTime   float64 // 外部时间（可能包含调度开销）
	TotalTime      float64 // 总计时间
	TimeDifference float64 // 时间差异
	TaskCounts     int     // 处理的任务数量
	WaitFrameCount int     // 等待帧的数量
	WaitMainCount  int     // 等待主线程的数量
	NextCount      int     // 下一帧队列数量
	GCCount        int     // GC次数
	LoopIterations int     // 循环迭代次数
}

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

	waiting   map[Thread]bool
	debug     bool
	waitMutex sync.Mutex
	waitCond  sync.Cond
}

const (
	waitStatusAdd = iota
	waitStatusDelete
	waitStatusBlock
	waitStatusIdle
	waitNotify
)

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
		suspended: make(map[Thread]bool),
		waiting:   make(map[Thread]bool),
		debug:     false,
	}
	p.cond.L = &p.mutex
	p.curQueue = NewQueue[*WaitJob]()
	p.nextQueue = NewQueue[*WaitJob]()
	p.hasInited = false
	p.waitCond.L = &p.waitMutex
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
			p.setWaitStatus(id, waitStatusDelete)
			p.sema.Unlock()
			if e := recover(); e != nil {
				if e != ErrAbortThread {
					panic(e)
				}
			}
		}()
		p.setWaitStatus(id, waitStatusAdd)
		fn(id)
	}()
	if start {
		runtime.Gosched()
	}
	return id
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

	p.waitNotify()

	p.sema.Lock()

	p.setCurrent(me)
	if me.stopped_ { // check stopped
		panic(ErrAbortThread)
	}
}

// Resume resumes a suspended coroutine.
func (p *Coroutines) Resume(me Thread) {
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

func (p *Coroutines) addWaitJob(job *WaitJob, isFront bool) {
	p.waitMutex.Lock()
	if isFront {
		p.curQueue.PushFront(job)
	} else {
		p.curQueue.PushBack(job)
	}
	p.waitCond.Signal()
	p.waitMutex.Unlock()
}

func (p *Coroutines) waitNotify() {
	p.waitMutex.Lock()
	p.waitCond.Signal()
	p.waitMutex.Unlock()
}

func (p *Coroutines) setWaitStatus(me *threadImpl, typeId int) {
	p.waitMutex.Lock()
	switch typeId {
	case waitStatusDelete:
		delete(p.waiting, me)
	case waitStatusAdd:
		p.waiting[me] = false
	case waitStatusBlock:
		p.waiting[me] = true
	case waitStatusIdle:
		p.waiting[me] = false
	}
	p.waitCond.Signal()
	p.waitMutex.Unlock()
}

func (p *Coroutines) Wait(t float64) {
	id := atomic.AddInt64(&p.curId, 1)
	me := p.Current()
	dstTime := time.TimeSinceLevelLoad() + t
	go func() {
		done := make(chan int)
		job := &WaitJob{
			Id:   id,
			Type: waitTypeTime,
			Call: func() {
				p.setWaitStatus(me, waitStatusIdle)
				done <- 1
			},
			Time: dstTime,
		}
		p.addWaitJob(job, false)
		<-done
		p.Resume(me)
	}()
	p.setWaitStatus(me, waitStatusBlock)
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
				p.setWaitStatus(me, waitStatusIdle)
				done <- 1
			},
			Frame: frame,
		}
		p.addWaitJob(job, false)
		<-done
		p.Resume(me)
	}()
	p.setWaitStatus(me, waitStatusBlock)
	p.Yield(me)
}

func (p *Coroutines) WaitMainThread(call func()) {
	id := atomic.AddInt64(&p.curId, 1)
	done := make(chan int)
	job := &WaitJob{
		Id:   id,
		Type: waitTypeMainThread,
		Call: func() {
			call()
			done <- 1
		},
	}
	// main thread call's priority is higher than other wait jobs
	p.addWaitJob(job, true)
	<-done
}

func (p *Coroutines) WaitToDo(fn func()) {
	me := p.Current()
	go func() {
		fn()
		p.setWaitStatus(me, waitStatusIdle)
		p.Resume(me)
	}()
	p.setWaitStatus(me, waitStatusBlock)
	p.Yield(me)
}

func (p *Coroutines) UpdateJobs() {
	timestamp := time.RealTimeSinceStart()
	curQueue := p.curQueue
	nextQueue := p.nextQueue
	curFrame := time.Frame()
	curTime := time.TimeSinceLevelLoad()
	debugStartTime := time.RealTimeSinceStart()
	waitFrameCount := 0
	waitMainCount := 0
	for {
		if !p.hasInited {
			if curQueue.Count() == 0 {
				time.Sleep(0.05) // 0.05ms
				continue
			}
		} else {
			done := false
			isContinue := false
			p.waitMutex.Lock()
			if curQueue.Count() == 0 {
				activeCount := 0
				for _, val := range p.waiting {
					if !val {
						activeCount++
					}
				}
				if activeCount == 0 {
					done = true
				} else {
					p.waitCond.Wait()
					isContinue = true
				}
			}
			p.waitMutex.Unlock()
			if done {
				break
			}
			if isContinue {
				continue
			}
		}

		task := curQueue.PopFront()
		switch task.Type {
		case waitTypeFrame:
			if task.Frame >= curFrame {
				nextQueue.PushBack(task)
			} else {
				task.Call()
				waitFrameCount++
			}
		case waitTypeTime:
			if task.Time >= curTime {
				nextQueue.PushBack(task)
			} else {
				task.Call()
			}
		case waitTypeMainThread:
			task.Call()
			waitMainCount++
		}
		if time.RealTimeSinceStart()-debugStartTime > 1 {
			println("Warning: engine update > 1 seconds, please check your code ! waitMainCount=", waitMainCount)
			break
		}
	}
	nextCount := nextQueue.Count()
	curQueue.Move(nextQueue)
	delta := (time.RealTimeSinceStart() - timestamp) * 1000
	fmt.Printf("curFrame %d,useTime %fms,fps %d, taskCount %d,curTime %f , moveCount %d \n", curFrame, delta, int(time.FPS()), waitFrameCount, curTime, nextCount)

}

// 全局变量，存储最近一次更新的统计信息
var lastUpdateStats UpdateJobsStats

// GetLastUpdateStats 返回最近一次更新的统计信息
func (p *Coroutines) GetLastUpdateStats() UpdateJobsStats {
	return lastUpdateStats
}

// -------------------------------------------------------------------------------------
