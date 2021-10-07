// Package logx is a convenience package for initializing pre-configured loggers on the command line.
package logx

import (
	"context"
	"flag"
	"fmt"
	stdlog "log"

	"go.atoms.co/lib/log"
	"go.atoms.co/lib/log/zap"
)

var (
	logger = flag.String("logger", "stackdriver", "Logger to use")
	level  = flag.String("log-level", "", "Log severity cutoff (default: no cutoff)")
)

// Init initializes the global logger as configured via flags.
func Init(ctx context.Context) {
	log.SetLogger(filter(load()))
	log.Debugf(ctx, "Initialized logger: %v, level: %v", *logger, *level)
}

func filter(l log.Logger) log.Logger {
	sev, err := log.ParseSeverity(*level)
	if err != nil {
		panic(fmt.Sprintf("Invalid log-level: %v", err))
	}
	return log.Filter(l, sev)
}

func load() log.Logger {
	switch *logger {
	case "", "standard":
		stdlog.SetFlags(stdlog.Ldate | stdlog.Lmicroseconds | stdlog.Lshortfile)
		return &log.Standard{Color: true}

	case "stackdriver":
		return zap.NewStackdriver(zap.WithContextFields)

	default:
		panic(fmt.Sprintf("Unknown logger: %v", *logger))
	}
}
