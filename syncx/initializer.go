package syncx

import "sync"

// Initializer is a Lock wrapper that ensures on-demand one-time initialization as well as reset.
type Initializer struct {
	lock *Lock
	done bool
	mu   sync.Mutex
}

// Lock returns the current initialization lock. The caller is expected to check IsComplete first.
func (l *Initializer) Lock() *Lock {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.lock == nil {
		l.lock = NewLock(1)
	}
	if l.done {
		l.lock.Close()
	}
	return l.lock
}

// IsComplete returns true iff initialization is done.
func (l *Initializer) IsComplete() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.done
}

// Complete completes initialization. The caller is assumed to hold the lock.
func (l *Initializer) Complete() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.done = true
	if l.lock != nil {
		l.lock.Close()
	}
}

// Reset uncompletes initialization, if done.
func (l *Initializer) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.done {
		l.done = false
		l.lock = nil
	}
}

func (l *Initializer) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.lock != nil {
		l.lock.Close()
	}
}
