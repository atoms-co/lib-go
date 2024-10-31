// Package iox contains various io utilities.
package iox

import (
	"context"
	"sync/atomic"
)

// AsyncCloser is an async closer that supports a quit chan as a close notification mechanism. Thread-safe.
type AsyncCloser interface {
	RAsyncCloser
	WAsyncCloser
}

// RAsyncCloser is an async closer that supports a read-only quit chan as a close notification mechanism. Thread-safe.
type RAsyncCloser interface {
	// IsClosed returns true iff the instance is closed.
	IsClosed() bool
	// Closed returns a quit chan that is closed iff the instance is closed.
	Closed() <-chan struct{}
}

// WAsyncCloser is an async closer that supports a quit chan as a close mechanism. Thread-safe.
type WAsyncCloser interface {
	// Close closes the instance. No error is returned as it is usually called in a defer.
	Close()
}

type asyncCloser struct {
	quit   chan struct{}
	closed atomic.Bool
}

// NewAsyncCloser returns a new open closer.
func NewAsyncCloser() AsyncCloser {
	return &asyncCloser{quit: make(chan struct{})}
}

func (c *asyncCloser) IsClosed() bool {
	return c.closed.Load()
}

func (c *asyncCloser) Closed() <-chan struct{} {
	return c.quit
}

func (c *asyncCloser) Close() {
	if c.closed.CompareAndSwap(false, true) {
		close(c.quit)
	}
}

// WithCancel closes the closer on context closure. Returns original closure for convenience.
func WithCancel(ctx context.Context, closer AsyncCloser) AsyncCloser {
	return WithQuit(ctx.Done(), closer)
}

// WithQuit closes the closer, if the quit channel if closed. Returns original closure for convenience.
func WithQuit(quit <-chan struct{}, closer AsyncCloser) AsyncCloser {
	go func() {
		select {
		case <-quit:
			closer.Close()
		case <-closer.Closed():
		}
	}()
	return closer
}

// WhenClosed closes the child closer, when the parent closes
func WhenClosed(parent, child AsyncCloser) {
	go func() {
		<-parent.Closed()
		child.Close()
	}()
}
