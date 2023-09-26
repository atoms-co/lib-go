// Package chanx contains convenience utilities for working with channels.
package chanx

import (
	"go.atoms.co/lib/mathx"
	"go.atoms.co/slicex"
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

// Append injects the given set of messages. The returned chan is closed when the input is closed.
func Append[T any](list []T, in <-chan T) <-chan T {
	ret := make(chan T, len(list))
	for _, elm := range list {
		ret <- elm
	}
	go func() {
		defer close(ret)
		for elm := range in {
			ret <- elm
		}
	}()
	return ret
}

// AppendLast injects the given set of messages after the input chan messages. The returned chan is closed
// when the input is closed and the trailing messages are processed.
func AppendLast[T any](in <-chan T, list ...T) <-chan T {
	cp := slicex.Clone(list)

	ret := make(chan T)
	go func() {
		defer close(ret)

		for elm := range in {
			ret <- elm
		}
		for _, elm := range cp {
			ret <- elm
		}
	}()
	return ret
}

// Envelope adds a single header and trailer to a stream. Convenience function.
func Envelope[T any](header T, in <-chan T, trailer T) <-chan T {
	return Append([]T{header}, AppendLast(in, trailer))
}

// Map transforms each element of the chan async. The returned chan is closed when the input is closed.
func Map[T, U any](in <-chan T, fn func(t T) U) <-chan U {
	ret := make(chan U)
	go func() {
		defer close(ret)
		for elm := range in {
			ret <- fn(elm)
		}
	}()
	return ret
}

// MapIf transforms selected element of the chan async. The returned chan is closed when the input is closed.
func MapIf[T, U any](in <-chan T, fn func(t T) (U, bool)) <-chan U {
	ret := make(chan U)
	go func() {
		defer close(ret)
		for elm := range in {
			if u, ok := fn(elm); ok {
				ret <- u
			}
		}
	}()
	return ret
}

// MapAppend transforms each element of the chan async, after injecting the given set of messages. The
// returned chan is closed when the input is closed.
func MapAppend[T, U any](list []U, in <-chan T, fn func(t T) U) <-chan U {
	ret := make(chan U, len(list))
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

// Join combines the elements of the input chans async. The returned chan is closed after the inputs are closed.
func Join[T any](in1, in2 <-chan T) <-chan T {
	ret := make(chan T)
	go func() {
		defer close(ret)
		for {
			select {
			case m, ok := <-in1:
				if !ok {
					for m := range in2 {
						ret <- m
					}
					return
				}
				ret <- m
			case m, ok := <-in2:
				if !ok {
					for m := range in1 {
						ret <- m
					}
					return
				}
				ret <- m
			}
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
