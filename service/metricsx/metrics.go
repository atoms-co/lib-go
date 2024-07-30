package metricsx

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	promExporter "contrib.go.opencensus.io/exporter/prometheus"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	promethusOTel "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"

	"go.atoms.co/lib/log"
	"go.atoms.co/lib/metrics"
)

var (
	port = flag.Int("prometheus_port", 9090, "Prometheus metrics exporter port")
	srv  = &http.Server{}
)

// Init initializes metrics with the given application name and exports them on port 9090.
func Init(ctx context.Context, application string) {
	log.Infof(ctx, "Initializing prometheus metrics on :%v", *port)

	pe, err := promExporter.NewExporter(promExporter.Options{
		Registerer: prometheus.DefaultRegisterer,
		Gatherer:   prometheus.DefaultGatherer,
	})
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

	initOTelMetricExporter(ctx)

	if err := metrics.Init(application); err != nil {
		log.Exitf(ctx, "Failed to initialize metric: %v", err)
	}
}

func Shutdown(ctx context.Context) error {
	return srv.Shutdown(ctx)
}

func initOTelMetricExporter(ctx context.Context) {
	exporter, err := promethusOTel.New()
	if err != nil {
		log.Exitf(ctx, "Failed to initialize OTel metric exporter: %v", err)
	}
	meterProvider := metric.NewMeterProvider(metric.WithReader(exporter))
	otel.SetMeterProvider(meterProvider)
}
