package chanx_test

import (
	"go.atoms.co/lib/chanx"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDrain(t *testing.T) {
	t.Run("from empty channel", func(t *testing.T) {
		ch := make(chan int)

		chanx.Drain(ch)

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

		chanx.Drain(ch)

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
