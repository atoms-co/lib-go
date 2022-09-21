package assertx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Equal is a typed convenience wrapper over assert.Equal to make constants the correct type. Also
// uses the idiomatic "have before want" parameter ordering.
func Equal[T any](t *testing.T, actual, expected T, args ...any) {
	assert.Equal(t, expected, actual, args...)
}

// NotEqual is a typed convenience wrapper over assert.NotEqual to make constants the correct type. Also
// uses the idiomatic "have before want" parameter ordering.
func NotEqual[T any](t *testing.T, actual, expected T, args ...any) {
	assert.NotEqual(t, expected, actual, args...)
}
