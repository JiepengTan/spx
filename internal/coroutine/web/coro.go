//go:build js
// +build js

package web

import (
	"sync"
	"sync/atomic"
	"time"
)

// State represents the coroutine state
type State int

const (
	StateRunning State = iota
	StateWaiting
	StateStopped
)

// CoroutineFunc defines the coroutine function type
type CoroutineFunc func(Thread) int

// Thread interface defines interaction methods between coroutine and scheduler
type Thread interface {
	WaitNextFrame() bool
	Wait(seconds float64) bool
	Current() *Coroutine
}

// Coroutine represents a coroutine
type Coroutine struct {
	id            int           // Unique identifier for the coroutine
	fn            CoroutineFunc // Coroutine function
	state         State         // Current state of the coroutine
	waitUntil     time.Time     // Time until which the coroutine waits
	owner         interface{}   // Owner of the coroutine
	stopped       bool          // Whether the coroutine has been stopped
	result        int           // Result of coroutine execution
	resumeCh      chan struct{} // Channel for suspending and resuming the coroutine
	started       bool          // Flag indicating if the coroutine has started
	goroutineDone bool          // Flag indicating if the goroutine has completed
}

// Coroutines is the coroutine scheduler
type Coroutines struct {
	readyQueue   []*Coroutine // Ready queue
	waitingQueue []*Coroutine // Waiting queue
	current      atomic.Value // Currently running coroutine
	nextID       int          // Next coroutine ID
	mu           sync.Mutex   // Mutex to protect queue operations
	timer        *Timer       // Timer for simulating Unity's frame loop
}

// StopFilterFunc defines the coroutine stop filter function type
type StopFilterFunc func(*Coroutine) bool

// New creates a new coroutine scheduler
func New() *Coroutines {
	c := &Coroutines{
		readyQueue:   make([]*Coroutine, 0),
		waitingQueue: make([]*Coroutine, 0),
		nextID:       1,
	}
	c.current.Store((*Coroutine)(nil))
	return c
}

// SetTimer sets the frame loop timer
func (c *Coroutines) SetTimer(timer *Timer) {
	c.timer = timer
	timer.AddCallback(c.Update)
}

// Create creates a new coroutine
func (c *Coroutines) Create(owner interface{}, fn CoroutineFunc, start bool) *Coroutine {
	c.mu.Lock()

	coro := &Coroutine{
		id:            c.nextID,
		fn:            fn,
		state:         StateStopped,
		owner:         owner,
		stopped:       false,
		resumeCh:      make(chan struct{}, 1),
		started:       false,
		goroutineDone: false,
	}
	c.nextID++

	if start {
		coro.state = StateRunning
		// If a new coroutine is created within a coroutine, put it at the front of the ready queue
		if c.Current() != nil {
			c.readyQueue = append([]*Coroutine{coro}, c.readyQueue...)
			// Store the currently running coroutine
			currentCoro := c.Current()
			// Temporarily clear the current coroutine so we can run the newly created one
			c.current.Store((*Coroutine)(nil))
			c.mu.Unlock()

			// Immediately run the newly created coroutine
			c.RunNext()

			// Restore the current coroutine
			c.mu.Lock()
			c.current.Store(currentCoro)
		} else {
			c.readyQueue = append(c.readyQueue, coro)
		}
	}

	c.mu.Unlock()
	return coro
}

// Current gets the currently running coroutine
func (c *Coroutines) Current() *Coroutine {
	return c.current.Load().(*Coroutine)
}

// WaitNextFrame moves the current coroutine to the waiting queue, waiting for the next frame
func (c *Coroutines) WaitNextFrame() bool {
	current := c.Current()
	if current == nil || current.stopped {
		return false
	}

	c.mu.Lock()
	// Check if the coroutine has been stopped
	if current.stopped || current.goroutineDone {
		c.mu.Unlock()
		return false
	}

	current.state = StateWaiting
	current.waitUntil = time.Now().Add(time.Nanosecond) // Resume immediately next frame
	c.waitingQueue = append(c.waitingQueue, current)
	c.current.Store((*Coroutine)(nil))
	c.mu.Unlock()

	// Block until resumed by the scheduler or timeout
	select {
	case <-current.resumeCh:
		// Check if the coroutine was stopped during waiting
		if current.stopped || current.goroutineDone {
			return false
		}
		return true
	case <-time.After(time.Millisecond * 500):
		// In a test environment, we might need a shorter timeout
		// Check if the coroutine was stopped during waiting
		if current.stopped || current.goroutineDone {
			return false
		}
		// Try to resume the coroutine again
		return true
	}
}

// Wait makes the current coroutine wait for the specified time
func (c *Coroutines) Wait(seconds float64) bool {
	current := c.Current()
	if current == nil || current.stopped {
		return false
	}

	c.mu.Lock()
	// Check if the coroutine has been stopped
	if current.stopped || current.goroutineDone {
		c.mu.Unlock()
		return false
	}

	current.state = StateWaiting
	current.waitUntil = time.Now().Add(time.Duration(seconds * float64(time.Second)))
	c.waitingQueue = append(c.waitingQueue, current)
	c.current.Store((*Coroutine)(nil))
	c.mu.Unlock()

	// Block until resumed by the scheduler
	select {
	case <-current.resumeCh:
		return true
	case <-time.After(time.Second * 2):
		// Timeout handling to prevent blocking after the coroutine is stopped
		return false
	}
}

// Update updates the coroutine scheduler, checks the waiting queue, and moves ready coroutines
func (c *Coroutines) Update() {
	c.mu.Lock()

	// Check the waiting queue and move ready coroutines to the ready queue
	now := time.Now()
	remainingWaiting := make([]*Coroutine, 0, len(c.waitingQueue))

	for _, coro := range c.waitingQueue {
		if coro.stopped || coro.goroutineDone {
			continue // Ignore stopped coroutines
		}

		if now.After(coro.waitUntil) || now.Equal(coro.waitUntil) {
			coro.state = StateRunning
			c.readyQueue = append(c.readyQueue, coro)
		} else {
			remainingWaiting = append(remainingWaiting, coro)
		}
	}

	c.waitingQueue = remainingWaiting
	c.mu.Unlock()

	// Execute coroutines in the ready queue
	c.RunNext()
}

// RunNext executes the next coroutine in the ready queue
func (c *Coroutines) RunNext() {
	// If there's a coroutine currently running, do nothing
	if c.Current() != nil {
		return
	}

	c.mu.Lock()
	if len(c.readyQueue) == 0 {
		c.mu.Unlock()
		return
	}

	// Take the first coroutine from the ready queue
	coro := c.readyQueue[0]
	c.readyQueue = c.readyQueue[1:]

	// If the coroutine has already stopped, do not run it
	if coro.stopped || coro.state == StateStopped || coro.goroutineDone {
		c.mu.Unlock()
		return
	}

	// Set the currently running coroutine
	c.current.Store(coro)

	// If the coroutine hasn't started yet, start it in a goroutine
	if !coro.started {
		coro.started = true
		// Ensure resumeCh is a buffered channel to prevent deadlock
		if coro.resumeCh == nil {
			coro.resumeCh = make(chan struct{}, 1)
		}
		c.mu.Unlock()

		go func() {
			// Execute the coroutine function
			coro.result = coro.fn(c)

			// Mark the coroutine as completed
			c.mu.Lock()
			coro.goroutineDone = true
			coro.state = StateStopped
			coro.stopped = true
			if c.Current() == coro {
				c.current.Store((*Coroutine)(nil))
			}
			c.mu.Unlock()

			// Check if there are more coroutines to execute
			c.mu.Lock()
			hasMoreCoroutines := len(c.readyQueue) > 0
			c.mu.Unlock()

			if hasMoreCoroutines {
				c.RunNext()
			}
		}()
	} else {
		// Ensure resumeCh is a buffered channel to prevent deadlock
		if coro.resumeCh == nil {
			coro.resumeCh = make(chan struct{}, 1)
		}
		c.mu.Unlock()

		// Send signal to resume coroutine execution
		select {
		case coro.resumeCh <- struct{}{}:
			// Successfully sent resume signal
		default:
			// Do not block if the channel is full or the coroutine has stopped
		}
	}
}

// Stop stops the specified coroutine
func (c *Coroutines) Stop(coro *Coroutine) {
	if coro == nil {
		return
	}

	c.mu.Lock()
	// Mark the coroutine as stopped
	coro.stopped = true
	coro.state = StateStopped
	coro.goroutineDone = true

	// Remove from waiting queue
	for i, waitingCoro := range c.waitingQueue {
		if waitingCoro == coro {
			c.waitingQueue = append(c.waitingQueue[:i], c.waitingQueue[i+1:]...)
			break
		}
	}
	c.mu.Unlock()

	// If it's the current coroutine, clear it
	current := c.Current()
	if current == coro {
		c.current.Store((*Coroutine)(nil))
	}

	// Remove from ready and waiting queues
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove from ready queue
	newReadyQueue := make([]*Coroutine, 0, len(c.readyQueue))
	for _, cr := range c.readyQueue {
		if cr != coro {
			newReadyQueue = append(newReadyQueue, cr)
		}
	}
	c.readyQueue = newReadyQueue

	// Remove from waiting queue
	newWaitingQueue := make([]*Coroutine, 0, len(c.waitingQueue))
	for _, cr := range c.waitingQueue {
		if cr != coro {
			newWaitingQueue = append(newWaitingQueue, cr)
		}
	}
	c.waitingQueue = newWaitingQueue

	// Ensure resumeCh is initialized
	if coro.resumeCh == nil {
		coro.resumeCh = make(chan struct{}, 1)
	}

	// Try to send signal to resumeCh to help coroutine recover from blocking
	// Clear the channel to ensure no old signals
	select {
	case <-coro.resumeCh:
		// Clear old signals
	default:
		// Channel is empty, continue
	}

	// Send new signal
	select {
	case coro.resumeCh <- struct{}{}:
		// Successfully sent signal
	default:
		// Channel is full or closed, ignore
	}
}

// IsRunning checks if the coroutine is running
func (c *Coroutines) IsRunning(coro *Coroutine) bool {
	if coro == nil {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if the coroutine has already stopped or completed
	if coro.stopped || coro.state == StateStopped || coro.goroutineDone {
		return false
	}

	// Check if the coroutine is in the ready queue
	for _, cr := range c.readyQueue {
		if cr == coro {
			return true
		}
	}

	// Check if the coroutine is in the waiting queue
	for _, cr := range c.waitingQueue {
		if cr == coro {
			return true
		}
	}

	// Check if it is the currently running coroutine
	current := c.Current()
	return current == coro
}

// StopIf stops coroutines based on a filter function
func (c *Coroutines) StopIf(filter StopFilterFunc) {
	if filter == nil {
		return
	}

	c.mu.Lock()
	// Get the list of all coroutines
	allCoroutines := make([]*Coroutine, 0, len(c.readyQueue)+len(c.waitingQueue))
	allCoroutines = append(allCoroutines, c.readyQueue...)
	allCoroutines = append(allCoroutines, c.waitingQueue...)

	// Currently running coroutine
	current := c.Current()
	if current != nil {
		found := false
		for _, coro := range allCoroutines {
			if coro == current {
				found = true
				break
			}
		}
		if !found {
			allCoroutines = append(allCoroutines, current)
		}
	}
	c.mu.Unlock()

	// Apply filter function to stop coroutines
	for _, coro := range allCoroutines {
		if filter(coro) {
			c.Stop(coro)
		}
	}
}

// StopAll stops all coroutines
func (c *Coroutines) StopAll() {
	c.StopIf(func(_ *Coroutine) bool {
		return true
	})
}

// StopAllOwnedBy stops all coroutines owned by a specific owner
func (c *Coroutines) StopAllOwnedBy(owner interface{}) {
	c.StopIf(func(coro *Coroutine) bool {
		return coro.owner == owner
	})
}

// Count returns the number of active coroutines
func (c *Coroutines) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Count non-stopped coroutines in the ready queue
	readyCount := 0
	for _, coro := range c.readyQueue {
		if !coro.stopped && coro.state != StateStopped {
			readyCount++
		}
	}

	// Count non-stopped coroutines in the waiting queue
	waitingCount := 0
	for _, coro := range c.waitingQueue {
		if !coro.stopped && coro.state != StateStopped {
			waitingCount++
		}
	}

	// Count the currently running coroutine
	currentCount := 0
	current := c.Current()
	if current != nil && !current.stopped && current.state != StateStopped {
		currentCount = 1
	}

	return readyCount + waitingCount + currentCount
}
