// Package backoffx is a convenience wrapper over the backoff package.
package backoffx

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"

	"go.atoms.co/lib/clock"
)

type BackOff = backoff.BackOff
type BackOffContext = backoff.BackOffContext

// Option represents an option for backoff.
type Option struct {
	fn  func(*backoff.ExponentialBackOff)
	max int
}

// WithClock uses the provided clock for backoff. Default: system clock.
func WithClock(c clock.Clock) Option {
	return Option{fn: func(b *backoff.ExponentialBackOff) {
		b.Clock = c
	}}
}

// WithInitialInterval sets the initial backoff interval. Default: 500ms.
func WithInitialInterval(interval time.Duration) Option {
	return Option{fn: func(b *backoff.ExponentialBackOff) {
		b.InitialInterval = interval
	}}
}

// WithMaxInterval sets the maximum interval. Default: 5s.
func WithMaxInterval(interval time.Duration) Option {
	return Option{fn: func(b *backoff.ExponentialBackOff) {
		b.MaxInterval = interval
	}}
}

// WithMaxRetries set the maximum number of retries regardless of the time used. Default: no maximum.
func WithMaxRetries(n int) Option {
	return Option{max: n}
}

// WithContext returns a BackOffContext with context ctx
//
// ctx must not be nil
func WithContext(b BackOff, ctx context.Context) BackOffContext {
	return backoff.WithContext(b, ctx)
}

// NewUnlimited returns a new exponential backoff without a max elapsed time.
func NewUnlimited(opts ...Option) BackOff {
	return NewLimited(0, opts...)
}

// NewLimited returns a new exponential backoff with the given max elapsed time.
func NewLimited(max time.Duration, opts ...Option) BackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 500 * time.Millisecond
	b.MaxInterval = 5 * time.Second
	b.MaxElapsedTime = max

	n := 0
	for _, opt := range opts {
		switch {
		case opt.fn != nil:
			opt.fn(b)
		case opt.max > 0:
			n = opt.max
		}
	}
	b.Reset()

	if n > 0 {
		return backoff.WithMaxRetries(b, uint64(n))
	}
	return b
}

// Retry retries the given function per the backoff policy.
func Retry(b BackOff, fn func() error) error {
	return backoff.Retry(fn, b)
}

// Retry1 retries the given function per the backoff policy.
// Note that this method will not pass the value back if the operation returned an error: it will return
// zero value.
func Retry1[T1 any](b BackOff, fn func() (T1, error)) (T1, error) {
	var t1 T1
	var err error
	err = backoff.Retry(func() error {
		t1, err = fn()
		return err
	}, b)
	if err != nil {
		var nil1 T1
		return nil1, err
	}
	return t1, nil
}

// Retry2 retries the given function per the backoff policy.
// Note that this method will not pass the values back if the operation returned an error: it will return
// zero values.
func Retry2[T1, T2 any](b BackOff, fn func() (T1, T2, error)) (T1, T2, error) {
	var t1 T1
	var t2 T2
	var err error
	err = backoff.Retry(func() error {
		t1, t2, err = fn()
		return err
	}, b)
	if err != nil {
		var nil1 T1
		var nil2 T2
		return nil1, nil2, err
	}
	return t1, t2, nil
}

// Retry3 retries the given function per the backoff policy.
// Note that this method will not pass the values back if the operation returned an error: it will return
// zero values.
func Retry3[T1, T2, T3 any](b BackOff, fn func() (T1, T2, T3, error)) (T1, T2, T3, error) {
	var t1 T1
	var t2 T2
	var t3 T3
	var err error
	err = backoff.Retry(func() error {
		t1, t2, t3, err = fn()
		return err
	}, b)
	if err != nil {
		var nil1 T1
		var nil2 T2
		var nil3 T3
		return nil1, nil2, nil3, err
	}
	return t1, t2, t3, nil
}

// ErrPermanent signals a permanent error and halts the retry attempts, if returned.
func ErrPermanent(err error) error {
	return backoff.Permanent(err)
}
