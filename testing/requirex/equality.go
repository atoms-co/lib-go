package requirex

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
)

// Equal is a typed convenience wrapper over require.Equal to make constants the correct type. Also
// uses the idiomatic "have before want" parameter ordering.
func Equal[T any](t *testing.T, actual, expected T, args ...any) {
	t.Helper()
	require.Equal(t, expected, actual, args...)
}

// NotEqual is a typed convenience wrapper over require.NotEqual to make constants the correct type. Also
// uses the idiomatic "have before want" parameter ordering.
func NotEqual[T any](t *testing.T, actual, expected T, args ...any) {
	t.Helper()
	require.NotEqual(t, expected, actual, args...)
}

func EqualProtobuf[T proto.Message](t *testing.T, actual, expected T, args ...any) {
	t.Helper()
	require.Equal(t, proto.MarshalTextString(expected), proto.MarshalTextString(actual), args)
}
