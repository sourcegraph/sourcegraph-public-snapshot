package httpcache

import (
	"context"
	"sync"
)

// NewMemoryCache returns a new Cache that will store items in an in-memory map
func NewMemoryCache() Cache {
	return &MemoryCache{
		items: make(map[string][]byte),
	}
}

// MemoryCache is an implemtation of Cache that stores responses in an in-memory map.
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string][]byte
}

// Get returns the []byte representation of the response and true if present, false if not
func (c *MemoryCache) Get(ctx context.Context, key string) (resp []byte, ok bool) {
	c.mu.RLock()
	resp, ok = c.items[key]
	c.mu.RUnlock()
	return resp, ok
}

// Set saves response resp to the cache with key
func (c *MemoryCache) Set(ctx context.Context, key string, resp []byte) {
	c.mu.Lock()
	c.items[key] = resp
	c.mu.Unlock()
}

// Delete removes key from the cache
func (c *MemoryCache) Delete(ctx context.Context, key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}
