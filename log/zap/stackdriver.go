package zap

import (
	"os"

	z "go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.cloudkitchens.org/lib/log"
)

// NewStackdriver returns a new zap-based log.Logger with Stackdriver-compatible formatting.
func NewStackdriver(opts ...FieldOption) log.Logger {
	core := NewStackdriverCore(zapcore.Lock(os.Stderr))

	// Enable call information. Enable zap-managed stack trace on errors and above.
	return New(z.New(core, z.AddCaller(), z.AddStacktrace(zapcore.ErrorLevel)), opts...)
}

// NewStackdriverCore creates a new Zap core with Stackdriver formatting.
func NewStackdriverCore(out zapcore.WriteSyncer) zapcore.Core {
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "severity",
		NameKey:       "logger",
		MessageKey:    "message",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel: func(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(toSeverity(level))
		},
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	})

	core := zapcore.NewCore(encoder, out, z.DebugLevel)
	return NewTransformCore(z.DebugLevel, core, toSourceLocation)
}

// toSourceLocation transforms Caller information to match Stackdriver expectations as defined by
// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogEntrySourceLocation. Disables Caller.
func toSourceLocation(entry zapcore.Entry, fields []zapcore.Field) (zapcore.Entry, []zapcore.Field) {
	if !entry.Caller.Defined {
		return entry, fields // nop: no caller information
	}
	entry.Caller.Defined = false

	const (
		key     = "logging.googleapis.com/sourceLocation"
		fileKey = "file"
		lineKey = "line"
	)

	source := z.Object(key, zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		enc.AddString(fileKey, entry.Caller.File)
		enc.AddInt(lineKey, entry.Caller.Line)
		return nil
	}))

	return entry, append(fields, source)
}

// toSeverity converts a zap Level to a Stackdriver-format string.
// Based on https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
func toSeverity(l zapcore.Level) string {
	switch l {
	case zapcore.DebugLevel:
		return "DEBUG"
	case zapcore.InfoLevel:
		return "INFO"
	case zapcore.WarnLevel:
		return "WARNING"
	case zapcore.ErrorLevel:
		return "ERROR"
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return "CRITICAL"
	default:
		return "DEFAULT"
	}
}
