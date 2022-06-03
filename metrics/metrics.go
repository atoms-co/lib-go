// Package metrics provides basic metrics recording functionality.
package metrics

import (
	"context"
	"fmt"
	"time"

	"go.cloudkitchens.org/lib/statshandlerx"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/runmetrics"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc"
)

type Distribution int

const (
	Exponential Distribution = iota
	Uniform
	UserDefined
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
	// LatencyUnit - "time.Second" or "time.Millisecond".
	LatencyUnit time.Duration
	// Distribution
	DistributionType Distribution
	// Required for UserDefined distribution, all other options will be ignored.
	// the length should be < maxBuckets
	UserDefinedBuckets []float64
}

// Key is a metric tag key.
type Key string

// Tag represents the metric tag with a key and a value.
// Example tag: Tag{Key: "serviceName", Value: "fooService"}
type Tag struct {
	Key   Key
	Value string
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

// Histogram sets the value for appropriate histogram.
type Histogram interface {
	// Observe adds a single observation to the histogram.
	// values is a map of all the tag keys with their
	// corresponding values.
	Observe(ctx context.Context, duration time.Duration, tags ...Tag)
}

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

// NewHistogram instantiates a histogram type for the given metric name, description and
// options which can be used to specify the bucket boundaries and the metric tag
// keys, if any.
func NewHistogram(name Name, description string, bucketOptions *BucketOptions, tagKeys ...Key) Histogram {
	return newHistogram(name, description, bucketOptions, tagKeys)
}

// WithGrpcStatsHandler sets up the grpc stats handler.
func WithGrpcStatsHandler() grpc.ServerOption {
	return grpc.StatsHandler(&statshandlerx.ServerHandler{})
}

func Init(appName string) error {
	// set the default app value for all metrics.
	initAppName(appName)

	err := view.Register(ocgrpc.DefaultServerViews...)
	if err != nil {
		return err
	}

	return runmetrics.Enable(runmetrics.RunMetricOptions{
		EnableCPU:    true,
		EnableMemory: true,
	})
}
