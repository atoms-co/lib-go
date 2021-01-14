// Package zap contains an adaptor for a Zap logger backend.
package zap

import (
	"context"
	"go.atoms.co/lib/log"
	z "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// FieldOption enriches a log message with Zap fields, usually extracted from the context.
type FieldOption func(ctx context.Context, sev log.Severity, calldepth int, msg string) []z.Field

type logger struct {
	l    *z.Logger
	opts []FieldOption
}

// New returns a Logger wrapper over a Zap logger.
func New(l *z.Logger, opts ...FieldOption) log.Logger {
	if l == nil {
		panic("zap logger cannot be nil")
	}
	return &logger{l: l, opts: opts}
}

func (l logger) Log(ctx context.Context, sev log.Severity, calldepth int, msg string) {
	zl := l.l.WithOptions(z.AddCallerSkip(calldepth + 1)) // +1 for this frame
	if ce := zl.Check(toLevel(sev), msg); ce != nil {
		if sev == log.SevFatal {
			// Let the log package handle the side-effect, because both panic and os.Exit is
			// represented by SevFatal.

			ce = ce.Should(ce.Entry, zapcore.WriteThenNoop)
		}

		var fields []z.Field
		for _, opt := range l.opts {
			fields = append(fields, opt(ctx, sev, calldepth, msg)...)
		}

		ce.Write(fields...)
	}
}

func (l logger) Flush(ctx context.Context) error {
	return l.l.Sync()
}

func toLevel(sev log.Severity) zapcore.Level {
	switch sev {
	case log.SevUnspecified, log.SevDebug:
		return zapcore.DebugLevel
	case log.SevWarn:
		return zapcore.WarnLevel
	case log.SevError:
		return zapcore.ErrorLevel
	case log.SevFatal:
		return zapcore.FatalLevel

	default:
		return zapcore.InfoLevel
	}
}
