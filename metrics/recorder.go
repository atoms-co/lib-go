package metrics

import (
	"math"
	"time"
)

const (
	// maxBuckets is the maximum bucket size. We restrict the number of buckets to keep metrics cardinality in check.
	maxBuckets = 25
)

var (
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
	n = min(max(n, 2), maxBuckets)

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
	n = min(max(n, 2), maxBuckets)

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

// getBuckets uses an underlying utility function to get buckets.
func getBuckets(opt *BucketOptions, unitType UnitType) []float64 {
	if opt == nil {
		opt = defaultBucketOptions
	}

	var unit, start, end float64
	switch unitType {
	case UnitSeconds:
		unit = 1.0
		if opt.LatencyUnit != 0 {
			unit = float64(opt.LatencyUnit) / float64(time.Second)
		}
		start = opt.Start * unit
		end = opt.End * unit
	default:
		unit = 1.0
		start = opt.Start
		end = opt.End
	}

	if opt.DistributionType == Exponential {
		return getExponentialBuckets(start, end, opt.NumBuckets)
	} else if opt.DistributionType == UserDefined {
		return getUserDefinedBuckets(opt.UserDefinedBuckets, unit)
	} else {
		return getUniformBuckets(start, end, opt.NumBuckets)
	}
}
