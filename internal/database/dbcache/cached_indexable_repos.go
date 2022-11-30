package dbcache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// indexableReposMaxAge is how long we cache the list of indexable repos. The list
// changes very rarely, so we can cache for a while.
const indexableReposMaxAge = time.Minute

type cachedRepos struct {
	minimalRepos []types.MinimalRepo
	fetched      time.Time
}

// repos returns the current cached repos and boolean value indicating
// whether an update is required
func (c *cachedRepos) repos() ([]types.MinimalRepo, bool) {
	if c == nil {
		return nil, true
	}
	if c.minimalRepos == nil {
		return nil, true
	}
	return append([]types.MinimalRepo{}, c.minimalRepos...), time.Since(c.fetched) > indexableReposMaxAge
}

var globalReposCache = reposCache{}

func NewIndexableReposLister(logger log.Logger, store database.RepoStore) *IndexableReposLister {
	return &IndexableReposLister{
		logger:     logger,
		store:      store,
		reposCache: &globalReposCache,
	}
}

type reposCache struct {
	cacheAllRepos atomic.Value
	mu            sync.Mutex
}

// IndexableReposLister holds the list of indexable repos which are cached for
// indexableReposMaxAge.
type IndexableReposLister struct {
	logger log.Logger
	store  database.RepoStore
	*reposCache
}

// List lists ALL indexable repos. These include all repos with a minimum number of stars.
//
// The values are cached for up to indexableReposMaxAge. If the cache has expired, we return
// stale data and start a background refresh.
func (s *IndexableReposLister) List(ctx context.Context) (results []types.MinimalRepo, err error) {
	cache := &(s.cacheAllRepos)

	cached, _ := cache.Load().(*cachedRepos)
	repos, needsUpdate := cached.repos()
	if !needsUpdate {
		return repos, nil
	}

	// We don't have any repos yet, fetch them
	if len(repos) == 0 {
		return s.refreshCache(ctx)
	}

	// We have existing repos, return the stale data and start background refresh
	go func() {
		newCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		_, err := s.refreshCache(newCtx)
		if err != nil {
			s.logger.Error("Refreshing indexable repos cache", log.Error(err))
		}
	}()
	return repos, nil
}

func (s *IndexableReposLister) refreshCache(ctx context.Context) ([]types.MinimalRepo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cache := &(s.cacheAllRepos)

	// Check whether another routine already did the work
	cached, _ := cache.Load().(*cachedRepos)
	repos, needsUpdate := cached.repos()
	if !needsUpdate {
		return repos, nil
	}

	opts := database.ListSourcegraphDotComIndexableReposOptions{
		// Zoekt can only index a repo which has been cloned.
		CloneStatus: types.CloneStatusCloned,
	}
	repos, err := s.store.ListSourcegraphDotComIndexableRepos(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "querying for indexable repos")
	}

	cache.Store(&cachedRepos{
		// Copy since repos will be mutated by the caller
		minimalRepos: append([]types.MinimalRepo{}, repos...),
		fetched:      time.Now(),
	})

	return repos, nil
}
