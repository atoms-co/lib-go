package assertx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Equal is a typed convenience wrapper over assert.Equals to make constants the correct type. Also
// uses the idoimatic "have before want" parameter ordering.
func Equal[T any](t *testing.T, actual, expected T, args ...any) {
	assert.Equal(t, expected, actual, args...)
}
