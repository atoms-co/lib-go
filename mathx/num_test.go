package mathx_test

import (
	"fmt"
	"testing"

	"go.atoms.co/lib/testing/requirex"
	"go.atoms.co/lib/mathx"
)

func TestIsPowerOf2(t *testing.T) {
	tests := []struct {
		value    int
		expected bool
	}{
		{-3, false},
		{0, false},
		{1, true},
		{2, true},
		{1024, true},
		{1025, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.value), func(t *testing.T) {
			requirex.Equal(t, mathx.IsPowerOf2(tt.value), tt.expected)
		})
	}
}
