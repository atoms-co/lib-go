// Package chanx contains convenience utilities for working with channels.
package chanx

import (
	"go.atoms.co/lib/mathx"
	"sync"
	"time"
)

// NewFixed returns a new closed chan with the given elements.
func NewFixed[T any](elms ...T) <-chan T {
	ret := make(chan T, len(elms))
	for _, elm := range elms {
		ret <- elm
	}
	close(ret)
	return ret
}

// ToList extracts all chan elements to a slice. Blocking.
func ToList[T any](ch <-chan T) []T {
	var ret []T
	for elm := range ch {
		ret = append(ret, elm)
	}
	return ret
}

// Map transforms each element of the chan async. The returned chan is closed when the input is closed.
func Map[T, U any](in <-chan T, size int, fn func(t T) U) <-chan U {
	ret := make(chan U, size)
	go func() {
		defer close(ret)
		for elm := range in {
			ret <- fn(elm)
		}
	}()
	return ret
}

// MapAppend transforms each element of the chan async, after injecting the given set of messages. The
// returned chan is closed when the input is closed.
func MapAppend[T, U any](list []U, in <-chan T, size int, fn func(t T) U) <-chan U {
	ret := make(chan U, mathx.Max(size, len(list)))
	for _, elm := range list {
		ret <- elm
	}
	go func() {
		defer close(ret)
		for elm := range in {
			ret <- fn(elm)
		}
	}()
	return ret
}

// Process processes all elements with the given level of concurrency. Blocking.
func Process[T any](in <-chan T, n int, fn func(t T)) {
	n = mathx.MaxInt(1, n)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for elm := range in {
				fn(elm)
			}
		}()
	}
	wg.Wait()
}

// TryRead reads an element, waiting up to the given timeout, Returns false otherwise.
func TryRead[T any](ch <-chan T, timeout time.Duration) (T, bool) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case elm, ok := <-ch:
		if ok {
			return elm, true
		}
	case <-timer.C:
	}

	var t T
	return t, false
}

// TryWrite writes an element, waiting up to the given timeout, Returns false otherwise.
func TryWrite[T any](ch chan<- T, t T, timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case ch <- t:
		return true
	case <-timer.C:
		return false
	}
}

// Drain removes all elements from the channel
func Drain[T any](ch <-chan T) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}
