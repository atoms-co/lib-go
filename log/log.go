// Package log is a lightweight logging system with context and a re-targetable backend.
package log

import (
	"context"
	"fmt"
	"os"
)

// Logger is a context-aware logging backend. The richer context allows for more sophisticated logging setups.
// Must be concurrency safe. Should respect ordering.
type Logger interface {
	// Log logs the message in some implementation-dependent way. Log should always return regardless of the severity.
	Log(ctx context.Context, sev Severity, calldepth int, msg string)

	// Flush forces a flush of all buffered log messages to the underlying storage, even if it incurs some delay.
	// Usually called right before a panic or shutdown, when the last messages are the most important ones. Even if
	// an error is returned, the logger should not drop messages or abandon any background tasks.
	Flush(ctx context.Context) error
}

var (
	logger Logger = &Standard{}
)

// SetLogger sets the global Logger. Intended to be called during initialization only.
func SetLogger(l Logger) {
	if l == nil {
		panic("Logger cannot be nil")
	}
	logger = l
}

// Output logs the given message to the global logger. Calldepth is the count of the number of frames to skip when
// computing the file name and line number.
func Output(ctx context.Context, sev Severity, calldepth int, msg string) {
	logger.Log(ctx, sev, calldepth+1, msg) // +1 for this frame
}

// Flush forces a flushes of log messages buffered by the global logger.
func Flush(ctx context.Context) error {
	return logger.Flush(ctx)
}

// User-facing logging functions.

// Debug writes the fmt.Sprint-formatted arguments to the global logger with debug severity.
func Debug(ctx context.Context, v ...interface{}) {
	Output(ctx, SevDebug, 1, fmt.Sprint(v...))
}

// Debugf writes the fmt.Sprintf-formatted arguments to the global logger with debug severity.
func Debugf(ctx context.Context, format string, v ...interface{}) {
	Output(ctx, SevDebug, 1, fmt.Sprintf(format, v...))
}

// Debugln writes the fmt.Sprintln-formatted arguments to the global logger with debug severity.
func Debugln(ctx context.Context, v ...interface{}) {
	Output(ctx, SevDebug, 1, fmt.Sprintln(v...))
}

// Info writes the fmt.Sprint-formatted arguments to the global logger with info severity.
func Info(ctx context.Context, v ...interface{}) {
	Output(ctx, SevInfo, 1, fmt.Sprint(v...))
}

// Infof writes the fmt.Sprintf-formatted arguments to the global logger with info severity.
func Infof(ctx context.Context, format string, v ...interface{}) {
	Output(ctx, SevInfo, 1, fmt.Sprintf(format, v...))
}

// Infoln writes the fmt.Sprintln-formatted arguments to the global logger with info severity.
func Infoln(ctx context.Context, v ...interface{}) {
	Output(ctx, SevInfo, 1, fmt.Sprintln(v...))
}

// Warn writes the fmt.Sprint-formatted arguments to the global logger with warn severity.
func Warn(ctx context.Context, v ...interface{}) {
	Output(ctx, SevWarn, 1, fmt.Sprint(v...))
}

// Warnf writes the fmt.Sprintf-formatted arguments to the global logger with warn severity.
func Warnf(ctx context.Context, format string, v ...interface{}) {
	Output(ctx, SevWarn, 1, fmt.Sprintf(format, v...))
}

// Warnln writes the fmt.Sprintln-formatted arguments to the global logger with warn severity.
func Warnln(ctx context.Context, v ...interface{}) {
	Output(ctx, SevWarn, 1, fmt.Sprintln(v...))
}

// Error writes the fmt.Sprint-formatted arguments to the global logger with error severity.
func Error(ctx context.Context, v ...interface{}) {
	Output(ctx, SevError, 1, fmt.Sprint(v...))
}

// Errorf writes the fmt.Sprintf-formatted arguments to the global logger with  error severity.
func Errorf(ctx context.Context, format string, v ...interface{}) {
	Output(ctx, SevError, 1, fmt.Sprintf(format, v...))
}

// Errorln writes the fmt.Sprintln-formatted arguments to the global logger with error severity.
func Errorln(ctx context.Context, v ...interface{}) {
	Output(ctx, SevError, 1, fmt.Sprintln(v...))
}

// Fatal writes the fmt.Sprint-formatted arguments to the global logger with fatal severity then panics.
func Fatal(ctx context.Context, v ...interface{}) {
	msg := fmt.Sprint(v...)
	Output(ctx, SevFatal, 1, msg)
	flush(ctx, msg)
	panic(msg)
}

// Fatalf writes the fmt.Sprintf-formatted arguments to the global logger with fatal severity then panics.
func Fatalf(ctx context.Context, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	Output(ctx, SevFatal, 1, msg)
	flush(ctx, msg)
	panic(msg)
}

// Fatalln writes the fmt.Sprintln-formatted arguments to the global logger with fatal severity then panics.
func Fatalln(ctx context.Context, v ...interface{}) {
	msg := fmt.Sprintln(v...)
	Output(ctx, SevFatal, 1, msg)
	flush(ctx, msg)
	panic(msg)
}

// Exit writes the fmt.Sprint-formatted arguments to the global logger with fatal severity then exits.
func Exit(ctx context.Context, v ...interface{}) {
	msg := fmt.Sprint(v...)
	Output(ctx, SevFatal, 1, msg)
	flush(ctx, msg)
	os.Exit(1)
}

// Exitf writes the fmt.Sprintf-formatted arguments to the global logger with fatal severity then exits.
func Exitf(ctx context.Context, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	Output(ctx, SevFatal, 1, msg)
	flush(ctx, msg)
	os.Exit(1)
}

// Exitln writes the fmt.Sprintln-formatted arguments to the global logger with fatal severity then exits.
func Exitln(ctx context.Context, v ...interface{}) {
	msg := fmt.Sprintln(v...)
	Output(ctx, SevFatal, 1, msg)
	flush(ctx, msg)
	os.Exit(1)
}

func flush(ctx context.Context, msg string) {
	if err := Flush(ctx); err != nil {
		// If Flush fails, we have no good options left before panic or os.Exit. Last ditch effort is print to stderr.
		println("Failed to flushes logs after Fatal/Exit: ", msg)
	}
}
