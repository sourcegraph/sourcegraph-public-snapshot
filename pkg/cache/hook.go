package cache

import "github.com/golang/groupcache/lru"

// Hook returns a wrapper around cache that calls hit or miss on cache hits or
// misses respectively.
func Hook(c Cache, hit, miss func()) Cache {
	if hit == nil {
		hit = func() {}
	}
	if miss == nil {
		miss = func() {}
	}
	return &hookCache{cache: c, hit: hit, miss: miss}
}

type hookCache struct {
	cache     Cache
	hit, miss func()
}

func (c *hookCache) Get(key lru.Key) (value interface{}, ok bool) {
	value, ok = c.cache.Get(key)
	if ok {
		c.hit()
	} else {
		c.miss()
	}
	return
}

func (c *hookCache) Add(key lru.Key, value interface{}) {
	c.cache.Add(key, value)
}
