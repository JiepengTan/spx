package coroutine

import (
	"errors"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	stime "time"
	"unsafe"

	"github.com/goplus/spx/internal/engine/platform"
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
	mutex    sync.Mutex // Mutex for this thread's condition variable
	cond     *sync.Cond // Per-thread condition variable for targeted wake-up
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
	id.cond = sync.NewCond(&id.mutex) // Initialize the thread's condition variable
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
	p.mutex.Unlock()

	// Wait on the thread's own condition variable instead of the shared one
	me.mutex.Lock()
	for p.isSuspended(me) {
		me.cond.Wait()
	}
	me.mutex.Unlock()

	p.waitNotify()

	p.sema.Lock()

	p.setCurrent(me)
	if me.stopped_ { // check stopped
		panic(ErrAbortThread)
	}
}

// isSuspended checks if a thread is suspended (thread-safe)
func (p *Coroutines) isSuspended(me Thread) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.suspended[me]
}

// Resume resumes a suspended coroutine.
func (p *Coroutines) Resume(me Thread) {
	for {
		done := false
		p.mutex.Lock()
		if p.suspended[me] {
			p.suspended[me] = false
			done = true
		}
		p.mutex.Unlock()

		if done {
			// Signal only the specific thread's condition variable
			me.mutex.Lock()
			me.cond.Signal()
			me.mutex.Unlock()
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
	me := p.Current()
	dstTime := time.TimeSinceLevelLoad() + t

	// Set up the job directly without a goroutine
	job := &WaitJob{
		Id:   atomic.AddInt64(&p.curId, 1),
		Type: waitTypeTime,
		Call: func() {
			// Mark thread as idle and resume it directly
			p.setWaitStatus(me, waitStatusIdle)
			p.Resume(me)
		},
		Time: dstTime,
	}

	// Add the job to the queue
	p.addWaitJob(job, false)

	// Mark the thread as blocked and yield control
	p.setWaitStatus(me, waitStatusBlock)
	p.Yield(me)
}

func (p *Coroutines) WaitNextFrame() {
	me := p.Current()
	frame := time.Frame()

	// Set up the job directly instead of creating a goroutine
	job := &WaitJob{
		Id:   atomic.AddInt64(&p.curId, 1),
		Type: waitTypeFrame,
		Call: func() {
			// Mark thread as idle and resume it directly
			p.setWaitStatus(me, waitStatusIdle)
			p.Resume(me)
		},
		Frame: frame,
	}

	// Add the job to the queue
	p.addWaitJob(job, false)

	// Mark the thread as blocked and yield control
	p.setWaitStatus(me, waitStatusBlock)
	p.Yield(me)
}

func (p *Coroutines) WaitMainThread(call func()) {
	if platform.IsWeb() {
		call()
		return
	}
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

	// Create a goroutine that executes the function
	// This is necessary since fn() could be a long-running task
	go func() {
		// Execute the function
		fn()
		// When done, mark thread as idle and resume it
		p.setWaitStatus(me, waitStatusIdle)
		p.Resume(me)
	}()

	// Mark the thread as blocked and yield control
	p.setWaitStatus(me, waitStatusBlock)
	p.Yield(me)
}

// 全局变量，存储最近一次更新的统计信息
var lastUpdateStats UpdateJobsStats

// GetLastUpdateStats 返回最近一次更新的统计信息
func (p *Coroutines) GetLastUpdateStats() UpdateJobsStats {
	return lastUpdateStats
}

func (p *Coroutines) UpdateJobs() {
	// 总计时开始
	start := stime.Now()

	// 记录GC信息
	var gcStatsBefore debug.GCStats
	debug.ReadGCStats(&gcStatsBefore)

	// 初始化统计信息
	stats := UpdateJobsStats{}

	// 初始化阶段开始
	initStart := stime.Now()
	curQueue := p.curQueue
	nextQueue := p.nextQueue
	curFrame := time.Frame()
	curTime := time.RealTimeSinceStart()
	debugStartTime := time.RealTimeSinceStart()
	waitFrameCount := 0
	waitMainCount := 0
	// 初始化阶段结束
	stats.InitTime = stime.Since(initStart).Seconds() * 1000

	// 主循环开始
	loopStart := stime.Now()
	// 循环迭代计数器
	loopIterCount := 0
	for {
		// 记录每次循环迭代的开始时间
		_ = stime.Now() // 不再使用loopIterStart变量
		loopIterCount++

		if !p.hasInited {
			if curQueue.Count() == 0 {
				waitStart := stime.Now()
				time.Sleep(0.05) // 0.05ms
				stats.WaitTime += stime.Since(waitStart).Seconds() * 1000
				continue
			}
		} else {
			done := false
			isContinue := false

			waitStart := stime.Now()
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
			stats.WaitTime += stime.Since(waitStart).Seconds() * 1000

			if done {
				break
			}
			if isContinue {
				continue
			}
		}

		// 任务处理开始
		taskStart := stime.Now()
		task := curQueue.PopFront()
		stats.TaskCounts++

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
		stats.TaskProcessing += stime.Since(taskStart).Seconds() * 1000

		if time.RealTimeSinceStart()-debugStartTime > 1 {
			println("Warning: engine update > 1 seconds, please check your code ! waitMainCount=", waitMainCount)
			break
		}
	}
	// 主循环结束
	_ = stime.Now() // 不再使用loopEndTime变量
	stats.LoopTime = stime.Since(loopStart).Seconds() * 1000

	// 队列移动开始
	moveStart := stime.Now()
	stats.NextCount = nextQueue.Count()
	curQueue.Move(nextQueue)
	stats.MoveTime = stime.Since(moveStart).Seconds() * 1000

	// 更新统计信息
	stats.WaitFrameCount = waitFrameCount
	stats.WaitMainCount = waitMainCount

	// 获取GC统计
	var gcStatsAfter debug.GCStats
	debug.ReadGCStats(&gcStatsAfter)
	stats.GCCount = int(gcStatsAfter.NumGC - gcStatsBefore.NumGC)
	stats.GCPauses = float64(gcStatsAfter.PauseTotal-gcStatsBefore.PauseTotal) / float64(stime.Millisecond)

	// 计算总时间
	_ = stime.Now() // 不再使用totalEndTime变量
	delta := stime.Since(start).Seconds() * 1000

	// 计算实际测量的时间与各部分时间之和的差异
	measuredTotal := delta
	sumParts := stats.InitTime + stats.LoopTime + stats.MoveTime
	timeDiff := measuredTotal - sumParts

	// 计算主循环外部的时间（可能包含Go运行时调度开销）
	externalTime := delta - sumParts

	// 更新统计信息
	stats.ExternalTime = externalTime
	stats.LoopIterations = loopIterCount
	stats.TotalTime = delta
	stats.TimeDifference = timeDiff

	// 保存统计信息供外部访问
	lastUpdateStats = stats

	// 输出基本信息
	//fmt.Printf("curFrame %d,useTime %.3fms,fps %d, taskCount %d,curTime %.3f, moveCount %d\n",curFrame, delta, int(time.FPS()), waitFrameCount, curTime, stats.NextCount)
}

// -------------------------------------------------------------------------------------
