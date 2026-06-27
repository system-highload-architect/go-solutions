// Package sampler provides a simple probabilistic sampler.
package sampler

import (
	"math/rand"
	"sync"
)

// Sampler randomly returns true with a configured probability.
// It is safe for concurrent use.
type Sampler struct {
	mu   sync.Mutex
	rng  *rand.Rand
	rate float64
}

// NewSampler creates a new Sampler with the given rate (0.0 – 1.0).
func NewSampler(rate float64) *Sampler {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	return &Sampler{
		rng:  rand.New(rand.NewSource(rand.Int63())),
		rate: rate,
	}
}

// Sample returns true with the configured probability.
func (s *Sampler) Sample() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rng.Float64() < s.rate
}
