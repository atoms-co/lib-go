package requirex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.cloudkitchens.org/lib/chanx"
)

const (
	chanWait = 2 * time.Second
)

// ChanEmpty verifies that the channel is empty. Fails if the channel is not empty after some time
func ChanEmpty[T any](t *testing.T, c <-chan T) {
	t.Helper()
	require.Eventually(t, func() bool { return len(c) == 0 }, chanWait, 100*time.Millisecond)
}

// ChanNonEmpty verifies that the channel has elements. Fails if the channel is empty after some time
func ChanNonEmpty[T any](t *testing.T, c <-chan T) {
	t.Helper()
	require.Eventually(t, func() bool { return len(c) > 0 }, chanWait, 100*time.Millisecond)
}

// Element requires an element in the given channel within a grace period.
func Element[T any](t *testing.T, ch <-chan T, args ...any) T {
	t.Helper()

	elm, ok := chanx.TryRead(ch, chanWait)
	if len(args) == 0 {
		args = []any{"no element in channel"}
	}
	require.True(t, ok, args...)
	return elm
}
