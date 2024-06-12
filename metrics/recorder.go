package metrics

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	"go.cloudkitchens.org/lib/log"
	"go.cloudkitchens.org/lib/mathx"
)

const (
	// maxBuckets is the maximum bucket size. We restrict the number of buckets to keep metrics cardinality in check.
	maxBuckets = 25
)

// recorder holds all the metric measures along with their appropriate registeredKeys (or tags) and their values.
type recorder struct {
	registeredKeys map[Key]bool // map of all the registered tag keys.
	measure        *stats.Float64Measure
}

var (
	recorders = map[Name]*recorder{}
	lock      sync.Mutex

	defaultTag = Tag{Key: AppTagKey} // default tag recorded on all metrics.

	defaultBucketOptions = &BucketOptions{
		Start:       0.001, // 1ms
		End:         300,   // 5m
		NumBuckets:  20,
		LatencyUnit: time.Second,
	}

	// SlowBucketOptions sets latency histogram from 1ms to 6h, instead of the default 5m. Useful for operations or
	// flows that may become very slow during outages.
	SlowBucketOptions = &BucketOptions{
		Start:       0.001, // 1ms
		End:         21600, // 6h
		NumBuckets:  25,
		LatencyUnit: time.Second,
	}
)

// initAppName sets up the default tag used for all metrics.
func initAppName(appName string) {
	defaultTag.Value = appName
}

func (r *recorder) Increment(ctx context.Context, delta int, tags ...Tag) {
	stats.Record(getTagCtx(ctx, r.registeredKeys, tags), r.measure.M(float64(delta)))
}

func newCounter(name Name, description string, tagKeys []Key) Counter {
	lock.Lock()
	defer lock.Unlock()

	_, existed := recorders[name]
	if existed {
		panic(fmt.Sprintf("Counter \"%v\" is already registered", name))
	}

	count := stats.Float64(name, description, stats.UnitDimensionless)
	r := &recorder{
		measure:        count,
		registeredKeys: make(map[Key]bool),
	}

	tags := setupTags(r, tagKeys)

	// register both Count & Sum aggregation
	// Ref: https://godoc.org/go.opencensus.io/stats/view#Aggregation
	err := view.Register(
		&view.View{
			Name:        fmt.Sprintf("%s_count", name),
			Description: description,
			Measure:     count,
			Aggregation: view.Count(),
			TagKeys:     tags,
		},
		&view.View{
			Name:        fmt.Sprintf("%s_sum", name),
			Description: description,
			Measure:     count,
			Aggregation: view.Sum(),
			TagKeys:     tags,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to register counter: %v", err))
	}

	recorders[name] = r
	return r
}

func newSingleViewCounter(name Name, description string, tagKeys []Key) Counter {
	lock.Lock()
	defer lock.Unlock()

	_, existed := recorders[name]
	if existed {
		panic(fmt.Sprintf("Counter \"%v\" is already registered", name))
	}

	count := stats.Float64(name, description, stats.UnitDimensionless)
	r := &recorder{
		measure:        count,
		registeredKeys: make(map[Key]bool),
	}

	tags := setupTags(r, tagKeys)

	err := view.Register(
		&view.View{
			Name:        name,
			Description: description,
			Measure:     count,
			Aggregation: view.Sum(),
			TagKeys:     tags,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to register counter: %v", err))
	}

	recorders[name] = r
	return r
}

func (r *recorder) Set(ctx context.Context, value float64, tags ...Tag) {
	stats.Record(getTagCtx(ctx, r.registeredKeys, tags), r.measure.M(value))
}

func newGauge(name Name, description string, tagKeys []Key) Gauge {
	lock.Lock()
	defer lock.Unlock()

	_, existed := recorders[name]
	if existed {
		panic(fmt.Sprintf("Gauge \"%v\" is already registered", name))
	}

	recorderM := stats.Float64(name, description, stats.UnitDimensionless)

	r := &recorder{
		measure:        recorderM,
		registeredKeys: make(map[Key]bool),
	}

	tags := setupTags(r, tagKeys)

	// register view along with tags
	err := view.Register(
		&view.View{
			Name:        name,
			Description: description,
			Measure:     recorderM,
			Aggregation: view.LastValue(),
			TagKeys:     tags,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to register gauge metrics: %v", err))
	}

	recorders[name] = r
	return r
}

// getExponentialBuckets calculates the exponential growth factor based on the start, end and num buckets
// and returns the buckets. We thus want, for given start, end and N:
//
//	end = start * factor^(N-1)
//
// After computing 'factor', the bucket boundaries become:
//
//	boundary[i] = start * factor^i
//
// for i in [0; N-1]. Note that factor^0 = 1, so boundary[0] = start.
func getExponentialBuckets(start, end float64, n int) []float64 {
	if start <= 0 {
		start = 1.0
	}
	n = mathx.MinInt(mathx.MaxInt(n, 2), maxBuckets)

	buckets := make([]float64, n)
	factor := math.Pow(end/start, 1.0/float64(n-1))

	buckets[0] = start
	for i := 1; i < n-1; i++ {
		buckets[i] = start * math.Pow(factor, float64(i))
	}
	buckets[n-1] = end

	return buckets
}

// getUniformBuckets splits buckets evenly based on the start, end and num buckets
// and returns the buckets. We thus want, for given start, end and N:
//
// for i in [0; N-1], end - start are evenly divided. boundary[0] = start.
func getUniformBuckets(start, end float64, n int) []float64 {
	if start < 0 {
		start = 1.0
	}
	n = mathx.MinInt(mathx.MaxInt(n, 2), maxBuckets)

	buckets := make([]float64, n)
	step := (end - start) / float64(n-1)

	buckets[0] = start
	buckets[n-1] = end
	for i := 1; i < n-1; i++ {
		buckets[i] = start + math.Round(step*float64(i))
	}

	return dropNonPosBuckets(buckets)
}

func getUserDefinedBuckets(buckets []float64, unit float64) []float64 {
	if len(buckets) < 2 {
		panic("user-defined bucket size must be >= 2")
	}
	var ret []float64
	for _, b := range buckets {
		ret = append(ret, b*unit)
	}
	return ret
}

func dropNonPosBuckets(input []float64) []float64 {
	for i, v := range input {
		if v > 0 {
			return input[i:]
		}
	}
	return []float64{}
}

// getBuckets uses an underlying utility function to get the exponential buckets.
func getBuckets(opt *BucketOptions) []float64 {
	if opt == nil {
		opt = defaultBucketOptions
	}

	// convert units to seconds
	var unit, start, end float64
	switch opt.LatencyUnit {
	case time.Millisecond:
		unit = 1.0
		start = opt.Start
		end = opt.End
	default:
		// anything other than millisecond is reported in seconds, so scale buckets to that
		unit = float64(opt.LatencyUnit) / float64(time.Second)
		start = opt.Start * unit
		end = opt.End * unit
	}

	if opt.DistributionType == Exponential {
		return getExponentialBuckets(start, end, opt.NumBuckets)
	} else if opt.DistributionType == UserDefined {
		return getUserDefinedBuckets(opt.UserDefinedBuckets, unit)
	} else {
		return getUniformBuckets(start, end, opt.NumBuckets)
	}
}

func setupHistogram(name string, description string, bucketOptions *BucketOptions, tagKeys []Key) (*recorder, error) {
	r := &recorder{
		registeredKeys: make(map[Key]bool),
	}
	tags := setupTags(r, tagKeys)
	unit := stats.UnitSeconds
	if bucketOptions != nil && bucketOptions.LatencyUnit == time.Millisecond {
		unit = stats.UnitMilliseconds
	}
	m := stats.Float64(name, description, unit)
	buckets := getBuckets(bucketOptions)
	err := view.Register(&view.View{
		Name:        name,
		Description: description,
		Measure:     m,
		Aggregation: view.Distribution(buckets...),
		TagKeys:     tags,
	})
	if err != nil {
		return nil, err
	}

	r.measure = m
	return r, nil
}

func (r *recorder) Observe(ctx context.Context, elapsed time.Duration, tags ...Tag) {
	var dur float64
	switch r.measure.Unit() {
	case stats.UnitMilliseconds:
		dur = float64(elapsed.Milliseconds())
	default:
		dur = elapsed.Seconds()
	}
	stats.Record(getTagCtx(ctx, r.registeredKeys, tags), r.measure.M(dur))
}

func newHistogram(name Name, description string, bucketOptions *BucketOptions, tagKeys []Key) Histogram {
	lock.Lock()
	defer lock.Unlock()
	_, existed := recorders[name]
	if existed {
		panic(fmt.Sprintf("Histogram \"%v\" is already registered", name))
	}

	r, err := setupHistogram(name, description, bucketOptions, tagKeys)
	if err != nil {
		panic(fmt.Sprintf("Failed to register histogram: %v", err))
	}
	recorders[name] = r

	return r
}

// setupTags sets up the tag keys and returns the corresponding []tag.Key
// required for registration.
func setupTags(r *recorder, tagKeys []Key) []tag.Key {
	// always have the default tag.
	ret := []tag.Key{tag.MustNewKey(string(defaultTag.Key))}
	for _, t := range tagKeys {
		// update the map so that we can cross check during the actual
		// stats Record.
		r.registeredKeys[t] = true
		ret = append(ret, tag.MustNewKey(string(t)))
	}

	return ret
}

func getTagCtx(ctx context.Context, registeredKeys map[Key]bool, tags []Tag) context.Context {
	var mutations []tag.Mutator
	// get the tags passed in tags now overwriting any defaults.
	for _, t := range tags {
		// check if the key is registered.
		if _, ok := registeredKeys[t.Key]; !ok {
			log.Errorf(ctx, "Metrics tag with Key \"%v\" is not registered", t.Key)
		}
		mutations = append(mutations, tag.Upsert(tag.MustNewKey(string(t.Key)), t.Value))
	}

	// make sure to have the default tag too.
	mutations = append(mutations, tag.Upsert(tag.MustNewKey(string(defaultTag.Key)), defaultTag.Value))
	ctx, _ = tag.New(ctx, mutations...)
	return ctx
}
