package codycontext

import (
	lru "github.com/hashicorp/golang-lru/v2"
)

type safeCache[K comparable, V any] struct {
	cache *lru.Cache[K, V]
}

func newSafeCache[K comparable, V any](maxSize int) safeCache[K, V] {
	c, _ := lru.New[K, V](maxSize)
	return safeCache[K, V]{cache: c}
}

func (c *safeCache[K, V]) Add(key K, value V) (evicted bool) {
	if c.cache != nil {
		return c.cache.Add(key, value)
	}
	return false
}

func (c *safeCache[K, V]) Get(key K) (V, bool) {
	if c.cache != nil {
		return c.cache.Get(key)

	}
	var empty V
	return empty, false
}

func (c *safeCache[K, V]) Clear() {
	if c.cache != nil {
		c.cache.Purge()
	}
}
