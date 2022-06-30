package metrics

import (
	"math"
	"testing"
)

func Test_GetUniformBuckets(t *testing.T) {
	buckets := getUniformBuckets(0, 1000, 20)
	expected := []float64{53.0, 105.0, 158.0, 211.0, 263.0, 316.0, 368.0, 421.0, 474.0, 526.0, 579.0, 632.0,
		684.0, 737.0, 789.0, 842.0, 895.0, 947.0, 1000.0}

	for i := range buckets {
		if buckets[i] != expected[i] {
			t.Errorf("Expected %v, received %v", expected[i], buckets[i])
		}
	}
}

func Test_GetExponentialBuckets(t *testing.T) {
	buckets := getExponentialBuckets(0, 100, 10)
	expected := []float64{1, 1.668, 2.782, 4.642, 7.743, 12.915, 21.544, 35.938, 59.948, 100}

	tolerance := 0.001
	for i := range buckets {
		if math.Abs(buckets[i]-expected[i]) > tolerance {
			t.Errorf("Expected %v, received %v", expected[i], buckets[i])
		}
	}
}

func Test_GetUserDefinedBuckets(t *testing.T) {
	buckets := getUserDefinedBuckets([]float64{1, 1.668, 2.782, 4.642, 7.743, 12.915})
	expected := []float64{1, 1.668, 2.782, 4.642, 7.743, 12.915}

	tolerance := 0.001
	for i := range buckets {
		if math.Abs(buckets[i]-expected[i]) > tolerance {
			t.Errorf("Expected %v, received %v", expected[i], buckets[i])
		}
	}
}
