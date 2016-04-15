package cache

import (
	"time"

	"github.com/golang/groupcache/lru"
)

// TTL returns a wrapper around cache that will expire entries. Note it does
// not evict the entry from the underlying cache, rather it just won't return
// it if the entry has expired.
func TTL(cache Cache, ttl time.Duration) Cache {
	return &ttlCache{cache: cache, ttl: ttl}
}

type ttlCache struct {
	cache Cache
	ttl   time.Duration
}

type ttlEntry struct {
	value    interface{}
	deadline time.Time
}

func (c *ttlCache) Get(key lru.Key) (interface{}, bool) {
	e, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	entry := e.(ttlEntry)
	if time.Now().After(entry.deadline) {
		return nil, false
	}
	return entry.value, true
}

func (c *ttlCache) Add(key lru.Key, value interface{}) {
	c.cache.Add(key, ttlEntry{value: value, deadline: time.Now().Add(c.ttl)})
}
