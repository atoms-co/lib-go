package syncx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type s struct {
	id int
}

func TestGenericMap_Simple(t *testing.T) {
	m := Map[int, *s]{}

	// (1) Store
	m.Store(100, &s{id: 1})

	value, loaded := m.Load(100)
	assert.Equal(t, true, loaded)
	assert.Equal(t, 1, value.id)

	// (2) Delete
	m.Delete(100)
	value, loaded = m.Load(100)
	assert.Equal(t, false, loaded)
	assert.Nil(t, value)
}

func TestGenericMap_Advanced(t *testing.T) {
	m := Map[string, *s]{}

	// (1) Load for empty map
	value, loaded := m.Load("a")
	assert.Equal(t, false, loaded)
	assert.Nil(t, value)

	// (2) LoadOrStore for empty map
	value, loaded = m.LoadOrStore("a", &s{id: 1})
	assert.Equal(t, false, loaded)
	assert.Equal(t, 1, value.id)

	value, loaded = m.LoadOrStore("a", &s{id: 2})
	assert.Equal(t, true, loaded)
	assert.Equal(t, 1, value.id) // unchanged

	// (3) Load for non-empty map
	value, loaded = m.Load("a")
	assert.Equal(t, true, loaded)
	assert.Equal(t, 1, value.id)

	// (4) LoadAndDelete for non-empty map
	value, loaded = m.LoadAndDelete("a")
	assert.Equal(t, true, loaded)
	assert.Equal(t, 1, value.id)

	value, loaded = m.Load("a")
	assert.Equal(t, false, loaded)
	assert.Nil(t, value)

	// (5) Deletes for non-existing keys
	m.Delete("non-existing")
	value, loaded = m.LoadAndDelete("non-existing")
	assert.Equal(t, false, loaded)
	assert.Nil(t, value)
}

func TestGenericMapRange(t *testing.T) {
	m := Map[string, *s]{}

	m.LoadOrStore("a", &s{id: 1})
	m.LoadOrStore("b", &s{id: 2})
	m.LoadOrStore("c", &s{id: 3})

	foundA := false
	foundB := false
	foundC := false
	m.Range(func(key string, value *s) bool {
		switch key {
		case "a":
			assert.Equal(t, 1, value.id)
			foundA = true
		case "b":
			assert.Equal(t, 2, value.id)
			foundB = true
		case "c":
			assert.Equal(t, 3, value.id)
			foundC = true
		default:
			assert.Fail(t, "unexpected key")
		}
		return true
	})

	assert.True(t, foundA)
	assert.True(t, foundB)
	assert.True(t, foundC)
}
