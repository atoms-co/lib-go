package syncx

import (
	"sync/atomic"

	"go.cloudkitchens.org/lib/iox"
)

// Lock is a chan-based closeable semaphore with N concurrent locks granted. If N=1, it acts as a mutex.
// Using channels allows callers to block and be interrupted by other events via select.
type Lock struct {
	iox.AsyncCloser

	guard   chan bool    // chan of lock grants
	pending atomic.Int64 // #unprocessed unlocks
	pulse   chan bool    // 1-buf trigger to process unlock
}

func NewLock(n int) *Lock {
	n = max(1, n)
	ret := &Lock{
		AsyncCloser: iox.NewAsyncCloser(),
		guard:       make(chan bool, n),
		pulse:       make(chan bool, 1),
	}
	for i := 0; i < n; i++ {
		ret.guard <- true
	}
	go ret.process()

	return ret
}

// TryLock returns the grant chan. If the result is true the lock is granted. If false, the
// lock has been closed (and hence a chan read returns the default value).
func (l *Lock) TryLock() <-chan bool {
	return l.guard
}

// Unlock frees a (previously held) lock. If called too many times, it cannot exceed the initial
// limit. There is a slight delay for the lock to be re-enqueued, so a Unlock + default-select
// TryLock may fail to acquire the released lock.
func (l *Lock) Unlock() {
	l.pending.Add(1)

	select {
	case l.pulse <- true:
	default:
	}
}

func (l *Lock) process() {
	defer close(l.guard)

	for {
		select {
		case <-l.Closed():
			// Close guard and stop processing unlocks.
			return

		case <-l.pulse:
			n := int(l.pending.Load())
			for i := 0; i < n; i++ {
				l.pending.Add(-1)

				if l.IsClosed() {
					return // ensure Close + Unlock does not allow a lock
				}

				select {
				case l.guard <- true:
				default:
				}
			}
		}
	}
}
