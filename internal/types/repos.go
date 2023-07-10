package types

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type RepoNameIndex map[api.RepoName]struct{}

type ThreadsafeRepoNameCache struct {
	mu    sync.RWMutex
	index RepoNameIndex
}

func NewThreadsafeRepoNameCache(index RepoNameIndex) *ThreadsafeRepoNameCache {
	if index == nil {
		index = make(RepoNameIndex)
	}

	return &ThreadsafeRepoNameCache{
		index: index,
	}
}

func (c *ThreadsafeRepoNameCache) Contains(name api.RepoName) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, ok := c.index[name]
	return ok
}

func (c *ThreadsafeRepoNameCache) Overwrite(index RepoNameIndex) {
	c.mu.Lock()
	c.index = index
	c.mu.Unlock()
}

func (c *ThreadsafeRepoNameCache) Index() RepoNameIndex {
	return c.index
}
