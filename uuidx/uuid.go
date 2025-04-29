package uuidx

import (
	"bytes"
	"crypto/md5"
	"math/big"

	"github.com/google/uuid"
)

// RetainBits retains specified number of bits at the beginning of the UUID and sets the rest to zeros.
// The returned UUID has version 4 with variant 2.
func RetainBits(id uuid.UUID, bits uint) uuid.UUID {
	rt := uuid.Nil
	bytes := bits / 8
	copy(rt[:bytes], id[:bytes])
	b := bits % 8
	if b > 0 {
		rt[bytes] = id[bytes] & (((1 << b) - 1) << (8 - b))
	}
	setUUIDRandomVersion(rt[:])
	return rt
}

// Hash hashes the provided byte slice and returns the result as UUID v4 (randomly generated).
func Hash(value []byte) uuid.UUID {
	rt := md5.Sum(value)
	setUUIDRandomVersion(rt[:])
	return rt
}

func setUUIDRandomVersion(value []byte) {
	// Set UUID version to 4
	value[6] = (value[6] & 0x0F) | 0x40
	// Set variant to 2
	value[8] = (value[8] & 0x3F) | 0x80
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
