package container_test

import (
	"go.cloudkitchens.org/lib/container"
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
}
