package yamlx_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"go.cloudkitchens.org/lib/encoding/yamlx"
	"go.cloudkitchens.org/lib/testing/requirex"
)

type testStruct struct {
	Field string `yaml:"field"`
}

func TestUnmarshal(t *testing.T) {
	t.Run("doesn't fail with unknown field", func(t *testing.T) {
		data := []byte("field: value\nunknown: value")
		s, err := yamlx.Unmarshal[testStruct](data)
		require.NoError(t, err)
		requirex.Equal(t, testStruct{Field: "value"}, s)
	})
}

func TestUnmarshalStrict(t *testing.T) {
	t.Run("fails with unknown field", func(t *testing.T) {
		data := []byte("field: value\nunknown: value")
		_, err := yamlx.UnmarshalStrict[testStruct](data)
		require.Error(t, err)
	})

	t.Run("deserialize", func(t *testing.T) {
		data := []byte("field: value")
		s, err := yamlx.UnmarshalStrict[testStruct](data)
		require.NoError(t, err)
		requirex.Equal(t, testStruct{Field: "value"}, s)
	})
}

func TestUnmarshalFromFile(t *testing.T) {
	tmpdir := t.TempDir()
	f, err := os.CreateTemp(tmpdir, "yamlx-test")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	})

	t.Run("doesn't fail with unknown field", func(t *testing.T) {
		err = os.WriteFile(f.Name(), []byte("field: value\nunknown: value"), 0644)
		require.NoError(t, err)

		s, err := yamlx.UnmarshalFromFile[testStruct](f.Name())
		require.NoError(t, err)
		requirex.Equal(t, testStruct{Field: "value"}, s)
	})
}

func TestUnmarshalStrictFromFile(t *testing.T) {
	tmpdir := t.TempDir()
	f, err := os.CreateTemp(tmpdir, "yamlx-test")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	})

	t.Run("fails with unknown field", func(t *testing.T) {
		err = os.WriteFile(f.Name(), []byte("field: value\nunknown: value"), 0644)
		require.NoError(t, err)

		_, err := yamlx.UnmarshalStrictFromFile[testStruct](f.Name())
		require.Error(t, err)
	})

	t.Run("deserialize", func(t *testing.T) {
		err = os.WriteFile(f.Name(), []byte("field: value"), 0644)
		require.NoError(t, err)

		s, err := yamlx.UnmarshalStrictFromFile[testStruct](f.Name())
		require.NoError(t, err)
		requirex.Equal(t, testStruct{Field: "value"}, s)
	})
}
