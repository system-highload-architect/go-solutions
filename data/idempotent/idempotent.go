package idempotent

import (
	"time"

	"github.com/system-highload-architect/go-solutions/data/timedcache"
)

// Store remembers keys of any comparable type K for a configurable TTL.
type Store[K comparable] struct {
	cache *timedcache.Cache[K, struct{}]
}

// NewStore creates a new Store with the given TTL.
func NewStore[K comparable](ttl time.Duration) *Store[K] {
	return &Store[K]{cache: timedcache.New[K, struct{}](ttl)}
}

// Check returns true if the key is new (not seen within TTL).
// It remembers the key for future calls.
func (s *Store[K]) Check(key K) bool {
	if _, exists := s.cache.Get(key); exists {
		return false
	}
	s.cache.Set(key, struct{}{})
	return true
}

// Stop gracefully stops the underlying cache.
func (s *Store[K]) Stop() {
	s.cache.Stop()
}
