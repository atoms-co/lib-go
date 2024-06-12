package syncx_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.cloudkitchens.org/lib/syncx"
)

func TestLock(t *testing.T) {
	l := syncx.NewLock(1)
	assert.False(t, l.IsClosed())

	// (1) Limit

	ok := <-l.TryLock()
	assert.True(t, ok)

	select {
	case <-l.TryLock():
		require.Fail(t, "Double lock")
	case <-time.After(20 * time.Millisecond):
	}

	l.Unlock()
	select {
	case <-l.TryLock():
	case <-time.After(time.Second):
		require.Fail(t, "Failed to grant")
	}

	// (2) Close unblocks with "false"

	time.AfterFunc(10*time.Millisecond, l.Close)

	select {
	case ok := <-l.TryLock():
		assert.False(t, ok)
	case <-time.After(time.Second):
		require.Fail(t, "Failed to unblock")
	}

	assert.True(t, l.IsClosed())
}

func TestLock3(t *testing.T) {
	l := syncx.NewLock(3)

	// (1) Limit 3

	for i := 0; i < 3; i++ {
		ok := <-l.TryLock()
		assert.True(t, ok)
	}
	select {
	case <-l.TryLock():
		require.Fail(t, "Double lock")
	case <-time.After(20 * time.Millisecond):
	}

	// (2) Unlock can't exceed limit of 3

	for i := 0; i < 5; i++ {
		l.Unlock()
	}
	time.Sleep(20 * time.Millisecond)

	for i := 0; i < 3; i++ {
		ok := <-l.TryLock()
		assert.True(t, ok)
	}
	select {
	case <-l.TryLock():
		require.Fail(t, "Double lock")
	case <-time.After(20 * time.Millisecond):
	}
}
