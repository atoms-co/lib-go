package container

import (
	"fmt"
	"maps"
	"time"
)

// node is a doubly-linked cache node. Next is newer.
type node[K, V any] struct {
	key        K
	value      V
	size       int
	inserted   time.Time
	prev, next *node[K, V]
}

// Cache is a size- and time-restricted FIFO cache with support for manual eviction. Not thread-safe.
type Cache[K comparable, V any] struct {
	elements map[K]*node[K, V]
	sentinel *node[K, V]
	sizeFn   func(K, V) int

	limit, size, count int
	duration           time.Duration
}

func NewCache[K comparable, V any](limit int, duration time.Duration, sizeFn func(K, V) int) *Cache[K, V] {
	ret := &Cache[K, V]{
		sizeFn:   sizeFn,
		elements: map[K]*node[K, V]{},
		sentinel: &node[K, V]{},
		limit:    limit,
		duration: duration,
	}
	ret.sentinel.prev = ret.sentinel
	ret.sentinel.next = ret.sentinel

	return ret
}

// Size returns the total size of the cached keys and values, usually in bytes.
func (c *Cache[K, V]) Size() int {
	return c.size
}

// Len returns the number of cached elements.
func (c *Cache[K, V]) Len() int {
	return c.count
}

// Head returns the oldest cached element, if any.
func (c *Cache[K, V]) Head() (K, V, bool) {
	if c.size == 0 {
		var k K
		var v V
		return k, v, false
	}

	head := c.sentinel.next
	return head.key, head.value, true
}

// Find returns the value with the given key, if present.
func (c *Cache[K, V]) Find(key K) (V, bool) {
	n, ok := c.elements[key]
	if !ok {
		var v V
		return v, false
	}
	return n.value, true
}

// Add adds an element to the cache, if not already present. It may evict elements to satisfy cache constraints.
// To update the insertion timestamp, Remove then Add.
func (c *Cache[K, V]) Add(key K, value V, now time.Time) {
	if _, ok := c.elements[key]; ok {
		return
	}

	fresh := &node[K, V]{
		key:      key,
		value:    value,
		size:     c.sizeFn(key, value),
		inserted: now,
		prev:     c.sentinel.prev,
		next:     c.sentinel,
	}
	c.sentinel.prev = fresh
	fresh.prev.next = fresh

	c.elements[key] = fresh
	c.size += fresh.size
	c.count++

	c.Trim(now)
}

// Remove removes an element to the cache, if present.
func (c *Cache[K, V]) Remove(key K) {
	elm, ok := c.elements[key]
	if !ok {
		return
	}

	elm.prev.next = elm.next
	elm.next.prev = elm.prev

	delete(c.elements, key)
	c.size -= elm.size
	c.count--
}

// Trim removes the oldest elements to fit the size and duration limits.
func (c *Cache[K, V]) Trim(now time.Time) {
	cutoff := now.Add(-c.duration)

	for c.count > 0 {
		head := c.sentinel.next
		if c.size < c.limit && head.inserted.After(cutoff) {
			break // ok
		}
		c.Remove(head.key)
	}
}

// Compact reallocates the internal maps due to potentially high turnover leaking memory. Called infrequently.
func (c *Cache[K, V]) Compact() {
	c.elements = maps.Clone(c.elements)
}

func (c *Cache[K, V]) String() string {
	return fmt.Sprintf("{size=%v/%v (%v%%), len=%v}", c.size, c.limit, (c.size*100)/c.limit, c.count)
}
