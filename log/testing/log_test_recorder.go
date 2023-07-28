package log_testing

import (
	"context"

	"go.atoms.co/lib/log"
)

type Call struct {
	Sev       log.Severity
	Calldepth int
	Fields    []log.Field
	Msg       string
}

// TestRecorder is a simple test Logger that records the invocations. This is useful for testing to verify
// logged data is as expected
type TestRecorder struct {
	Calls   []Call
	Flushes int
}

func (l *TestRecorder) Log(ctx context.Context, sev log.Severity, calldepth int, msg string) {
	l.Calls = append(l.Calls, Call{Sev: sev, Calldepth: calldepth, Msg: msg, Fields: log.FromContext(ctx)})
}

func (l *TestRecorder) Flush(ctx context.Context) error {
	l.Flushes++
	return nil
}

func (l *TestRecorder) Reset() ([]Call, int) {
	calls, flushes := l.Calls, l.Flushes
	l.Calls = nil
	l.Flushes = 0
	return calls, flushes
}
