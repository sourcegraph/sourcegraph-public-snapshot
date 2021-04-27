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
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
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

type DefaultRepoStore struct {
	*basestore.Store

	cache atomic.Value
	once  sync.Once

	mu sync.Mutex
}

// DefaultRepos instantiates and returns a new DefaultRepoStore with prepared statements.
func DefaultRepos(db dbutil.DB) *DefaultRepoStore {
	return &DefaultRepoStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// NewDefaultRepoStoreWithDB instantiates and returns a new DefaultRepoStore using the other store handle.
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

// ensureStore instantiates a basestore.Store if necessary, using the dbconn.Global handle.
// This function ensures access to dbconn happens after the rest of the code or tests have
// initialized it.
func (s *DefaultRepoStore) ensureStore() {
	s.once.Do(func() {
		if s.Store == nil {
			s.Store = basestore.NewWithDB(dbconn.Global, sql.TxOptions{})
		}
	})
}

func (s *DefaultRepoStore) List(ctx context.Context) (results []types.RepoName, err error) {
	s.ensureStore()

	cached, _ := s.cache.Load().(*cachedRepos)
	repos, needsUpdate := cached.Repos()
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
			log15.Error("Refreshing default repos cache", "error", err)
		}
	}()
	return repos, nil
}

func (s *DefaultRepoStore) refreshCache(ctx context.Context) ([]types.RepoName, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check whether another routine already did the work
	cached, _ := s.cache.Load().(*cachedRepos)
	repos, needsUpdate := cached.Repos()
	if !needsUpdate {
		return repos, nil
	}

	repos, err := ReposWith(s).ListDefaultRepos(ctx, ListDefaultReposOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "querying for default repos")
	}

	s.cache.Store(&cachedRepos{
		// Copy since repos will be mutated by the caller
		repos:   append([]types.RepoName{}, repos...),
		fetched: time.Now(),
	})

	return repos, nil
}

func (s *DefaultRepoStore) resetCache() {
	s.cache.Store(&cachedRepos{})
}
