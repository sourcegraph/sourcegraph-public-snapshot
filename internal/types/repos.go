package types

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/collections"
)

type RepoURISet struct {
	mu    sync.RWMutex
	index collections.Set[string]
}

func NewRepoURISet(index collections.Set[string]) *RepoURISet {
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

// NewEmptyRepoURISet is syntactical sugar to initialise a RepoURISet with a nil set used in tests.
func NewEmptyRepoURISet() *RepoURISet { return NewRepoURISet(nil) }
