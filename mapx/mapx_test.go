package mapx_test

import (
	"go.cloudkitchens.org/lib/mapx"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEquals(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		require.True(t, mapx.Equals(map[int]int{}, map[int]int{}))
	})

	t.Run("left empty", func(t *testing.T) {
		require.False(t, mapx.Equals(map[int]int{}, map[int]int{1: 1}))
	})

	t.Run("right empty", func(t *testing.T) {
		require.False(t, mapx.Equals(map[int]int{1: 1}, map[int]int{}))
	})

	t.Run("equal elements", func(t *testing.T) {
		require.True(t, mapx.Equals(map[int]int{1: 1, 2: 2}, map[int]int{2: 2, 1: 1}))
	})

	t.Run("not equal elements", func(t *testing.T) {
		require.False(t, mapx.Equals(map[int]int{1: 2, 2: 1}, map[int]int{2: 2, 1: 1}))
	})
}
