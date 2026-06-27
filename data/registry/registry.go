// Package registry provides a type‑safe generic handler registry.
package registry

import (
	"context"
	"fmt"
	"sync"
)

// Handler is a function that processes a request and returns a response or an error.
type Handler[Req, Resp any] func(ctx context.Context, req Req) (Resp, error)

// Registry stores handlers by key and dispatches requests.
// K is the key type (must be comparable), Req and Resp are the request/response types.
type Registry[K comparable, Req, Resp any] struct {
	mu       sync.RWMutex
	handlers map[K]Handler[Req, Resp]
}

// New creates a new empty Registry.
func New[K comparable, Req, Resp any]() *Registry[K, Req, Resp] {
	return &Registry[K, Req, Resp]{
		handlers: make(map[K]Handler[Req, Resp]),
	}
}

// Register adds or replaces a handler for the given key.
func (r *Registry[K, Req, Resp]) Register(key K, h Handler[Req, Resp]) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[key] = h
}

// Dispatch calls the handler for the given key with the provided request.
// If no handler is found, an error is returned.
func (r *Registry[K, Req, Resp]) Dispatch(ctx context.Context, key K, req Req) (Resp, error) {
	r.mu.RLock()
	h, ok := r.handlers[key]
	r.mu.RUnlock()
	if !ok {
		var zero Resp
		return zero, fmt.Errorf("registry: handler %v not found", key)
	}
	return h(ctx, req)
}

// Exists returns true if a handler is registered for the given key.
func (r *Registry[K, Req, Resp]) Exists(key K) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.handlers[key]
	return ok
}
