package metricsx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelexporter "go.opentelemetry.io/otel/exporters/prometheus"
	otelmetric "go.opentelemetry.io/otel/metric"

	"go.atoms.co/lib/metrics"
)

func TestPrometheusEndpointExportsOnlyOpenTelemetryMetrics(t *testing.T) {
	counter := metrics.NewCounter(
		"metricsx_direct_cutover",
		"Direct OTel cutover test",
	)

	registry := prometheus.NewRegistry()
	exporter, err := otelexporter.New(otelexporter.WithRegisterer(registry), otelexporter.WithoutCounterSuffixes(), otelexporter.WithoutUnits())
	if err != nil {
		t.Fatalf("create OTel Prometheus exporter: %v", err)
	}
	provider := newMeterProvider(exporter, nil)
	previousProvider := otel.GetMeterProvider()
	otel.SetMeterProvider(provider)
	t.Cleanup(func() {
		otel.SetMeterProvider(previousProvider)
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shut down meter provider: %v", err)
		}
	})

	counter.Increment(context.Background(), 3)
	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	response := httptest.NewRecorder()
	prometheusHandler(registry).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, name := range []string{
		"metricsx_direct_cutover_count",
		"metricsx_direct_cutover_sum",
	} {
		if !strings.Contains(body, name) {
			t.Errorf("metrics endpoint does not contain %q:\n%s", name, body)
		}
	}
	if strings.Contains(body, "metricsx_direct_cutover_count_total") {
		t.Errorf("metrics endpoint added an unexpected counter suffix:\n%s", body)
	}
}

func TestMeterProviderPreservesUnlimitedCardinality(t *testing.T) {
	const cardinality = 2101

	registry := prometheus.NewRegistry()
	exporter, err := otelexporter.New(otelexporter.WithRegisterer(registry))
	if err != nil {
		t.Fatalf("create OTel Prometheus exporter: %v", err)
	}
	provider := newMeterProvider(exporter, nil)
	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shut down meter provider: %v", err)
		}
	})

	counter, err := provider.Meter("metricsx.cardinality.test").Int64Counter(
		"metricsx_unlimited_cardinality",
	)
	if err != nil {
		t.Fatalf("create OTel counter: %v", err)
	}
	for i := range cardinality {
		counter.Add(context.Background(), 1, otelmetric.WithAttributes(attribute.String("id", strconv.Itoa(i))))
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}
	for _, family := range families {
		if family.GetName() != "metricsx_unlimited_cardinality_total" {
			continue
		}
		if got := len(family.GetMetric()); got != cardinality {
			t.Fatalf("metric cardinality = %d, want %d", got, cardinality)
		}
		for _, sample := range family.GetMetric() {
			for _, label := range sample.GetLabel() {
				if label.GetName() == "otel_metric_overflow" || label.GetName() == "otel.metric.overflow" {
					t.Fatalf("unexpected cardinality overflow label: %s=%s", label.GetName(), label.GetValue())
				}
			}
		}
		return
	}
	t.Fatal("unlimited-cardinality metric was not gathered")
}
