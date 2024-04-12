// Package mathx contains various math utilities.
package mathx

import "golang.org/x/exp/constraints"

// MinInt returns the smallest of the given numbers.
// Deprecated: use built-in min
func MinInt(a, b int) int {
	return min(a, b)
}

// MaxInt returns the largest of the given numbers.
// Deprecated: use built-in max
func MaxInt(a, b int) int {
	return max(a, b)
}

// Min returns the smallest of the given numbers.
// Deprecated: use built-in min
func Min[T constraints.Ordered](a, b T) T {
	return min(a, b)
}

// Max returns the largest of the given numbers.
// Deprecated: use built-in max
func Max[T constraints.Ordered](a, b T) T {
	return max(a, b)
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
