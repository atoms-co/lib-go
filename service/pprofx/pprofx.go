package pprofx

import (
	"context"
	"fmt"
	"net/http"

	"go.atoms.co/lib/log"
)

type options struct {
	port int
}

// Option is a test server option.
type Option func(*options)

func WithPort(port int) Option {
	return func(o *options) {
		o.port = port
	}
}

// Start the default pprof handlers on the pprof port: https://golang.org/pkg/net/http/pprof/.
func Start(ctx context.Context, opts ...Option) {
	o := options{
		port: 6060,
	}
	for _, fn := range opts {
		fn(&o)
	}

	log.Infof(ctx, "Setting up pprof on port: %v", o.port)
	log.Fatal(ctx, http.ListenAndServe(fmt.Sprintf(":%v", o.port), nil))
}
