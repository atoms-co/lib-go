package metrics

import (
	"context"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	otelexporter "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func fakeRuntimeSnapshot() runtimeSnapshot {
	return runtimeSnapshot{
		memStats: runtime.MemStats{
			Alloc:         101,
			TotalAlloc:    102,
			Sys:           103,
			Lookups:       104,
			Mallocs:       105,
			Frees:         106,
			HeapAlloc:     107,
			HeapSys:       108,
			HeapIdle:      109,
			HeapInuse:     110,
			HeapReleased:  111,
			HeapObjects:   112,
			StackInuse:    113,
			StackSys:      114,
			MSpanInuse:    115,
			MSpanSys:      116,
			MCacheInuse:   117,
			MCacheSys:     118,
			GCSys:         119,
			OtherSys:      120,
			NumGC:         121,
			NumForcedGC:   122,
			NextGC:        123,
			LastGC:        uint64(124 * time.Millisecond),
			PauseTotalNs:  uint64(125 * time.Millisecond),
			GCCPUFraction: 0.126,
		},
		numGoroutine: 127,
		numCgoCall:   128,
	}
}

func TestOTelRuntimeMetricsPreserveLegacySchemaAndValues(t *testing.T) {
	setTestApp(t, "runtime-test-app")
	meter, reader := testMeter(t)
	readCalls := 0
	registration, err := newOTelRuntimeMetrics(meter, func() runtimeSnapshot {
		readCalls++
		return fakeRuntimeSnapshot()
	})
	if err != nil {
		t.Fatalf("create runtime metrics: %v", err)
	}
	t.Cleanup(func() {
		if err := registration.Unregister(); err != nil {
			t.Errorf("unregister runtime callback: %v", err)
		}
	})

	collected := collectMetrics(t, reader)
	if readCalls != 1 {
		t.Errorf("runtime snapshot reads = %d, want 1 per collection", readCalls)
	}
	if len(collected) != 28 {
		t.Fatalf("collected metric count = %d, want 28: %v", len(collected), collected)
	}

	wantInt64Values := map[Name]int64{
		"process/memory_alloc":               101,
		"process/total_memory_alloc":         102,
		"process/sys_memory_alloc":           103,
		"process/memory_lookups":             104,
		"process/memory_malloc":              105,
		"process/memory_frees":               106,
		"process/heap_alloc":                 107,
		"process/sys_heap":                   108,
		"process/heap_idle":                  109,
		"process/heap_inuse":                 110,
		"process/heap_release":               111,
		"process/heap_objects":               112,
		"process/stack_inuse":                113,
		"process/sys_stack":                  114,
		"process/stack_mspan_inuse":          115,
		"process/sys_stack_mspan":            116,
		"process/stack_mcache_inuse":         117,
		"process/sys_stack_mcache":           118,
		"process/gc_sys":                     119,
		"process/other_sys":                  120,
		"process/num_gc":                     121,
		"process/num_forced_gc":              122,
		"process/next_gc_heap_size":          123,
		"process/last_gc_finished_timestamp": 124,
		"process/pause_total":                125,
		"process/cpu_goroutines":             127,
		"process/cpu_cgo_calls":              128,
	}
	wantAttributes := map[string]string{
		string(AppTagKey):      "runtime-test-app",
		otelSourceAttributeKey: otelSourceAttributeValue,
	}
	cumulativeCount := 0
	for _, definition := range runtimeInt64MetricDefinitions {
		data, ok := collected[definition.name]
		if !ok {
			t.Errorf("runtime metric %q was not collected", definition.name)
			continue
		}
		requireMetricMetadata(t, data, definition.description, string(definition.unit))

		var value int64
		var attributes map[string]string
		if definition.cumulative {
			cumulativeCount++
			sum, point := requireOnlyInt64SumPoint(t, data)
			if !sum.IsMonotonic || sum.Temporality != metricdata.CumulativeTemporality {
				t.Errorf("%q aggregation = monotonic:%t temporality:%v, want monotonic cumulative", definition.name, sum.IsMonotonic, sum.Temporality)
			}
			value = point.Value
			attributes = attributesAsStrings(t, point.Attributes)
		} else {
			gauge, ok := data.Data.(metricdata.Gauge[int64])
			if !ok {
				t.Errorf("aggregation for %q is %T, want metricdata.Gauge[int64]", definition.name, data.Data)
				continue
			}
			if len(gauge.DataPoints) != 1 {
				t.Errorf("data point count for %q = %d, want 1", definition.name, len(gauge.DataPoints))
				continue
			}
			value = gauge.DataPoints[0].Value
			attributes = attributesAsStrings(t, gauge.DataPoints[0].Attributes)
		}
		if value != wantInt64Values[definition.name] {
			t.Errorf("%q value = %d, want %d", definition.name, value, wantInt64Values[definition.name])
		}
		if !reflect.DeepEqual(attributes, wantAttributes) {
			t.Errorf("%q attributes = %v, want exactly %v", definition.name, attributes, wantAttributes)
		}
	}
	if cumulativeCount != 8 {
		t.Errorf("cumulative runtime metric count = %d, want 8", cumulativeCount)
	}

	for _, definition := range runtimeFloat64MetricDefinitions {
		data, ok := collected[definition.name]
		if !ok {
			t.Errorf("runtime metric %q was not collected", definition.name)
			continue
		}
		requireMetricMetadata(t, data, definition.description, string(definition.unit))
		point := requireOnlyFloat64GaugePoint(t, data)
		if point.Value != 0.126 {
			t.Errorf("%q value = %v, want 0.126", definition.name, point.Value)
		}
		if got := attributesAsStrings(t, point.Attributes); !reflect.DeepEqual(got, wantAttributes) {
			t.Errorf("%q attributes = %v, want exactly %v", definition.name, got, wantAttributes)
		}
	}
}

func TestOTelRuntimeMetricsPrometheusNamesAndLabels(t *testing.T) {
	setTestApp(t, "runtime-prometheus-app")
	registry := prometheus.NewRegistry()

	opts := []otelexporter.Option{
		otelexporter.WithRegisterer(registry),
		otelexporter.WithoutCounterSuffixes(),
		otelexporter.WithoutUnits(),
		otelexporter.WithoutScopeInfo(),
		otelexporter.WithoutTargetInfo(),
	}
	otelExporter, err := otelexporter.New(opts...)
	if err != nil {
		t.Fatalf("create OTel Prometheus exporter: %v", err)
	}
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(otelExporter))
	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown OTel meter provider: %v", err)
		}
	})
	registration, err := newOTelRuntimeMetrics(provider.Meter(otelMeterName), fakeRuntimeSnapshot)
	if err != nil {
		t.Fatalf("create OTel runtime metrics: %v", err)
	}
	t.Cleanup(func() {
		if err := registration.Unregister(); err != nil {
			t.Errorf("unregister runtime callback: %v", err)
		}
	})

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather runtime metrics: %v", err)
	}
	wantKinds := make(map[string]string, 28)
	for _, definition := range runtimeInt64MetricDefinitions {
		kind := "GAUGE"
		if definition.cumulative {
			kind = "COUNTER"
		}
		wantKinds[strings.ReplaceAll(definition.name, "/", "_")] = kind
	}
	for _, definition := range runtimeFloat64MetricDefinitions {
		wantKinds[strings.ReplaceAll(definition.name, "/", "_")] = "GAUGE"
	}

	foundFamilies := make(map[string]bool, len(wantKinds))
	for _, family := range families {
		name := family.GetName()
		if name == "otel_scope_info" || name == "target_info" {
			t.Errorf("unexpected OTel exporter metadata family %q", name)
			continue
		}
		wantKind, wanted := wantKinds[name]
		if !wanted {
			if strings.HasPrefix(name, "process_") {
				t.Errorf("unexpected or name-mutated runtime family %q", name)
			}
			continue
		}
		foundFamilies[name] = true
		if gotKind := family.GetType().String(); gotKind != wantKind {
			t.Errorf("%q type = %s, want %s", name, gotKind, wantKind)
		}

		if len(family.GetMetric()) != 1 {
			t.Errorf("%q samples = %d, want 1", name, len(family.GetMetric()))
			continue
		}
		sample := family.GetMetric()[0]
		labels := make(map[string]string, len(sample.GetLabel()))
		for _, label := range sample.GetLabel() {
			labels[label.GetName()] = label.GetValue()
		}
		want := map[string]string{
			string(AppTagKey):      "runtime-prometheus-app",
			otelSourceAttributeKey: otelSourceAttributeValue,
		}
		if !reflect.DeepEqual(labels, want) {
			t.Errorf("%q labels = %v, want exactly %v", name, labels, want)
		}
	}

	if len(foundFamilies) != len(wantKinds) {
		var missing []string
		for name := range wantKinds {
			if !foundFamilies[name] {
				missing = append(missing, name)
			}
		}
		t.Errorf("found %d of %d exact runtime families; missing %v", len(foundFamilies), len(wantKinds), missing)
	}
}
