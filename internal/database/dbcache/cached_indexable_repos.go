package dbcache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// indexableReposMaxAge is how long we cache the list of indexable repos. The list
// changes very rarely, so we can cache for a while.
const indexableReposMaxAge = time.Minute

type cachedRepos struct {
	repos   []types.MinimalRepo
	fetched time.Time
}

// Repos returns the current cached repos and boolean value indicating
// whether an update is required
func (c *cachedRepos) Repos() ([]types.MinimalRepo, bool) {
	if c == nil {
		return nil, true
	}
	if c.repos == nil {
		return nil, true
	}
	return append([]types.MinimalRepo{}, c.repos...), time.Since(c.fetched) > indexableReposMaxAge
}

var globalReposCache = reposCache{}

func NewIndexableReposLister(store database.RepoStore) *IndexableReposLister {
	return &IndexableReposLister{
		store:      store,
		reposCache: &globalReposCache,
	}
}

type reposCache struct {
	cacheAllRepos    atomic.Value
	cachePublicRepos atomic.Value
	mu               sync.Mutex
}

// IndexableReposLister holds the list of indexable repos which are cached for
// indexableReposMaxAge.
type IndexableReposLister struct {
	store database.RepoStore
	*reposCache
}

// List lists ALL indexable repos. These include all repos with a minimum number of stars,
// user added repos (both public and private) as well as any repos added
// to the user_public_repos table.
//
// The values are cached for up to indexableReposMaxAge. If the cache has expired, we return
// stale data and start a background refresh.
func (s *IndexableReposLister) List(ctx context.Context) (results []types.MinimalRepo, err error) {
	return s.list(ctx, false)
}

// ListPublic is similar to List except that it only includes public repos.
func (s *IndexableReposLister) ListPublic(ctx context.Context) (results []types.MinimalRepo, err error) {
	return s.list(ctx, true)
}

func (s *IndexableReposLister) list(ctx context.Context, onlyPublic bool) (results []types.MinimalRepo, err error) {
	cache := &(s.cacheAllRepos)
	if onlyPublic {
		cache = &(s.cachePublicRepos)
	}

	cached, _ := cache.Load().(*cachedRepos)
	repos, needsUpdate := cached.Repos()
	if !needsUpdate {
		return repos, nil
	}

	// We don't have any repos yet, fetch them
	if len(repos) == 0 {
		return s.refreshCache(ctx, onlyPublic)
	}

	// We have existing repos, return the stale data and start background refresh
	go func() {
		newCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		_, err := s.refreshCache(newCtx, onlyPublic)
		if err != nil {
			log15.Error("Refreshing indexable repos cache", "error", err)
		}
	}()
	return repos, nil
}

func (s *IndexableReposLister) refreshCache(ctx context.Context, onlyPublic bool) ([]types.MinimalRepo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cache := &(s.cacheAllRepos)
	if onlyPublic {
		cache = &(s.cachePublicRepos)
	}

	// Check whether another routine already did the work
	cached, _ := cache.Load().(*cachedRepos)
	repos, needsUpdate := cached.Repos()
	if !needsUpdate {
		return repos, nil
	}

	opts := database.ListIndexableReposOptions{}
	if !onlyPublic {
		opts.IncludePrivate = true
	}
	repos, err := s.store.ListIndexableRepos(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "querying for indexable repos")
	}

	cache.Store(&cachedRepos{
		// Copy since repos will be mutated by the caller
		repos:   append([]types.MinimalRepo{}, repos...),
		fetched: time.Now(),
	})

	return repos, nil
}
