package mapx

import (
	"fmt"
	"sync"
)

type SerializeFn[V any] func(V) ([]byte, error)

// Internalizer offers shared, unique representations of immutable values, usually reference types. It requires
// the caller to pick a (deterministic and complete) serialization function and not mutate internalized values.
type Internalizer[V any] struct {
	m  map[string]V
	fn SerializeFn[V]
	mu sync.Mutex
}

func NewInternalizer[V any](fn SerializeFn[V]) *Internalizer[V] {
	return &Internalizer[V]{
		m:  map[string]V{},
		fn: fn,
	}
}

// Internalize returns a shared equivalent value of v.
func (i *Internalizer[V]) Internalize(v V) V {
	key, err := i.fn(v)
	if err != nil {
		return v // bad value: return itself
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	if ret, ok := i.m[string(key)]; ok {
		return ret
	}
	i.m[string(key)] = v
	return v
}

func (i *Internalizer[V]) Reset() {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.m = map[string]V{}
}

func (i *Internalizer[V]) String() string {
	i.mu.Lock()
	defer i.mu.Unlock()

	return fmt.Sprintf("{size=%v}", len(i.m))
}
