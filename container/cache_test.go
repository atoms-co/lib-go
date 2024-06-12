package container_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.cloudkitchens.org/lib/testing/assertx"
	"go.cloudkitchens.org/lib/container"
)

func TestCache(t *testing.T) {
	now := time.Now()

	t.Run("basic", func(t *testing.T) {
		c := container.NewCache[int, int](100, 10*time.Second, func(k int, v int) int {
			return v
		})

		assertx.Equal(t, c.Len(), 0)
		assertx.Equal(t, c.Size(), 0)
		_, _, ok := c.Head()
		assert.False(t, ok)

		// (1) Add and find elements

		c.Add(1, 10, now)
		c.Add(2, 20, now.Add(time.Second))

		assertx.Equal(t, c.Len(), 2)
		assertx.Equal(t, c.Size(), 30)
		k, v, ok := c.Head()
		assert.True(t, ok)
		assertx.Equal(t, k, 1)
		assertx.Equal(t, v, 10)

		v, ok = c.Find(1)
		assert.True(t, ok)
		assertx.Equal(t, v, 10)

		v, ok = c.Find(2)
		assert.True(t, ok)
		assertx.Equal(t, v, 20)

		_, ok = c.Find(3)
		assert.False(t, ok)

		// (2) Remove the first

		c.Remove(1)

		assertx.Equal(t, c.Len(), 1)
		assertx.Equal(t, c.Size(), 20)
		k, v, ok = c.Head()
		assert.True(t, ok)
		assertx.Equal(t, k, 2)
		assertx.Equal(t, v, 20)
	})

	t.Run("trim", func(t *testing.T) {
		c := container.NewCache[int, int](100, 10*time.Second, func(k int, v int) int {
			return v
		})

		// (1) Cache is trimmed to fit size

		c.Add(1, 10, now.Add(time.Second))
		c.Add(2, 20, now.Add(2*time.Second))
		c.Add(3, 30, now.Add(4*time.Second))
		c.Add(4, 40, now.Add(6*time.Second))
		c.Add(5, 50, now.Add(8*time.Second))

		assertx.Equal(t, c.Len(), 2)
		assertx.Equal(t, c.Size(), 90)

		_, ok := c.Find(2)
		assert.False(t, ok)
		k, v, ok := c.Head()
		assert.True(t, ok)
		assertx.Equal(t, k, 4)
		assertx.Equal(t, v, 40)

		// (3) Trim removes element 2 after 10+6s

		c.Trim(now)
		assertx.Equal(t, c.Len(), 2)

		c.Trim(now.Add(15 * time.Second))
		assertx.Equal(t, c.Len(), 2)

		c.Trim(now.Add(17 * time.Second))
		assertx.Equal(t, c.Len(), 1)
		assertx.Equal(t, c.Size(), 50)

		// (4) Time may empty the cache based on time alone

		c.Trim(now.Add(time.Minute))
		assertx.Equal(t, c.Len(), 0)
		assertx.Equal(t, c.Size(), 0)
	})
}
