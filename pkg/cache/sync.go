package cache

import (
	"sync"

	"github.com/golang/groupcache/lru"
)

// Sync returns a wrapper around cache that synchronizes access to it.
func Sync(c Cache) Cache {
	return &syncedCache{cache: c}
}

type syncedCache struct {
	mu    sync.Mutex
	cache Cache
}

func (c *syncedCache) Get(key lru.Key) (value interface{}, ok bool) {
	c.mu.Lock()
	value, ok = c.cache.Get(key)
	c.mu.Unlock()
	return
}

func (c *syncedCache) Add(key lru.Key, value interface{}) {
	c.mu.Lock()
	c.cache.Add(key, value)
	c.mu.Unlock()
}
