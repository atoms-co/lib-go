package uuidx_test

import (
	"testing"

	"go.cloudkitchens.org/lib/uuidx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
