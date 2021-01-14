package log

import (
	"fmt"
	"strings"
)

// Severity represents the severity of the log message.
type Severity int

const (
	SevUnspecified Severity = iota
	SevDebug
	SevInfo
	SevWarn
	SevError
	SevFatal
)

func (s Severity) String() string {
	switch s {
	case SevUnspecified:
		return "unspecified"
	case SevDebug:
		return "debug"
	case SevInfo:
		return "info"
	case SevWarn:
		return "warn"
	case SevError:
		return "error"
	case SevFatal:
		return "fatal"
	default:
		return fmt.Sprintf("<unknown:%v>", int(s))
	}
}

// ParseSeverity returns a Severity from a string. Case-insensitive.
func ParseSeverity(v string) (Severity, error) {
	switch strings.ToLower(v) {
	case "", "unspecified":
		return SevUnspecified, nil
	case "debug":
		return SevDebug, nil
	case "info":
		return SevInfo, nil
	case "warn", "warning":
		return SevWarn, nil
	case "error":
		return SevError, nil
	case "fatal":
		return SevFatal, nil
	default:
		return SevUnspecified, fmt.Errorf("invalid severity: '%v'", v)
	}
}
