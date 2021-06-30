// Package iox contains various io utilities.
package iox

import (
	"context"

	"go.uber.org/atomic"
)

// AsyncCloser is an async closer that support a quit chan as a close notification mechanism. Thread-safe.
type AsyncCloser interface {
	// IsClosed returns true iff the instance is closed.
	IsClosed() bool
	// Closed returns a quit chan that is closed iff the instance is closed.
	Closed() <-chan struct{}

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
	if c.closed.CAS(false, true) {
		close(c.quit)
	}
}

// WithCancel closes the closer on context closure. Returns original closure for convenience.
func WithCancel(ctx context.Context, closer AsyncCloser) AsyncCloser {
	go func() {
		select {
		case <-ctx.Done():
			closer.Close()
		case <-closer.Closed():
		}
	}()
	return closer
}

// WithCascade closes the closer, if the upstream if closed. Returns original closure for convenience.
func WithCascade(upstream, closer AsyncCloser) AsyncCloser {
	go func() {
		select {
		case <-upstream.Closed():
			closer.Close()
		case <-upstream.Closed():
		}
	}()
	return closer
}
