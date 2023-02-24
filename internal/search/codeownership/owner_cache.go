package codeownership

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
)

type OwnerKey struct {
	backend.OwnerResolutionContext
	email  string
	handle string
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

// func (c *OwnerCache) GetFromCacheOrFetch(ctx context.Context,  resCtx backend.OwnerResolutionContext, owners ...*codeownerspb.Owner) (ros codeowners.ResolvedOwners, err error) {
// 	ros = make(codeowners.ResolvedOwners, len(owners))
// 	lookup := make([]*codeownerspb.Owner, 0, len(owners)/2)
// 	for _, o := range owners {
// 		c.mu.RLock()
// 		key := OwnerKey{repoName, o.Email, o.Handle}
// 		if _, ok := c.owners[key]; ok {
// 			if c.owners[key].Err != nil {
// 				defer c.mu.RUnlock()
// 				return ros, err
// 			}
// 			ros.Add(c.owners[key].Owner)
// 		} else {
// 			lookup = append(lookup, o)
// 		}
// 		c.mu.RUnlock()
// 	}

// 	// TODO: cache logic feels wonky. Should recheck entry conditions etc.
// 	c.mu.Lock()
// defer c.mu.Unlock()
// 	owners, err := c.ownService.ResolveOwnersWithType(ctx, lookup, resCtx)
// 	for _, o := range owners {
// 		// Recheck condition.
// 		key := OwnerKey{repoName, o.Email, o.Handle}
// 		if _, ok := c.owners[key]; !ok {
// 			owners, err := c.ownService.ResolveOwnersWithType(ctx, []*codeownerspb.Owner{owner})
// 			if err != nil {
// 				c.owners[key] = OwnerCacheEntry{Err: err}
// 				return nil, err
// 			}
// 			e := OwnerCacheEntry{}
// 			if len(owners) == 1 {
// 				e.Owner = owners[0]
// 			}
// 			c.owners[key] = e
// 		}

// 	}
// 	}

// 	c.mu.Lock()
// 	defer c.mu.Unlock()
// 	// Recheck condition.
// 	if _, ok := c.owners[key]; !ok {
// 		owners, err := c.ownService.ResolveOwnersWithType(ctx, []*codeownerspb.Owner{owner})
// 		if err != nil {
// 			c.owners[key] = OwnerCacheEntry{Err: err}
// 			return nil, err
// 		}
// 		e := OwnerCacheEntry{}
// 		if len(owners) == 1 {
// 			e.Owner = owners[0]
// 		}
// 		c.owners[key] = e
// 	}
// 	return ros, nil
// }
