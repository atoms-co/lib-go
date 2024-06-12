package container_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.atoms.co/lib/testing/assertx"
	"go.atoms.co/lib/container"
)

func TestRingBuf(t *testing.T) {

	t.Run("basic", func(t *testing.T) {
		buf := container.NewRingBuf[int](3)
		assert.True(t, buf.IsEmpty())
		assertx.Equal(t, buf.Len(), 0)

		_, ok := buf.Peek()
		assert.False(t, ok)

		// (1) Enqueue [1,2,3]

		assert.True(t, buf.Enqueue(1))
		assertx.Equal(t, buf.Len(), 1)
		assert.True(t, buf.Enqueue(2))
		assertx.Equal(t, buf.Len(), 2)
		assert.True(t, buf.Enqueue(3))
		assertx.Equal(t, buf.Len(), 3)

		assert.False(t, buf.Enqueue(4))
		assert.False(t, buf.IsEmpty())
		assertx.Equal(t, buf.Len(), 3)

		// (2) Dequeue, enqueue, dequeue. Wraparound is handled.

		one, ok := buf.Dequeue()
		assert.True(t, ok)
		assertx.Equal(t, one, 1)
		assertx.Equal(t, buf.Len(), 2)

		peek, ok := buf.Peek()
		assert.True(t, ok)
		assertx.Equal(t, peek, 2)

		assert.True(t, buf.Enqueue(5))
		assert.False(t, buf.Enqueue(6))
		assertx.Equal(t, buf.Len(), 3)

		two, ok := buf.Dequeue()
		assert.True(t, ok)
		assertx.Equal(t, two, 2)
		assertx.Equal(t, buf.Len(), 2)

		// (3) Empty queue

		three, ok := buf.Dequeue()
		assert.True(t, ok)
		assertx.Equal(t, three, 3)
		assertx.Equal(t, buf.Len(), 1)

		five, ok := buf.Dequeue()
		assert.True(t, ok)
		assertx.Equal(t, five, 5)

		assert.True(t, buf.IsEmpty())
		_, ok = buf.Dequeue()
		assert.False(t, ok)
	})

	t.Run("zero", func(t *testing.T) {
		buf := container.NewRingBuf[int](0)
		assert.True(t, buf.IsEmpty())

		assert.False(t, buf.Enqueue(1))
		_, ok := buf.Dequeue()
		assert.False(t, ok)
	})

}
