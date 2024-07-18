// Package uuidx holds UUID utility functionality, notably logic for splitting UUID ranges.
package uuidx

import (
	"bytes"
	"fmt"
	"math/big"

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

// Divide returns a/b * (uuid.Max+1), i.e, the UUID of the a'th of b partitions.
func Divide(a, b int64) (uuid.UUID, error) {
	if b <= a || b < 1 {
		return uuid.UUID{}, fmt.Errorf("invalid")
	}

	maxValue := Domain.To()
	end := big.NewInt(0).Add(big.NewInt(0).SetBytes(maxValue[:]), big.NewInt(1))
	ret := big.NewInt(0).Div(big.NewInt(0).Mul(end, big.NewInt(a)), big.NewInt(b))

	return uuid.FromBytes(ret.FillBytes(make([]byte, 16)))
}

// Compare compares the given UUIDs a and b. The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
func Compare(a, b uuid.UUID) int {
	// Note: UUID is [16]byte and bytes.Compare needs a []byte.
	return bytes.Compare(a[:], b[:])
}

// Less returns a < b. For convenience in sorting
func Less(a, b uuid.UUID) bool {
	return Compare(a, b) < 0
}

// Inc returns the next uuid
func Inc(n uuid.UUID) uuid.UUID {
	next := big.NewInt(0).Add(big.NewInt(0).SetBytes(n[:]), big.NewInt(1))
	ret, _ := uuid.FromBytes(next.FillBytes(make([]byte, 16)))
	return ret
}
