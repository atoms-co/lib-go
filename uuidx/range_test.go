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
