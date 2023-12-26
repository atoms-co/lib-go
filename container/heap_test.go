package container_test

import (
	"go.atoms.co/lib/container"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeapInt(t *testing.T) {
	heap := container.NewHeap[int](func(a, b int) bool { return a < b })
	values := []int{6, 5, 4, 8, 9, 10, 13, 12, 11, 7}

	// Push
	for _, v := range values {
		heap.Push(v)
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	i := 0
	for heap.Len() > 0 {
		val := heap.Pop()
		require.Equal(t, values[i], val, "item at index %v", i)
		i++
	}

	heap.Push(14)
	heap.Push(15)

	ok := heap.Remove(eq(1))
	require.False(t, ok)
	require.Equal(t, heap.Len(), 2)
	require.Equal(t, heap.Peek(), 14)

	ok = heap.Remove(eq(14))
	require.True(t, ok)
	require.Equal(t, heap.Len(), 1)
	require.Equal(t, heap.Peek(), 15)

	ok = heap.Remove(eq(15))
	require.True(t, ok)
	require.Equal(t, heap.Len(), 0)
}

func eq(x int) func(a int)bool  {
	return func(a int) bool {
		return x == a
	}
}
