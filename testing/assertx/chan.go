package assertx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.cloudkitchens.org/lib/chanx"
)

const (
	chanWait = 100 * time.Millisecond
)

// Element requires an element in the given channel within a grace period.
func Element[T any](t *testing.T, ch <-chan T, args ...any) T {
	t.Helper()

	elm, ok := chanx.TryRead(ch, chanWait)
	assert.True(t, ok, append([]any{"no chan element:"}, args...))
	return elm
}

// NoElement requires no element in the given channel in a grace period.
func NoElement[T any](t *testing.T, ch <-chan T, args ...any) {
	t.Helper()

	elm, ok := chanx.TryRead(ch, chanWait)
	assert.False(t, ok, append([]any{"unexpected chan element: ", elm}, args...))
}

// Closed requires a channel to be closed within a grace period
func Closed[T any](t *testing.T, ch <-chan T, args ...any) {
	t.Helper()

	ok := chanx.TryDrain(ch, chanWait)
	assert.True(t, ok, append([]any{"channel not closed:"}, args...))
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
