package resolvers

import (
	"context"

	"github.com/pkg/errors"
)

type cachedCommitChecker struct {
	gitserverClient GitserverClient
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
	if _, ok := c.cache[repositoryID]; !ok {
		c.cache[repositoryID] = map[string]bool{}
	}

	c.cache[repositoryID][commit] = true
}

// exists determines if the given commit is resolvable for the given repository. If
// we do not know the answer from a previous call to set or exists, we ask gitserver
// to resolve the commit and store the result for a subsequent call.
func (c *cachedCommitChecker) exists(ctx context.Context, repositoryID int, commit string) (bool, error) {
	if _, ok := c.cache[repositoryID]; !ok {
		c.cache[repositoryID] = map[string]bool{}
	}

	if exists, ok := c.cache[repositoryID][commit]; ok {
		return exists, nil
	}

	exists, err := c.gitserverClient.CommitExists(ctx, repositoryID, commit)
	if err != nil {
		return false, errors.Wrap(err, "gitserverClient.CommitExists")
	}

	c.cache[repositoryID][commit] = exists
	return exists, nil
}
