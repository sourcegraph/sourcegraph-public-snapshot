package types

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/collections"
)

type RepoURICache struct {
	mu    sync.RWMutex
	index collections.Set[string]
}

func NewRepoURICache(index collections.Set[string]) *RepoURICache {
	if index == nil {
		index = make(collections.Set[string])
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

func (c *RepoURICache) Overwrite(index collections.Set[string]) {
	c.mu.Lock()
	c.index = index
	c.mu.Unlock()
}

func (c *RepoURICache) Index() collections.Set[string] {
	return c.index
}
