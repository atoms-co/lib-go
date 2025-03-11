package stringx_test

import (
	"testing"

	"go.atoms.co/lib/testing/requirex"
	"go.atoms.co/lib/stringx"
)

func TestToString(t *testing.T) {
	type TestString string

	requirex.Equal(t, stringx.ToString(TestString("test")), "test")
}
