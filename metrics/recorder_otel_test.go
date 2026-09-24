package metrics

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

const (
	methodKey Key = "method"
	resultKey Key = "result"
	statusKey Key = "status"
	typeKey   Key = "type"
)

func testMeter(t *testing.T) (metric.Meter, *sdkmetric.ManualReader) {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown meter provider: %v", err)
		}
	})
	return provider.Meter(otelMeterName), reader
}

func collectMetrics(t *testing.T, reader *sdkmetric.ManualReader) map[string]metricdata.Metrics {
	t.Helper()
	var resourceMetrics metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &resourceMetrics); err != nil {
		t.Fatalf("collect metrics: %v", err)
	}

	metrics := make(map[string]metricdata.Metrics)
	for _, scopeMetrics := range resourceMetrics.ScopeMetrics {
		for _, data := range scopeMetrics.Metrics {
			metrics[data.Name] = data
		}
	}
	return metrics
}

func setTestApp(t *testing.T, app string) {
	t.Helper()
	previous := defaultTag.Value
	initAppName(app)
	t.Cleanup(func() { initAppName(previous) })
}

func attributesAsStrings(t *testing.T, set attribute.Set) map[string]string {
	t.Helper()
	attrs := make(map[string]string, set.Len())
	for _, attr := range set.ToSlice() {
		if attr.Value.Type() != attribute.STRING {
			t.Fatalf("attribute %q is not a string", attr.Key)
		}
		attrs[string(attr.Key)] = attr.Value.AsString()
	}
	return attrs
}

func requireMetricMetadata(t *testing.T, data metricdata.Metrics, description, unit string) {
	t.Helper()
	if data.Description != description {
		t.Errorf("description = %q, want %q", data.Description, description)
	}
	if data.Unit != unit {
		t.Errorf("unit = %q, want %q", data.Unit, unit)
	}
}

func requireOnlyInt64SumPoint(t *testing.T, data metricdata.Metrics) (metricdata.Sum[int64], metricdata.DataPoint[int64]) {
	t.Helper()
	sum, ok := data.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatalf("aggregation for %q is %T, want metricdata.Sum[int64]", data.Name, data.Data)
	}
	if len(sum.DataPoints) != 1 {
		t.Fatalf("data point count for %q = %d, want 1", data.Name, len(sum.DataPoints))
	}
	return sum, sum.DataPoints[0]
}

func requireOnlyFloat64GaugePoint(t *testing.T, data metricdata.Metrics) metricdata.DataPoint[float64] {
	t.Helper()
	gauge, ok := data.Data.(metricdata.Gauge[float64])
	if !ok {
		t.Fatalf("aggregation for %q is %T, want metricdata.Gauge[float64]", data.Name, data.Data)
	}
	if len(gauge.DataPoints) != 1 {
		t.Fatalf("data point count for %q = %d, want 1", data.Name, len(gauge.DataPoints))
	}
	return gauge.DataPoints[0]
}

func TestDefaultConstructorsUseOpenTelemetry(t *testing.T) {
	setTestApp(t, "default-constructor-app")
	meter, reader := testMeter(t)
	previousMeter := otelMeter
	otelMeter = meter
	t.Cleanup(func() { otelMeter = previousMeter })

	ctx := context.Background()
	NewCounter("default_counter", "Default counter", resultKey).Increment(ctx, 2, Tag{Key: resultKey, Value: "ok"})
	NewGauge("default_gauge", "Default gauge").Set(ctx, 3)
	NewHistogram("default_duration", "Default duration", &BucketOptions{
		DistributionType:   UserDefined,
		UserDefinedBuckets: []float64{1, 2},
	}).Observe(ctx, 1500*time.Millisecond)
	NewDimensionlessHistogram("default_dimensionless", "Default dimensionless histogram", &BucketOptions{
		DistributionType:   UserDefined,
		UserDefinedBuckets: []float64{1, 2},
	}).Observe(ctx, 1.5)
	NewByteHistogram("default_bytes", "Default byte histogram", &BucketOptions{
		DistributionType:   UserDefined,
		UserDefinedBuckets: []float64{1, 2},
	}).Observe(ctx, 2)
	NewSingleViewCounter("default_single", "Default single-view counter").Increment(ctx, 4)

	collected := collectMetrics(t, reader)
	for _, name := range []string{
		"default_counter_count",
		"default_counter_sum",
		"default_gauge",
		"default_duration",
		"default_dimensionless",
		"default_bytes",
		"default_single",
	} {
		if _, ok := collected[name]; !ok {
			t.Errorf("default constructor metric %q was not collected through OTel; got %v", name, collected)
		}
	}
}

func TestOTelCounterPreservesLegacyCountAndSum(t *testing.T) {
	setTestApp(t, "counter-test-app")
	meter, reader := testMeter(t)
	counter := newOTelCounter(meter, "legacy_requests", "Legacy request count", []Key{resultKey})

	tags := []Tag{
		{Key: resultKey, Value: "success"},
		{Key: "route", Value: "/v1/orders"}, // Undeclared in the legacy OC view; filtered for cardinality parity.
		{Key: AppTagKey, Value: "caller-cannot-replace-app"},
	}
	counter.Increment(context.Background(), 3, tags...)
	counter.Increment(context.Background(), 5, tags...)

	metrics := collectMetrics(t, reader)
	if len(metrics) != 2 {
		t.Fatalf("metric count = %d, want 2: %v", len(metrics), metrics)
	}

	countData, ok := metrics["legacy_requests_count"]
	if !ok {
		t.Fatal("legacy_requests_count was not collected")
	}
	requireMetricMetadata(t, countData, "Legacy request count", string(UnitDimensionless))
	count, countPoint := requireOnlyInt64SumPoint(t, countData)
	if !count.IsMonotonic || count.Temporality != metricdata.CumulativeTemporality {
		t.Errorf("count aggregation = monotonic:%t temporality:%v, want monotonic cumulative", count.IsMonotonic, count.Temporality)
	}
	if countPoint.Value != 2 {
		t.Errorf("legacy_requests_count = %d, want 2", countPoint.Value)
	}
	wantAttrs := map[string]string{
		string(AppTagKey):      "counter-test-app",
		string(resultKey):      "success",
		otelSourceAttributeKey: otelSourceAttributeValue,
	}
	if got := attributesAsStrings(t, countPoint.Attributes); !reflect.DeepEqual(got, wantAttrs) {
		t.Errorf("count attributes = %v, want %v", got, wantAttrs)
	}

	sumData, ok := metrics["legacy_requests_sum"]
	if !ok {
		t.Fatal("legacy_requests_sum was not collected")
	}
	requireMetricMetadata(t, sumData, "Legacy request count", string(UnitDimensionless))
	sum, sumPoint := requireOnlyInt64SumPoint(t, sumData)
	if !sum.IsMonotonic || sum.Temporality != metricdata.CumulativeTemporality {
		t.Errorf("sum aggregation = monotonic:%t temporality:%v, want monotonic cumulative", sum.IsMonotonic, sum.Temporality)
	}
	if sumPoint.Value != 8 {
		t.Errorf("legacy_requests_sum = %d, want 8", sumPoint.Value)
	}
}

func TestOTelSingleViewCounterUsesExactNameAndDelta(t *testing.T) {
	setTestApp(t, "single-counter-app")
	meter, reader := testMeter(t)
	counter := newOTelSingleViewCounter(meter, "legacy_exact_counter", "Exact counter", []Key{methodKey})
	counter.Increment(context.Background(), 7, Tag{Key: methodKey, Value: "Create"})

	metrics := collectMetrics(t, reader)
	if len(metrics) != 1 {
		t.Fatalf("metric count = %d, want 1: %v", len(metrics), metrics)
	}
	data, ok := metrics["legacy_exact_counter"]
	if !ok {
		t.Fatal("exact legacy counter name was not collected")
	}
	requireMetricMetadata(t, data, "Exact counter", string(UnitDimensionless))
	_, point := requireOnlyInt64SumPoint(t, data)
	if point.Value != 7 {
		t.Errorf("legacy_exact_counter = %d, want 7", point.Value)
	}
	if got := attributesAsStrings(t, point.Attributes); got[string(methodKey)] != "Create" || got[string(AppTagKey)] != "single-counter-app" || got[otelSourceAttributeKey] != otelSourceAttributeValue {
		t.Errorf("counter attributes = %v, want method, app, and OTel source", got)
	}
}

func TestOTelGaugeUsesLastValueAndFixedLabelSchema(t *testing.T) {
	setTestApp(t, "gauge-test-app")
	meter, reader := testMeter(t)
	gauge := newOTelGauge(meter, "legacy_queue_depth", "Current queue depth", []Key{statusKey})

	// Status is deliberately omitted. OpenCensus registered tag columns are
	// fixed, so the OTel primitive retains the label with an empty value. Queue
	// is undeclared and therefore filtered to preserve OC cardinality.
	gauge.Set(context.Background(), 4, Tag{Key: "queue", Value: "primary"})
	gauge.Set(context.Background(), 9, Tag{Key: "queue", Value: "primary"})

	data := collectMetrics(t, reader)["legacy_queue_depth"]
	requireMetricMetadata(t, data, "Current queue depth", string(UnitDimensionless))
	point := requireOnlyFloat64GaugePoint(t, data)
	if point.Value != 9 {
		t.Errorf("legacy_queue_depth = %v, want last value 9", point.Value)
	}
	wantAttrs := map[string]string{
		string(AppTagKey):      "gauge-test-app",
		string(statusKey):      "",
		otelSourceAttributeKey: otelSourceAttributeValue,
	}
	if got := attributesAsStrings(t, point.Attributes); !reflect.DeepEqual(got, wantAttrs) {
		t.Errorf("gauge attributes = %v, want %v", got, wantAttrs)
	}
}

func TestOTelHistogramsPreserveKindsUnitsBoundariesAndValues(t *testing.T) {
	setTestApp(t, "histogram-test-app")
	meter, reader := testMeter(t)

	durationOptions := &BucketOptions{
		DistributionType:   UserDefined,
		UserDefinedBuckets: []float64{100, 500},
		LatencyUnit:        time.Millisecond,
	}
	duration := newOTelDurationHistogram(meter, "legacy_latency", "Legacy latency", durationOptions, []Key{methodKey})
	duration.Observe(context.Background(), 250*time.Millisecond, Tag{Key: methodKey, Value: "GET"})

	dimensionlessOptions := &BucketOptions{
		DistributionType:   UserDefined,
		UserDefinedBuckets: []float64{1, 3},
	}
	dimensionless := newOTelFloatHistogram(meter, "legacy_batch_size", "Legacy batch size", UnitDimensionless, dimensionlessOptions, nil)
	dimensionless.Observe(context.Background(), 2)

	byteOptions := &BucketOptions{
		DistributionType:   UserDefined,
		UserDefinedBuckets: []float64{10, 20},
	}
	bytes := newOTelFloatHistogram(meter, "legacy_payload_size", "Legacy payload size", UnitBytes, byteOptions, nil)
	bytes.Observe(context.Background(), 25)

	metrics := collectMetrics(t, reader)
	tests := []struct {
		name        string
		description string
		unit        UnitType
		bounds      []float64
		wantSum     float64
		wantBuckets []uint64
	}{
		{
			name:        "legacy_latency",
			description: "Legacy latency",
			unit:        UnitMilliseconds,
			bounds:      []float64{100, 500},
			wantSum:     250,
			wantBuckets: []uint64{0, 1, 0},
		},
		{
			name:        "legacy_batch_size",
			description: "Legacy batch size",
			unit:        UnitDimensionless,
			bounds:      []float64{1, 3},
			wantSum:     2,
			wantBuckets: []uint64{0, 1, 0},
		},
		{
			name:        "legacy_payload_size",
			description: "Legacy payload size",
			unit:        UnitBytes,
			bounds:      []float64{10, 20},
			wantSum:     25,
			wantBuckets: []uint64{0, 0, 1},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, ok := metrics[test.name]
			if !ok {
				t.Fatalf("%q was not collected", test.name)
			}
			requireMetricMetadata(t, data, test.description, string(test.unit))
			histogram, ok := data.Data.(metricdata.Histogram[float64])
			if !ok {
				t.Fatalf("aggregation for %q is %T, want metricdata.Histogram[float64]", test.name, data.Data)
			}
			if len(histogram.DataPoints) != 1 {
				t.Fatalf("data point count = %d, want 1", len(histogram.DataPoints))
			}
			point := histogram.DataPoints[0]
			if point.Count != 1 || point.Sum != test.wantSum {
				t.Errorf("count/sum = %d/%v, want 1/%v", point.Count, point.Sum, test.wantSum)
			}
			if !reflect.DeepEqual(point.Bounds, test.bounds) {
				t.Errorf("bounds = %v, want %v", point.Bounds, test.bounds)
			}
			if !reflect.DeepEqual(point.BucketCounts, test.wantBuckets) {
				t.Errorf("bucket counts = %v, want %v", point.BucketCounts, test.wantBuckets)
			}
			attrs := attributesAsStrings(t, point.Attributes)
			if attrs[string(AppTagKey)] != "histogram-test-app" {
				t.Errorf("histogram app attribute = %q, want histogram-test-app", attrs[string(AppTagKey)])
			}
			if attrs[otelSourceAttributeKey] != otelSourceAttributeValue {
				t.Errorf("histogram source attribute = %q, want %q", attrs[otelSourceAttributeKey], otelSourceAttributeValue)
			}
		})
	}
}

func TestOTelHistogramCanonicalizesBoundsLikeOpenCensus(t *testing.T) {
	setTestApp(t, "canonical-bounds-app")
	meter, reader := testMeter(t)
	histogram := newOTelFloatHistogram(meter, "legacy_canonical_bounds", "Canonical bounds", UnitDimensionless, &BucketOptions{
		DistributionType:   UserDefined,
		UserDefinedBuckets: []float64{3, 0, 1},
	}, nil)
	histogram.Observe(context.Background(), 2)

	data := collectMetrics(t, reader)["legacy_canonical_bounds"]
	aggregation, ok := data.Data.(metricdata.Histogram[float64])
	if !ok || len(aggregation.DataPoints) != 1 {
		t.Fatalf("aggregation = %T with %d points, want one float64 histogram point", data.Data, len(aggregation.DataPoints))
	}
	if got, want := aggregation.DataPoints[0].Bounds, []float64{1, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("canonical bounds = %v, want %v", got, want)
	}
}

func TestOTelMillisecondHistogramPreservesFractionalObservation(t *testing.T) {
	setTestApp(t, "fractional-histogram-app")
	meter, reader := testMeter(t)
	histogram := newOTelFloatHistogram(meter, "legacy_fractional_milliseconds", "Fractional millisecond latency", UnitMilliseconds, &BucketOptions{
		DistributionType:   UserDefined,
		UserDefinedBuckets: []float64{0.1, 1},
	}, nil)
	histogram.Observe(context.Background(), 0.125)

	data := collectMetrics(t, reader)["legacy_fractional_milliseconds"]
	requireMetricMetadata(t, data, "Fractional millisecond latency", string(UnitMilliseconds))
	aggregation, ok := data.Data.(metricdata.Histogram[float64])
	if !ok || len(aggregation.DataPoints) != 1 {
		t.Fatalf("aggregation = %T, want one float64 histogram point", data.Data)
	}
	point := aggregation.DataPoints[0]
	if point.Count != 1 || point.Sum != 0.125 {
		t.Errorf("count/sum = %d/%v, want 1/0.125", point.Count, point.Sum)
	}
}

func TestOTelPrimitivesDoNotFilterAttributeSets(t *testing.T) {
	setTestApp(t, "cardinality-test-app")
	meter, reader := testMeter(t)
	counter := newOTelSingleViewCounter(meter, "legacy_unbounded_counter", "Unbounded counter", []Key{"id"})

	const attributeSetCount = 2100
	for i := range attributeSetCount {
		counter.Increment(context.Background(), 1, Tag{Key: "id", Value: fmt.Sprintf("id-%d", i)})
	}

	data := collectMetrics(t, reader)["legacy_unbounded_counter"]
	sum, ok := data.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatalf("aggregation is %T, want metricdata.Sum[int64]", data.Data)
	}
	if len(sum.DataPoints) != attributeSetCount {
		t.Errorf("data point count = %d, want %d", len(sum.DataPoints), attributeSetCount)
	}
}

func TestOTelInstrumentCreationErrorsPanic(t *testing.T) {
	meter, _ := testMeter(t)

	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("newOTelGauge did not panic for an invalid instrument name")
		}
		if message := fmt.Sprint(got); !strings.Contains(message, "Failed to create OTel gauge") {
			t.Errorf("panic = %q, want OTel gauge construction context", message)
		}
	}()
	newOTelGauge(meter, "", "invalid gauge", nil)
}
