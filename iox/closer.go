// Package iox contains various io utilities.
package iox

import "go.uber.org/atomic"

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
