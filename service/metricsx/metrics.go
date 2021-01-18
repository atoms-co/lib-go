package metricsx

import (
	"context"
	"flag"
	"fmt"

	promExporter "contrib.go.opencensus.io/exporter/prometheus"
	"go.cloudkitchens.org/lib/log"
	"go.cloudkitchens.org/lib/metrics"
	"github.com/gorilla/mux"
	"net/http"
)

var (
	port = flag.Int("prometheus_port", 9090, "Prometheus metrics exporter port")
)

// Init initializes metrics with the givne application name and exports them on port 9090.
func Init(ctx context.Context, application string) {
	log.Infof(ctx, "Initializing prometheus metrics on :%v", *port)

	pe, err := promExporter.NewExporter(promExporter.Options{})
	if err != nil {
		log.Exitf(ctx, "Failed to create Prometheus exporter: %v", err)
	}

	r := mux.NewRouter()
	r.Handle("/metrics", pe)
	go func() {
		log.Fatalf(ctx, "Metrics server exited: %v", http.ListenAndServe(fmt.Sprintf(":%v", *port), r))
	}()

	if err := metrics.Init(application); err != nil {
		log.Exitf(ctx, "Failed to initialize metric: %v", err)
	}
}
