package log

import "context"

type filter struct {
	l      Logger
	cutoff Severity
}

// Filter is a wrapper that drops any logs below the given severity, such as debug or info.
func Filter(l Logger, cutoff Severity) Logger {
	if l == nil {
		panic("nil logger")
	}
	if cutoff <= SevUnspecified {
		return l // nop: no filtering needed
	}
	return &filter{l: l, cutoff: cutoff}
}

func (f *filter) Log(ctx context.Context, sev Severity, calldepth int, msg string) {
	if sev < f.cutoff {
		return
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
