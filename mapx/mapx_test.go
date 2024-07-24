package mapx_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.atoms.co/lib/testing/requirex"
	"go.atoms.co/lib/mapx"
)

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

func TestGetOnly(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		k, v, ok := mapx.GetOnly(map[int]int{})
		requirex.Equal(t, k, 0)
		requirex.Equal(t, v, 0)
		require.False(t, ok)
	})

	t.Run("non-empty", func(t *testing.T) {
		k, v, ok := mapx.GetOnly(map[int]int{1: 2})
		requirex.Equal(t, k, 1)
		requirex.Equal(t, v, 2)
		require.True(t, ok)
	})
}
