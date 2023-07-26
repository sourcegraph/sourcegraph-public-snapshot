package types

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/collections"
)

type RepoURISet struct {
	mu    sync.RWMutex
	index collections.Set[string]
}

func NewRepoURICache(index collections.Set[string]) *RepoURISet {
	if index == nil {
		index = make(collections.Set[string])
	}

	return &RepoURISet{
		index: index,
	}
}

func (c *RepoURISet) Contains(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, ok := c.index[name]
	return ok
}

func (c *RepoURISet) Overwrite(index collections.Set[string]) {
	c.mu.Lock()
	c.index = index
	c.mu.Unlock()
}
