package log_test

import (
	"context"
	"testing"

	"go.atoms.co/lib/log"
	logtesting "go.atoms.co/lib/log/testing"
)

// TestNonFatal validates that non-fatal user-facing logging functions call the backend as expected.
func TestNonFatal(t *testing.T) {
	ctx := context.Background()
	rec := logtesting.TestRecorder{}
	log.SetLogger(&rec)

	tests := []struct {
		name string
		fn   func()
		sev  log.Severity
		msg  string
	}{
		{"Debug", func() { log.Debug(ctx, "foo", "bar") }, log.SevDebug, "foobar"},
		{"Debugf", func() { log.Debugf(ctx, "%v-%v", "foo", "bar") }, log.SevDebug, "foo-bar"},
		{"Debugln", func() { log.Debugln(ctx, "foo", "bar") }, log.SevDebug, "foo bar\n"},

		{"Info", func() { log.Info(ctx, "foo", "bar") }, log.SevInfo, "foobar"},
		{"Infof", func() { log.Infof(ctx, "%v-%v", "foo", "bar") }, log.SevInfo, "foo-bar"},
		{"Infoln", func() { log.Infoln(ctx, "foo", "bar") }, log.SevInfo, "foo bar\n"},

		{"Warn", func() { log.Warn(ctx, "foo", "bar") }, log.SevWarn, "foobar"},
		{"Warnf", func() { log.Warnf(ctx, "%v-%v", "foo", "bar") }, log.SevWarn, "foo-bar"},
		{"Warnln", func() { log.Warnln(ctx, "foo", "bar") }, log.SevWarn, "foo bar\n"},

		{"Error", func() { log.Error(ctx, "foo", "bar") }, log.SevError, "foobar"},
		{"Errorf", func() { log.Errorf(ctx, "%v-%v", "foo", "bar") }, log.SevError, "foo-bar"},
		{"Errorln", func() { log.Errorln(ctx, "foo", "bar") }, log.SevError, "foo bar\n"},
	}

	for _, test := range tests {
		test.fn()

		calls, flushes := rec.Reset()
		if len(calls) != 1 || flushes > 0 {
			t.Fatalf("log.%v invoked the log backend incorrectly: (%v Log, %v Flush), want (1,0)", test.name, len(calls), flushes)
		}
		if calls[0].Sev != test.sev || calls[0].Msg != test.msg {
			t.Errorf("log.%v invoked Log with (%v, %v), want (%v, %v)", test.name, calls[0].Sev, calls[0].Msg, test.sev, test.msg)
		}
	}
}

// TestFatal validates that the Fatal user-facing logging functions call the backend as expected and then panics.
func TestFatal(t *testing.T) {
	ctx := context.Background()
	rec := logtesting.TestRecorder{}
	log.SetLogger(&rec)

	tests := []struct {
		name string
		fn   func()
		msg  string
	}{
		{"Fatal", func() { log.Fatal(ctx, "foo", "bar") }, "foobar"},
		{"Fatalf", func() { log.Fatalf(ctx, "%v-%v", "foo", "bar") }, "foo-bar"},
		{"Fatalln", func() { log.Fatalln(ctx, "foo", "bar") }, "foo bar\n"},
	}

	for _, test := range tests {
		msg, panicked := invokeAndRecover(test.fn)
		if !panicked {
			t.Errorf("log.%v failed to panic", test.name)
		}

		calls, flushes := rec.Reset()
		if len(calls) != 1 || flushes != 1 {
			t.Fatalf("log.%v invoked the log backend incorrectly: (%v Log, %v Flush), want (1,1)", test.name, len(calls), flushes)
		}
		if calls[0].Msg != test.msg || msg != test.msg {
			t.Errorf("log.%v invoked Log/panic with message %v/%v, want %v", test.name, calls[0].Msg, msg, test.msg)
		}
	}
}

func invokeAndRecover(fn func()) (msg string, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			msg = r.(string)
			panicked = true
		}
	}()

	fn()
	return "", false
}
