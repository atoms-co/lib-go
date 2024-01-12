package chanx

import (
	"context"
	"fmt"
	"sync"
)

// Provider provides a T value, if present. Thread-safe.
type Provider[T any] interface {
	V() (T, bool)
}

type provider[T any] struct {
	t  T
	ok bool
	mu sync.RWMutex
}

// NewProvider transforms a chan to a provider of the latest value. Uses context cancellation.
func NewProvider[T any](ctx context.Context, in <-chan T) Provider[T] {
	ret := &provider[T]{}
	go ret.process(ctx, in)
	return ret
}

func (p *provider[T]) V() (T, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.t, p.ok
}

func (p *provider[T]) process(ctx context.Context, in <-chan T) {
	for {
		select {
		case t, ok := <-in:
			if !ok {
				return // keep last value
			}
			p.mu.Lock()
			p.t = t
			p.ok = true
			p.mu.Unlock()

		case <-ctx.Done():
			return
		}
	}
}

func (p *provider[T]) String() string {
	if v, ok := p.V(); ok {
		return fmt.Sprintf("%v", v)
	}
	return "-"
}
