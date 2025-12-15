// Package mathx contains various math utilities.
package mathx

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
