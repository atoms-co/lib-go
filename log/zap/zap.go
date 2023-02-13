// Package zap contains an adaptor for a Zap logger backend.
package zap

import (
	"context"

	"go.cloudkitchens.org/lib/log"
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

func WithContextFields(ctx context.Context, sev log.Severity, calldepth int, msg string) []z.Field {
	generic := log.FromContext(ctx)
	zfields := make([]z.Field, 0, len(generic))
	for _, field := range generic {
		zfields = append(zfields, z.Field{
			Key:       field.Key,
			Type:      zapType(field.Type),
			Integer:   field.Integer,
			String:    field.String,
			Interface: field.Interface,
		})
	}
	return zfields
}

func zapType(logType log.FieldType) zapcore.FieldType {
	switch logType {
	// BinaryType indicates that the field carries an opaque binary blob.
	case log.BinaryType:
		return zapcore.BinaryType
	case log.BoolType:
		return zapcore.BoolType
	case log.ByteStringType:
		return zapcore.ByteStringType
	case log.Complex128Type:
		return zapcore.Complex128Type
	case log.Complex64Type:
		return zapcore.Complex64Type
	case log.DurationType:
		return zapcore.DurationType
	case log.Float64Type:
		return zapcore.Float64Type
	case log.Float32Type:
		return zapcore.Float32Type
	case log.Int64Type:
		return zapcore.Int64Type
	case log.Int32Type:
		return zapcore.Int32Type
	case log.Int16Type:
		return zapcore.Int16Type
	case log.Int8Type:
		return zapcore.Int8Type
	case log.StringType:
		return zapcore.StringType
	case log.TimeType:
		return zapcore.TimeType
	case log.TimeFullType:
		return zapcore.TimeFullType
	case log.Uint64Type:
		return zapcore.Uint64Type
	case log.Uint32Type:
		return zapcore.Uint32Type
	case log.Uint16Type:
		return zapcore.Uint16Type
	case log.Uint8Type:
		return zapcore.Uint8Type
	case log.UintptrType:
		return zapcore.UintptrType
	case log.StringerType:
		return zapcore.StringerType
	case log.ErrorType:
		return zapcore.ErrorType
	default:
		// UnknownType is the default field type. Attempting to add it to an encoder will panic.
		return zapcore.UnknownType
	}
}
