package timex

import (
	"time"

	"go.atoms.co/lib/chanx"
)

// Timer is a wrapper for the standard timer that drains the C channel on Reset and Stop. Not threadsafe.
type Timer struct {
	timer *time.Timer
	ttl   time.Time
	C     <-chan time.Time
}

// NewTimer creates and starts a timer with the specified duration
func NewTimer(d time.Duration) *Timer {
	timer := time.NewTimer(d)
	return &Timer{
		timer: timer,
		C:     timer.C,
		ttl:   time.Now().Add(d),
	}
}

// AfterFunc starts a timer with the specified duration and function
func AfterFunc(d time.Duration, fn func()) *Timer {
	timer := time.AfterFunc(d, fn)
	return &Timer{
		timer: timer,
		C:     timer.C,
		ttl:   time.Now().Add(d),
	}
}

func (t *Timer) TTL() time.Time {
	return t.ttl
}

func (t *Timer) Reset(d time.Duration) {
	if !t.timer.Stop() {
		chanx.Clear(t.timer.C)
	}
	t.timer.Reset(d)
	t.ttl = time.Now().Add(d)
}

func (t *Timer) Stop() {
	if !t.timer.Stop() {
		chanx.Clear(t.timer.C)
	}
	t.ttl = time.Now()
}
