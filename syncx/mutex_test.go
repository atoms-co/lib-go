package syncx_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.cloudkitchens.org/lib/testing/assertx"
	"go.cloudkitchens.org/lib/syncx"
)

func TestMutex_Lock(t *testing.T) {
	mu := syncx.NewMutex()

	// Test basic lock/unlock

	select {
	case <-mu.Lock():
	default:
		assert.Fail(t, "failed to lock")
	}
	select {
	case <-mu.Lock():
		assert.Fail(t, "locked twice")
	default:
	}

	mu.Unlock()

	select {
	case <-mu.Lock():
	default:
		assert.Fail(t, "failed to lock")
	}
}

func TestMutex_Guard(t *testing.T) {
	ctx := context.Background()
	mu := syncx.NewMutex()

	// Test concurrent guards and check exclusivity

	<-mu.Lock()

	var excl atomic.Bool
	var n int

	var wg sync.WaitGroup
	for i := 0; i < 1_000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := syncx.Guard0(ctx, mu, func() error {
				assert.False(t, excl.Load())
				assert.True(t, excl.CompareAndSwap(false, true))

				n++

				assert.True(t, excl.CompareAndSwap(true, false))
				assert.False(t, excl.Load())
				return nil
			})
			assert.NoError(t, err)
		}()
	}

	mu.Unlock()
	wg.Wait()

	assertx.Equal(t, n, 1_000)
}
