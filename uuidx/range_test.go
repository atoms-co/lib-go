package uuidx_test

import (
	"fmt"
	"testing"

	"go.cloudkitchens.org/lib/uuidx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzTestSplitByWeights(f *testing.F) {
	f.Add(50, 40, 10)

	f.Fuzz(func(t *testing.T, a, b, c int) {
		weights := []int{a, b, c}

		if a+b+c != 100 || a < 0 || b < 0 || c < 0 {
			t.SkipNow()
			return
		}

		shards, err := uuidx.SplitByWeights(uuidx.Domain, weights...)
		assert.NoError(t, err)
		assert.Equal(t, len(weights), len(shards))
		assert.True(t, uuidx.IsRangeSerializable(shards...))

		distributions := uuidx.RangeDistributions(shards...)
		for i, distribution := range distributions {
			assert.Equal(t, weights[i], int(distribution), fmt.Sprintf("actual distribution: %+v | expected: %+v", distributions, weights))
		}
	})
}

func TestSplitByWeights(t *testing.T) {
	tests := []struct {
		weights       []int
		expectedError bool
	}{
		{weights: []int{100}},
		{weights: []int{50, 50}},
		{weights: []int{25, 25, 25, 25}},
		{weights: []int{50, 25, 25}},
		{weights: []int{25, 25, 50}},
		{weights: []int{10, 90}},
		{weights: []int{90, 10}},
		{weights: []int{99, 1}},
		{weights: []int{95, 5}},
		{weights: []int{90, 5, 5}},
		{weights: []int{23, 44, 33}},
		{weights: []int{90, 0, 10}},
		{weights: []int{58, 26, 16}},
	}

	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%v", test.weights), func(t *testing.T) {
			shards, err := uuidx.SplitByWeights(uuidx.Domain, test.weights...)
			if test.expectedError && err == nil {
				t.Fatal("expected error but got <nil>")
			} else if !test.expectedError && err != nil {
				t.Fatalf("expected no error but got %s", err)
			}

			assert.Equal(t, len(test.weights), len(shards))
			assert.True(t, uuidx.IsRangeSerializable(shards...))

			distributions := uuidx.RangeDistributions(shards...)
			for i, distribution := range distributions {
				assert.Equal(t, test.weights[i], int(distribution), fmt.Sprintf("actual distribution: %+v | expected: %+v", distributions, test.weights))
			}
		})
	}
}

func TestSplit(t *testing.T) {
	from := uuid.MustParse("80000000-0000-0000-0000-000000000000")
	to := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")

	assert.Equal(t, -1, uuidx.Compare(from, to))

	r, err := uuidx.NewRange(from, to)
	assert.NoError(t, err)

	size := 6
	ranges, err := uuidx.Split(r, size)
	assert.NoError(t, err)
	assert.Equal(t, size, len(ranges))
	assert.Equal(t, from, ranges[0].From())
	assert.Equal(t, to, ranges[len(ranges)-1].To())
}

// TestShards verifies the basic range splitting logic.
func TestShards(t *testing.T) {
	from := uuid.MustParse("dc9076e9-2fda-4019-bd2c-900a8284b9c4")
	to := uuid.MustParse("ef8076e9-2fda-4019-bd2c-900a8284b9c4")
	tests := []struct {
		from          uuid.UUID
		to            uuid.UUID
		numPartitions int
	}{
		{from: uuidx.Domain.From(), to: uuidx.Domain.To(), numPartitions: 32}, // full range - even number of partitions
		{from: uuidx.Domain.From(), to: uuidx.Domain.To(), numPartitions: 27}, // full range - odd number of partitions
		{from: from, to: to, numPartitions: 16},                               // even number of partitions
		{from: from, to: to, numPartitions: 5},                                // odd number of partitions
	}

	for _, tc := range tests {
		s, err := uuidx.NewRange(tc.from, tc.to)
		require.NoError(t, err, "failed to get a shard")
		ranges, err := uuidx.Split(s, tc.numPartitions)
		require.NoError(t, err, "failed to split ranges")
		assert.Equal(t, len(ranges), tc.numPartitions, "number of partitions is not the same")

		prev := uuidx.Range{}
		for i, r := range ranges {
			if i == 0 {
				prev = r
				continue
			}
			assert.Greaterf(t, uuidx.Compare(r.To(), prev.To()), 0, "UUIDs not sorted")
			// reset prev
			prev = r
		}
	}
}

func TestIntersect(t *testing.T) {
	a := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	b := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	c := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	d := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	e := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")

	s, _ := uuidx.NewRange(b, d)

	assert.Equal(t, false, s.Contains(a))
	assert.Equal(t, true, s.Contains(b))
	assert.Equal(t, true, s.Contains(c))
	assert.Equal(t, false, s.Contains(d))

	r, _ := uuidx.NewRange(a, b)

	// First verify our expected sematics match the established ones
	xr, _ := uuidx.NewRange(a, b)
	assert.Equal(t, true, xr.Contains(a))
	assert.Equal(t, false, xr.Contains(b))

	assert.Equal(t, true, r.Contains(a))
	assert.Equal(t, false, r.Contains(b))

	_, ok := r.Intersects(s)
	assert.False(t, ok)

	r, _ = uuidx.NewRange(b, c)

	ir, ok := r.Intersects(s)
	assert.True(t, ok)
	assert.Equal(t, b, ir.From())
	assert.Equal(t, c, ir.To())

	r, _ = uuidx.NewRange(a, c)

	ir, ok = r.Intersects(s)
	assert.True(t, ok)
	assert.Equal(t, b, ir.From())
	assert.Equal(t, c, ir.To())

	r, _ = uuidx.NewRange(a, d)

	ir, ok = r.Intersects(s)
	assert.True(t, ok)
	assert.Equal(t, b, ir.From())
	assert.Equal(t, d, ir.To())

	r, _ = uuidx.NewRange(b, d)

	ir, ok = r.Intersects(s)
	assert.True(t, ok)
	assert.Equal(t, b, ir.From())
	assert.Equal(t, d, ir.To())

	r, _ = uuidx.NewRange(uuidx.Domain.From(), uuidx.Domain.To())

	ir, ok = r.Intersects(s)
	assert.True(t, ok)
	assert.Equal(t, b, ir.From())
	assert.Equal(t, d, ir.To())

	r, _ = uuidx.NewRange(c, e)

	ir, ok = r.Intersects(s)
	assert.True(t, ok)
	assert.Equal(t, c, ir.From())
	assert.Equal(t, d, ir.To())

	r, _ = uuidx.NewRange(d, e)

	_, ok = r.Intersects(s)
	assert.False(t, ok)
}

func TestDivisible(t *testing.T) {
	quants := []int{8, 16, 32, 64}

	for _, quant := range quants {
		ranges, err := uuidx.Split(uuidx.Domain, quant)
		assert.Nil(t, err)
		ranges2, err := uuidx.Split(uuidx.Domain, quant*2)
		assert.Nil(t, err)
		for ii := 0; ii < quant; ii++ {
			assert.Equal(t, ranges[ii].From().String(), ranges2[ii*2].From().String(), "%v shards doesn't subdivide cleanly", quant)
			assert.Equal(t, ranges[ii].To().String(), ranges2[ii*2+1].To().String(), "%v shards doesn't subdivide cleanly", quant)
		}
	}
}

// Test if we pad big int representations correctly for a magnitudes of numbers
func TestMagnitudes(t *testing.T) {
	ranges, err := uuidx.Split(uuidx.Domain, 1)
	assert.Nil(t, err)
	assert.Len(t, ranges, 1)

	ranges, err = uuidx.Split(uuidx.Domain, 10)
	assert.Nil(t, err)
	assert.Len(t, ranges, 10)

	ranges, err = uuidx.Split(uuidx.Domain, 100)
	assert.Nil(t, err)
	assert.Len(t, ranges, 100)

	ranges, err = uuidx.Split(uuidx.Domain, 1000)
	assert.Nil(t, err)
	assert.Len(t, ranges, 1000)

	ranges, err = uuidx.Split(uuidx.Domain, 10000)
	assert.Nil(t, err)
	assert.Len(t, ranges, 10000)

	ranges, err = uuidx.Split(uuidx.Domain, 100000)
	assert.Nil(t, err)
	assert.Len(t, ranges, 100000)
}

// TestShardsInvalid tests some basic invalid inputs.
func TestShardsInvalid(t *testing.T) {
	_, err := uuidx.NewRange(uuidx.Domain.To(), uuidx.Domain.From())
	require.Error(t, err, "to uuid > from uuid should be invalid")

	ranges, err := uuidx.Split(uuidx.Domain, 0)
	require.Error(t, err, "num partitions of 0 should be invalid")
	assert.Equal(t, len(ranges), 0, "num ranges should be empty")

	s, err := uuidx.NewRange(uuid.MustParse("dc9076e9-2fda-4019-bd2c-900a8284b9c4"), uuid.MustParse("dc9076e9-2fda-4019-bd2c-900a8284b9c5"))
	require.NoError(t, err, "valid UUID range should not error")
	ranges, err = uuidx.Split(s, 5)
	require.Error(t, err, "num partitions > num UUIDs should be invalid")
}

func TestInc(t *testing.T) {
	id := uuid.MustParse("ef8076e9-2fda-4019-bd2c-900a8284b9c4")
	assert.Equal(t, uuidx.Inc(id).String(), "ef8076e9-2fda-4019-bd2c-900a8284b9c5")
}
