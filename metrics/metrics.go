// Package metrics provides basic metrics recording functionality.
package metrics

import (
	"context"
	"fmt"
	"time"

	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/runmetrics"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

type Distribution int

const (
	Exponential Distribution = iota
	Uniform
	UserDefined
	version = "1.0.1"
)

type UnitType string

const (
	UnitDimensionless UnitType = stats.UnitDimensionless
	UnitBytes         UnitType = stats.UnitBytes
	UnitMilliseconds  UnitType = stats.UnitMilliseconds
	UnitSeconds       UnitType = stats.UnitSeconds
)

// Name is the name of the metric.
type Name = string

// BucketOptions is used to specify the histogram buckets.
type BucketOptions struct {
	// Start Bucket (>0).
	Start float64
	// End Bucket.
	End float64
	// Number of Buckets (>0).
	NumBuckets int
	// LatencyUnit is the unit of the bucket boundaries for duration-based histograms.
	// "time.Second" or "time.Millisecond". time.Second is the default.
	LatencyUnit time.Duration
	// Distribution
	DistributionType Distribution
	// Required for UserDefined distribution, all other options will be ignored.
	UserDefinedBuckets []float64
}

var (
	// JavaBucketOptions are the options used by Java clients
	JavaBucketOptions = &BucketOptions{
		UserDefinedBuckets: []float64{
			2.0,
			5.0,
			11.0,
			23.0,
			51.0,
			112.0,
			245.0,
			537.0,
			1179.0,
			2588.0,
			5679.0,
			12461.0,
			27344.0,
			60000.0,
		},
		DistributionType: UserDefined,
		LatencyUnit:      time.Millisecond,
	}
)

// Key is a metric tag key.
type Key string

// Commonly-used metric key names
const (
	ActionKey      Key = "action"
	MethodKey      Key = "method"
	MessageTypeKey Key = "message_type"
	ResultKey      Key = "result"
	SegmentKey     Key = "segment"
	StatusKey      Key = "status"
	TableKey       Key = "table"
	TypeKey        Key = "type"
)

// Tag represents the metric tag with a key and a value.
// Example tag: Tag{Key: "serviceName", Value: "fooService"}
type Tag struct {
	Key   Key
	Value string
}

func NewTag(key Key, value any) Tag {
	return Tag{Key: key, Value: fmt.Sprintf("%v", value)}
}

func (t Tag) String() string {
	return fmt.Sprintf("%v:%v", t.Key, t.Value)
}

const (
	AppTagKey Key = "app"
)

// Counter is an interface for reporting counter values.
type Counter interface {
	// Increment increases the counter with the given delta
	// values is a map of all the tag keys with their
	// corresponding values.
	Increment(ctx context.Context, delta int, tags ...Tag)
}

// Gauge is an interface for reporting Gauge metrics.
type Gauge interface {
	// Set the gauge to the given arbitrary value
	// values is a map of all the tag keys with their
	// corresponding values.
	Set(ctx context.Context, value float64, tags ...Tag)
}

// GenericHistogram sets a value for appropriate histogram.
type GenericHistogram[T float64 | time.Duration] interface {
	// Observe adds a single observation to the histogram.
	Observe(ctx context.Context, value T, tags ...Tag)
}

// Histogram sets a duration value for appropriate histogram.
type Histogram = GenericHistogram[time.Duration]

// NewCounter instantiates a counter type for the given metric name, description with
// the given metric tag keys, if any.
func NewCounter(name Name, description string, tagKeys ...Key) Counter {
	return newCounter(name, description, tagKeys)
}

// NewGauge instantiates a gauge type for the given metric name, description with
// the given metric tag keys, if any.
func NewGauge(name Name, description string, tagKeys ...Key) Gauge {
	return newGauge(name, description, tagKeys)
}

// NewHistogram instantiates a histogram with duration values for the given metric name, description and
// options which can be used to specify the bucket boundaries and the metric tag keys, if any.
func NewHistogram(name Name, description string, bucketOptions *BucketOptions, tagKeys ...Key) Histogram {
	return newDurationHistogram(name, description, bucketOptions, tagKeys)
}

// NewDimensionlessHistogram instantiates a histogram with float values for the given metric name, description and
// options which can be used to specify the bucket boundaries and the metric tag keys, if any.
func NewDimensionlessHistogram(name Name, description string, bucketOptions *BucketOptions, tagKeys ...Key) GenericHistogram[float64] {
	return newHistogram(name, description, UnitDimensionless, bucketOptions, tagKeys)
}

// NewByteHistogram instantiates a histogram with byte values for the given metric name, description and
// options which can be used to specify the bucket boundaries and the metric tag keys, if any.
func NewByteHistogram(name Name, description string, bucketOptions *BucketOptions, tagKeys ...Key) GenericHistogram[float64] {
	return newHistogram(name, description, UnitBytes, bucketOptions, tagKeys)
}

// NewSingleViewCounter instantiates a counter type for the given metrics name, description with
// the given metric tag keys, if any. It is different from NewCounter, so it only creates a single
// view to match the Java counterpart.
//
// Use NewSingleViewCounter if there's a Java counterpart producing counter metric with the same name
// and it's important for metric names to match. In other cases, prefer NewCounter.
func NewSingleViewCounter(name Name, description string, tagKeys ...Key) Counter {
	return newSingleViewCounter(name, description, tagKeys)
}

func Init(appName string) error {
	// set the default app value for all metrics.
	initAppName(appName)


	err := view.Register(ocgrpc.DefaultServerViews...)
	if err != nil {
		return err
	}

	// register extra views for gRPC metrics
	view.Register(ocgrpc.ClientStartedRPCsView, ocgrpc.ServerStartedRPCsView)

	return runmetrics.Enable(runmetrics.RunMetricOptions{
		EnableCPU:            true,
		EnableMemory:         true,
		UseDerivedCumulative: true,
	})
}
