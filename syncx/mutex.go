package syncx

import (
	"context"
	"sync"

	"go.atoms.co/lib/contextx"
)

// Mutex is a channel-based mutex for select locking. The lock chan is never closed
// and a value received indicates ownership until released.
type Mutex struct {
	ch chan bool
}

func NewMutex() *Mutex {
	ch := make(chan bool, 1)
	ch <- true
	return &Mutex{ch: ch}
}

// Lock returns the ownership chan. If a value is read, the lock is granted until Unlock is called.
func (m *Mutex) Lock() <-chan bool {
	return m.ch
}

// Unlock releases a previously held lock.
func (m *Mutex) Unlock() {
	select {
	case m.ch <- true:
	default:
	}
}

// Guard0 runs the given function in a critical section protected by the mutex. Context cancelable.
func Guard0(ctx context.Context, m *Mutex, fn func() error) error {
	if contextx.IsCancelled(ctx) {
		return ctx.Err()
	}

	select {
	case <-m.Lock():
		defer m.Unlock()

		return fn()

	case <-ctx.Done():
		return ctx.Err()
	}
}

// Guard1 runs the given function in a critical section protected by the mutex. Context cancelable.
func Guard1[T any](ctx context.Context, m *Mutex, fn func() (T, error)) (T, error) {
	if !contextx.IsCancelled(ctx) {
		select {
		case <-m.Lock():
			defer m.Unlock()

			return fn()

		case <-ctx.Done():
		}
	}

	var zero T
	return zero, ctx.Err()
}

// Guard2 runs the given function in a critical section protected by the mutex. Context cancelable.
func Guard2[T1, T2 any](ctx context.Context, m *Mutex, fn func() (T1, T2, error)) (T1, T2, error) {
	if !contextx.IsCancelled(ctx) {
		select {
		case <-m.Lock():
			defer m.Unlock()

			return fn()

		case <-ctx.Done():
		}
	}

	var zero1 T1
	var zero2 T2
	return zero1, zero2, ctx.Err()

}

// MutexMap creates mutexes on demand for a given key.
type MutexMap[T comparable] struct {
	m  map[T]*Mutex
	mu sync.Mutex
}

func NewMutexMap[T comparable]() *MutexMap[T] {
	return &MutexMap[T]{
		m: map[T]*Mutex{},
	}
}

// V returns the mutex for the given key. Created if not already exist.
func (m *MutexMap[T]) V(k T) *Mutex {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ret, ok := m.m[k]; ok {
		return ret
	}

	ret := NewMutex()
	m.m[k] = ret
	return ret
}

// Delete removes the mutex if unlocked. Returns true if removed. It also locks the abandoned mutex
// to block out any pending callers, so -- if recreated -- exclusivity is preserved.
func (m *MutexMap[T]) Delete(k T) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ret, ok := m.m[k]; ok {
		select {
		case <-ret.ch:
			// Delete mutex in locked state only, if present.
		default:
			return false
		}
	}

	delete(m.m, k)
	return true
}
