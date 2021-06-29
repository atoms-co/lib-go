package metricsx

import (
	"context"
	"flag"
	"fmt"

	promExporter "contrib.go.opencensus.io/exporter/prometheus"
	"github.com/gorilla/mux"
	"net/http"

	"go.cloudkitchens.org/lib/log"
	"go.cloudkitchens.org/lib/metrics"
)

var (
	port = flag.Int("prometheus_port", 9090, "Prometheus metrics exporter port")
	srv  = &http.Server{}
)

// Init initializes metrics with the given application name and exports them on port 9090.
func Init(ctx context.Context, application string) {
	log.Infof(ctx, "Initializing prometheus metrics on :%v", *port)

	pe, err := promExporter.NewExporter(promExporter.Options{})
	if err != nil {
		log.Exitf(ctx, "Failed to create Prometheus exporter: %v", err)
	}

	r := mux.NewRouter()
	r.Handle("/metrics", pe)
	srv = &http.Server{Addr: fmt.Sprintf(":%v", *port), Handler: r}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf(ctx, "Metrics server exited unexpectedly: %v", err)
		}
	}()

	if err := metrics.Init(application); err != nil {
		log.Exitf(ctx, "Failed to initialize metric: %v", err)
	}
}

func Shutdown(ctx context.Context) error {
	return srv.Shutdown(ctx)
}
