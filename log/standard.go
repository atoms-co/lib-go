package log

import (
	"context"
	"fmt"
	stdlog "log"
)

const (
	DebugColor   = 90 // Dark gray
	WarningColor = 93 // Light yellow
	ErrorColor   = 91 // Light red
)

// Standard is a wrapper over the standard Go log package.
type Standard struct {
	// Color enables colorful log output
	Color bool
}

// Log writes the message to stdlog, optionally using 8-ANSI colors for different log levels.
func (s *Standard) Log(ctx context.Context, sev Severity, calldepth int, msg string) {
	if s.Color && sev != SevInfo {
		color := DebugColor
		switch sev {
		case SevWarn:
			color = WarningColor
		case SevError:
			color = ErrorColor
		case SevFatal:
			color = ErrorColor
		}
		msg = fmt.Sprintf("\033[%vm%v\033[0m", color, msg)
	}
	_ = stdlog.Output(calldepth+2, msg) // +2 for this frame and stdlog.Output
}

func (s *Standard) Flush(ctx context.Context) error {
	return nil // nop
}
