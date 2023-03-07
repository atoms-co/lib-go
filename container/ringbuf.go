package container

import "fmt"

// RingBuf is a fixed-size ring buffer. See: https://en.wikipedia.org/wiki/Circular_buffer.
type RingBuf[T any] struct {
	buf         []T
	read, write int
}

func NewRingBuf[T any](size int) *RingBuf[T] {
	return &RingBuf[T]{
		buf: make([]T, size+1, size+1),
	}
}

// IsEmpty returns true iff the ring buffer is empty.
func (r *RingBuf[T]) IsEmpty() bool {
	return r.read == r.write
}

// Len returns the number of elements in the buffer.
func (r *RingBuf[T]) Len() int {
	if r.read <= r.write {
		return r.write - r.read
	}
	return r.write + (len(r.buf) - r.read)
}

// Enqueue enqueues an element. Returns false iff full.
func (r *RingBuf[T]) Enqueue(t T) bool {
	next := (r.write + 1) % len(r.buf)
	if next == r.read {
		return false // full
	}

	r.buf[r.write] = t
	r.write = next
	return true
}

// Dequeue dequeues an element, if any. Returns false iff empty.
func (r *RingBuf[T]) Dequeue() (T, bool) {
	var def T
	if r.IsEmpty() {
		return def, false // empty
	}

	ret := r.buf[r.read]
	r.buf[r.read] = def // don't keep a reference
	r.read = (r.read + 1) % len(r.buf)
	return ret, true
}

// Peek returns the next element, if any, without removing it.
func (r *RingBuf[T]) Peek() (T, bool) {
	var def T
	if r.IsEmpty() {
		return def, false // empty
	}
	return r.buf[r.read], true
}

func (r *RingBuf[T]) String() string {
	if next, ok := r.Peek(); ok {
		return fmt.Sprintf("[next=%v, len=%v]", next, r.Len())
	}
	return "{}"
}
