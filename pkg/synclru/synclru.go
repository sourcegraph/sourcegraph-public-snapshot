// Package synclru provides a synchronized LRU cache.
package synclru

import (
	"sync"

	"github.com/golang/groupcache/lru"
)

type Cache interface {
	Get(key lru.Key) (value interface{}, ok bool)
	Add(key lru.Key, value interface{})
}

// New returns a wrapper around cache that synchronizes access to it.
func New(c Cache) Cache {
	return &cache{cache: c}
}

type cache struct {
	mu    sync.Mutex
	cache Cache
}

func (c *cache) Get(key lru.Key) (value interface{}, ok bool) {
	c.mu.Lock()
	value, ok = c.cache.Get(key)
	c.mu.Unlock()
	return
}

func (c *cache) Add(key lru.Key, value interface{}) {
	c.mu.Lock()
	c.cache.Add(key, value)
	c.mu.Unlock()
}
