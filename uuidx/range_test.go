package uuidx_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.atoms.co/lib/uuidx"
)

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

	// This test is an example of splitting a full range of UUID by number of shards that is not power of 2.
	// Such splitting is correct, but can introduce unnecessary complications when used by humans. For example,
	// given a UUID close to the range border, it would be harder for a person to figure out which range it belongs to
	// (e.g. try to determine whether `071c71c7-1c71-c71c-71c7-1c17c71c71c7` is in the first range or second for
	// 36 shards vs. `081c71c7-1c71-c71c-71c7-1c71c71c71c7` for 32).
	t.Run("split by not power of 2", func(t *testing.T) {
		s, _ := uuidx.NewRange(uuidx.Min, uuidx.Max)
		ranges, _ := uuidx.Split(s, 36)
		assert.Equal(t, ranges[0].From(), uuidx.Min)
		assert.Equal(t, ranges[0].To().String(), "071c71c7-1c71-c71c-71c7-1c71c71c71c7")

		ranges, _ = uuidx.Split(s, 32)
		assert.Equal(t, ranges[0].From(), uuidx.Min)
		assert.Equal(t, ranges[0].To().String(), "08000000-0000-0000-0000-000000000000")
	})
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
