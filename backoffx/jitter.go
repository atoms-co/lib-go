package backoffx

import (
	"math/rand"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// JitterBackoff adds jitter to an underlying backoff (like
// backoff.ConstantBackoff)
type JitterBackoff struct {
	b         backoff.BackOff
	maxJitter time.Duration
}

func WithJitter(b backoff.BackOff, maxJitter time.Duration) *JitterBackoff {
	return &JitterBackoff{
		b:         b,
		maxJitter: maxJitter,
	}
}

func (j *JitterBackoff) Reset() { j.b.Reset() }

func (j *JitterBackoff) NextBackOff() time.Duration {
	next := j.b.NextBackOff()
	if next == backoff.Stop {
		return next
	}

	return next + time.Duration(rand.Int63n(int64(j.maxJitter)))
}
