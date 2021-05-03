package database

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// defaultReposMaxAge is how long we cache the list of default repos. The list
// changes very rarely, so we can cache for a while.
const defaultReposMaxAge = time.Minute

type cachedRepos struct {
	repos   []types.RepoName
	fetched time.Time
}

// Repos returns the current cached repos and boolean value indicating
// whether an update is required
func (c *cachedRepos) Repos() ([]types.RepoName, bool) {
	if c == nil {
		return nil, true
	}
	if c.repos == nil {
		return nil, true
	}
	return append([]types.RepoName{}, c.repos...), time.Since(c.fetched) > defaultReposMaxAge
}

// DefaultRepoStore holds the list of default repos which are cached for
// defaultReposMaxAge.
type DefaultRepoStore struct {
	*basestore.Store

	cacheAllRepos    atomic.Value
	cachePublicRepos atomic.Value

	mu sync.Mutex
}

// DefaultRepos instantiates and returns a new DefaultRepoStore with prepared statements.
func DefaultRepos(db dbutil.DB) *DefaultRepoStore {
	return &DefaultRepoStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// DefaultReposWith instantiates and returns a new DefaultRepoStore using the other store handle.
func DefaultReposWith(other basestore.ShareableStore) *DefaultRepoStore {
	return &DefaultRepoStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *DefaultRepoStore) With(other basestore.ShareableStore) *DefaultRepoStore {
	return &DefaultRepoStore{Store: s.Store.With(other)}
}

func (s *DefaultRepoStore) Transact(ctx context.Context) (*DefaultRepoStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &DefaultRepoStore{Store: txBase}, err
}

// List lists ALL default repos. These include anything in the default_repos
// table, user added repos (both public and private) as well as any repos added
// to the user_public_repos table.
//
// The values are cached for up to defaultReposMaxAge. If the cache has expired, we return
// stale data and start a background refresh.
func (s *DefaultRepoStore) List(ctx context.Context) (results []types.RepoName, err error) {
	return s.list(ctx, false)
}

// ListPublic is similar to List except that it only includes public repos.
func (s *DefaultRepoStore) ListPublic(ctx context.Context) (results []types.RepoName, err error) {
	return s.list(ctx, true)
}

func (s *DefaultRepoStore) list(ctx context.Context, onlyPublic bool) (results []types.RepoName, err error) {
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
			log15.Error("Refreshing default repos cache", "error", err)
		}
	}()
	return repos, nil
}

func (s *DefaultRepoStore) refreshCache(ctx context.Context, onlyPublic bool) ([]types.RepoName, error) {
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

	opts := ListDefaultReposOptions{}
	if !onlyPublic {
		opts.IncludePrivate = true
	}
	repos, err := ReposWith(s).ListDefaultRepos(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "querying for default repos")
	}

	cache.Store(&cachedRepos{
		// Copy since repos will be mutated by the caller
		repos:   append([]types.RepoName{}, repos...),
		fetched: time.Now(),
	})

	return repos, nil
}

func (s *DefaultRepoStore) resetCache() {
	s.cacheAllRepos.Store(&cachedRepos{})
	s.cachePublicRepos.Store(&cachedRepos{})
}
