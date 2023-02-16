// Package slicex contains convenience utilities for working with slices. Some functionality here
// is expected to be subsumed by the standard library at some point.
package slicex

// New returns a slice from zero or more values. Useful in non-vararg contexts.
func New[T any](ts ...T) []T {
	return ts
}

// Map transforms all elements.
func Map[T, U any](list []T, fn func(T) U) []U {
	var ret []U
	for _, v := range list {
		ret = append(ret, fn(v))
	}
	return ret
}

// FlatMap transforms all elements into a list and flattens it.
func FlatMap[T, U any](list []T, fn func(T) []U) []U {
	var ret []U
	for _, v := range list {
		ret = append(ret, fn(v)...)
	}
	return ret
}

// MapIf transforms selected elements.
func MapIf[T, U any](list []T, fn func(T) (U, bool)) []U {
	var ret []U
	for _, v := range list {
		if u, ok := fn(v); ok {
			ret = append(ret, u)
		}
	}
	return ret
}

// TryMap transforms all elements using the provided function or returns the first error.
func TryMap[T, U any](list []T, fn func(T) (U, error)) ([]U, error) {
	var ret []U
	for _, v := range list {
		u, err := fn(v)
		if err != nil {
			return []U{}, err
		}
		ret = append(ret, u)
	}
	return ret, nil
}

// Flatten flattens a slice of slices.
func Flatten[T any](list [][]T) []T {
	var ret []T
	for _, v := range list {
		ret = append(ret, v...)
	}
	return ret
}

// Clone makes a copy of the slice (with value copy of elements).
func Clone[T any](list []T) []T {
	return append([]T{}, list...)
}

// CopyAppend makes a copy of the slice (with value copy of elements) and appends the elements to the copy.
func CopyAppend[T any](list []T, elms ...T) []T {
	return append(Clone(list), elms...)
}

// Count returns the number of elements satisfying the predicate.
func Count[T any](list []T, fn func(T) bool) int {
	ret := 0
	for _, v := range list {
		if fn(v) {
			ret++
		}
	}
	return ret
}

// Contains returns true if at least one element satisfying the predicate is found.
func Contains[T any](list []T, fn func(T) bool) bool {
	for _, v := range list {
		if fn(v) {
			return true
		}
	}
	return false
}

// ContainsT returns true if any of the elements are present in the list.
func ContainsT[T comparable](list []T, elms ...T) bool {
	m := map[T]bool{}
	for _, elm := range elms {
		m[elm] = true
	}
	for _, v := range list {
		if m[v] {
			return true
		}
	}
	return false
}

// First returns true and the element if at least one element satisfying the predicate is found.
func First[T any](list []T, fn func(T) bool) (T, bool) {
	for _, v := range list {
		if fn(v) {
			return v, true
		}
	}
	var ret T
	return ret, false
}

// Filter returns elements matching the filter function.
func Filter[T any](list []T, fn func(T) bool) []T {
	var ret []T
	for _, v := range list {
		if fn(v) {
			ret = append(ret, v)
		}
	}
	return ret
}
