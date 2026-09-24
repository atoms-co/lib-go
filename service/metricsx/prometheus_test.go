package metricsx

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	otelexporter "go.opentelemetry.io/otel/exporters/prometheus"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func TestDefaultPrometheusNames(t *testing.T) {
	provider, registry := newTestMeterProvider(t)
	ctx := context.Background()

	// Repository-owned instruments declare their final Prometheus names.
	nativeMeter := provider.Meter("native")
	nativeCounter, err := nativeMeter.Int64Counter("native_requests_total")
	if err != nil {
		t.Fatalf("create native counter: %v", err)
	}
	nativeCounter.Add(ctx, 1)
	nativeHistogram, err := nativeMeter.Float64Histogram("native_latency_seconds", otelmetric.WithUnit("s"))
	if err != nil {
		t.Fatalf("create native histogram: %v", err)
	}
	nativeHistogram.Record(ctx, 0.1)

	// External auto-instrumentation is renamed by an exact scoped View.
	autoMeter := provider.Meter(otelHTTPScope)
	autoHistogram, err := autoMeter.Float64Histogram("http.server.request.duration", otelmetric.WithUnit("s"))
	if err != nil {
		t.Fatalf("create auto-instrumented histogram: %v", err)
	}
	autoHistogram.Record(ctx, 0.1)

	// NewOTelCounter callers also provide final, Prometheus-safe base names.
	compatibilityMeter := provider.Meter("css.com/lib-go/pkg/metrics")
	compatibilityCount, err := compatibilityMeter.Int64Counter("css_com_dasher_dashboard_reconciler_total_runs_count")
	if err != nil {
		t.Fatalf("create compatible count: %v", err)
	}
	compatibilityCount.Add(ctx, 1)
	compatibilitySum, err := compatibilityMeter.Int64Counter("css_com_dasher_dashboard_reconciler_total_runs_sum")
	if err != nil {
		t.Fatalf("create compatible sum: %v", err)
	}
	compatibilitySum.Add(ctx, 1)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather Prometheus metrics: %v", err)
	}
	names := make(map[string]bool, len(families))
	nativeScopeLabel := false
	targetSDKLabel := false
	for _, family := range families {
		names[family.GetName()] = true
		for _, sample := range family.GetMetric() {
			for _, label := range sample.GetLabel() {
				if family.GetName() == "native_requests_total" &&
					label.GetName() == "otel_scope_name" && label.GetValue() == "native" {
					nativeScopeLabel = true
				}
				if family.GetName() == "target_info" &&
					label.GetName() == "telemetry_sdk_name" && label.GetValue() == "opentelemetry" {
					targetSDKLabel = true
				}
			}
		}
	}

	for _, name := range []string{
		"native_requests_total",
		"native_latency_seconds",
		"http_server_request_duration_seconds",
		"css_com_dasher_dashboard_reconciler_total_runs_count",
		"css_com_dasher_dashboard_reconciler_total_runs_sum",
		"otel_scope_info",
		"target_info",
	} {
		if !names[name] {
			t.Errorf("expected Prometheus metric %q; got %v", name, names)
		}
	}
	for _, name := range []string{
		"native_requests_total_total",
		"native_latency_seconds_seconds",
		"http_server_request_duration",
		"css_com_dasher_dashboard_reconciler_total_runs_count_total",
		"css_com_dasher_dashboard_reconciler_total_runs_sum_total",
	} {
		if names[name] {
			t.Errorf("unexpected Prometheus metric %q", name)
		}
	}
	if !nativeScopeLabel {
		t.Error("native metric is missing the otel_scope_name label used by OTel canaries")
	}
	if !targetSDKLabel {
		t.Error("target_info is missing telemetry_sdk_name=opentelemetry used by OTel canaries")
	}
}

func TestExplicitPrometheusNameViewsComposeWithCustomViews(t *testing.T) {
	filter := attribute.NewDenyKeysFilter("server.address")
	custom := sdkmetric.NewView(sdkmetric.Instrument{Name: "http.server.request.body.size"}, sdkmetric.Stream{AttributeFilter: filter})
	views := explicitPrometheusNameViews([]sdkmetric.View{custom})
	instrument := sdkmetric.Instrument{
		Name: "http.server.request.body.size",
		Kind: sdkmetric.InstrumentKindHistogram,
		Unit: "By",
	}
	instrument.Scope.Name = otelHTTPScope

	matchedStreams := 0
	for _, view := range views {
		stream, matched := view(instrument)
		if !matched {
			continue
		}
		matchedStreams++
		if got, want := stream.Name, "http_server_request_body_size_bytes"; got != want {
			t.Errorf("custom view stream name = %q, want %q", got, want)
		}
		if stream.AttributeFilter(attribute.String("server.address", "bucket.example.com")) {
			t.Error("custom view no longer filters server.address")
		}
		if !stream.AttributeFilter(attribute.String("http.request.method", "GET")) {
			t.Error("custom view unexpectedly filters http.request.method")
		}
	}
	if matchedStreams != 1 {
		t.Errorf("matching streams = %d, want exactly 1", matchedStreams)
	}
}

func TestExplicitPrometheusNameViewsOverrideCustomName(t *testing.T) {
	custom := sdkmetric.NewView(sdkmetric.Instrument{Name: "http.server.request.body.size"}, sdkmetric.Stream{Name: "custom_http_body_bytes"})
	views := explicitPrometheusNameViews([]sdkmetric.View{custom})
	instrument := sdkmetric.Instrument{
		Name: "http.server.request.body.size",
		Kind: sdkmetric.InstrumentKindHistogram,
		Unit: "By",
	}
	instrument.Scope.Name = otelHTTPScope

	stream, matched := views[0](instrument)
	if !matched {
		t.Fatal("custom view did not match instrument")
	}
	if got, want := stream.Name, "http_server_request_body_size_bytes"; got != want {
		t.Errorf("custom view stream name = %q, want %q", got, want)
	}
	if _, matched := views[1](instrument); matched {
		t.Error("fallback exact-name view must not duplicate a custom stream")
	}
}

func TestCustomViewPrometheusIntegration(t *testing.T) {
	filter := attribute.NewDenyKeysFilter("server.address", "server.port")
	custom := sdkmetric.NewView(sdkmetric.Instrument{Name: "http.server.request.body.size"}, sdkmetric.Stream{AttributeFilter: filter})
	provider, registry := newTestMeterProvider(t, custom)

	meter := provider.Meter(otelHTTPScope)
	bodySize, err := meter.Int64Histogram("http.server.request.body.size", otelmetric.WithUnit("By"))
	if err != nil {
		t.Fatalf("create body-size histogram: %v", err)
	}
	attrs := otelmetric.WithAttributes(attribute.String("server.address", "bucket.example.com"), attribute.Int("server.port", 443), attribute.String("http.request.method", "GET"))
	bodySize.Record(context.Background(), 123, attrs)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather Prometheus metrics: %v", err)
	}
	found := false
	for _, family := range families {
		if family.GetName() == "http_server_request_body_size" {
			t.Error("found unrenamed body-size metric family")
		}
		if family.GetName() != "http_server_request_body_size_bytes" {
			continue
		}
		if found {
			t.Fatal("found duplicate renamed body-size metric family")
		}
		found = true
		keptMethod := false
		for _, sample := range family.GetMetric() {
			for _, label := range sample.GetLabel() {
				switch label.GetName() {
				case "server_address", "server_port":
					t.Errorf("filtered label %q remains on body-size metric", label.GetName())
				case "http_request_method":
					keptMethod = true
				}
			}
		}
		if !keptMethod {
			t.Error("body-size metric lost retained http_request_method label")
		}
	}
	if !found {
		t.Error("renamed body-size metric family not found")
	}
}

func TestExplicitPrometheusMetricNames(t *testing.T) {
	tests := []struct {
		source string
		scope  string
		name   string
		unit   string
		kind   sdkmetric.InstrumentKind
		want   string
	}{
		// otelhttp v0.63.0.
		{source: "otelhttp", scope: otelHTTPScope, name: "http.server.request.body.size", unit: "By", kind: sdkmetric.InstrumentKindHistogram, want: "http_server_request_body_size_bytes"},
		{source: "otelhttp", scope: otelHTTPScope, name: "http.server.response.body.size", unit: "By", kind: sdkmetric.InstrumentKindHistogram, want: "http_server_response_body_size_bytes"},
		{source: "otelhttp", scope: otelHTTPScope, name: "http.server.request.duration", unit: "s", kind: sdkmetric.InstrumentKindHistogram, want: "http_server_request_duration_seconds"},
		{source: "otelhttp", scope: otelHTTPScope, name: "http.client.request.body.size", unit: "By", kind: sdkmetric.InstrumentKindHistogram, want: "http_client_request_body_size_bytes"},
		{source: "otelhttp", scope: otelHTTPScope, name: "http.client.request.duration", unit: "s", kind: sdkmetric.InstrumentKindHistogram, want: "http_client_request_duration_seconds"},

		// otelgrpc v0.61.0.
		{source: "otelgrpc", scope: otelGRPCScope, name: "rpc.server.duration", unit: "ms", kind: sdkmetric.InstrumentKindHistogram, want: "rpc_server_duration_milliseconds"},
		{source: "otelgrpc", scope: otelGRPCScope, name: "rpc.server.request.size", unit: "By", kind: sdkmetric.InstrumentKindHistogram, want: "rpc_server_request_size_bytes"},
		{source: "otelgrpc", scope: otelGRPCScope, name: "rpc.server.response.size", unit: "By", kind: sdkmetric.InstrumentKindHistogram, want: "rpc_server_response_size_bytes"},
		{source: "otelgrpc", scope: otelGRPCScope, name: "rpc.server.requests_per_rpc", unit: "{count}", kind: sdkmetric.InstrumentKindHistogram, want: "rpc_server_requests_per_rpc"},
		{source: "otelgrpc", scope: otelGRPCScope, name: "rpc.server.responses_per_rpc", unit: "{count}", kind: sdkmetric.InstrumentKindHistogram, want: "rpc_server_responses_per_rpc"},
		{source: "otelgrpc", scope: otelGRPCScope, name: "rpc.client.duration", unit: "ms", kind: sdkmetric.InstrumentKindHistogram, want: "rpc_client_duration_milliseconds"},
		{source: "otelgrpc", scope: otelGRPCScope, name: "rpc.client.request.size", unit: "By", kind: sdkmetric.InstrumentKindHistogram, want: "rpc_client_request_size_bytes"},
		{source: "otelgrpc", scope: otelGRPCScope, name: "rpc.client.response.size", unit: "By", kind: sdkmetric.InstrumentKindHistogram, want: "rpc_client_response_size_bytes"},
		{source: "otelgrpc", scope: otelGRPCScope, name: "rpc.client.requests_per_rpc", unit: "{count}", kind: sdkmetric.InstrumentKindHistogram, want: "rpc_client_requests_per_rpc"},
		{source: "otelgrpc", scope: otelGRPCScope, name: "rpc.client.responses_per_rpc", unit: "{count}", kind: sdkmetric.InstrumentKindHistogram, want: "rpc_client_responses_per_rpc"},

		// Runtime instrumentation v0.53.0, enabled by patterns-service. Hicar
		// uses its separate OTLP/Collector pipeline and is intentionally absent.
		{source: "runtime", scope: otelRuntimeScope, name: "runtime.uptime", unit: "ms", kind: sdkmetric.InstrumentKindObservableCounter, want: "runtime_uptime_milliseconds_total"},
		{source: "runtime", scope: otelRuntimeScope, name: "process.runtime.go.goroutines", kind: sdkmetric.InstrumentKindObservableUpDownCounter, want: "process_runtime_go_goroutines"},
		{source: "runtime", scope: otelRuntimeScope, name: "process.runtime.go.cgo.calls", kind: sdkmetric.InstrumentKindObservableUpDownCounter, want: "process_runtime_go_cgo_calls"},
		{source: "runtime", scope: otelRuntimeScope, name: "process.runtime.go.mem.heap_alloc", unit: "By", kind: sdkmetric.InstrumentKindObservableUpDownCounter, want: "process_runtime_go_mem_heap_alloc_bytes"},
		{source: "runtime", scope: otelRuntimeScope, name: "process.runtime.go.mem.heap_idle", unit: "By", kind: sdkmetric.InstrumentKindObservableUpDownCounter, want: "process_runtime_go_mem_heap_idle_bytes"},
		{source: "runtime", scope: otelRuntimeScope, name: "process.runtime.go.mem.heap_inuse", unit: "By", kind: sdkmetric.InstrumentKindObservableUpDownCounter, want: "process_runtime_go_mem_heap_inuse_bytes"},
		{source: "runtime", scope: otelRuntimeScope, name: "process.runtime.go.mem.heap_objects", kind: sdkmetric.InstrumentKindObservableUpDownCounter, want: "process_runtime_go_mem_heap_objects"},
		{source: "runtime", scope: otelRuntimeScope, name: "process.runtime.go.mem.heap_released", unit: "By", kind: sdkmetric.InstrumentKindObservableUpDownCounter, want: "process_runtime_go_mem_heap_released_bytes"},
		{source: "runtime", scope: otelRuntimeScope, name: "process.runtime.go.mem.heap_sys", unit: "By", kind: sdkmetric.InstrumentKindObservableUpDownCounter, want: "process_runtime_go_mem_heap_sys_bytes"},
		{source: "runtime", scope: otelRuntimeScope, name: "process.runtime.go.mem.lookups", kind: sdkmetric.InstrumentKindObservableCounter, want: "process_runtime_go_mem_lookups_total"},
		{source: "runtime", scope: otelRuntimeScope, name: "process.runtime.go.mem.live_objects", kind: sdkmetric.InstrumentKindObservableUpDownCounter, want: "process_runtime_go_mem_live_objects"},
		{source: "runtime", scope: otelRuntimeScope, name: "process.runtime.go.gc.count", kind: sdkmetric.InstrumentKindObservableCounter, want: "process_runtime_go_gc_count_total"},
		{source: "runtime", scope: otelRuntimeScope, name: "process.runtime.go.gc.pause_total_ns", kind: sdkmetric.InstrumentKindObservableCounter, want: "process_runtime_go_gc_pause_total_ns_total"},
		{source: "runtime", scope: otelRuntimeScope, name: "process.runtime.go.gc.pause_ns", kind: sdkmetric.InstrumentKindHistogram, want: "process_runtime_go_gc_pause_ns"},
	}

	if got, want := len(explicitPrometheusMetricNames), len(tests); got != want {
		t.Fatalf("explicit metric mappings = %d, want %d", got, want)
	}
	for _, test := range tests {
		t.Run(test.source+"/"+test.name, func(t *testing.T) {
			instrument := sdkmetric.Instrument{Name: test.name, Unit: test.unit, Kind: test.kind}
			instrument.Scope.Name = test.scope
			got, ok := explicitPrometheusMetricName(instrument)
			if !ok {
				t.Fatal("exact metric mapping not found")
			}
			if got != test.want {
				t.Errorf("exact metric name = %q, want %q", got, test.want)
			}
		})
	}
}

func TestExplicitPrometheusMetricNameRequiresExactInstrument(t *testing.T) {
	base := sdkmetric.Instrument{
		Name: "http.server.request.body.size",
		Kind: sdkmetric.InstrumentKindHistogram,
		Unit: "By",
	}
	base.Scope.Name = otelHTTPScope

	tests := []struct {
		name       string
		instrument sdkmetric.Instrument
	}{
		{name: "unknown scope", instrument: withInstrumentScope(base, "example.com/custom")},
		{name: "unknown name", instrument: withInstrumentName(base, "http.server.unknown")},
		{name: "changed kind", instrument: withInstrumentKind(base, sdkmetric.InstrumentKindCounter)},
		{name: "changed unit", instrument: withInstrumentUnit(base, "s")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if name, ok := explicitPrometheusMetricName(test.instrument); ok {
				t.Errorf("unexpected exact mapping %q", name)
			}
		})
	}
}

func newTestMeterProvider(t *testing.T, customViews ...sdkmetric.View) (*sdkmetric.MeterProvider, *prometheus.Registry) {
	t.Helper()
	registry := prometheus.NewRegistry()
	cfg := defaultConfig()
	cfg.otelExporterOptions = append(cfg.otelExporterOptions, otelexporter.WithRegisterer(registry))

	exporter, err := otelexporter.New(cfg.otelExporterOptions...)
	if err != nil {
		t.Fatalf("create Prometheus exporter: %v", err)
	}
	providerOptions := []sdkmetric.Option{sdkmetric.WithReader(exporter)}
	for _, view := range explicitPrometheusNameViews(customViews) {
		providerOptions = append(providerOptions, sdkmetric.WithView(view))
	}
	provider := sdkmetric.NewMeterProvider(providerOptions...)
	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shut down meter provider: %v", err)
		}
	})
	return provider, registry
}

func withInstrumentScope(instrument sdkmetric.Instrument, scope string) sdkmetric.Instrument {
	instrument.Scope.Name = scope
	return instrument
}

func withInstrumentName(instrument sdkmetric.Instrument, name string) sdkmetric.Instrument {
	instrument.Name = name
	return instrument
}

func withInstrumentKind(instrument sdkmetric.Instrument, kind sdkmetric.InstrumentKind) sdkmetric.Instrument {
	instrument.Kind = kind
	return instrument
}

func withInstrumentUnit(instrument sdkmetric.Instrument, unit string) sdkmetric.Instrument {
	instrument.Unit = unit
	return instrument
}
