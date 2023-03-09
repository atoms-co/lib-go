// Package mapx contains convenience utilities for working with maps. Some functionality here
// is expected to be subsumed by the standard library at some point.
package mapx

// New returns a map from intrinsically keyed values.
func New[K comparable, V any](values []V, keyOf func(V) K) map[K]V {
	ret := map[K]V{}
	for _, v := range values {
		ret[keyOf(v)] = v
	}
	return ret
}

// MapNew returns a map from transformed values.
func MapNew[K comparable, V any, T any](values []T, fn func(T) (K, V)) map[K]V {
	ret := map[K]V{}
	for _, t := range values {
		k, v := fn(t)
		ret[k] = v
	}
	return ret
}

// Keys extracts all keys to a slice.
func Keys[K comparable, V any](m map[K]V) []K {
	var ret []K
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}

// Values extracts all values to a slice.
func Values[K comparable, V any](m map[K]V) []V {
	var ret []V
	for _, v := range m {
		ret = append(ret, v)
	}
	return ret
}

// ValuesIf extracts all values to a slice, if they satisfy the given predicate.
func ValuesIf[K comparable, V any](m map[K]V, fn func(V) bool) []V {
	var ret []V
	for _, v := range m {
		if fn(v) {
			ret = append(ret, v)
		}
	}
	return ret
}

// MapValues extracts all transformed values to a slice.
func MapValues[K comparable, V, T any](m map[K]V, fn func(V) T) []T {
	var ret []T
	for _, v := range m {
		ret = append(ret, fn(v))
	}
	return ret
}

// MapValuesIf extracts selected transformed values to a slice.
func MapValuesIf[K comparable, V, T any](m map[K]V, fn func(V) (T, bool)) []T {
	var ret []T
	for _, v := range m {
		if w, ok := fn(v); ok {
			ret = append(ret, w)
		}
	}
	return ret
}

// MapToSlice extracts all transformed entries to a slice.
func MapToSlice[K comparable, V, T any](m map[K]V, fn func(k K, v V) T) []T {
	var ret []T
	for k, v := range m {
		ret = append(ret, fn(k, v))
	}
	return ret
}

// Flatten extracts all value elements of a multi-map to a single slice.
func Flatten[K comparable, V any](m map[K][]V) []V {
	var ret []V
	for _, v := range m {
		ret = append(ret, v...)
	}
	return ret
}

// Clone makes a copy of the map (with value copy of keys and values).
func Clone[K comparable, V any](m map[K]V) map[K]V {
	ret := map[K]V{}
	for k, v := range m {
		ret[k] = v
	}
	return ret
}

// Contains returns true if the given key is in the map.
func Contains[K comparable, V any](m map[K]V, k K) bool {
	if m != nil {
		_, ok := m[k]
		return ok
	}
	return false
}

// FilterKeys returns elements with Keys matching the filter function.
func FilterKeys[K comparable, V any](m map[K]V, fn func(K) bool) map[K]V {
	ret := make(map[K]V)
	for k, v := range m {
		if fn(k) {
			ret[k] = v
		}
	}
	return ret
}
