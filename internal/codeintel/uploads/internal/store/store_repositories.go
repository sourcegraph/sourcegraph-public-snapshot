package store

import (
	"context"
	"errors"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

// SetRepositoriesForRetentionScan returns a set of repository identifiers with live code intelligence
// data and a fresh associated commit graph. Repositories that were returned previously from this call
// within the  given process delay are not returned.
func (s *store) SetRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) (_ []int, err error) {
	ctx, _, endObservation := s.operations.setRepositoriesForRetentionScan.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	now := timeutil.Now()

	return basestore.ScanInts(s.db.Query(ctx, sqlf.Sprintf(
		repositoryIDsForRetentionScanQuery,
		now,
		int(processDelay/time.Second),
		limit,
		now,
		now,
	)))
}

func (s *store) SetRepositoriesForRetentionScanWithTime(ctx context.Context, processDelay time.Duration, limit int, now time.Time) (_ []int, err error) {
	ctx, _, endObservation := s.operations.setRepositoriesForRetentionScan.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return basestore.ScanInts(s.db.Query(ctx, sqlf.Sprintf(
		repositoryIDsForRetentionScanQuery,
		now,
		int(processDelay/time.Second),
		limit,
		now,
		now,
	)))
}

const repositoryIDsForRetentionScanQuery = `
-- source: internal/codeintel/uploads/internal/store/store_repositories.go:setRepositoriesForRetentionScan
WITH candidate_repositories AS (
	SELECT DISTINCT u.repository_id AS id
	FROM lsif_uploads u
	WHERE u.state = 'completed'
),
repositories AS (
	SELECT cr.id
	FROM candidate_repositories cr
	LEFT JOIN lsif_last_retention_scan lrs ON lrs.repository_id = cr.id
	JOIN lsif_dirty_repositories dr ON dr.repository_id = cr.id

	-- Ignore records that have been checked recently. Note this condition is
	-- true for a null last_retention_scan_at (which has never been checked).
	WHERE (%s - lrs.last_retention_scan_at > (%s * '1 second'::interval)) IS DISTINCT FROM FALSE
	AND dr.update_token = dr.dirty_token
	ORDER BY
		lrs.last_retention_scan_at NULLS FIRST,
		cr.id -- tie breaker
	LIMIT %s
)
INSERT INTO lsif_last_retention_scan (repository_id, last_retention_scan_at)
SELECT r.id, %s::timestamp FROM repositories r
ON CONFLICT (repository_id) DO UPDATE
SET last_retention_scan_at = %s
RETURNING repository_id
`

// SetRepositoryAsDirty marks the given repository's commit graph as out of date.
func (s *store) SetRepositoryAsDirty(ctx context.Context, repositoryID int) (err error) {
	ctx, _, endObservation := s.operations.setRepositoryAsDirty.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(setRepositoryAsDirtyQuery, repositoryID))
}

// SetRepositoryAsDirtyWithTx marks the given repository's commit graph as out of date.
func (s *store) setRepositoryAsDirtyWithTx(ctx context.Context, repositoryID int, tx *basestore.Store) (err error) {
	ctx, _, endObservation := s.operations.setRepositoryAsDirty.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	return tx.Exec(ctx, sqlf.Sprintf(setRepositoryAsDirtyQuery, repositoryID))
}

const setRepositoryAsDirtyQuery = `
-- source: internal/codeintel/uploads/internal/stores/store_repositories.go:SetRepositoryAsDirty
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
-- source: internal/codeintel/uploads/internal/store/store_repositories.go:GetDirtyRepositories
SELECT ldr.repository_id, ldr.dirty_token
  FROM lsif_dirty_repositories ldr
    INNER JOIN repo ON repo.id = ldr.repository_id
  WHERE ldr.dirty_token > ldr.update_token
    AND repo.deleted_at IS NULL
	AND repo.blocked IS NULL
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
-- source: internal/codeintel/uploads/internal/store/store_repositories.go:MaxStaleAge
SELECT EXTRACT(EPOCH FROM NOW() - ldr.set_dirty_at)::integer AS age
  FROM lsif_dirty_repositories ldr
    INNER JOIN repo ON repo.id = ldr.repository_id
  WHERE ldr.dirty_token > ldr.update_token
    AND repo.deleted_at IS NULL
    AND repo.blocked IS NULL
  ORDER BY age DESC
  LIMIT 1
`

// ErrUnknownRepository occurs when a repository does not exist.
var ErrUnknownRepository = errors.New("unknown repository")

// RepoName returns the name for the repo with the given identifier.
func (s *store) RepoName(ctx context.Context, repositoryID int) (_ string, err error) {
	ctx, _, endObservation := s.operations.repoName.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	name, exists, err := basestore.ScanFirstString(s.db.Query(ctx, sqlf.Sprintf(repoNameQuery, repositoryID)))
	if err != nil {
		return "", err
	}
	if !exists {
		return "", ErrUnknownRepository
	}
	return name, nil
}

const repoNameQuery = `
-- source: internal/codeintel/uploads/internal/store/store_repositories.go:RepoName
SELECT name FROM repo WHERE id = %s
`

// RepoNames returns a map from repository id to names.
func (s *store) RepoNames(ctx context.Context, repositoryIDs ...int) (_ map[int]string, err error) {
	ctx, _, endObservation := s.operations.repoName.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numRepositories", len(repositoryIDs)),
	}})
	defer endObservation(1, observation.Args{})

	return scanRepoNames(s.db.Query(ctx, sqlf.Sprintf(repoNamesQuery, pq.Array(repositoryIDs))))
}

const repoNamesQuery = `
-- source: internal/codeintel/uploads/internal/store/store_repositories.go:RepoNames
SELECT id, name FROM repo WHERE id = ANY(%s)
`

// HasRepository determines if there is LSIF data for the given repository.
func (s *store) HasRepository(ctx context.Context, repositoryID int) (_ bool, err error) {
	ctx, _, endObservation := s.operations.hasRepository.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	_, found, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(hasRepositoryQuery, repositoryID)))
	return found, err
}

const hasRepositoryQuery = `
-- source: internal/codeintel/stores/dbstore/commits.go:HasRepository
SELECT 1 FROM lsif_uploads WHERE state NOT IN ('deleted', 'deleting') AND repository_id = %s LIMIT 1
`
