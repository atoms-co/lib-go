package metricsx

import "go.opentelemetry.io/otel/sdk/metric"

const (
	otelHTTPScope    = "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	otelGRPCScope    = "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	otelRuntimeScope = "go.opentelemetry.io/contrib/instrumentation/runtime"
)

type otelInstrumentName struct {
	scope string
	name  string
	kind  metric.InstrumentKind
	unit  string
}

// explicitPrometheusMetricNames contains only metrics declared by external
// OTel instrumentation libraries, where this repository cannot change the
// instrument declaration itself. Repository-owned instruments declare their
// final Prometheus names at the source.
//
// These names are the output produced by the pinned Prometheus exporter before
// metricsx began using WithoutCounterSuffixes and WithoutUnits.
var explicitPrometheusMetricNames = map[otelInstrumentName]string{
	// otelhttp v0.63.0.
	{scope: otelHTTPScope, name: "http.server.request.body.size", kind: metric.InstrumentKindHistogram, unit: "By"}:  "http_server_request_body_size_bytes",
	{scope: otelHTTPScope, name: "http.server.response.body.size", kind: metric.InstrumentKindHistogram, unit: "By"}: "http_server_response_body_size_bytes",
	{scope: otelHTTPScope, name: "http.server.request.duration", kind: metric.InstrumentKindHistogram, unit: "s"}:    "http_server_request_duration_seconds",
	{scope: otelHTTPScope, name: "http.client.request.body.size", kind: metric.InstrumentKindHistogram, unit: "By"}:  "http_client_request_body_size_bytes",
	{scope: otelHTTPScope, name: "http.client.request.duration", kind: metric.InstrumentKindHistogram, unit: "s"}:    "http_client_request_duration_seconds",

	// otelgrpc v0.61.0.
	{scope: otelGRPCScope, name: "rpc.server.duration", kind: metric.InstrumentKindHistogram, unit: "ms"}:               "rpc_server_duration_milliseconds",
	{scope: otelGRPCScope, name: "rpc.server.request.size", kind: metric.InstrumentKindHistogram, unit: "By"}:           "rpc_server_request_size_bytes",
	{scope: otelGRPCScope, name: "rpc.server.response.size", kind: metric.InstrumentKindHistogram, unit: "By"}:          "rpc_server_response_size_bytes",
	{scope: otelGRPCScope, name: "rpc.server.requests_per_rpc", kind: metric.InstrumentKindHistogram, unit: "{count}"}:  "rpc_server_requests_per_rpc",
	{scope: otelGRPCScope, name: "rpc.server.responses_per_rpc", kind: metric.InstrumentKindHistogram, unit: "{count}"}: "rpc_server_responses_per_rpc",
	{scope: otelGRPCScope, name: "rpc.client.duration", kind: metric.InstrumentKindHistogram, unit: "ms"}:               "rpc_client_duration_milliseconds",
	{scope: otelGRPCScope, name: "rpc.client.request.size", kind: metric.InstrumentKindHistogram, unit: "By"}:           "rpc_client_request_size_bytes",
	{scope: otelGRPCScope, name: "rpc.client.response.size", kind: metric.InstrumentKindHistogram, unit: "By"}:          "rpc_client_response_size_bytes",
	{scope: otelGRPCScope, name: "rpc.client.requests_per_rpc", kind: metric.InstrumentKindHistogram, unit: "{count}"}:  "rpc_client_requests_per_rpc",
	{scope: otelGRPCScope, name: "rpc.client.responses_per_rpc", kind: metric.InstrumentKindHistogram, unit: "{count}"}: "rpc_client_responses_per_rpc",

	// Runtime instrumentation v0.53.0.
	{scope: otelRuntimeScope, name: "runtime.uptime", kind: metric.InstrumentKindObservableCounter, unit: "ms"}:                             "runtime_uptime_milliseconds_total",
	{scope: otelRuntimeScope, name: "process.runtime.go.goroutines", kind: metric.InstrumentKindObservableUpDownCounter}:                    "process_runtime_go_goroutines",
	{scope: otelRuntimeScope, name: "process.runtime.go.cgo.calls", kind: metric.InstrumentKindObservableUpDownCounter}:                     "process_runtime_go_cgo_calls",
	{scope: otelRuntimeScope, name: "process.runtime.go.mem.heap_alloc", kind: metric.InstrumentKindObservableUpDownCounter, unit: "By"}:    "process_runtime_go_mem_heap_alloc_bytes",
	{scope: otelRuntimeScope, name: "process.runtime.go.mem.heap_idle", kind: metric.InstrumentKindObservableUpDownCounter, unit: "By"}:     "process_runtime_go_mem_heap_idle_bytes",
	{scope: otelRuntimeScope, name: "process.runtime.go.mem.heap_inuse", kind: metric.InstrumentKindObservableUpDownCounter, unit: "By"}:    "process_runtime_go_mem_heap_inuse_bytes",
	{scope: otelRuntimeScope, name: "process.runtime.go.mem.heap_objects", kind: metric.InstrumentKindObservableUpDownCounter}:              "process_runtime_go_mem_heap_objects",
	{scope: otelRuntimeScope, name: "process.runtime.go.mem.heap_released", kind: metric.InstrumentKindObservableUpDownCounter, unit: "By"}: "process_runtime_go_mem_heap_released_bytes",
	{scope: otelRuntimeScope, name: "process.runtime.go.mem.heap_sys", kind: metric.InstrumentKindObservableUpDownCounter, unit: "By"}:      "process_runtime_go_mem_heap_sys_bytes",
	{scope: otelRuntimeScope, name: "process.runtime.go.mem.lookups", kind: metric.InstrumentKindObservableCounter}:                         "process_runtime_go_mem_lookups_total",
	{scope: otelRuntimeScope, name: "process.runtime.go.mem.live_objects", kind: metric.InstrumentKindObservableUpDownCounter}:              "process_runtime_go_mem_live_objects",
	{scope: otelRuntimeScope, name: "process.runtime.go.gc.count", kind: metric.InstrumentKindObservableCounter}:                            "process_runtime_go_gc_count_total",
	{scope: otelRuntimeScope, name: "process.runtime.go.gc.pause_total_ns", kind: metric.InstrumentKindObservableCounter}:                   "process_runtime_go_gc_pause_total_ns_total",
	{scope: otelRuntimeScope, name: "process.runtime.go.gc.pause_ns", kind: metric.InstrumentKindHistogram}:                                 "process_runtime_go_gc_pause_ns",
}

// explicitPrometheusNameViews applies exact names only to external
// auto-instrumentation. It never infers a name from an instrument's unit or
// kind. When a service supplies a custom View, the custom stream keeps all of
// its settings and receives the exact compatibility name.
func explicitPrometheusNameViews(customViews []metric.View) []metric.View {
	views := make([]metric.View, 0, len(customViews)+1)
	for _, customView := range customViews {
		view := customView
		views = append(views, func(instrument metric.Instrument) (metric.Stream, bool) {
			stream, matched := view(instrument)
			if !matched {
				return stream, false
			}
			if name, ok := explicitPrometheusMetricName(instrument); ok {
				stream.Name = name
			}
			return stream, true
		})
	}

	views = append(views, func(instrument metric.Instrument) (metric.Stream, bool) {
		for _, customView := range customViews {
			if _, matched := customView(instrument); matched {
				return metric.Stream{}, false
			}
		}

		name, ok := explicitPrometheusMetricName(instrument)
		if !ok {
			return metric.Stream{}, false
		}
		return metric.Stream{
			Name:        name,
			Description: instrument.Description,
			Unit:        instrument.Unit,
		}, true
	})

	return views
}

func explicitPrometheusMetricName(instrument metric.Instrument) (string, bool) {
	name, ok := explicitPrometheusMetricNames[otelInstrumentName{
		scope: instrument.Scope.Name,
		name:  instrument.Name,
		kind:  instrument.Kind,
		unit:  instrument.Unit,
	}]
	return name, ok
}
