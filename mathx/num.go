// Package mathx contains various math utilities.
package mathx

import "golang.org/x/exp/constraints"

// MinInt returns the smaller of the given numbers.
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxInt returns the larger of the given numbers.
func MaxInt(a, b int) int {
	if a < b {
		return b
	}
	return a
}

// Min returns the smaller of the given numbers.
func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger of the given numbers.
func Max[T constraints.Ordered](a, b T) T {
	if a < b {
		return b
	}
	return a
}

// CeilDivInt returns the ceiling of a/b, where a and b are positive integers. The positive requirement avoids the
// otherwise numerous corner cases. See e.g.: https://ericlippert.com/2013/01/28/integer-division-that-rounds-up/.
func CeilDivInt(a, b int) int {
	return (a-1)/b + 1
}

func IsPowerOf2(a int) bool {
	if a <= 0 {
		return false
	}
	return a&(a-1) == 0
}
