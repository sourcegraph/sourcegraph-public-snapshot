package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// SetRepositoryAsDirty marks the given repository's commit graph as out of date.
func (s *store) SetRepositoryAsDirty(ctx context.Context, repositoryID int, tx *basestore.Store) (err error) {
	ctx, _, endObservation := s.operations.setRepositoryAsDirty.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	return tx.Exec(ctx, sqlf.Sprintf(setRepositoryAsDirtyQuery, repositoryID))
}

const setRepositoryAsDirtyQuery = `
-- source: internal/codeintel/uploads/internal/stores/store_commits.go:SetRepositoryAsDirty
INSERT INTO lsif_dirty_repositories (repository_id, dirty_token, update_token)
VALUES (%s, 1, 0)
ON CONFLICT (repository_id) DO UPDATE SET
    dirty_token = lsif_dirty_repositories.dirty_token + 1,
    set_dirty_at = CASE
        WHEN lsif_dirty_repositories.update_token = lsif_dirty_repositories.dirty_token THEN NOW()
        ELSE lsif_dirty_repositories.set_dirty_at
    END
`

// GetDirtyRepositories returns a map from repository identifiers to a dirty token for each repository whose commit
// graph is out of date. This token should be passed to CalculateVisibleUploads in order to unmark the repository.
func (s *store) GetDirtyRepositories(ctx context.Context) (_ map[int]int, err error) {
	ctx, trace, endObservation := s.operations.getDirtyRepositories.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	repositories, err := scanIntPairs(s.db.Query(ctx, sqlf.Sprintf(dirtyRepositoriesQuery)))
	if err != nil {
		return nil, err
	}
	trace.Log(log.Int("numRepositories", len(repositories)))

	return repositories, nil
}

const dirtyRepositoriesQuery = `
-- source: internal/codeintel/uploads/internal/store/store_commits.go:GetDirtyRepositories
SELECT ldr.repository_id, ldr.dirty_token
  FROM lsif_dirty_repositories ldr
    INNER JOIN repo ON repo.id = ldr.repository_id
  WHERE ldr.dirty_token > ldr.update_token
    AND repo.deleted_at IS NULL
`

// GetRepositoriesMaxStaleAge returns the longest duration that a repository has been (currently) stale for. This method considers
// only repositories that would be returned by DirtyRepositories. This method returns a duration of zero if there
// are no stale repositories.
func (s *store) GetRepositoriesMaxStaleAge(ctx context.Context) (_ time.Duration, err error) {
	ctx, _, endObservation := s.operations.getRepositoriesMaxStaleAge.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	ageSeconds, ok, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(maxStaleAgeQuery)))
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, nil
	}

	return time.Duration(ageSeconds) * time.Second, nil
}

const maxStaleAgeQuery = `
-- source: internal/codeintel/stores/dbstore/commits.go:MaxStaleAge
SELECT EXTRACT(EPOCH FROM NOW() - ldr.set_dirty_at)::integer AS age
  FROM lsif_dirty_repositories ldr
    INNER JOIN repo ON repo.id = ldr.repository_id
  WHERE ldr.dirty_token > ldr.update_token
    AND repo.deleted_at IS NULL
  ORDER BY age DESC
  LIMIT 1
`
