package randx

import (
	"testing"
)

func Test_IntnRange(t *testing.T) {
	tests := []struct {
		name string
		min  int
		max  int
	}{
		{
			name: "min is 0 and max is 1",
			min:  0,
			max:  1,
		},
		{
			name: "min is 0 and max is 10",
			min:  0,
			max:  10,
		},
		{
			name: "min is 1 and max is 10",
			min:  1,
			max:  10,
		},
		{
			name: "min is 10 and max is 20",
			min:  10,
			max:  20,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				got := IntnRange(tt.min, tt.max)
				if got < tt.min || got >= tt.max {
					t.Errorf("IntnRange() = %v, want >= %v and < %v", got, tt.min, tt.max)
				}
			}
		})
	}
}
