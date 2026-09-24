package metrics

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"go.atoms.co/slicex"
)

const otelMeterName = "atoms.co/lib-go/metrics"

const (
	otelSourceAttributeKey   = "telemetry_source"
	otelSourceAttributeValue = "otel"
)

var otelMeter = otel.Meter(otelMeterName)

// otelAttributes owns the label schema for an OTel instrument.
type otelAttributes struct {
	declaredKeys     []Key
	declaredKeySet   map[Key]bool
	retainUndeclared bool
}

func newOTelAttributes(tagKeys []Key) otelAttributes {
	keys := slicex.CopyAppend(tagKeys)
	keySet := slicex.NewSet(keys...)
	return otelAttributes{declaredKeys: keys, declaredKeySet: keySet}
}

func (a otelAttributes) option(tags []Tag) metric.MeasurementOption {
	attrs := slicex.Map(a.declaredKeys, func(key Key) attribute.KeyValue {
		return attribute.String(string(key), "")
	})
	for _, tag := range tags {
		if a.declaredKeySet[tag.Key] || a.retainUndeclared {
			attrs = append(attrs, attribute.String(string(tag.Key), tag.Value))
		}
	}

	attrs = append(attrs, attribute.String(string(defaultTag.Key), defaultTag.Value), attribute.String(otelSourceAttributeKey, otelSourceAttributeValue))
	return metric.WithAttributeSet(attribute.NewSet(attrs...))
}

func panicOTelInstrument(kind string, name Name, err error) {
	panic(fmt.Sprintf("Failed to create OTel %s %q: %v", kind, name, err))
}

// otelCounter implements the Counter interface using OTel Counters. It
// creates _count and _sum instruments to match the OpenCensus NewCounter
// two-view behavior. The Prometheus exporter must be configured not to append
// counter suffixes if exact legacy metric names are required.
type otelCounter struct {
	count      metric.Int64Counter
	sum        metric.Int64Counter
	attributes otelAttributes
}

func newOTelCounter(meter metric.Meter, name Name, description string, tagKeys []Key) Counter {
	return newOTelCounterWithAttributes(meter, name, description, newOTelAttributes(tagKeys))
}

func newOTelCounterWithAttributes(meter metric.Meter, name Name, description string, attributes otelAttributes) Counter {
	countName := name + "_count"
	count, err := meter.Int64Counter(countName, metric.WithDescription(description), metric.WithUnit(string(UnitDimensionless)))
	if err != nil {
		panicOTelInstrument("counter", countName, err)
	}

	sumName := name + "_sum"
	sum, err := meter.Int64Counter(sumName, metric.WithDescription(description), metric.WithUnit(string(UnitDimensionless)))
	if err != nil {
		panicOTelInstrument("counter", sumName, err)
	}

	return &otelCounter{
		count:      count,
		sum:        sum,
		attributes: attributes,
	}
}

func (r *otelCounter) Increment(ctx context.Context, delta int, tags ...Tag) {
	opt := r.attributes.option(tags)
	r.count.Add(ctx, 1, opt)
	r.sum.Add(ctx, int64(delta), opt)
}

// otelSingleViewCounter implements the NewSingleViewCounter sum semantics
// with an exact, unsuffixed instrument name.
type otelSingleViewCounter struct {
	counter    metric.Int64Counter
	attributes otelAttributes
}

// newOTelSingleViewCounter creates the OTel equivalent of
// NewSingleViewCounter. The instrument name is exactly name and each
// Increment adds delta to its cumulative sum.
func newOTelSingleViewCounter(meter metric.Meter, name Name, description string, tagKeys []Key) Counter {
	counter, err := meter.Int64Counter(name, metric.WithDescription(description), metric.WithUnit(string(UnitDimensionless)))
	if err != nil {
		panicOTelInstrument("counter", name, err)
	}
	return &otelSingleViewCounter{
		counter:    counter,
		attributes: newOTelAttributes(tagKeys),
	}
}

func (r *otelSingleViewCounter) Increment(ctx context.Context, delta int, tags ...Tag) {
	r.counter.Add(ctx, int64(delta), r.attributes.option(tags))
}

// otelGauge implements the OpenCensus LastValue view semantics with an OTel
// synchronous gauge.
type otelGauge struct {
	gauge      metric.Float64Gauge
	attributes otelAttributes
}

func newOTelGauge(meter metric.Meter, name Name, description string, tagKeys []Key) Gauge {
	gauge, err := meter.Float64Gauge(name, metric.WithDescription(description), metric.WithUnit(string(UnitDimensionless)))
	if err != nil {
		panicOTelInstrument("gauge", name, err)
	}
	return &otelGauge{
		gauge:      gauge,
		attributes: newOTelAttributes(tagKeys),
	}
}

func (r *otelGauge) Set(ctx context.Context, value float64, tags ...Tag) {
	r.gauge.Record(ctx, value, r.attributes.option(tags))
}

type otelFloatHistogram struct {
	histogram  metric.Float64Histogram
	attributes otelAttributes
}

func newOTelFloatHistogram(meter metric.Meter, name Name, description string, unitType UnitType, bucketOptions *BucketOptions, tagKeys []Key) *otelFloatHistogram {
	boundaries := getBuckets(bucketOptions, unitType)
	// OpenCensus canonicalizes distribution bounds during view registration:
	// it sorts them and silently removes zero. Apply the same canonicalization
	// before constructing the OTel histogram so both exporters expose the same
	// bucket boundaries. The two SDKs still differ for observations exactly on
	// a boundary (OC is upper-exclusive; OTel/Prometheus is upper-inclusive),
	// which the OTel histogram API cannot configure.
	sort.Float64s(boundaries)
	boundaries = dropNonPosBuckets(boundaries)
	histogram, err := meter.Float64Histogram(name, metric.WithDescription(description), metric.WithUnit(string(unitType)), metric.WithExplicitBucketBoundaries(boundaries...))
	if err != nil {
		panicOTelInstrument("histogram", name, err)
	}
	return &otelFloatHistogram{
		histogram:  histogram,
		attributes: newOTelAttributes(tagKeys),
	}
}

func (r *otelFloatHistogram) Observe(ctx context.Context, value float64, tags ...Tag) {
	r.histogram.Record(ctx, value, r.attributes.option(tags))
}

type otelDurationHistogram struct {
	*otelFloatHistogram
	unit UnitType
}

func (r *otelDurationHistogram) Observe(ctx context.Context, value time.Duration, tags ...Tag) {
	var observation float64
	switch r.unit {
	case UnitMilliseconds:
		observation = float64(value.Milliseconds())
	default:
		observation = value.Seconds()
	}
	r.otelFloatHistogram.Observe(ctx, observation, tags...)
}

// newOTelDurationHistogram creates the OTel equivalent of NewHistogram, including
// the same duration unit conversion and explicit bucket boundaries.
func newOTelDurationHistogram(meter metric.Meter, name Name, description string, bucketOptions *BucketOptions, tagKeys []Key) Histogram {
	unitType := UnitSeconds
	if bucketOptions != nil && bucketOptions.LatencyUnit == time.Millisecond {
		unitType = UnitMilliseconds
	}
	return &otelDurationHistogram{
		otelFloatHistogram: newOTelFloatHistogram(meter, name, description, unitType, bucketOptions, tagKeys),
		unit:               unitType,
	}
}
