// Package contextx contains convenience utilities for working with contexts.
package contextx

import (
	"context"
	"time"
)

// WithQuitCancel returns a cancellable context. Additionally, if the quit channel is closed the context is cancelled.
func WithQuitCancel(ctx context.Context, quit <-chan struct{}) (context.Context, context.CancelFunc) {
	wctx, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-quit:
			cancel()
		case <-wctx.Done():
			// ok: caller cancelled the context
		}
	}()

	return wctx, cancel
}

// WithQuitTimeout returns a cancellable context that times out. Additionally, if the quit channel is closed the
// context is cancelled.
func WithQuitTimeout(ctx context.Context, quit <-chan struct{}, timeout time.Duration) (context.Context, context.CancelFunc) {
	wctx, cancel := context.WithTimeout(ctx, timeout)

	go func() {
		select {
		case <-quit:
			cancel()
		case <-wctx.Done():
			// ok: caller cancelled the context
		}
	}()

	return wctx, cancel
}
