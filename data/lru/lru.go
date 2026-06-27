// Package lru provides a generic, thread-safe LRU cache with fixed capacity.
package lru

import (
	"container/list"
	"sync"
)

// EvictCallback is called when an entry is evicted from the cache.
type EvictCallback[K comparable, V any] func(key K, value V)

// Cache is a generic LRU cache. It is safe for concurrent use.
// K must be comparable, V can be any type.
type Cache[K comparable, V any] struct {
	mu       sync.RWMutex
	capacity int
	items    map[K]*list.Element
	order    *list.List // front = most recently used, back = least recently used
	onEvict  EvictCallback[K, V]
}

// entry holds a key-value pair stored in the list.
type entry[K comparable, V any] struct {
	key   K
	value V
}

// Option is a functional option for Cache.
type Option[K comparable, V any] func(*Cache[K, V])

// WithEvictCallback sets a callback that is invoked when an entry is evicted.
func WithEvictCallback[K comparable, V any](cb EvictCallback[K, V]) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.onEvict = cb
	}
}

// New creates a new LRU cache with the given capacity and options.
// Capacity must be positive.
func New[K comparable, V any](capacity int, opts ...Option[K, V]) *Cache[K, V] {
	if capacity < 1 {
		capacity = 1
	}
	c := &Cache[K, V]{
		capacity: capacity,
		items:    make(map[K]*list.Element, capacity),
		order:    list.New(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Get returns the value associated with the key, and a boolean indicating
// whether it was found. If found, the key is moved to the front (most recent).
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	c.order.MoveToFront(elem)
	ent := elem.Value.(*entry[K, V])
	return ent.value, true
}

// Set adds or updates a key. If the cache is at capacity, the least recently
// used item is evicted. If the key already exists, it is moved to the front
// and its value is updated.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		ent := elem.Value.(*entry[K, V])
		ent.value = value
		return
	}

	// Evict the oldest if at capacity
	if c.order.Len() >= c.capacity {
		c.evictOldest()
	}

	ent := &entry[K, V]{key: key, value: value}
	elem := c.order.PushFront(ent)
	c.items[key] = elem
}

// Extend moves the key to the front if it exists, effectively extending its
// "recency". Returns true if the key existed, false otherwise.
func (c *Cache[K, V]) Extend(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return false
	}
	c.order.MoveToFront(elem)
	return true
}

// Delete removes a key from the cache. If the key does not exist, it is a no-op.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return
	}
	c.removeElement(elem)
}

// Len returns the current number of items in the cache.
func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.order.Len()
}

// evictOldest removes the least recently used item. Must be called with mu held.
func (c *Cache[K, V]) evictOldest() {
	elem := c.order.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

// removeElement removes a specific list element. Must be called with mu held.
func (c *Cache[K, V]) removeElement(elem *list.Element) {
	ent := elem.Value.(*entry[K, V])
	delete(c.items, ent.key)
	c.order.Remove(elem)
	if c.onEvict != nil {
		c.onEvict(ent.key, ent.value)
	}
}
