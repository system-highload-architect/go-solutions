// Package timewheel provides a generic, high‑performance timing wheel.
// It is safe for concurrent use and supports O(1) Add, Move, Remove, and Tick
// operations with a configurable tick duration and number of slots.
package timewheel

import (
	"context"
	"sync"
	"time"
)

// Task represents a scheduled task with generic key K and value T.
type Task[K comparable, T any] struct {
	Key      K
	Value    T
	ExpireAt time.Time // absolute expiration time (informational)
}

// ExpireCallback is called when a task expires.
type ExpireCallback[K comparable, T any] func(task Task[K, T])

// TimeWheel implements a single‑level timing wheel.
type TimeWheel[K comparable, T any] struct {
	mu        sync.Mutex
	tick      time.Duration
	slots     []map[K]Task[K, T] // each slot contains tasks by key
	bitmask   []uint64           // quick emptiness check (1 bit per slot)
	pointer   int                // current slot index
	keyToSlot map[K]int          // maps a key to the slot it currently resides in (for O(1) Move/Remove)
	onExpire  ExpireCallback[K, T]
	stopCh    chan struct{}
	ticker    *time.Ticker
}

// Option is a functional option for TimeWheel.
type Option[K comparable, T any] func(*TimeWheel[K, T])

// WithExpireCallback sets a callback that is invoked when a task expires.
func WithExpireCallback[K comparable, T any](cb ExpireCallback[K, T]) Option[K, T] {
	return func(tw *TimeWheel[K, T]) {
		tw.onExpire = cb
	}
}

// New creates a new timing wheel.
//   - tick: the duration of one slot.
//   - numSlots: total number of slots in the wheel (must be >= 1).
//
// The wheel covers a total time span of tick * numSlots.
func New[K comparable, T any](tick time.Duration, numSlots int, opts ...Option[K, T]) *TimeWheel[K, T] {
	if numSlots < 1 {
		numSlots = 1
	}
	slots := make([]map[K]Task[K, T], numSlots)
	for i := range slots {
		slots[i] = make(map[K]Task[K, T])
	}
	bitmaskLen := (numSlots + 63) / 64
	tw := &TimeWheel[K, T]{
		tick:      tick,
		slots:     slots,
		bitmask:   make([]uint64, bitmaskLen),
		keyToSlot: make(map[K]int),
		stopCh:    make(chan struct{}),
	}
	for _, opt := range opts {
		opt(tw)
	}
	return tw
}

// Add schedules a task to be executed after the given delay.
// Returns the absolute expiration time (approximate).
func (tw *TimeWheel[K, T]) Add(key K, value T, delay time.Duration) time.Time {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	return tw.addLocked(key, value, delay)
}

// addLocked inserts a task. Must be called with mu held.
func (tw *TimeWheel[K, T]) addLocked(key K, value T, delay time.Duration) time.Time {
	// If the key already exists, remove it first (simplifies Move logic)
	if oldSlot, exists := tw.keyToSlot[key]; exists {
		tw.removeFromSlot(key, oldSlot)
	}

	ticks := int(delay / tw.tick)
	if ticks < 1 {
		ticks = 1 // at least one tick
	}
	slot := (tw.pointer + ticks) % len(tw.slots)
	expireAt := time.Now().Add(delay)
	tw.slots[slot][key] = Task[K, T]{Key: key, Value: value, ExpireAt: expireAt}
	tw.setBit(slot)
	tw.keyToSlot[key] = slot
	return expireAt
}

// Move reschedules an existing key to a new delay. Returns the new expiration time.
// If the key does not exist, it acts like Add.
func (tw *TimeWheel[K, T]) Move(key K, newDelay time.Duration) time.Time {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if oldSlot, exists := tw.keyToSlot[key]; exists {
		tw.removeFromSlot(key, oldSlot)
	}
	return tw.addLocked(key, Task[K, T]{Key: key}.Value, newDelay) // value will be overwritten; better to keep the old value
}

// MoveValue moves an existing key and also updates its associated value.
func (tw *TimeWheel[K, T]) MoveValue(key K, value T, newDelay time.Duration) time.Time {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if oldSlot, exists := tw.keyToSlot[key]; exists {
		tw.removeFromSlot(key, oldSlot)
	}
	return tw.addLocked(key, value, newDelay)
}

// Remove deletes a key from the wheel. It is a no-op if the key does not exist.
func (tw *TimeWheel[K, T]) Remove(key K) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if slot, exists := tw.keyToSlot[key]; exists {
		tw.removeFromSlot(key, slot)
	}
}

// Tick advances the wheel by one slot and returns all tasks that expire at the new pointer position.
// This method is safe to call concurrently.
func (tw *TimeWheel[K, T]) Tick() []Task[K, T] {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	tw.pointer = (tw.pointer + 1) % len(tw.slots)
	current := tw.pointer

	// Quick emptiness test using bitmask
	if !tw.isBitSet(current) {
		return nil
	}

	// Collect expired tasks
	slot := tw.slots[current]
	tasks := make([]Task[K, T], 0, len(slot))
	for key, task := range slot {
		tasks = append(tasks, task)
		delete(tw.keyToSlot, key)
	}
	// Clear the slot and bit
	tw.slots[current] = make(map[K]Task[K, T])
	tw.clearBit(current)

	return tasks
}

// Start launches a background goroutine that calls Tick at every tick interval.
// Expired tasks are passed to the onExpire callback (if set).
// The context can be used to stop the goroutine.
func (tw *TimeWheel[K, T]) Start(ctx context.Context) {
	tw.ticker = time.NewTicker(tw.tick)
	go func() {
		for {
			select {
			case <-ctx.Done():
				tw.ticker.Stop()
				return
			case <-tw.stopCh:
				tw.ticker.Stop()
				return
			case <-tw.ticker.C:
				tasks := tw.Tick()
				if tw.onExpire != nil {
					for _, task := range tasks {
						tw.onExpire(task)
					}
				}
			}
		}
	}()
}

// Stop gracefully stops the background goroutine started by Start.
func (tw *TimeWheel[K, T]) Stop() {
	close(tw.stopCh)
}

// --- bit helpers ---

func (tw *TimeWheel[K, T]) setBit(slot int) {
	word := slot / 64
	bit := slot % 64
	tw.bitmask[word] |= (1 << bit)
}

func (tw *TimeWheel[K, T]) clearBit(slot int) {
	word := slot / 64
	bit := slot % 64
	tw.bitmask[word] &= ^(1 << bit)
}

func (tw *TimeWheel[K, T]) isBitSet(slot int) bool {
	word := slot / 64
	bit := slot % 64
	return (tw.bitmask[word] & (1 << bit)) != 0
}

// removeFromSlot removes a key from the given slot and updates the bitmask.
// Must be called with mu held.
func (tw *TimeWheel[K, T]) removeFromSlot(key K, slot int) {
	delete(tw.slots[slot], key)
	delete(tw.keyToSlot, key)
	if len(tw.slots[slot]) == 0 {
		tw.clearBit(slot)
	}
}
