// Package mapx contains convenience utilities for working with maps. Some functionality here
// is expected to be subsumed by the standard library at some point.
package mapx

// New returns a map from intrinsically keyed values.
func New[K comparable, V any](values []V, keyOf func(V) K) map[K]V {
	ret := make(map[K]V, len(values))
	for _, v := range values {
		ret[keyOf(v)] = v
	}
	return ret
}

// Clone makes a copy of the map (with value copy of keys and values).
// TODO(jhhurwitz): 09/16/24 Remove when we upgrade to go 1.23 (https://github.com/golang/go/issues/69110)
func Clone[K comparable, V any](m map[K]V) map[K]V {
	ret := make(map[K]V, len(m))
	for k, v := range m {
		ret[k] = v
	}
	return ret
}

// MapNew returns a map from transformed values.
func MapNew[K comparable, V any, T any](values []T, fn func(T) (K, V)) map[K]V {
	ret := make(map[K]V, len(values))
	for _, t := range values {
		k, v := fn(t)
		ret[k] = v
	}
	return ret
}

// Keys extracts all keys to a slice.
func Keys[K comparable, V any](m map[K]V) []K {
	if len(m) == 0 {
		return nil
	}
	ret := make([]K, 0, len(m))
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}

// Values extracts all values to a slice.
func Values[K comparable, V any](m map[K]V) []V {
	if len(m) == 0 {
		return nil
	}
	ret := make([]V, 0, len(m))
	for _, v := range m {
		ret = append(ret, v)
	}
	return ret
}

// ValuesIf extracts all values to a slice, if they satisfy the given predicate.
func ValuesIf[K comparable, V any](m map[K]V, fn func(V) bool) []V {
	if len(m) == 0 {
		return nil
	}
	var ret []V
	for _, v := range m {
		if fn(v) {
			ret = append(ret, v)
		}
	}
	return ret
}

// Map extracts all transformed keys and values to a map.
func Map[K, K1 comparable, V, V1 any](m map[K]V, fn func(K, V) (K1, V1)) map[K1]V1 {
	ret := make(map[K1]V1, len(m))
	for k, v := range m {
		k1, v1 := fn(k, v)
		ret[k1] = v1
	}
	return ret
}

// TryMap extracts all transformed keys and values to a map.
func TryMap[K, K1 comparable, V, V1 any](m map[K]V, fn func(K, V) (K1, V1, error)) (map[K1]V1, error) {
	ret := make(map[K1]V1, len(m))
	for k, v := range m {
		k1, v1, err := fn(k, v)
		if err != nil {
			return nil, err
		}
		ret[k1] = v1
	}
	return ret, nil
}

// MapIf extracts all transformed keys and values to a map, if they satisfy a given predicate
func MapIf[K, K1 comparable, V, V1 any](m map[K]V, fn func(K, V) (K1, V1, bool)) map[K1]V1 {
	ret := make(map[K1]V1)
	for k, v := range m {
		if k1, v1, ok := fn(k, v); ok {
			ret[k1] = v1
		}
	}
	return ret
}

// TryMapIf extracts all transformed keys and values to a map, if they satisfy a given predicate
func TryMapIf[K, K1 comparable, V, V1 any](m map[K]V, fn func(K, V) (K1, V1, bool, error)) (map[K1]V1, error) {
	ret := make(map[K1]V1)
	for k, v := range m {
		k1, v1, ok, err := fn(k, v)
		if err != nil {
			return nil, err
		}
		if ok {
			ret[k1] = v1
		}
	}
	return ret, nil
}

// MapValues extracts all transformed values to a slice.
func MapValues[K comparable, V, T any](m map[K]V, fn func(V) T) []T {
	if len(m) == 0 {
		return nil
	}
	ret := make([]T, 0, len(m))
	for _, v := range m {
		ret = append(ret, fn(v))
	}
	return ret
}

// MapValuesIf extracts selected transformed values to a slice.
func MapValuesIf[K comparable, V, T any](m map[K]V, fn func(V) (T, bool)) []T {
	if len(m) == 0 {
		return nil
	}
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
	if len(m) == 0 {
		return nil
	}
	ret := make([]T, 0, len(m))
	for k, v := range m {
		ret = append(ret, fn(k, v))
	}
	return ret
}

// TryMapToSlice extracts all transformed entries to a slice.
func TryMapToSlice[K comparable, V, T any](m map[K]V, fn func(k K, v V) (T, error)) ([]T, error) {
	if len(m) == 0 {
		return nil, nil
	}
	ret := make([]T, 0, len(m))
	for k, v := range m {
		e, err := fn(k, v)
		if err != nil {
			return nil, err
		}
		ret = append(ret, e)
	}
	return ret, nil
}

// Flatten extracts all value elements of a multi-map to a single slice.
func Flatten[K comparable, V any](m map[K][]V) []V {
	if len(m) == 0 {
		return nil
	}
	var ret []V
	for _, v := range m {
		ret = append(ret, v...)
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

// Merge combines multiple maps into a single map. Values with identical keys are overridden by the last one.
func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	sz := 0
	for _, m := range maps {
		sz += len(m)
	}
	rt := make(map[K]V, sz)
	for _, m := range maps {
		for k, v := range m {
			rt[k] = v
		}
	}
	return rt
}

// GetOnly returns the only key-value pair in the map. When used with a map that has more than one element,
// returns an arbitrary pair and true. When used with an empty map, returns the zero values and false.
func GetOnly[K comparable, V any](m map[K]V) (K, V, bool) {
	for k, v := range m {
		return k, v, true
	}
	var k K
	var v V
	return k, v, false
}
