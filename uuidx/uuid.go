package uuidx

import (
	"crypto/md5"

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
