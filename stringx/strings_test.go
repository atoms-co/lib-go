package stringx_test

import (
	"testing"

	"go.cloudkitchens.org/lib/testing/requirex"
	"go.cloudkitchens.org/lib/stringx"
)

func TestToString(t *testing.T) {
	type TestString string

	requirex.Equal(t, stringx.ToString(TestString("test")), "test")
}
