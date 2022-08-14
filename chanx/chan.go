// Package chanx contains convenience utilities for working with channels.
package chanx

import (
	"go.cloudkitchens.org/lib/mathx"
	"sync"
)

// NewFixed returns a new closed chan with the given elements.
func NewFixed[T any](elms ...T) chan T {
	ret := make(chan T, len(elms))
	for _, elm := range elms {
		ret <- elm
	}
	close(ret)
	return ret
}

// ToList extracts all chan elements to a slice. Blocking.
func ToList[T any](ch chan T) []T {
	var ret []T
	for elm := range ch {
		ret = append(ret, elm)
	}
	return ret
}

// Process processes all elements with the given level of concurrency. Blocking.
func Process[T any](in chan T, n int, fn func(t T)) {
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
