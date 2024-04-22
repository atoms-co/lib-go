package slicex_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"go.cloudkitchens.org/lib/slicex"
)

var result int

func benchmarkMap(n int, b *testing.B) {
	a := make([]int, n)
	for i := 0; i < n; i++ {
		a[i] = i
	}
	b.ResetTimer()

	c := 0
	for i := 0; i < b.N; i++ {
		b := slicex.Map(a, strconv.Itoa)
		c += len(b)
	}

	result = c
}

func BenchmarkMap1000(b *testing.B)   { benchmarkMap(1000, b) }
func BenchmarkMap2000(b *testing.B)   { benchmarkMap(2000, b) }
func BenchmarkMap5000(b *testing.B)   { benchmarkMap(5000, b) }
func BenchmarkMap10000(b *testing.B)  { benchmarkMap(10000, b) }
func BenchmarkMap15000(b *testing.B)  { benchmarkMap(15000, b) }
func BenchmarkMap20000(b *testing.B)  { benchmarkMap(20000, b) }
func BenchmarkMap25000(b *testing.B)  { benchmarkMap(25000, b) }
func BenchmarkMap30000(b *testing.B)  { benchmarkMap(30000, b) }
func BenchmarkMap35000(b *testing.B)  { benchmarkMap(35000, b) }
func BenchmarkMap40000(b *testing.B)  { benchmarkMap(40000, b) }
func BenchmarkMap45000(b *testing.B)  { benchmarkMap(45000, b) }
func BenchmarkMap50000(b *testing.B)  { benchmarkMap(50000, b) }
func BenchmarkMap60000(b *testing.B)  { benchmarkMap(60000, b) }
func BenchmarkMap70000(b *testing.B)  { benchmarkMap(70000, b) }
func BenchmarkMap80000(b *testing.B)  { benchmarkMap(80000, b) }
func BenchmarkMap90000(b *testing.B)  { benchmarkMap(90000, b) }
func BenchmarkMap100000(b *testing.B) { benchmarkMap(100000, b) }

func TestMap(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		require.Empty(t, slicex.Map([]int{}, strconv.Itoa))
	})

	t.Run("nonempty", func(t *testing.T) {
		require.Equal(t, slicex.Map([]int{1, 2, 3}, strconv.Itoa), []string{"1", "2", "3"})
	})
}

func TestFlatMap(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		require.Empty(t, slicex.FlatMap([]string{}, strings.Fields))
	})

	t.Run("nonempty", func(t *testing.T) {
		require.Equal(t, slicex.FlatMap([]string{"1 2", "3 4", "5"}, strings.Fields), []string{"1", "2", "3", "4", "5"})
	})
}

func TestMapIf(t *testing.T) {
	f := func(n int) (int, bool) {
		return n, n != 2
	}

	t.Run("empty", func(t *testing.T) {
		require.Empty(t, slicex.MapIf([]int{}, f))
	})

	t.Run("nonempty", func(t *testing.T) {
		require.Equal(t, slicex.MapIf([]int{1, 2, 3, 4}, f), []int{1, 3, 4})
	})
}

func TestTryMap(t *testing.T) {
	f := func(n int) (int, error) {
		if n == 2 {
			return 0, fmt.Errorf("error")
		}
		return n, nil
	}

	t.Run("empty", func(t *testing.T) {
		s, err := slicex.TryMap([]int{}, f)
		require.NoError(t, err)
		require.Empty(t, s)
	})

	t.Run("without errors", func(t *testing.T) {
		s, err := slicex.TryMap([]int{1, 3, 4, 5, 6}, f)
		require.NoError(t, err)
		require.Equal(t, s, []int{1, 3, 4, 5, 6})
	})

	t.Run("with errors", func(t *testing.T) {
		_, err := slicex.TryMap([]int{1, 2, 3, 4, 5, 6}, f)
		require.Equal(t, err, fmt.Errorf("error"))
	})
}

func TestFlatten(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		require.Empty(t, slicex.Flatten([][]string{}))
	})

	t.Run("nonempty", func(t *testing.T) {
		require.Equal(t, slicex.Flatten([][]string{{"1", "2"}, {"3", "4"}, {"5"}}), []string{"1", "2", "3", "4", "5"})
	})
}

func TestCopyAppend(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		require.Empty(t, slicex.CopyAppend([]string{}))
	})

	t.Run("empty append", func(t *testing.T) {
		a := []string{"1", "2"}
		b := slicex.CopyAppend(a)
		require.Equal(t, b, a)
		a[0] = "3"
		require.Equal(t, b, []string{"1", "2"})
	})

	t.Run("nonempty append", func(t *testing.T) {
		a := []string{"1", "2"}
		b := slicex.CopyAppend(a, "4", "5")
		require.Equal(t, b, []string{"1", "2", "4", "5"})
		a[0] = "3"
		require.Equal(t, b, []string{"1", "2", "4", "5"})
	})
}

func TestCount(t *testing.T) {
	f := func(n int) bool {
		return n == 2
	}

	t.Run("empty", func(t *testing.T) {
		require.Equal(t, slicex.Count([]int{}, f), 0)
	})

	t.Run("no matches", func(t *testing.T) {
		require.Equal(t, slicex.Count([]int{1, 3, 4}, f), 0)
	})

	t.Run("matches", func(t *testing.T) {
		require.Equal(t, slicex.Count([]int{1, 2, 3, 4, 2}, f), 2)
	})
}

func TestContains(t *testing.T) {
	f := func(n int) bool {
		return n == 2
	}

	t.Run("empty", func(t *testing.T) {
		require.False(t, slicex.Contains([]int{}, f))
	})

	t.Run("no matches", func(t *testing.T) {
		require.False(t, slicex.Contains([]int{1, 3, 4}, f))
	})

	t.Run("matches", func(t *testing.T) {
		require.True(t, slicex.Contains([]int{1, 3, 4, 2}, f))
	})
}

func TestContainsT(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		require.False(t, slicex.ContainsT([]int{}, 1))
	})

	t.Run("no matches", func(t *testing.T) {
		require.False(t, slicex.ContainsT([]int{1, 3, 4}, 2, 5))
	})

	t.Run("matches", func(t *testing.T) {
		require.True(t, slicex.ContainsT([]int{1, 3, 4, 2}, 2, 3))
	})
}

func TestFirst(t *testing.T) {
	f := func(n int) bool {
		return n == 2
	}

	t.Run("empty", func(t *testing.T) {
		_, ok := slicex.First([]int{}, f)
		require.False(t, ok)
	})

	t.Run("no matches", func(t *testing.T) {
		_, ok := slicex.First([]int{1, 3, 4}, f)
		require.False(t, ok)
	})

	t.Run("matches", func(t *testing.T) {
		_, ok := slicex.First([]int{1, 3, 4, 2}, f)
		require.True(t, ok)
	})
}

func TestFilter(t *testing.T) {
	f := func(n int) bool {
		return n%2 == 0
	}

	t.Run("empty", func(t *testing.T) {
		require.Empty(t, slicex.Filter([]int{}, f))
	})

	t.Run("no matches", func(t *testing.T) {
		require.Empty(t, slicex.Filter([]int{1, 3, 5}, f))
	})

	t.Run("matches", func(t *testing.T) {
		require.Equal(t, slicex.Filter([]int{1, 2, 3, 4, 5, 6}, f), []int{2, 4, 6})
	})
}

func TestSet(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		require.Empty(t, slicex.NewSet[int]())
	})

	t.Run("nonempty", func(t *testing.T) {
		require.Equal(t, slicex.NewSet(1, 2, 3, 3, 5), map[int]bool{1: true, 2: true, 3: true, 5: true})
	})
}
