// Package timedcache provides a generic, thread-safe in-memory cache with TTL.
// Elements are automatically evicted exactly when their TTL expires.
package timedcache

import (
	"sync"
	"time"
)

// node is an internal element of the doubly linked list.
type node[K comparable, V any] struct {
	key       K
	value     V
	expiresAt time.Time
	prev      *node[K, V]
	next      *node[K, V]
}

// Cache is a generic cache with TTL. It is safe for concurrent use.
// The list is ordered by expiresAt (head = latest, tail = earliest).
type Cache[K comparable, V any] struct {
	mu               sync.Mutex
	ttl              time.Duration
	items            map[K]*node[K, V]
	head             *node[K, V]
	tail             *node[K, V]
	finalizer        func(key K, value V)
	finalizerWorkers int
	finalizerBuf     int
	finalizeCh       chan *node[K, V]
	stopCh           chan struct{}
	wakeCh           chan struct{}
	stopped          bool
	now              func() time.Time // for testing
}

// Option is a functional option for Cache.
type Option[K comparable, V any] func(*Cache[K, V])

// WithFinalizer sets a callback that is invoked when an entry expires.
// The callback is executed in a separate goroutine pool.
func WithFinalizer[K comparable, V any](fn func(key K, value V)) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.finalizer = fn
	}
}

// WithFinalizerWorkers sets the number of goroutines that process finalizer jobs.
// Default is 4.
func WithFinalizerWorkers[K comparable, V any](n int) Option[K, V] {
	return func(c *Cache[K, V]) {
		if n > 0 {
			c.finalizerWorkers = n
		}
	}
}

// WithFinalizerBuffer sets the size of the finalizer channel buffer.
// Default is 256.
func WithFinalizerBuffer[K comparable, V any](size int) Option[K, V] {
	return func(c *Cache[K, V]) {
		if size > 0 {
			c.finalizerBuf = size
		}
	}
}

// WithNowFunc sets an alternative time source. Useful for testing.
func WithNowFunc[K comparable, V any](fn func() time.Time) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.now = fn
	}
}

// New creates a new Cache with the given TTL and options.
func New[K comparable, V any](ttl time.Duration, opts ...Option[K, V]) *Cache[K, V] {
	c := &Cache[K, V]{
		ttl:              ttl,
		items:            make(map[K]*node[K, V]),
		finalizerWorkers: 4,
		finalizerBuf:     256,
		now:              time.Now,
		stopCh:           make(chan struct{}),
		wakeCh:           make(chan struct{}, 1),
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.finalizer != nil {
		c.finalizeCh = make(chan *node[K, V], c.finalizerBuf)
		for i := 0; i < c.finalizerWorkers; i++ {
			go c.finalizeWorker()
		}
	}
	go c.daemon()
	return c
}

// Get returns the value associated with the key if it exists and is not expired.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	n, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	if c.now().After(n.expiresAt) {
		// already expired, remove now
		c.remove(n)
		var zero V
		return zero, false
	}
	// extend TTL: move to head with new expiresAt
	n.expiresAt = c.now().Add(c.ttl)
	c.moveToHead(n)
	return n.value, true
}

// Set adds or updates a key. The TTL is reset.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	if n, ok := c.items[key]; ok {
		n.value = value
		n.expiresAt = now.Add(c.ttl)
		c.moveToHead(n)
		return
	}
	n := &node[K, V]{
		key:       key,
		value:     value,
		expiresAt: now.Add(c.ttl),
	}
	c.items[key] = n
	if c.head == nil {
		c.head, c.tail = n, n
	} else {
		n.next = c.head
		c.head.prev = n
		c.head = n
	}
	c.wakeUp()
}

// Extend prolongs the TTL of an existing key. Returns false if the key was not found.
func (c *Cache[K, V]) Extend(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	n, ok := c.items[key]
	if !ok {
		return false
	}
	n.expiresAt = c.now().Add(c.ttl)
	c.moveToHead(n)
	return true
}

// Delete removes a key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n, ok := c.items[key]; ok {
		c.remove(n)
	}
}

// Stop gracefully shuts down the cache: stops the daemon and finalizer workers.
func (c *Cache[K, V]) Stop() {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return
	}
	c.stopped = true
	close(c.stopCh)
	if c.finalizeCh != nil {
		close(c.finalizeCh)
	}
	c.mu.Unlock()
}

// remove removes a node from the list and map. Must be called with mu held.
func (c *Cache[K, V]) remove(n *node[K, V]) {
	delete(c.items, n.key)
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		c.head = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	} else {
		c.tail = n.prev
	}
	if c.finalizer != nil {
		c.finalizeCh <- n
	}
}

// moveToHead moves the given node to the head of the list (most recent).
func (c *Cache[K, V]) moveToHead(n *node[K, V]) {
	if c.head == n {
		return // already head
	}
	// unlink
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		c.head = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	} else {
		c.tail = n.prev
	}
	// link to head
	n.prev = nil
	n.next = c.head
	if c.head != nil {
		c.head.prev = n
	}
	c.head = n
	if c.tail == nil {
		c.tail = n
	}
}

// wakeUp sends a non-blocking signal to the daemon.
func (c *Cache[K, V]) wakeUp() {
	select {
	case c.wakeCh <- struct{}{}:
	default:
	}
}

// daemon periodically checks for expired entries.
func (c *Cache[K, V]) daemon() {
	for {
		c.mu.Lock()
		if c.tail == nil {
			c.mu.Unlock()
			select {
			case <-c.stopCh:
				return
			case <-c.wakeCh:
				continue
			}
		}
		// compute sleep until tail expires
		sleep := c.tail.expiresAt.Sub(c.now())
		c.mu.Unlock()
		if sleep > 0 {
			select {
			case <-c.stopCh:
				return
			case <-c.wakeCh:
				// new element added, recalculate
			case <-time.After(sleep):
				// time to purge
			}
		}
		// purge expired
		c.mu.Lock()
		now := c.now()
		for c.tail != nil && !now.Before(c.tail.expiresAt) {
			c.remove(c.tail)
		}
		c.mu.Unlock()
	}
}

// finalizeWorker calls the finalizer on each node received.
func (c *Cache[K, V]) finalizeWorker() {
	for n := range c.finalizeCh {
		if c.finalizer != nil {
			c.finalizer(n.key, n.value)
		}
	}
}
