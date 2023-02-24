package codeownership

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
)

type OwnerKey struct {
	repoName api.RepoName
	email    string
	handle   string
}

type OwnerCacheEntry struct {
	Err   error
	Owner codeowners.ResolvedOwner
}

type OwnerCache struct {
	owners     map[OwnerKey]OwnerCacheEntry
	ownService backend.OwnService

	mu sync.RWMutex
}

func NewOwnerCache(ownService backend.OwnService) OwnerCache {
	return OwnerCache{
		owners:     make(map[OwnerKey]OwnerCacheEntry),
		ownService: ownService,
	}
}

func (c *OwnerCache) GetFromCacheOrFetch(ctx context.Context, repoName api.RepoName, owner *codeownerspb.Owner) (codeowners.ResolvedOwner, error) {
	c.mu.RLock()
	key := OwnerKey{repoName, owner.Email, owner.Handle}
	if _, ok := c.owners[key]; ok {
		defer c.mu.RUnlock()
		return c.owners[key].Owner, c.owners[key].Err
	}
	c.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	// Recheck condition.
	if _, ok := c.owners[key]; !ok {
		owners, err := c.ownService.ResolveOwnersWithType(ctx, []*codeownerspb.Owner{owner})
		if err != nil {
			c.owners[key] = OwnerCacheEntry{Err: err}
			return nil, err
		}
		e := OwnerCacheEntry{}
		if len(owners) == 1 {
			e.Owner = owners[0]
		}
		c.owners[key] = e
	}
	return c.owners[key].Owner, c.owners[key].Err
}
