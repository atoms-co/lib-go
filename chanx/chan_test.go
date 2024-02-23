package chanx_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.cloudkitchens.org/lib/testing/assertx"
	"go.cloudkitchens.org/lib/chanx"
	"go.cloudkitchens.org/lib/iox"
)

func TestDrain(t *testing.T) {
	t.Run("from empty channel", func(t *testing.T) {
		ch := make(chan int)

		wait := iox.NewAsyncCloser()
		go func() {
			chanx.Drain(ch)
			wait.Close()
		}()
		close(ch)
		<-wait.Closed()

		require.Empty(t, ch)
	})

	t.Run("from closed channel", func(t *testing.T) {
		ch := make(chan int)
		close(ch)

		chanx.Drain(ch)

		require.Empty(t, ch)
	})

	t.Run("from non-empty channel", func(t *testing.T) {
		ch := make(chan int, 5)
		ch <- 1
		ch <- 2

		wait := iox.NewAsyncCloser()
		go func() {
			chanx.Drain(ch)
			wait.Close()
		}()
		close(ch)
		<-wait.Closed()

		require.Empty(t, ch)
	})

	t.Run("from non-empty closed channel", func(t *testing.T) {
		ch := make(chan int, 5)
		ch <- 1
		ch <- 2
		close(ch)

		chanx.Drain(ch)

		require.Empty(t, ch)
	})
}

func TestClear(t *testing.T) {
	t.Run("from empty channel", func(t *testing.T) {
		ch := make(chan int)

		chanx.Clear(ch)

		require.Empty(t, ch)
	})

	t.Run("from closed channel", func(t *testing.T) {
		ch := make(chan int)
		close(ch)

		chanx.Clear(ch)

		require.Empty(t, ch)
	})

	t.Run("from non-empty channel", func(t *testing.T) {
		ch := make(chan int, 5)
		ch <- 1
		ch <- 2

		chanx.Clear(ch)

		require.Empty(t, ch)
	})

	t.Run("from non-empty closed channel", func(t *testing.T) {
		ch := make(chan int, 5)
		ch <- 1
		ch <- 2
		close(ch)

		chanx.Clear(ch)

		require.Empty(t, ch)
	})
}

func TestBreaker(t *testing.T) {
	t.Run("chan", func(t *testing.T) {

		ch := make(chan int, 5)
		ch <- 1

		quit := iox.NewAsyncCloser()
		out := chanx.Breaker(ch, quit, 5)

		one := assertx.Element(t, out)
		assertx.Equal(t, one, 1)

		ch <- 2

		two := assertx.Element(t, out)
		assertx.Equal(t, two, 2)

		close(ch)

		assertx.NoElement(t, out)
		assert.True(t, quit.IsClosed())
	})

	t.Run("closer", func(t *testing.T) {

		ch := make(chan int, 2)
		ch <- 1

		quit := iox.NewAsyncCloser()
		out := chanx.Breaker(ch, quit, 2)

		one := assertx.Element(t, out)
		assertx.Equal(t, one, 1)

		quit.Close()

		ch <- 2
		ch <- 3
		ch <- 4

		two := assertx.Element(t, out)
		assertx.Equal(t, two, 2)

		three := assertx.Element(t, out)
		assertx.Equal(t, three, 3)

		assertx.NoElement(t, out) // No room in buffer for 4
	})
}
