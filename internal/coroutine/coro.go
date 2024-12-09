package coroutine

import (
	"errors"
	"fmt"
	"runtime"
	"sort"
	"strings"
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
	id       int64
	Obj      ThreadObj
	frame    int
	stopped_ bool

	stack_       string
	stackSimple_ string
	timer_       float64
	duration_    float64
	isWaitFrame_ bool
	isActive_    bool
}

var (
	gThreadId    int64
	gCurThreadId int64
)

func newThreadImpl(isDebug bool, obj ThreadObj, frame int) *threadImpl {
	id := &threadImpl{Obj: obj, frame: frame}
	id.id = atomic.AddInt64(&gThreadId, 1)
	if isDebug {
		buf := make([]byte, 1024)
		n := runtime.Stack(buf, false)
		id.stack_ = fmt.Sprintf("%s\n", buf[:n])

		stackIdx := 4 // print the last stack
		lines := strings.Split(id.stack_, "\n")
		if stackIdx*2 <= len(lines) {
			id.stackSimple_ = lines[stackIdx*2-1] + " " + lines[stackIdx*2]
		}

	}
	return id
}

func (p *threadImpl) onFrameStart() {
	p.timer_ = time.RealTimeSinceStart()
}
func (p *threadImpl) onFrameStop() {
	p.duration_ = time.RealTimeSinceStart() - p.timer_
}
func (p *threadImpl) String() string {
	dt := p.duration_ * 1000
	if dt > 3 {
		return fmt.Sprintf("duration %fms isWaitFrame %t  stack:%s", dt, p.isWaitFrame_, p.stack_)
	}
	return fmt.Sprintf("duration %fms isWaitFrame %t  stack:%s ", dt, p.isWaitFrame_, p.stackSimple_)
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
	Id     int64
	Type   int
	call   func()
	Time   float64
	Frame  int64
	thread Thread
}

func (job *WaitJob) Call() {
	timer := debugTimerUs()
	job.call()
	printWithTimer("WaitJob::Call", (debugTimerUs() - timer))
}

// New creates a coroutine manager.
func New() *Coroutines {
	p := &Coroutines{
		suspended: make(map[Thread]bool),
		waiting:   make(map[Thread]bool),
		debug:     true,
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

// CreateAndStart creates and executes the new coroutine.
func (p *Coroutines) CreateAndStart(start bool, tobj ThreadObj, fn func(me Thread) int) Thread {
	id := newThreadImpl(p.debug, tobj, p.frame)
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
	gCurThreadId = me.id
	printTimerWithId(me.id, "Yield start")
	p.sema.Unlock()
	p.mutex.Lock()
	me.onFrameStop()
	p.suspended[me] = true
	for p.suspended[me] {
		p.cond.Wait()
	}
	me.onFrameStart()
	p.mutex.Unlock()

	p.sema.Lock()
	printTimerWithId(me.id, "Yield Done")
	gCurThreadId = me.id
	p.setCurrent(me)
	if me.stopped_ { // check stopped
		panic(ErrAbortThread)
	}
	printTimerWithId(me.id, "Yield waitNotify start")
	p.waitNotify()
	printTimerWithId(me.id, "Yield waitNotify done")
}

// Resume resumes a suspended coroutine.
func (p *Coroutines) Resume(me Thread) {
	printTimerWithId(me.id, "Resume start")
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
			printTimerWithId(me.id, "Resume Done")
			return
		}
		runtime.Gosched()
	}
}

func (p *Coroutines) addWaitJob(job *WaitJob, isFront bool) {
	printWithTimer("addWaitJob")
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
	printWithTimer("waitNotify")
	p.waitMutex.Lock()
	p.waitCond.Signal()
	p.waitMutex.Unlock()
}

func (p *Coroutines) setWaitStatus(me *threadImpl, typeId int) {
	printTimerWithId(me.id, "setWaitStatus ", typeId)
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
			thread: me,
			Id:     id,
			Type:   waitTypeTime,
			call: func() {
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
	printTimerWithId(me.id, "WaitNextFrame")
	frame := time.Frame()
	go func() {
		printTimerWithId(me.id, "WaitNextFrame go start")
		done := make(chan int)
		job := &WaitJob{
			thread: me,
			Id:     id,
			Type:   waitTypeFrame,
			call: func() {
				p.setWaitStatus(me, waitStatusIdle)
				done <- 1
			},
			Frame: frame,
		}
		p.addWaitJob(job, false)
		printTimerWithId(me.id, "WaitNextFrame go blocking")
		<-done
		printTimerWithId(me.id, "WaitNextFrame go block done")
		me.isWaitFrame_ = false
		p.Resume(me)
	}()
	me.isWaitFrame_ = true
	printTimerWithId(me.id, "WaitNextFrame setWaitStatus")
	p.setWaitStatus(me, waitStatusBlock)
	printTimerWithId(me.id, "WaitNextFrame setWaitStatus done")
	p.Yield(me)
}

func (p *Coroutines) WaitMainThread(call func()) {

	call()
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

func WaitForChan[T any](p *Coroutines, done chan T, data *T) {
	me := p.Current()
	go func() {
		*data = <-done
		p.setWaitStatus(me, waitStatusIdle)
		p.Resume(me)
	}()
	p.setWaitStatus(me, waitStatusBlock)
	p.Yield(me)
}

var sbuilder strings.Builder

func DebugLog(info ...any) {
	printWithTimer(" debugLog ", info...)
}
func printTimerWithId(id int64, msg string, args ...any) {
	argsStr := fmt.Sprint(args...)
	val := fmt.Sprint(id, " "+msg+"  ", debugTimerUs(), argsStr)
	sbuilder.WriteString(val)
	sbuilder.WriteString("\n")
}
func printWithTimer(msg string, args ...any) {
	printTimerWithId(gCurThreadId, msg, args...)
}

func debugTimerUs() int64 {
	return int64(time.RealTimeSinceStart() * 1000000)
}

func (p *Coroutines) UpdateJobs() {
	curQueue := p.curQueue
	nextQueue := p.nextQueue
	curFrame := time.Frame()
	curTime := time.TimeSinceLevelLoad()
	startTimer := time.RealTimeSinceStart()
	waitFrameCount := 0
	waitMainCount := 0
	waitSecCount := 0

	printWithTimer("UpdateJobs start")
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
					timer := debugTimerUs()
					printWithTimer("Wait bfeore 11")
					runtime.Gosched()
					p.waitCond.Wait()
					printWithTimer("Wait after ", (debugTimerUs() - timer))
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
		printTimerWithId(p.Current().id, "Updatetask start ", task.Type)
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
			waitSecCount++
		case waitTypeMainThread:
			task.Call()
			waitMainCount++
		}
		printTimerWithId(p.Current().id, "Updatetask done")
		if time.RealTimeSinceStart()-startTimer > 1 {
			println("Warning: engine update > 1 seconds, please check your code ! waitMainCount=", waitMainCount)
			p.printCorotineInfos()
			break
		}
	}

	printWithTimer("UpdateJobs done")
	nextCount := nextQueue.Count()
	curQueue.Move(nextQueue)
	dtms := (time.RealTimeSinceStart() - startTimer) * 1000
	if p.debug {
		if dtms > 10 {
			fmt.Printf("curFrame %d,useTime %fms,fps %d, curTime %f,\n waitFrameCount%d waitSecCount%d waitMainCount%d moveCount %d \n",
				curFrame, dtms, int(time.FPS()), curTime, waitFrameCount, waitFrameCount, waitSecCount, nextCount)
			p.printCorotineInfos()
		}
	}
	printWithTimer("UpdateJobs Done before print")
	// flush debug infos
	println(sbuilder.String())
	sbuilder.Reset()
}

func (p *Coroutines) printCorotineInfos() {
	var sb strings.Builder
	idx := 0
	p.waitMutex.Lock()
	ths := make([]*threadImpl, len(p.waiting))
	for thread, val := range p.waiting {
		thread.isActive_ = !val
		ths[idx] = thread
		idx++
	}
	p.waitMutex.Unlock()
	sort.Slice(ths, func(i, j int) bool {
		return ths[i].duration_ > ths[j].duration_
	})

	for i, thread := range ths {
		sb.WriteString(fmt.Sprintf(" %d, isActive %t, %s \n", i, thread.isActive_, thread.String()))
	}

	msg := fmt.Sprintf("printCorotineInfos coroCount= %d frame:%d deltaTime %d \n %s", len(p.waiting), time.Frame(), int(time.DeltaTime()*1000), sb.String())
	println(msg)
}

// -------------------------------------------------------------------------------------
