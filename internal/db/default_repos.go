package db

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// defaultReposMaxAge is how long we cache the list of default repos. The list
// changes very rarely, so we can cache for a while.
const defaultReposMaxAge = time.Minute

type cachedRepos struct {
	repos   []*types.RepoName
	fetched time.Time
}

// Repos returns the current cached repos and boolean value indicating
// whether an update is required
func (c *cachedRepos) Repos() ([]*types.RepoName, bool) {
	if c == nil {
		return nil, true
	}
	if c.repos == nil {
		return nil, true
	}
	return append([]*types.RepoName{}, c.repos...), time.Since(c.fetched) > defaultReposMaxAge
}

type defaultRepos struct {
	cache atomic.Value
	mu    sync.Mutex
}

func (s *defaultRepos) List(ctx context.Context) (results []*types.RepoName, err error) {
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

func (s *defaultRepos) refreshCache(ctx context.Context) ([]*types.RepoName, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check whether another routine already did the work
	cached, _ := s.cache.Load().(*cachedRepos)
	repos, needsUpdate := cached.Repos()
	if !needsUpdate {
		return repos, nil
	}

	// We can reset the slice here so we don't allocate another one
	if repos != nil {
		repos = repos[0:0]
	}

	const q = `
-- source: internal/db/default_repos.go:defaultRepos.List
SELECT
    id,
    name
FROM
    repo r
WHERE
    EXISTS (
        SELECT
        FROM
            external_service_repos sr
            INNER JOIN external_services s ON s.id = sr.external_service_id
        WHERE
			s.namespace_user_id IS NOT NULL
			AND s.deleted_at IS NULL
			AND r.id = sr.repo_id
            AND r.deleted_at IS NULL)
UNION
    SELECT
        repo.id,
        repo.name
    FROM
        default_repos
		JOIN repo ON default_repos.repo_id = repo.id
	WHERE
		deleted_at IS NULL
`
	rows, err := dbconn.Global.QueryContext(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "fetching repos")
	}
	defer rows.Close()
	for rows.Next() {
		var r types.RepoName
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return nil, errors.Wrap(err, "scanning row from default_repos table")
		}
		repos = append(repos, &r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "scanning rows from default_repos table")
	}

	s.cache.Store(&cachedRepos{
		// Copy since repos will be mutated by the caller
		repos:   append([]*types.RepoName{}, repos...),
		fetched: time.Now(),
	})

	return repos, nil
}

func (s *defaultRepos) resetCache() {
	s.cache.Store(&cachedRepos{})
}
