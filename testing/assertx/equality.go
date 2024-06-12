package assertx

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

// Equal is a typed convenience wrapper over assert.Equal to make constants the correct type. Also
// uses the idiomatic "have before want" parameter ordering.
func Equal[T any](t *testing.T, actual, expected T, args ...any) {
	t.Helper()
	assert.Equal(t, expected, actual, args...)
}

// NotEqual is a typed convenience wrapper over assert.NotEqual to make constants the correct type. Also
// uses the idiomatic "have before want" parameter ordering.
func NotEqual[T any](t *testing.T, actual, expected T, args ...any) {
	t.Helper()
	assert.NotEqual(t, expected, actual, args...)
}

func EqualProtobuf(t *testing.T, actual, expected proto.Message, args ...any) {
	t.Helper()
	assert.Equal(t, proto.MarshalTextString(actual), proto.MarshalTextString(expected), args)
}
