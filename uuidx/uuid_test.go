package uuidx_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"go.atoms.co/lib/testing/requirex"
	"go.atoms.co/lib/uuidx"
)

func TestRetainBits(t *testing.T) {
	u := uuidx.RetainBits(uuid.MustParse("dc9076e9-2fda-4019-bd2c-900a8284b9c4"), 4)
	requirex.Equal(t, u.String(), "d0000000-0000-4000-8000-000000000000")

	u = uuidx.RetainBits(uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"), 4)
	requirex.Equal(t, u.String(), "f0000000-0000-4000-8000-000000000000")

	u = uuidx.RetainBits(uuid.MustParse("00000000-0000-0000-0000-000000000000"), 4)
	requirex.Equal(t, u.String(), "00000000-0000-4000-8000-000000000000")

	u = uuidx.RetainBits(uuid.MustParse("dc9076e9-2fda-4019-bd2c-900a8284b9c4"), 8)
	requirex.Equal(t, u.String(), "dc000000-0000-4000-8000-000000000000")

	u = uuidx.RetainBits(uuid.MustParse("dc9076e9-2fda-4019-bd2c-900a8284b9c4"), 12)
	requirex.Equal(t, u.String(), "dc900000-0000-4000-8000-000000000000")
}

func TestHash(t *testing.T) {
	u := uuidx.Hash([]byte("dc9076e9-2fda-4019-bd2c-900a8284b9c4"))
	requirex.Equal(t, u.String(), "d3bf0a23-7923-4752-87d4-9882ded25008")

	u = uuidx.Hash([]byte("dc9076e9-2fda-4019-bd2c-900a8284b9c5"))
	requirex.Equal(t, u.String(), "d1110a8b-af13-431b-b3e4-4ff8007ccc74")

	a := uuid.MustParse("dc9076e9-2fda-4019-bd2c-900a8284b9c5")
	u = uuidx.Hash(a[:])
	requirex.Equal(t, u.String(), "e5c8ce1b-7ccf-4154-a779-064e8b483edd")
}

func TestInc(t *testing.T) {
	id := uuid.MustParse("ef8076e9-2fda-4019-bd2c-900a8284b9c4")
	assert.Equal(t, uuidx.Inc(id).String(), "ef8076e9-2fda-4019-bd2c-900a8284b9c5")
}
