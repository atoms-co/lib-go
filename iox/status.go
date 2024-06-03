package iox

import (
	"fmt"
	"time"
)

// Failure represents an I/O operation failure that can only be retried after a given time. Value type.
type Failure struct {
	Err   error
	Retry time.Time
}

func (f Failure) String() string {
	return fmt.Sprintf("[err=%v, retry=%v]", f.Err, f.Retry.Unix())
}

// Status represents the status of a pending, retryable I/O operation. The default value is
// an operation that has yet to be attempted.
type Status struct {
	Last     *Failure
	Attempts int
	Inflight bool
}

func (s Status) ToInflight() Status {
	return Status{
		Last:     s.Last,
		Attempts: s.Attempts,
		Inflight: true,
	}
}

func (s Status) ToFailed(err error, retry time.Time) Status {
	return Status{
		Last:     &Failure{Err: err, Retry: retry},
		Attempts: s.Attempts + 1,
		Inflight: false,
	}
}

func (s Status) IsReady(now time.Time) bool {
	if s.Inflight {
		return false
	}
	if s.Last != nil {
		return now.After(s.Last.Retry)
	}
	return true
}

func (s Status) String() string {
	status := "pending"
	if s.Inflight {
		status = "inflight"
	}
	if s.Last == nil {
		return status
	}
	return fmt.Sprintf("%v[attempts=%v, last=%v]", status, s.Attempts, *s.Last)
}
