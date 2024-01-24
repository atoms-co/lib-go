package requirex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ChanEmpty verifies that the channel is empty. Fails if the channel is not empty after given duration
func ChanEmpty[T any](t *testing.T, c <-chan T, limit time.Duration) {
	t.Helper()
	require.Eventually(t, func() bool { return len(c) == 0 }, limit, 100*time.Millisecond)
}

// ChanNonEmpty verifies that the channel has elements. Fails if the channel is empty after given duration
func ChanNonEmpty[T any](t *testing.T, c <-chan T, limit time.Duration) {
	t.Helper()
	require.Eventually(t, func() bool { return len(c) > 0 }, limit, 100*time.Millisecond)
}
