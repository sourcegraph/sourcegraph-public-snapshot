package repos

import (
	"context"
	"time"

	lru "github.com/hashicorp/golang-lru"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type MetadataCache struct {
	lru *lru.TwoQueueCache
}

type cacheEntry struct {
	repo       *types.Repo
	validUntil time.Time
}

// Get returns a cached *types.Repo for an id if it exists in the cache
// and hasn't expired.
func (c *MetadataCache) Get(id api.RepoID) (*types.Repo, bool) {
	ci, ok := c.lru.Get(id)
	if !ok {
		return nil, false
	}
	ce := ci.(cacheEntry)

	// Evict the entry if it has expired
	if time.Now().After(ce.validUntil) {
		c.lru.Remove(id)
		return nil, false
	}

	return ce.repo, true
}

func (c *MetadataCache) Add(id api.RepoID, repo *types.Repo) {
	e := cacheEntry{
		repo:       repo,
		validUntil: time.Now().Add(10 * time.Minute),
	}
	c.lru.Add(id, e)
}

func (c *MetadataCache) GetReposForEvent(ctx context.Context, db dbutil.DB, event streaming.SearchEvent) (map[api.RepoID]*types.Repo, error) {
	res := make(map[api.RepoID]*types.Repo, 10)
	uncachedIDs := make(map[api.RepoID]struct{})
	for _, match := range event.Results {
		id := match.RepoName().ID
		if repo, ok := c.Get(id); ok {
			res[id] = repo
		} else {
			uncachedIDs[id] = struct{}{}
		}
	}

	// All repos referenced in the event were populated from the cache,
	// so no need to hit the database
	if len(uncachedIDs) == 0 {
		return res, nil
	}

	uncachedIDSlice := make([]api.RepoID, 0, len(uncachedIDs))
	for id := range uncachedIDs {
		uncachedIDSlice = append(uncachedIDSlice, id)
	}

	// Retrieve all repo metadata that doesn't exist in the cache
	repos, err := database.Repos(db).GetByIDs(ctx, uncachedIDSlice...)
	if err != nil {
		return nil, err
	}

	// Update the cache and the result
	for _, repo := range repos {
		c.Add(repo.ID, repo)
		res[repo.ID] = repo
	}
	return res, nil
}

func NewMetadataCache(size int) MetadataCache {
	l, _ := lru.New2Q(size)
	return MetadataCache{
		lru: l,
	}
}
