// Package randx contains convenience utilities for the rand package.
package randx

import (
	"math/rand"
	"time"
)

// Duration returns a random duration upto the given one.
func Duration(duration time.Duration) time.Duration {
	return time.Duration(rand.Int63n(int64(duration)))
}

// Element returns a random element of the slice. If empty, returns the default value.
func Element[T any](list []T) T {
	if len(list) == 0 {
		var t T
		return t
	}
	return list[rand.Intn(len(list))]
}
