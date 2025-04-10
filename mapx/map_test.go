package mapx_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"go.atoms.co/lib/testing/requirex"
	"go.atoms.co/lib/mapx"
)

func TestNew(t *testing.T) {
	fn := func(v int) string { return fmt.Sprintf("%v", v) }

	t.Run("some", func(t *testing.T) {
		requirex.Equal(t, mapx.New([]int{3, 2, 1}, fn), map[string]int{"1": 1, "2": 2, "3": 3})
	})

	t.Run("empty", func(t *testing.T) {
		requirex.Equal(t, mapx.New(nil, fn), map[string]int{})
	})
}

func TestMapNew(t *testing.T) {
	fn := func(v int) (string, int) { return fmt.Sprintf("%v", v), v }

	t.Run("some", func(t *testing.T) {
		requirex.Equal(t, mapx.MapNew([]int{3, 2, 1}, fn), map[string]int{"1": 1, "2": 2, "3": 3})
	})

	t.Run("empty", func(t *testing.T) {
		requirex.Equal(t, mapx.MapNew(nil, fn), map[string]int{})
	})
}

func TestKeys(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		requirex.Equal(t, sorted(mapx.Keys(map[string]int{"1": 1, "2": 2, "3": 3})), []string{"1", "2", "3"})
	})

	t.Run("empty", func(t *testing.T) {
		require.Nil(t, mapx.Keys(map[string]int{}))
	})

	t.Run("nil", func(t *testing.T) {
		require.Nil(t, mapx.Keys[string, int](nil))
	})
}

func TestSortedKeys(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		requirex.Equal(t, mapx.SortedKeys(map[string]int{"1": 1, "2": 2, "3": 3}), []string{"1", "2", "3"})
	})

	t.Run("empty", func(t *testing.T) {
		require.Nil(t, mapx.SortedKeys(map[string]int{}))
	})

	t.Run("nil", func(t *testing.T) {
		require.Nil(t, mapx.SortedKeys[string, int](nil))
	})
}

func TestValues(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		requirex.Equal(t, sorted(mapx.Values(map[int]string{1: "1", 2: "2", 3: "3"})), []string{"1", "2", "3"})
	})

	t.Run("empty", func(t *testing.T) {
		require.Nil(t, mapx.Values(map[string]int{}))
	})

	t.Run("nil", func(t *testing.T) {
		require.Nil(t, mapx.Values[string, int](nil))
	})
}

func TestValuesIf(t *testing.T) {
	fn := func(v string) bool { return strings.Contains(v, "3") }
	t.Run("new", func(t *testing.T) {
		v := mapx.ValuesIf(map[int]string{1: "1", 2: "2", 3: "3", 33: "33"}, fn)
		requirex.Equal(t, sorted(v), []string{"3", "33"})
	})

	t.Run("empty", func(t *testing.T) {
		require.Nil(t, mapx.ValuesIf(map[int]string{}, fn))
	})

	t.Run("nil", func(t *testing.T) {
		require.Nil(t, mapx.ValuesIf[int, string](nil, fn))
	})
}

func TestMap(t *testing.T) {
	fn := func(k string, v int) (int, string) { return v, k }

	t.Run("some", func(t *testing.T) {
		m := mapx.Map(map[string]int{"1": 1, "2": 2, "3": 3}, fn)
		requirex.Equal(t, m, map[int]string{1: "1", 2: "2", 3: "3"})
	})

	t.Run("empty", func(t *testing.T) {
		requirex.Equal(t, mapx.Map(map[string]int{}, fn), map[int]string{})
	})

	t.Run("nil", func(t *testing.T) {
		requirex.Equal(t, mapx.Map(nil, fn), map[int]string{})
	})
}

func TestTryMap(t *testing.T) {
	fn := func(k string, v int) (int, string, error) {
		if v == 4 {
			return 0, "", fmt.Errorf("error")
		}
		return v, k, nil
	}

	t.Run("some", func(t *testing.T) {
		m, err := mapx.TryMap(map[string]int{"1": 1, "2": 2, "3": 3}, fn)
		require.NoError(t, err)
		requirex.Equal(t, m, map[int]string{1: "1", 2: "2", 3: "3"})
	})

	t.Run("error", func(t *testing.T) {
		m, err := mapx.TryMap(map[string]int{"1": 1, "2": 2, "3": 3, "4": 4}, fn)
		requirex.Equal(t, err.Error(), "error")
		require.Nil(t, m)
	})

	t.Run("empty", func(t *testing.T) {
		m, err := mapx.TryMap(map[string]int{}, fn)
		require.NoError(t, err)
		requirex.Equal(t, m, map[int]string{})
	})

	t.Run("nil", func(t *testing.T) {
		m, err := mapx.TryMap(nil, fn)
		require.NoError(t, err)
		requirex.Equal(t, m, map[int]string{})
	})
}

func TestMapValues(t *testing.T) {
	fn := func(v int) string { return fmt.Sprintf("%v", v+12) }

	t.Run("some", func(t *testing.T) {
		v := mapx.MapValues(map[string]int{"1": 1, "2": 2, "3": 3}, fn)
		requirex.Equal(t, sorted(v), []string{"13", "14", "15"})
	})

	t.Run("empty", func(t *testing.T) {
		require.Nil(t, mapx.MapValues(map[string]int{}, fn))
	})

	t.Run("nil", func(t *testing.T) {
		require.Nil(t, mapx.MapValues[string, int](nil, fn))
	})
}

func TestMapValuesIf(t *testing.T) {
	fn := func(v int) (string, bool) { return fmt.Sprintf("%v", v+12), v%2 == 1 }

	t.Run("some", func(t *testing.T) {
		v := mapx.MapValuesIf(map[string]int{"1": 1, "2": 2, "3": 3}, fn)
		requirex.Equal(t, sorted(v), []string{"13", "15"})
	})

	t.Run("empty", func(t *testing.T) {
		require.Nil(t, mapx.MapValuesIf(map[string]int{}, fn))
	})

	t.Run("nil", func(t *testing.T) {
		require.Nil(t, mapx.MapValuesIf[string, int](nil, fn))
	})
}

func TestMapToSlice(t *testing.T) {
	fn := func(k string, v int) string { return fmt.Sprintf("%v-%v", v, k) }

	t.Run("some", func(t *testing.T) {
		s := mapx.MapToSlice(map[string]int{"1": 1, "2": 2, "3": 3}, fn)
		requirex.Equal(t, sorted(s), []string{"1-1", "2-2", "3-3"})
	})

	t.Run("empty", func(t *testing.T) {
		require.Nil(t, mapx.MapToSlice(map[string]int{}, fn))
	})

	t.Run("nil", func(t *testing.T) {
		require.Nil(t, mapx.MapToSlice(nil, fn))
	})
}

func TestTryMapToSlice(t *testing.T) {
	fn := func(k string, v int) (string, error) {
		if v == 4 {
			return "", fmt.Errorf("error")
		}
		return fmt.Sprintf("%v-%v", v, k), nil
	}

	t.Run("some", func(t *testing.T) {
		s, err := mapx.TryMapToSlice(map[string]int{"1": 1, "2": 2, "3": 3}, fn)
		require.NoError(t, err)
		requirex.Equal(t, sorted(s), []string{"1-1", "2-2", "3-3"})
	})

	t.Run("error", func(t *testing.T) {
		m, err := mapx.TryMapToSlice(map[string]int{"1": 1, "2": 2, "3": 3, "4": 4}, fn)
		requirex.Equal(t, err.Error(), "error")
		require.Nil(t, m)
	})

	t.Run("empty", func(t *testing.T) {
		s, err := mapx.TryMapToSlice(map[string]int{}, fn)
		require.NoError(t, err)
		require.Nil(t, s)
	})

	t.Run("nil", func(t *testing.T) {
		s, err := mapx.TryMapToSlice(nil, fn)
		require.NoError(t, err)
		require.Nil(t, s)
	})
}

func TestFlatten(t *testing.T) {
	t.Run("some", func(t *testing.T) {
		requirex.Equal(t, sorted(mapx.Flatten(map[int][]string{1: {"11", "12"}, 3: {"31"}})), []string{"11", "12", "31"})
	})

	t.Run("empty", func(t *testing.T) {
		require.Nil(t, mapx.Flatten(map[int][]string{}))
	})

	t.Run("nil", func(t *testing.T) {
		require.Nil(t, mapx.Flatten[int, []string](nil))
	})
}

func TestContainsFunc(t *testing.T) {
	t.Run("some", func(t *testing.T) {
		require.True(t, mapx.ContainsFunc(map[string]int{"1": 1, "2": 2, "3": 3}, func(k string, v int) bool {
			return k == "1" && v == 1
		}))
	})
	t.Run("none", func(t *testing.T) {
		require.False(t, mapx.ContainsFunc(map[string]int{"1": 1, "2": 2, "3": 3}, func(k string, v int) bool {
			return k == "1" && v == 2
		}))
	})
}

func sorted(s []string) []string {
	sort.Strings(s)
	return s
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
