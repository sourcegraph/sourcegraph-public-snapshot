package store

import (
	"context"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GetRepositoriesForIndexScan returns a set of repository identifiers that should be considered
// for indexing jobs. Repositories that were returned previously from this call within the given
// process delay are not returned.
//
// If allowGlobalPolicies is false, then configuration policies that define neither a repository id
// nor a non-empty set of repository patterns wl be ignored. When true, such policies apply over all
// repositories known to the instance.
func (s *store) GetRepositoriesForIndexScan(ctx context.Context, table, column string, processDelay time.Duration, allowGlobalPolicies bool, repositoryMatchLimit *int, limit int, now time.Time) (_ []int, err error) {
	ctx, _, endObservation := s.operations.getRepositoriesForIndexScan.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Bool("allowGlobalPolicies", allowGlobalPolicies),
		log.Int("limit", limit),
	}})
	defer endObservation(1, observation.Args{})

	queries := make([]*sqlf.Query, 0, 3)
	if allowGlobalPolicies {
		limitExpression := sqlf.Sprintf("")
		if repositoryMatchLimit != nil {
			limitExpression = sqlf.Sprintf("LIMIT %s", *repositoryMatchLimit)
		}

		queries = append(queries, sqlf.Sprintf(
			getRepositoriesForIndexScanGlobalRepositoriesQuery,
			limitExpression,
		))
	}
	queries = append(queries, sqlf.Sprintf(getRepositoriesForIndexScanRepositoriesWithPolicyQuery))
	queries = append(queries, sqlf.Sprintf(getRepositoriesForIndexScanRepositoriesWithPolicyViaPatternQuery))

	for i, query := range queries {
		queries[i] = sqlf.Sprintf("(%s)", query)
	}

	replacer := strings.NewReplacer("{column_name}", column)
	return basestore.ScanInts(s.db.Query(ctx, sqlf.Sprintf(
		replacer.Replace(getRepositoriesForIndexScanQuery),
		sqlf.Join(queries, " UNION ALL "),
		quote(table),
		now,
		int(processDelay/time.Second),
		limit,
		quote(table),
		now,
		now,
	)))
}

func quote(s string) *sqlf.Query { return sqlf.Sprintf(s) }

const getRepositoriesForIndexScanQuery = `
WITH
-- This CTE will contain a single row if there is at least one global policy, and will return an empty
-- result set otherwise. If any global policy is for HEAD, the value for the column is_head_policy will
-- be true.
global_policy_descriptor AS MATERIALIZED (
	SELECT (p.type = 'GIT_COMMIT' AND p.pattern = 'HEAD') AS is_head_policy
	FROM lsif_configuration_policies p
	WHERE
		p.indexing_enabled AND
		p.repository_id IS NULL AND
		p.repository_patterns IS NULL
	ORDER BY is_head_policy DESC
	LIMIT 1
),
repositories_matching_policy AS (
	%s
),
repositories AS (
	SELECT rmp.id
	FROM repositories_matching_policy rmp
	LEFT JOIN %s lrs ON lrs.repository_id = rmp.id
	WHERE
		-- Records that have not been checked within the global reindex threshold are also eligible for
		-- indexing. Note that condition here is true for a record that has never been indexed.
		(%s - lrs.{column_name} > (%s * '1 second'::interval)) IS DISTINCT FROM FALSE OR

		-- Records that have received an update since their last scan are also eligible for re-indexing.
		-- Note that last_changed is NULL unless the repository is attached to a policy for HEAD.
		(rmp.last_changed > lrs.{column_name})
	ORDER BY
		lrs.{column_name} NULLS FIRST,
		rmp.id -- tie breaker
	LIMIT %s
)
INSERT INTO %s (repository_id, {column_name})
SELECT DISTINCT r.id, %s::timestamp FROM repositories r
ON CONFLICT (repository_id) DO UPDATE
SET {column_name} = %s
RETURNING repository_id
`

const getRepositoriesForIndexScanGlobalRepositoriesQuery = `
SELECT
	r.id,
	CASE
		-- Return non-NULL last_changed only for policies that are attached to a HEAD commit.
		-- We don't want to superfluously return the same repos becasue they had an update, but
		-- we only (for example) index a branch that doesn't have many active commits.
		WHEN gpd.is_head_policy THEN gr.last_changed
		ELSE NULL
	END AS last_changed
FROM repo r
JOIN gitserver_repos gr ON gr.repo_id = r.id
JOIN global_policy_descriptor gpd ON TRUE
WHERE
	r.deleted_at IS NULL AND
	r.blocked IS NULL AND
	gr.clone_status = 'cloned'
ORDER BY stars DESC NULLS LAST, id
%s
`

const getRepositoriesForIndexScanRepositoriesWithPolicyQuery = `
SELECT
	r.id,
	CASE
		-- Return non-NULL last_changed only for policies that are attached to a HEAD commit.
		-- We don't want to superfluously return the same repos becasue they had an update, but
		-- we only (for example) index a branch that doesn't have many active commits.
		WHEN p.type = 'GIT_COMMIT' AND p.pattern = 'HEAD' THEN gr.last_changed
		ELSE NULL
	END AS last_changed
FROM repo r
JOIN gitserver_repos gr ON gr.repo_id = r.id
JOIN lsif_configuration_policies p ON p.repository_id = r.id
WHERE
	r.deleted_at IS NULL AND
	r.blocked IS NULL AND
	p.indexing_enabled AND
	gr.clone_status = 'cloned'
`

const getRepositoriesForIndexScanRepositoriesWithPolicyViaPatternQuery = `
SELECT
	r.id,
	CASE
		-- Return non-NULL last_changed only for policies that are attached to a HEAD commit.
		-- We don't want to superfluously return the same repos becasue they had an update, but
		-- we only (for example) index a branch that doesn't have many active commits.
		WHEN p.type = 'GIT_COMMIT' AND p.pattern = 'HEAD' THEN gr.last_changed
		ELSE NULL
	END AS last_changed
FROM repo r
JOIN gitserver_repos gr ON gr.repo_id = r.id
JOIN lsif_configuration_policies_repository_pattern_lookup rpl ON rpl.repo_id = r.id
JOIN lsif_configuration_policies p ON p.id = rpl.policy_id
WHERE
	r.deleted_at IS NULL AND
	r.blocked IS NULL AND
	p.indexing_enabled AND
	gr.clone_status = 'cloned'
`

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

// GetDirtyRepositories returns a map from repository identifiers to a dirty token for each repository whose commit
// graph is out of date. This token should be passed to CalculateVisibleUploads in order to unmark the repository.
func (s *store) GetDirtyRepositories(ctx context.Context) (_ map[int]int, err error) {
	ctx, trace, endObservation := s.operations.getDirtyRepositories.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	repositories, err := scanIntPairs(s.db.Query(ctx, sqlf.Sprintf(dirtyRepositoriesQuery)))
	if err != nil {
		return nil, err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("numRepositories", len(repositories)))

	return repositories, nil
}

const dirtyRepositoriesQuery = `
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
