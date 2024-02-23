// Package contextx contains convenience utilities for working with contexts.
package contextx

import (
	"context"
	"time"
)

// IsCancelled returns true if the context is cancelled, i.e, if ctx.Done() is closed.
func IsCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

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

// WithQuitCancelDelay returns a cancellable context with a delay.
// Additionally, if the quit channel is closed the context is cancelled.
func WithQuitCancelDelay(ctx context.Context, quit <-chan struct{}, delay time.Duration) (context.Context, context.CancelFunc) {
	wctx, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-quit:
			time.AfterFunc(delay, cancel)
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
