package log

import "context"

const logCutoffKey logCtxKeyType = "log_cutoff"

// WithCutoff returns a context with the given severity cutoff, used by the Filter logger. The severity
// will take effect where the context flows, incl. helper functions and libraries that may be called in
// various contexts. The contextual setting thus allows a precise filtering than the class-scoped filtering
// approach typically used in OO-centric languages like Java.
func WithCutoff(ctx context.Context, cutoff Severity) context.Context {
	return context.WithValue(ctx, logCutoffKey, cutoff)
}

// CutoffValue returns an effective severity set by WithCutoff, if present.
func CutoffValue(ctx context.Context) (Severity, bool) {
	sev, ok := ctx.Value(logCutoffKey).(Severity)
	return sev, ok
}

type filter struct {
	l      Logger
	cutoff Severity
}

// Filter is a wrapper that drops any logs below the given severity, such as debug or info. The effective severity
// may also be passed in the context to disable/enable noisy parts of a program, notably external libraries.
func Filter(l Logger, cutoff Severity) Logger {
	if l == nil {
		panic("nil logger")
	}
	return &filter{l: l, cutoff: cutoff}
}

func (f *filter) Log(ctx context.Context, sev Severity, calldepth int, msg string) {
	if s, ok := CutoffValue(ctx); ok {
		if sev < s {
			return // omit: cutoff override
		}
	} else if sev < f.cutoff {
		return // omit: default cutoff, no context override
	}
	f.l.Log(ctx, sev, calldepth+1, msg) // +1 for this frame
}

func (f *filter) Flush(ctx context.Context) error {
	return f.l.Flush(ctx)
}

type dynamicFilter struct {
	l         Logger
	shouldLog func(ctx context.Context, sev Severity) bool
}

// DynamicFilter is a wrapper that drops any logs given a dynamic condition
// shouldLog must be thread-safe
func DynamicFilter(l Logger, shouldLog func(ctx context.Context, sev Severity) bool) Logger {
	if l == nil {
		panic("nil logger")
	}
	return &dynamicFilter{l: l, shouldLog: shouldLog}
}

func (f *dynamicFilter) Log(ctx context.Context, sev Severity, calldepth int, msg string) {
	if !f.shouldLog(ctx, sev) {
		return
	}
	f.l.Log(ctx, sev, calldepth+1, msg) // +1 for this frame
}

func (f *dynamicFilter) Flush(ctx context.Context) error {
	return f.l.Flush(ctx)
}
