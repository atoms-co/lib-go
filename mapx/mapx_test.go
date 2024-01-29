package mapx_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.atoms.co/lib/testing/requirex"
	"go.atoms.co/lib/mapx"
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

func TestMerge(t *testing.T) {
	t.Run("merge", func(t *testing.T) {
		m1 := map[int]int{1: 1, 2: 2}
		m2 := map[int]int{1: 3, 3: 3}

		m := mapx.Merge(m1, m2)
		requirex.Equal(t, m, map[int]int{1: 3, 2: 2, 3: 3})
		requirex.Equal(t, m1, map[int]int{1: 1, 2: 2})
		requirex.Equal(t, m2, map[int]int{1: 3, 3: 3})

		m = mapx.Merge(m2, m1)
		requirex.Equal(t, m, map[int]int{1: 1, 2: 2, 3: 3})
		requirex.Equal(t, m1, map[int]int{1: 1, 2: 2})
		requirex.Equal(t, m2, map[int]int{1: 3, 3: 3})
	})
}
