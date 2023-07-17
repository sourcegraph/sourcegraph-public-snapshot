package types

import (
	"sync"
)

type repoURIIndex map[string]struct{}

type RepoURICache struct {
	mu    sync.RWMutex
	index repoURIIndex
}

func NewRepoURICache(index repoURIIndex) *RepoURICache {
	if index == nil {
		index = make(repoURIIndex)
	}

	return &RepoURICache{
		index: index,
	}
}

func (c *RepoURICache) Contains(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, ok := c.index[name]
	return ok
}

func (c *RepoURICache) Overwrite(index repoURIIndex) {
	c.mu.Lock()
	c.index = index
	c.mu.Unlock()
}

func (c *RepoURICache) Index() repoURIIndex {
	return c.index
}
