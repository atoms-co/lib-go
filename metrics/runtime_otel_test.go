package metrics

import (
	"context"
	"math"
	"reflect"
	"runtime"
	rtmetrics "runtime/metrics"
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
		heapAlloc:     101,
		totalAlloc:    102,
		sys:           103,
		mallocs:       105,
		frees:         106,
		heapSys:       108,
		heapIdle:      109,
		heapInuse:     110,
		heapReleased:  111,
		heapObjects:   112,
		stackInuse:    113,
		stackSys:      114,
		mSpanInuse:    115,
		mSpanSys:      116,
		mCacheInuse:   117,
		mCacheSys:     118,
		gcSys:         119,
		otherSys:      120,
		numGC:         121,
		numForcedGC:   122,
		nextGC:        123,
		pauseTotal:    125 * time.Millisecond,
		gcCPUFraction: 0.126,
		numGoroutine:  127,
		numCgoCall:    128,
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
	if len(collected) != 26 {
		t.Fatalf("collected metric count = %d, want 26: %v", len(collected), collected)
	}

	wantInt64Values := map[Name]int64{
		"process/memory_alloc":       101,
		"process/total_memory_alloc": 102,
		"process/sys_memory_alloc":   103,
		"process/memory_malloc":      105,
		"process/memory_frees":       106,
		"process/heap_alloc":         101,
		"process/sys_heap":           108,
		"process/heap_idle":          109,
		"process/heap_inuse":         110,
		"process/heap_release":       111,
		"process/heap_objects":       112,
		"process/stack_inuse":        113,
		"process/sys_stack":          114,
		"process/stack_mspan_inuse":  115,
		"process/sys_stack_mspan":    116,
		"process/stack_mcache_inuse": 117,
		"process/sys_stack_mcache":   118,
		"process/gc_sys":             119,
		"process/other_sys":          120,
		"process/num_gc":             121,
		"process/num_forced_gc":      122,
		"process/next_gc_heap_size":  123,
		"process/pause_total":        125,
		"process/cpu_goroutines":     127,
		"process/cpu_cgo_calls":      128,
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
	if cumulativeCount != 7 {
		t.Errorf("cumulative runtime metric count = %d, want 7", cumulativeCount)
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
	wantKinds := make(map[string]string, 26)
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

func TestRuntimeMetricNamesAreSupported(t *testing.T) {
	supported := make(map[string]rtmetrics.ValueKind)
	for _, description := range rtmetrics.All() {
		supported[description.Name] = description.Kind
	}
	for _, name := range runtimeMetricNames {
		if _, ok := supported[name]; !ok {
			t.Errorf("runtime/metrics does not support %q; its value would silently read as 0", name)
		}
	}
}

func TestHistogramTotalDuration(t *testing.T) {
	if got := histogramTotalDuration(nil); got != 0 {
		t.Errorf("histogramTotalDuration(nil) = %v, want 0", got)
	}
	histogram := &rtmetrics.Float64Histogram{
		Buckets: []float64{math.Inf(-1), 0, 0.001, 0.003, math.Inf(1)},
		Counts:  []uint64{5, 2, 3, 1},
	}
	// -Inf bucket counts as 0, [0,1ms) as 0.5ms, [1ms,3ms) as 2ms, [3ms,+Inf) as 3ms.
	want := 2*500*time.Microsecond + 3*2*time.Millisecond + 3*time.Millisecond
	if got := histogramTotalDuration(histogram); got != want {
		t.Errorf("histogramTotalDuration = %v, want %v", got, want)
	}
}

// TestReadRuntimeSnapshotMatchesMemStats checks the runtime/metrics derivation
// against runtime.ReadMemStats, which it replaces.
func TestReadRuntimeSnapshotMatchesMemStats(t *testing.T) {
	runtime.GC()
	runtime.GC()

	snapshot := readRuntimeSnapshot()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// GC counters cannot move between the two reads without another GC.
	if snapshot.numGC != uint64(memStats.NumGC) {
		t.Errorf("numGC = %d, MemStats.NumGC = %d", snapshot.numGC, memStats.NumGC)
	}
	if snapshot.numForcedGC != uint64(memStats.NumForcedGC) {
		t.Errorf("numForcedGC = %d, MemStats.NumForcedGC = %d", snapshot.numForcedGC, memStats.NumForcedGC)
	}
	if snapshot.nextGC != memStats.NextGC {
		t.Errorf("nextGC = %d, MemStats.NextGC = %d", snapshot.nextGC, memStats.NextGC)
	}
	if diff := math.Abs(snapshot.gcCPUFraction - memStats.GCCPUFraction); diff > 1e-9 {
		t.Errorf("gcCPUFraction = %v, MemStats.GCCPUFraction = %v", snapshot.gcCPUFraction, memStats.GCCPUFraction)
	}
	wantPause := time.Duration(memStats.PauseTotalNs)
	if diff := snapshot.pauseTotal - wantPause; diff < -wantPause/4 || diff > wantPause/4 {
		t.Errorf("pauseTotal = %v, MemStats.PauseTotalNs = %v; want within 25%%", snapshot.pauseTotal, wantPause)
	}

	// Memory values can drift slightly between the two reads. heapReleased is
	// not compared: the background scavenger can release megabytes between the
	// reads. It is a direct 1:1 read of /memory/classes/heap/released:bytes, and
	// heapIdle (free + released) still covers it.
	approx := []struct {
		name string
		got  uint64
		want uint64
	}{
		{"totalAlloc", snapshot.totalAlloc, memStats.TotalAlloc},
		{"sys", snapshot.sys, memStats.Sys},
		{"mallocs", snapshot.mallocs, memStats.Mallocs},
		{"frees", snapshot.frees, memStats.Frees},
		{"heapAlloc", snapshot.heapAlloc, memStats.HeapAlloc},
		{"heapSys", snapshot.heapSys, memStats.HeapSys},
		{"heapIdle", snapshot.heapIdle, memStats.HeapIdle},
		{"heapInuse", snapshot.heapInuse, memStats.HeapInuse},
		{"heapObjects", snapshot.heapObjects, memStats.HeapObjects},
		{"stackInuse", snapshot.stackInuse, memStats.StackInuse},
		{"stackSys", snapshot.stackSys, memStats.StackSys},
		{"mSpanInuse", snapshot.mSpanInuse, memStats.MSpanInuse},
		{"mSpanSys", snapshot.mSpanSys, memStats.MSpanSys},
		{"mCacheInuse", snapshot.mCacheInuse, memStats.MCacheInuse},
		{"mCacheSys", snapshot.mCacheSys, memStats.MCacheSys},
		{"gcSys", snapshot.gcSys, memStats.GCSys},
		{"otherSys", snapshot.otherSys, memStats.OtherSys},
	}
	for _, value := range approx {
		if value.want == 0 {
			continue
		}
		ratio := float64(value.got) / float64(value.want)
		if ratio < 0.8 || ratio > 1.2 {
			t.Errorf("%s = %d, MemStats = %d; want within 20%%", value.name, value.got, value.want)
		}
	}
}
