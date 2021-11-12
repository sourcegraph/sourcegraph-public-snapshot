package resolvers

import (
	"context"
	"sync"

	"github.com/cockroachdb/errors"
)

type cachedCommitChecker struct {
	gitserverClient GitserverClient
	mutex           sync.RWMutex
	cache           map[int]map[string]bool
}

func newCachedCommitChecker(gitserverClient GitserverClient) *cachedCommitChecker {
	return &cachedCommitChecker{
		gitserverClient: gitserverClient,
		cache:           map[int]map[string]bool{},
	}
}

// set marks the given repository and commit as valid and resolvable by gitserver.
func (c *cachedCommitChecker) set(repositoryID int, commit string) {
	c.setValue(repositoryID, commit, true)
}

// exists determines if the given commit is resolvable for the given repository. If
// we do not know the answer from a previous call to set or exists, we ask gitserver
// to resolve the commit and store the result for a subsequent call.
func (c *cachedCommitChecker) exists(ctx context.Context, repositoryID int, commit string) (bool, error) {
	c.mutex.RLock()
	if repositoryMap, ok := c.cache[repositoryID]; ok {
		if exists, ok := repositoryMap[commit]; ok {
			c.mutex.RUnlock()
			return exists, nil
		}
	}
	c.mutex.RUnlock()

	// Perform heavy work outside of critical section
	exists, err := c.gitserverClient.CommitExists(ctx, repositoryID, commit)
	if err != nil {
		return false, errors.Wrap(err, "gitserverClient.CommitExists")
	}

	c.setValue(repositoryID, commit, exists)
	return exists, nil
}

func (c *cachedCommitChecker) setValue(repositoryID int, commit string, exists bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.cache[repositoryID]; !ok {
		c.cache[repositoryID] = map[string]bool{}
	}

	c.cache[repositoryID][commit] = exists
}
