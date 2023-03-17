package chanx

import (
	"context"
	"fmt"
	"sync"

	"go.cloudkitchens.org/lib/iox"
)

// Broadcaster is a utility struct for broadcasting a single channel to multiple channels
// On Connect, consumers receive the latest message
type Broadcaster[T any] struct {
	iox.AsyncCloser

	inject   chan func()
	in       <-chan T
	outs     map[int]chan T
	latest   T
	isLatest bool

	idx int
}

// NewBroadcaster creates and starts a new Broadcaster
func NewBroadcaster[T any](ctx context.Context) *Broadcaster[T] {
	b := &Broadcaster[T]{
		AsyncCloser: iox.NewAsyncCloser(),
		inject:      make(chan func()),
		in:          make(chan T),
		outs:        make(map[int]chan T),
	}
	go b.process(ctx)
	return b
}

// Connect connects a consumer to the Broadcaster.
// Returns a closer to allow the consumer to disconnect from broadcasts
func (b *Broadcaster[T]) Connect(ctx context.Context) (<-chan T, iox.AsyncCloser) {
	ret := make(chan T, 1)

	id, err := txn1(ctx, b, func() (int, error) {
		b.idx++
		b.outs[b.idx] = ret
		if b.isLatest {
			ret <- b.latest // initialize channel with latest message
		}
		return b.idx, nil
	})
	quit := iox.NewAsyncCloser()
	go func() {
		<-quit.Closed()
		_ = b.txn(ctx, func() error {
			out := b.outs[id]
			delete(b.outs, id)
			close(out)
			return nil
		})
	}()
	if err != nil {
		quit.Close()
		return ret, quit
	}
	return ret, quit
}

// Forward begins forwarding a new channel to all connected consumers. Waits on the previous message to finish sending.
func (b *Broadcaster[T]) Forward(ctx context.Context, in <-chan T) {
	_ = b.txn(ctx, func() error {
		var t T
		b.in = in
		b.latest = t
		b.isLatest = false
		return nil
	})
}

// process fans out to multiple other channels.
func (b *Broadcaster[T]) process(ctx context.Context) {
	defer b.Close()

	for {
		select {
		case t, ok := <-b.in:
			if !ok {
				return
			}
			b.latest = t
			b.isLatest = true
			for _, out := range b.outs {
				select {
				case <-out:
				default:
				}
				out <- t
			}
		case fn := <-b.inject:
			fn()
		case <-b.Closed():
			return
		case <-ctx.Done():
			return
		}
	}
}

// txn runs the given function in the main thread sync. Any signal that triggers a complex action must
// perform I/O or expensive parts outside txn and potentially use multiple txn calls.
func (b *Broadcaster[T]) txn(ctx context.Context, fn func() error) error {
	var wg sync.WaitGroup
	var err error

	wg.Add(1)
	select {
	case b.inject <- func() {
		defer wg.Done()
		err = fn()
	}:
		wg.Wait()
		return err
	case <-b.Closed():
		return fmt.Errorf("closed")
	case <-ctx.Done():
		return fmt.Errorf("closed")
	}
}

// txn1 is a convenience wrapper for r.txn with 1 return value of type T.
func txn1[T, V any](ctx context.Context, b *Broadcaster[T], fn func() (V, error)) (V, error) {
	var ret V
	var err error

	err = b.txn(ctx, func() error {
		ret, err = fn()
		return err
	})
	return ret, err
}
