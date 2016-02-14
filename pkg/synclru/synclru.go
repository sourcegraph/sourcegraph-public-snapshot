// Package synclru provides a synchronized LRU cache.
package synclru

import (
	"sync"

	"github.com/golang/groupcache/lru"
)

// New returns a wrapper around lru that synchronizes access to it.
func New(lru *lru.Cache) *Cache {
	return &Cache{lru: lru}
}

type Cache struct {
	mu  sync.Mutex
	lru *lru.Cache
}

func (c *Cache) Get(key lru.Key) (value interface{}, ok bool) {
	c.mu.Lock()
	value, ok = c.lru.Get(key)
	c.mu.Unlock()
	return
}

func (c *Cache) Add(key lru.Key, value interface{}) {
	c.mu.Lock()
	c.lru.Add(key, value)
	c.mu.Unlock()
}

// Clear removes all elements.
func (c *Cache) Clear() {
	c.mu.Lock()
	for c.lru.Len() > 0 {
		c.lru.RemoveOldest()
	}
	c.mu.Unlock()
}
