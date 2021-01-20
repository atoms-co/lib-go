// Package workqueue contains a workqueue with bounded concurrency.
package workqueue

import (
	"go.cloudkitchens.org/lib/mathx"
	"go.uber.org/atomic"
)

// WorkQueue is a simple work queue with bounded concurrency, optionally buffered. Intended for IO operations. The
// queue will work through work items with the given level of concurrency.
type WorkQueue struct {
	pending   chan func()
	completed chan bool
	quit      chan struct{}
	closed    atomic.Bool

	active, cap atomic.Int64
}

func New(cap, buffer int) *WorkQueue {
	p := &WorkQueue{
		pending:   make(chan func(), mathx.MaxInt(0, buffer)),
		completed: make(chan bool),
		quit:      make(chan struct{}),
	}
	p.Resize(cap)

	go p.process()

	return p
}

// Chan returns the ingestion channel for incoming work.
func (p *WorkQueue) Chan() chan<- func() {
	return p.pending
}

// Size returns the number of go routines in use.
func (p *WorkQueue) Size() int {
	return int(p.active.Load())
}

// Resize sets the maximum number of go routines to use. At least 1.
func (p *WorkQueue) Resize(cap int) {
	if cap > 0 {
		p.cap.Store(int64(cap))
	}
}

// Capacity returns the maximum number of go routines the pool can use.
func (p *WorkQueue) Capacity() int {
	return int(p.cap.Load())
}

func (p *WorkQueue) Close() {
	if p.closed.CAS(false, true) {
		close(p.quit)
	}
}

func (p *WorkQueue) process() {
	for {
		if p.active.Load() >= p.cap.Load() {
			// The pool has no available capacity, so wait for something to complete. We don't want to take work off the
			// backlog, if we can't process it to maintain back pressure.

			<-p.completed
			p.active.Dec()
			continue // check for capacity again, in case pool was resized
		}

		select {
		case fn := <-p.pending:
			p.active.Inc()
			go p.run(fn)

		case <-p.completed:
			p.active.Dec()

		case <-p.quit:
			// Let active routines terminate to avoid leaking them. Then exit.

			for p.active.Load() > 0 {
				<-p.completed
				p.active.Dec()
			}
			return
		}
	}
}

func (p *WorkQueue) run(fn func()) {
	fn()
	p.completed <- true
}
