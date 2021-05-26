// Package backoffx is a convenience wrapper over the backoff package.
package backoffx

import (
	"time"

	"go.atoms.co/lib/clock"
	"github.com/cenkalti/backoff/v4"
)

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

// NewUnlimited returns a new exponential backoff without a max elapsed time.
func NewUnlimited(opts ...Option) backoff.BackOff {
	return NewLimited(0, opts...)
}

// NewLimited returns a new exponential backoff with the given max elapsed time.
func NewLimited(max time.Duration, opts ...Option) backoff.BackOff {
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

	if n > 0 {
		return backoff.WithMaxRetries(b, uint64(n))
	}
	return b
}

// Retry retries the given functions per the backoff policy.
func Retry(b backoff.BackOff, fn func() error) error {
	return backoff.Retry(fn, b)
}

// ErrPermanent signals a permanent error and halts the retry attempts, if returned.
func ErrPermanent(err error) error {
	return backoff.Permanent(err)
}
