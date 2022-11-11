// Package uuidx holds UUID utility functionality, notably logic for splitting UUID ranges.
package uuidx

import (
	"bytes"
	"fmt"
	"math"
	"math/big"

	"go.cloudkitchens.org/lib/css/slices"
	"github.com/google/uuid"
)

var (
	Min = uuid.Nil // Provide Min to be consistent with Max
	Max = uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
)

// Domain is a the full UUID range.
var Domain = Range{from: uuid.Nil, to: Max}

// Range represents a half-open UUID range [from, to).
type Range struct {
	from, to uuid.UUID
}

// NewRange returns a shard from the given from and to UUID.
func NewRange(from, to uuid.UUID) (Range, error) {
	if Compare(from, to) >= 0 {
		return Range{}, fmt.Errorf("range start UUID must be less than the end UUID")
	}

	return Range{
		from: from,
		to:   to,
	}, nil
}

func MustNewRange(from, to uuid.UUID) Range {
	return Range{
		from: from,
		to:   to,
	}
}

func (s Range) From() uuid.UUID {
	return s.from
}

func (s Range) To() uuid.UUID {
	return s.to
}

// Contains returns true iff the given key is in the half-open range.
func (s Range) Contains(key uuid.UUID) bool {
	return Compare(s.from, key) <= 0 && Compare(key, s.to) < 0
}

// Intersects returns the range intersection of two ranges, and if it intersects.
func (s Range) Intersects(r Range) (Range, bool) {
	if Compare(r.from, s.to) >= 0 {
		return Range{}, false
	}
	if Compare(r.to, s.from) <= 0 {
		return Range{}, false
	}
	ret := r
	if Compare(r.from, s.from) < 0 {
		// this end of the range is for less than the full shard
		ret.from = s.from
	}
	if Compare(r.to, s.to) > 0 {
		// this end of the range is for less than the full shard
		ret.to = s.to
	}
	return ret, true
}

// Size returns the number of UUIDs in the range.
func (s Range) Size() *big.Int {
	return big.NewInt(0).Sub(big.NewInt(0).SetBytes(s.to[:]), big.NewInt(0).SetBytes(s.from[:]))
}

func (s Range) String() string {
	return fmt.Sprintf("[%s;%s)", s.from.String(), s.to.String())
}

// SplitByWeights shards a UUID range by a set of weights that determine the size of each shard.
// For example, weights of '(50, 50)' would give 2 shards of equal size whereas weights of (50, 25, 10, 15) would give
// 4 shards of various different sizes.
func SplitByWeights(s Range, weights ...int) ([]Range, error) {
	var (
		ranges       = make([]Range, len(weights))
		start        = s.From()
		end          = s.To()
		tokenRange   = big.NewInt(0)
		total        = big.NewInt(0)
		weightsTotal = 0
	)

	total = total.Sub(
		big.NewInt(0).SetBytes(end[:]),
		big.NewInt(0).SetBytes(start[:]),
	).
		Add(total, big.NewInt(1))

	// Keep track of the original position of each weight so that
	// weights that are 0 can remain in-place without affecting
	// the UUID splitting logic.
	originalPos := make(map[int]int)

	// Calculate how many UUIDs should be allocate to each shard based on
	// the weights given and the total number of UUIDs available in the current
	// range.
	uuidsPerWeight := make([]*big.Int, 0)
	for i, weight := range weights {
		// Any weights that are zero require an emtpy range to be created
		// as the ordering needs to be preserved.
		if weight == 0 {
			ranges[i] = Range{
				from: Min,
				to:   Min,
			}

			continue
		}

		// Number of UUIDs allocated per shard = Total / weight * 100
		div := big.NewInt(0).Mul(total, big.NewInt(int64(weight)))
		div.Div(div, big.NewInt(100))

		uuidsPerWeight = append(uuidsPerWeight, div)

		// Keep track of the index that was just inserted as zero weights are
		// not inserted into `uuidsPerWeight` but still need to be tracked.
		originalPos[len(uuidsPerWeight)-1] = i

		// Sum all the weights to ensure we allocate 100%.
		weightsTotal += weight
	}

	if weightsTotal != 100 {
		return nil, fmt.Errorf("expected the weights to sum to 100 but got %v", weightsTotal)
	}

	var to uuid.UUID
	var err error
	for i, weightedPartSize := range uuidsPerWeight {
		// Ensure the last index always allocates up to ffff-fffff..
		if i == len(uuidsPerWeight)-1 {
			to = end
		} else {
			tokenRange.Add(weightedPartSize, big.NewInt(0).SetBytes(start[:]))

			tokenRangeBytes := make([]byte, 16)
			if to, err = uuid.FromBytes(tokenRange.FillBytes(tokenRangeBytes)); err != nil {
				return []Range{}, fmt.Errorf("partition range: %v", err)
			}
		}

		split, err := NewRange(start, to)
		if err != nil {
			return nil, fmt.Errorf("range [%s-%s] is invalid: %s", start, to, err)
		}

		pos := originalPos[i]
		ranges[pos] = split
		start = to
	}

	return ranges, nil
}

// Split uniformly splits the given range into N sub-ranges of equal size.
func Split(s Range, numPartitions int) ([]Range, error) {
	start := s.From()
	end := s.To()
	if numPartitions <= 0 {
		return []Range{}, fmt.Errorf("number of partitions should atleast be 1")
	}

	ranges := make([]Range, numPartitions)

	// size of each partition = ((end - start + 1) / numPartitions).
	tokenRange := big.NewInt(0)
	// (end - start + 1)
	total := big.NewInt(0)
	total = total.Sub(big.NewInt(0).SetBytes(end[:]), big.NewInt(0).SetBytes(start[:])).Add(total, big.NewInt(1))
	// ((end - start + 1) / numPartitions)
	partSize := big.NewInt(0)
	partSize = partSize.Div(total, big.NewInt(int64(numPartitions)))
	// each partition should atleast hold atleast one UUID.
	if partSize.Cmp(big.NewInt(1)) < 0 {
		return []Range{}, fmt.Errorf("number of partitions > total number of UUIDs in range: (%v)", s.String())
	}

	var to uuid.UUID
	var err error
	for partition := 0; partition < numPartitions; partition++ {
		if partition == numPartitions-1 {
			to = end
		} else {
			// `start` keeps getting re-extended so simply keep adding the partition size
			// to the previous `start` to get the correct UUID boundary.
			tokenRange.Add(partSize, big.NewInt(0).SetBytes(start[:]))

			// must be 16 bytes to work with uuid
			tokenRangeBytes := make([]byte, 16)
			if to, err = uuid.FromBytes(tokenRange.FillBytes(tokenRangeBytes)); err != nil {
				return []Range{}, fmt.Errorf("partition range: %v", err)
			}
		}

		split, err := NewRange(start, to)
		if err != nil {
			return nil, fmt.Errorf("range [%s-%s] is invalid: %s", start, to, err)
		}

		ranges[partition] = split
		start = to
	}

	return ranges, nil
}

// Compare compares the given UUIDs a and b. The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
func Compare(a, b uuid.UUID) int {
	// Note: UUID is [16]byte and bytes.Compare needs a []byte.
	return bytes.Compare(a[:], b[:])
}

// Inc returns the next uuid
func Inc(n uuid.UUID) uuid.UUID {
	next := big.NewInt(0).Add(big.NewInt(0).SetBytes(n[:]), big.NewInt(1))
	ret, _ := uuid.FromBytes(next.FillBytes(make([]byte, 16)))
	return ret
}

// IsSerializable verifies whether a set of ranges can be serialized (i.e., strictly ordered)
// and contains the full "Domain" UUID keyspace.
//
// The given `ranges` don't need to be ordered as long as the totality can be ordered.
func IsRangeSerializable(ranges ...Range) bool {
	var (
		startRanges = make(map[uuid.UUID]Range)
		start       = uuid.Nil
		end         = Max
	)

	for _, r := range ranges {
		// Skip any zeroed ranges
		if r.from == Min && r.to == Min {
			continue
		}

		startRanges[r.From()] = r
	}

	for start != end {
		next, ok := startRanges[start]
		if !ok {
			return false
		}

		start = next.To()
		if start == uuid.Nil {
			return false
		}
	}

	return true
}

// RangeDistributions takes a list of UUID ranges and calculates the distribution %
// based on how many UUIDs are allocated in each. This is typically used in conjunction with
// the `SplitByWeights` methods that create uneven allocations.
//
// The map that is returned is keyed by the index of the `ranges` that were inputed and the
// value is a weighted percentage that the given range index has allocated.
func RangeDistributions(ranges ...Range) map[int]int64 {
	// Not all situations have a full UUID range so need to collect the total amount of UUIDs
	// that can be allocated across the set of ranges given.
	totalSize := slices.Fold(ranges, big.NewInt(0), func(total *big.Int, r Range) *big.Int {
		return total.Add(total, r.Size())
	})

	// Map key is the index of the `ranges` element.
	allocs := make(map[int]int64, len(ranges))

	for index, rng := range ranges {
		totalSizeF := big.NewFloat(0).SetInt(totalSize)
		sizeF := big.NewFloat(0).SetInt(rng.Size())

		// Handle zero weights
		if rng.Size().Cmp(big.NewInt(0)) == 0 {
			allocs[index] = 0
			continue
		}

		// Size of allocation / Total size of UUID ranges
		res := sizeF.Quo(sizeF, totalSizeF)
		p, _ := res.Float64()

		// Round is needed for some edge cases where values are 0.9999
		//
		// Normalized Percentage = Percentage of UUID Allocations * 100
		allocs[index] = int64(math.Round(p * 100))
	}

	return allocs
}
