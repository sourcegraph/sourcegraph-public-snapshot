package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel/attribute"

	autoindexingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
INSERT INTO lsif_dirty_repositories (repository_id, dirty_token, update_token)
VALUES (%s, 1, 0)
ON CONFLICT (repository_id) DO UPDATE SET
    dirty_token = lsif_dirty_repositories.dirty_token + 1,
    set_dirty_at = CASE
        WHEN lsif_dirty_repositories.update_token = lsif_dirty_repositories.dirty_token THEN NOW()
        ELSE lsif_dirty_repositories.set_dirty_at
    END
`

var scanDirtyRepositories = basestore.NewSliceScanner(func(s dbutil.Scanner) (dr shared.DirtyRepository, _ error) {
	err := s.Scan(&dr.RepositoryID, &dr.RepositoryName, &dr.DirtyToken)
	return dr, err
})

// GetDirtyRepositories returns list of repositories whose commit graph is out of date. The dirty token should be
// passed to CalculateVisibleUploads in order to unmark the repository.
func (s *store) GetDirtyRepositories(ctx context.Context) (_ []shared.DirtyRepository, err error) {
	ctx, trace, endObservation := s.operations.getDirtyRepositories.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	repositories, err := scanDirtyRepositories(s.db.Query(ctx, sqlf.Sprintf(dirtyRepositoriesQuery)))
	if err != nil {
		return nil, err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("numRepositories", len(repositories)))

	return repositories, nil
}

const dirtyRepositoriesQuery = `
SELECT ldr.repository_id, repo.name, ldr.dirty_token
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
SELECT 1 FROM lsif_uploads WHERE state NOT IN ('deleted', 'deleting') AND repository_id = %s LIMIT 1
`

func (s *store) NumRepositoriesWithCodeIntelligence(ctx context.Context) (_ int, err error) {
	ctx, _, endObservation := s.operations.numRepositoriesWithCodeIntelligence.With(ctx, &err, observation.Args{LogFields: []log.Field{}})
	defer endObservation(1, observation.Args{})

	numRepositoriesWithCodeIntelligence, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(countRepositoriesQuery)))
	if err != nil {
		return 0, err
	}

	return numRepositoriesWithCodeIntelligence, err
}

const countRepositoriesQuery = `
WITH candidate_repositories AS (
	SELECT
	DISTINCT uvt.repository_id AS id
	FROM lsif_uploads_visible_at_tip uvt
	WHERE is_default_branch
)
SELECT COUNT(*)
FROM candidate_repositories s
JOIN repo r ON r.id = s.id
WHERE
	r.deleted_at IS NULL AND
	r.blocked IS NULL
`

func (s *store) RepositoryIDsWithErrors(ctx context.Context, offset, limit int) (_ []autoindexingshared.RepositoryWithCount, totalCount int, err error) {
	ctx, _, endObservation := s.operations.repositoryIDsWithErrors.With(ctx, &err, observation.Args{LogFields: []log.Field{}})
	defer endObservation(1, observation.Args{})

	return scanRepositoryWithCounts(s.db.Query(ctx, sqlf.Sprintf(repositoriesWithErrorsQuery, limit, offset)))
}

var scanRepositoryWithCounts = basestore.NewSliceWithCountScanner(func(s dbutil.Scanner) (rc autoindexingshared.RepositoryWithCount, count int, _ error) {
	err := s.Scan(&rc.RepositoryID, &rc.Count, &count)
	return rc, count, err
})

const repositoriesWithErrorsQuery = `
WITH

-- Return unique (repository, root, indexer) triples for each "project" (root/indexer pair)
-- within a repository that has a failing record without a newer completed record shadowing
-- it. Group these by the project triples so that we only return one row for the count we
-- perform below.
candidates_from_uploads AS (
	SELECT u.repository_id
	FROM lsif_uploads u
	WHERE
		u.state = 'failed' AND
		NOT EXISTS (
			SELECT 1
			FROM lsif_uploads u2
			WHERE
				u2.state = 'completed' AND
				u2.repository_id = u.repository_id AND
				u2.root = u.root AND
				u2.indexer = u.indexer AND
				u2.finished_at > u.finished_at
		)
	GROUP BY u.repository_id, u.root, u.indexer
),

-- Same as above for index records
candidates_from_indexes AS (
	SELECT u.repository_id
	FROM lsif_indexes u
	WHERE
		u.state = 'failed' AND
		NOT EXISTS (
			SELECT 1
			FROM lsif_indexes u2
			WHERE
				u2.state = 'completed' AND
				u2.repository_id = u.repository_id AND
				u2.root = u.root AND
				u2.indexer = u.indexer AND
				u2.finished_at > u.finished_at
		)
	GROUP BY u.repository_id, u.root, u.indexer
),

candidates AS (
	SELECT * FROM candidates_from_uploads UNION ALL
	SELECT * FROM candidates_from_indexes
),
grouped_candidates AS (
	SELECT
		r.repository_id,
		COUNT(*) AS num_failures
	FROM candidates r
	GROUP BY r.repository_id
)
SELECT
	r.repository_id,
	r.num_failures,
	COUNT(*) OVER() AS count
FROM grouped_candidates r
ORDER BY num_failures DESC, repository_id
LIMIT %s
OFFSET %s
`
