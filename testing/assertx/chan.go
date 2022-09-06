package assertx

import (
	"go.cloudkitchens.org/lib/chanx"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	chanWait = 50 * time.Millisecond
)

// Element requires an element in the given channel within a grace period.
func Element[T any](t *testing.T, ch <-chan T, args ...any) T {
	elm, ok := chanx.TryRead(ch, chanWait)
	assert.True(t, ok, append([]any{"no chan element:"}, args...))
	return elm
}

// NoElement requires no element in the given channel in a grace period.
func NoElement[T any](t *testing.T, ch <-chan T, args ...any) {
	elm, ok := chanx.TryRead(ch, chanWait)
	assert.False(t, ok, append([]any{"unexpected chan element: ", elm}, args...))
}

// Drain returns all elements of the given channel until no element appear in a grace period.
func Drain[T any](ch <-chan T) []T {
	var ret []T
	for {
		elm, ok := chanx.TryRead(ch, chanWait)
		if !ok {
			return ret
		}
		ret = append(ret, elm)
	}
}
