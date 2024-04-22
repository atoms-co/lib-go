package container

import (
	"container/heap"

	"slices"

	"go.cloudkitchens.org/lib/slicex"
)

type heapStore[T any] struct {
	data []T
	comp func(a, b T) bool
}

func (hs *heapStore[T]) Len() int {
	return len(hs.data)
}

func (hs *heapStore[T]) Less(i, j int) bool {
	return hs.comp(hs.data[i], hs.data[j])
}

func (hs *heapStore[T]) Swap(i, j int) {
	tmp := hs.data[i]
	hs.data[i] = hs.data[j]
	hs.data[j] = tmp
}

func (hs *heapStore[T]) Push(x any) {
	t, ok := x.(T)
	if !ok {
		panic("Unexpected type passed to Push")
	}
	hs.data = append(hs.data, t)
}

func (hs *heapStore[T]) Pop() any {
	tmp := hs.data[len(hs.data)-1]
	hs.data = hs.data[:len(hs.data)-1]
	return tmp
}

type Heap[T any] struct {
	store *heapStore[T]
}

func NewHeap[T any](comp func(a, b T) bool) *Heap[T] {
	return &Heap[T]{
		store: &heapStore[T]{
			data: make([]T, 0),
			comp: comp,
		},
	}
}

func (h *Heap[T]) Push(t T) {
	heap.Push(h.store, t)
}

func (h *Heap[T]) Pop() T {
	x := heap.Pop(h.store)
	ret, ok := x.(T)
	if !ok {
		panic("Unexpected type returned by Pop")
	}
	return ret
}

func (h *Heap[T]) Peek() T {
	return h.store.data[0]
}

func (h *Heap[T]) Len() int {
	return h.store.Len()
}

// Remove removes the element satisfying the predicate.
func (h *Heap[T]) Remove(fn func(x T) bool) bool {
	idx := slices.IndexFunc(h.store.data, fn)
	if idx != -1 {
		heap.Remove(h.store, idx)
	}
	return idx != -1
}

// Elements return a list of elements stored in the heap
func (h *Heap[T]) Elements() []T {
	return slicex.New(h.store.data...)
}
