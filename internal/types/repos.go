package types

import (
	"sync"
)

// TODO: Design decision needed: Which is better between caching the repo name vs the repo URI? The
// repo name is the name of the repo as it appears on the codehost and the repo URI is the name as
// it appears on the Sourcegraph instance, which may be different if a repositoryPathPattern is
// being used.
//
// We want to ensure we do lookups against repoURI always. But not yet sure if we need to use
// repoName instead in any scenario. Probably not, the more I write this comment the clearer it
// gets. But I'm going to hold on to this TODO comment until I've tested the code with a
// repositoryPathPattern.
//
// FIXME: This TODO comment should not be part of the merged PR.
//
// UPDATE: We want to store repo URI here because users will enter the name of the repos as they appear on the code host.
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
