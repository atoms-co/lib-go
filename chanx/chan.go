// Package chanx contains convenience utilities for working with channels.
package chanx

import (
	"slices"
	"sync"
	"time"

	"go.cloudkitchens.org/lib/clock"
	"go.cloudkitchens.org/lib/iox"
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

// Prepend injects the given set of elements at the beginning of the channel. The returned chan is closed when the input is closed.
func Prepend[T any](in <-chan T, list ...T) <-chan T {
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

// Append injects the given set of messages after the input chan messages. The returned chan is closed
// when the input is closed and the appended messages are processed.
func Append[T any](in <-chan T, list ...T) <-chan T {
	cp := slices.Clone(list)

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
	return Prepend(Append(in, trailer), header)
}

// Breaker is a cancellable, buffered forwarder. It bidirectionally ties chan closure to the given async
// closer: if the chan is closed, the closer is closed; and if the closer is closed, the returned chan
// is closed and the input chan is drained.
func Breaker[T any](in <-chan T, closer iox.AsyncCloser, size int) <-chan T {
	buf := make(chan T, size)

	go func() {
		defer Drain(in)
		defer close(buf)
		defer closer.Close()

		for !closer.IsClosed() {
			select {
			case msg, ok := <-in:
				if !ok {
					return
				}

				// Delay checking for closure, if there is room in the buffer. That ensures all
				// messages in an unbuffered input chan are forwarded.

				select {
				case buf <- msg:
					// ok
				default:
					select {
					case buf <- msg:
						// ok
					case <-closer.Closed():
						return
					}
				}
			case <-closer.Closed():
				break
			}
		}

		// Flush input to buffer on closure (best-effort)
		for {
			select {
			case msg, ok := <-in:
				if !ok {
					return
				}
				select {
				case buf <- msg:
				default:
					return
				}
			default:
				return
			}
		}
	}()

	return buf
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
	n = max(1, n)

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
func TryRead[T any](ch <-chan T, cl clock.Clock, timeout time.Duration) (T, bool) {
	timer := cl.NewTimer(timeout)
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
func TryWrite[T any](ch chan<- T, t T, cl clock.Clock, timeout time.Duration) bool {
	timer := cl.NewTimer(timeout)
	defer timer.Stop()

	select {
	case ch <- t:
		return true
	case <-timer.C:
		return false
	}
}

// TryWriteWithCloser writes an element, waiting until either to write succeeds or the closer is closed.
// Returns false if the closer is closed before write succeeds.
func TryWriteWithCloser[T any](closer iox.RAsyncCloser, ch chan<- T, t T) bool {
	select {
	case ch <- t:
		return true
	case <-closer.Closed():
		return false
	}
}

// TryDrain reads, waiting up to the given timeout. Returns true if channel is closed, false otherwise.
func TryDrain[T any](ch <-chan T, cl clock.Clock, timeout time.Duration) bool {
	timer := cl.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return true
			}
		case <-timer.C:
			return false
		}
	}
}

// Drain removes all elements from the channel until it's closed.
func Drain[T any](ch <-chan T) {
	for range ch {
	}
}

// Clear removes all elements from the channel until it's empty at the time of the last check.
// Note that if there is a concurrent write to this channel, the channel may not be empty after this method is finished.
func Clear[T any](ch <-chan T) {
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
