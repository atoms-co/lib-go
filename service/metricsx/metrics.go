package metricsx

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	otelexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"

	"go.atoms.co/lib/log"
	"go.atoms.co/lib/metrics"
)

var (
	port = flag.Int("prometheus_port", 9090, "Prometheus metrics exporter port")
	srv  = &http.Server{}
)

// Option configures the metrics initialization.
type Option func(*config)

type config struct {
	otelExporterOptions []otelexporter.Option
	views               []metric.View
}

// WithOTelExporterOptions passes additional options to the OTel Prometheus
// exporter. metricsx disables exporter-added counter and unit suffixes by
// default. Repository-owned instruments must declare their final Prometheus
// names; exact Views preserve names from supported auto-instrumentation.
// Scope and target metadata remain enabled unless a service opts out.
func WithOTelExporterOptions(opts ...otelexporter.Option) Option {
	return func(c *config) {
		c.otelExporterOptions = append(c.otelExporterOptions, opts...)
	}
}

// WithViews registers SDK Views on the global OTel MeterProvider. Use this to
// rename, re-aggregate, or filter attributes off specific instruments (e.g.,
// drop high-cardinality attributes from auto-instrumented HTTP metrics).
// For supported auto-instrumented metrics, metricsx enforces the established
// Prometheus name while preserving the View's other stream settings.
func WithViews(views ...metric.View) Option {
	return func(c *config) {
		c.views = append(c.views, views...)
	}
}

// Init initializes metrics with the given application name and exports them on port 9090.
func Init(ctx context.Context, application string, opts ...Option) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}

	log.Infof(ctx, "Initializing prometheus metrics on :%v", *port)

	initOTelMetricExporter(ctx, cfg)

	if err := metrics.Init(application); err != nil {
		log.Exitf(ctx, "Failed to initialize metric: %v", err)
	}

	m := http.NewServeMux()
	m.Handle("/metrics", prometheusHandler(prometheus.DefaultGatherer))
	srv = &http.Server{Addr: fmt.Sprintf(":%v", *port), Handler: m}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf(ctx, "Metrics server exited unexpectedly: %v", err)
		}
	}()
}

func Shutdown(ctx context.Context) error {
	return srv.Shutdown(ctx)
}

func initOTelMetricExporter(ctx context.Context, cfg *config) {
	log.Infof(ctx, "Initializing OTel metric exporter with %d option(s)", len(cfg.otelExporterOptions))
	exporter, err := otelexporter.New(cfg.otelExporterOptions...)
	if err != nil {
		log.Exitf(ctx, "Failed to initialize OTel metric exporter: %v", err)
	}
	views := explicitPrometheusNameViews(cfg.views)
	if len(views) > 0 {
		log.Infof(ctx, "Registering %d OTel metric view(s)", len(views))
	}
	meterProvider := newMeterProvider(exporter, views)
	otel.SetMeterProvider(meterProvider)
}

func prometheusHandler(gatherer prometheus.Gatherer) http.Handler {
	return promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})
}

func newMeterProvider(reader metric.Reader, views []metric.View) *metric.MeterProvider {
	mpOpts := []metric.Option{
		metric.WithReader(reader),
		metric.WithCardinalityLimit(0),
	}
	for _, v := range views {
		mpOpts = append(mpOpts, metric.WithView(v))
	}
	return metric.NewMeterProvider(mpOpts...)
}

func defaultConfig() *config {
	return &config{
		otelExporterOptions: []otelexporter.Option{
			otelexporter.WithoutCounterSuffixes(),
			otelexporter.WithoutUnits(),
		},
	}
}
