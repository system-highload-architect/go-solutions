// Package ratelimit provides a collection of rate‑limiting algorithms
// suitable for high‑throughput systems. All implementations are safe for
// concurrent use.
package ratelimit

import (
	"sync"
	"time"
)

// ----------------------------------------------------------------------------
// Token Bucket
// ----------------------------------------------------------------------------

// TokenBucket implements the classic token bucket algorithm.
// Tokens are added at a fixed rate up to a maximum burst.
type TokenBucket struct {
	rate   float64 // tokens per second
	burst  float64
	tokens float64
	last   time.Time
	mu     sync.Mutex
}

// NewTokenBucket creates a new TokenBucket with the given rate (tokens/sec)
// and burst size.
func NewTokenBucket(rate, burst float64) *TokenBucket {
	return &TokenBucket{
		rate:  rate,
		burst: burst,
		last:  time.Now(),
	}
}

// Allow consumes a single token. Returns true if the request is allowed,
// false otherwise.
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

// AllowN attempts to consume n tokens. It returns true if enough tokens
// are available, false otherwise.
func (tb *TokenBucket) AllowN(n float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.last).Seconds()
	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.burst {
		tb.tokens = tb.burst
	}
	tb.last = now

	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}
	return false
}

// ----------------------------------------------------------------------------
// Leaky Bucket
// ----------------------------------------------------------------------------

// LeakyBucket implements a leaky bucket rate limiter.
// Requests are added to the bucket and processed at a fixed rate.
// If the bucket is full, the request is rejected.
type LeakyBucket struct {
	rate     float64 // requests per second
	capacity float64
	water    float64
	last     time.Time
	mu       sync.Mutex
}

// NewLeakyBucket creates a new LeakyBucket.
//
//	rate: requests per second that leak out
//	capacity: maximum burst the bucket can hold
func NewLeakyBucket(rate, capacity float64) *LeakyBucket {
	return &LeakyBucket{
		rate:     rate,
		capacity: capacity,
		last:     time.Now(),
	}
}

// Allow returns true if the request can be added to the bucket without
// exceeding capacity.
func (lb *LeakyBucket) Allow() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(lb.last).Seconds()
	lb.water -= elapsed * lb.rate
	if lb.water < 0 {
		lb.water = 0
	}
	lb.last = now

	if lb.water < lb.capacity {
		lb.water++
		return true
	}
	return false
}

// ----------------------------------------------------------------------------
// Sliding Window Log
// ----------------------------------------------------------------------------

// SlidingWindowLog keeps a precise log of recent request timestamps and
// limits the number of requests within a rolling window.
type SlidingWindowLog struct {
	limit  int
	window time.Duration
	times  []time.Time
	mu     sync.Mutex
}

// NewSlidingWindowLog creates a new log‑based sliding window.
//
//	limit: maximum number of requests allowed in the window
//	window: duration of the sliding window
func NewSlidingWindowLog(limit int, window time.Duration) *SlidingWindowLog {
	return &SlidingWindowLog{
		limit:  limit,
		window: window,
		times:  make([]time.Time, 0, limit),
	}
}

// Allow returns true if the request is within the limit for the current
// sliding window.
func (sw *SlidingWindowLog) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	// prune old entries
	idx := 0
	for idx < len(sw.times) && sw.times[idx].Before(cutoff) {
		idx++
	}
	sw.times = sw.times[idx:]

	if len(sw.times) < sw.limit {
		sw.times = append(sw.times, now)
		return true
	}
	return false
}

// ----------------------------------------------------------------------------
// Sliding Window Counter
// ----------------------------------------------------------------------------

// SlidingWindowCounter uses a probabilistic counter that divides the window
// into smaller buckets. It is more memory‑efficient than the log‑based version.
type SlidingWindowCounter struct {
	limit    int
	window   time.Duration
	buckets  []int
	lastIdx  int
	lastTime time.Time
	mu       sync.Mutex
}

// NewSlidingWindowCounter creates a counter‑based sliding window.
//
//	limit: maximum requests in the full window
//	window: total window duration
//	slots: number of sub‑buckets (e.g., 10 for a 1‑second window gives 100ms resolution)
func NewSlidingWindowCounter(limit int, window time.Duration, slots int) *SlidingWindowCounter {
	if slots <= 0 {
		slots = 1
	}
	return &SlidingWindowCounter{
		limit:    limit,
		window:   window,
		buckets:  make([]int, slots),
		lastTime: time.Now(),
	}
}

// Allow returns true if the request is allowed.
func (swc *SlidingWindowCounter) Allow() bool {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(swc.lastTime)
	slotDuration := swc.window / time.Duration(len(swc.buckets))
	slotsPassed := int(elapsed / slotDuration)
	if slotsPassed > len(swc.buckets) {
		slotsPassed = len(swc.buckets)
	}

	// clear expired slots
	for i := 1; i <= slotsPassed; i++ {
		idx := (swc.lastIdx + i) % len(swc.buckets)
		swc.buckets[idx] = 0
	}
	if slotsPassed > 0 {
		swc.lastIdx = (swc.lastIdx + slotsPassed) % len(swc.buckets)
		swc.lastTime = now
	}

	// count current requests in the window
	total := 0
	for _, v := range swc.buckets {
		total += v
	}
	if total >= swc.limit {
		return false
	}
	swc.buckets[swc.lastIdx]++
	return true
}

// ----------------------------------------------------------------------------
// Fixed Window (for completeness, not recommended for production)
// ----------------------------------------------------------------------------

// FixedWindow limits requests within a fixed time window (e.g., 100 req/min).
// Note: This algorithm suffers from boundary‑burst issues.
type FixedWindow struct {
	limit     int
	window    time.Duration
	counter   int
	resetTime time.Time
	mu        sync.Mutex
}

// NewFixedWindow creates a new FixedWindow limiter.
func NewFixedWindow(limit int, window time.Duration) *FixedWindow {
	return &FixedWindow{
		limit:     limit,
		window:    window,
		resetTime: time.Now().Add(window),
	}
}

// Allow returns true if the request is within the limit for the current window.
func (fw *FixedWindow) Allow() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now()
	if now.After(fw.resetTime) {
		fw.counter = 0
		fw.resetTime = now.Add(fw.window)
	}
	if fw.counter < fw.limit {
		fw.counter++
		return true
	}
	return false
}

// ----------------------------------------------------------------------------
// Generic Limiter (helper for per‑key limiting using any of the above)
// ----------------------------------------------------------------------------

// Limiter wraps any rate limiter that supports the Allow method.
type Limiter interface {
	Allow() bool
}

// KeyedLimiter provides a simple way to maintain per‑key limiters.
type KeyedLimiter struct {
	mu       sync.Mutex
	factory  func() Limiter
	limiters map[string]Limiter
}

// NewKeyedLimiter creates a keyed limiter. factory is called to create a new
// limiter for each key.
func NewKeyedLimiter(factory func() Limiter) *KeyedLimiter {
	return &KeyedLimiter{
		factory:  factory,
		limiters: make(map[string]Limiter),
	}
}

// Allow reports whether the operation is allowed for the given key.
func (kl *KeyedLimiter) Allow(key string) bool {
	kl.mu.Lock()
	lim, ok := kl.limiters[key]
	if !ok {
		lim = kl.factory()
		kl.limiters[key] = lim
	}
	kl.mu.Unlock()
	return lim.Allow()
}
