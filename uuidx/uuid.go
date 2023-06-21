package uuidx

import (
	"crypto/md5"
	"github.com/google/uuid"
)

// RetainBits retains specified number of bits at the beginning of the UUID and sets the rest to zeros.
func RetainBits(id uuid.UUID, bits uint) uuid.UUID {
	rt := uuid.Nil
	bytes := bits / 8
	copy(rt[:bytes], id[:bytes])
	bits %= 8
	if bits > 0 {
		rt[bytes] = id[bytes] & (((1 << (bits + 1)) - 1) << (8 - bits))
	}
	return rt
}

// Hash hashes the provided byte slice and returns the result as UUID v4 (randomly generated).
func Hash(value []byte) uuid.UUID {
	rt := md5.Sum(value)
	// Set UUID version to 4
	rt[6] = (rt[6] & 0x0F) | 0x40
	// Set variant to 2
	rt[8] = (rt[8] & 0x3F) | 0x80
	return rt
}
