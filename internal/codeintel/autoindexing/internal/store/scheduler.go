package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetRepositoriesForIndexScan returns a set of repository identifiers that should be considered
// for indexing jobs. Repositories that were returned previously from this call within the given
// process delay are not returned.
//
// If allowGlobalPolicies is false, then configuration policies that define neither a repository id
// nor a non-empty set of repository patterns wl be ignored. When true, such policies apply over all
// repositories known to the instance.
func (s *store) GetRepositoriesForIndexScan(
	ctx context.Context,
	processDelay time.Duration,
	allowGlobalPolicies bool,
	repositoryMatchLimit *int,
	limit int,
	now time.Time,
) (_ []int, err error) {
	var repositoryMatchLimitValue int
	if repositoryMatchLimit != nil {
		repositoryMatchLimitValue = *repositoryMatchLimit
	}

	ctx, _, endObservation := s.operations.getRepositoriesForIndexScan.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Bool("allowGlobalPolicies", allowGlobalPolicies),
		attribute.Int("repositoryMatchLimit", repositoryMatchLimitValue),
		attribute.Int("limit", limit),
	}})
	defer endObservation(1, observation.Args{})

	queries := make([]*sqlf.Query, 0, 3)
	if allowGlobalPolicies {
		queries = append(queries, sqlf.Sprintf(
			getRepositoriesForIndexScanGlobalRepositoriesQuery,
			optionalLimit(repositoryMatchLimit),
		))
	}
	queries = append(queries, sqlf.Sprintf(getRepositoriesForIndexScanRepositoriesWithPolicyQuery))
	queries = append(queries, sqlf.Sprintf(getRepositoriesForIndexScanRepositoriesWithPolicyViaPatternQuery))

	for i, query := range queries {
		queries[i] = sqlf.Sprintf("(%s)", query)
	}

	return basestore.ScanInts(s.db.Query(ctx, sqlf.Sprintf(
		getRepositoriesForIndexScanQuery,
		sqlf.Join(queries, " UNION ALL "),
		now,
		int(processDelay/time.Second),
		limit,
		now,
		now,
	)))
}

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
	LEFT JOIN lsif_last_index_scan lrs ON lrs.repository_id = rmp.id
	WHERE
		-- Records that have not been checked within the global reindex threshold are also eligible for
		-- indexing. Note that condition here is true for a record that has never been indexed.
		(%s - lrs.last_index_scan_at > (%s * '1 second'::interval)) IS DISTINCT FROM FALSE OR

		-- Records that have received an update since their last scan are also eligible for re-indexing.
		-- Note that last_changed is NULL unless the repository is attached to a policy for HEAD.
		(rmp.last_changed > lrs.last_index_scan_at)
	ORDER BY
		lrs.last_index_scan_at NULLS FIRST,
		rmp.id -- tie breaker
	LIMIT %s
)
INSERT INTO lsif_last_index_scan (repository_id, last_index_scan_at)
SELECT DISTINCT r.id, %s::timestamp FROM repositories r
ON CONFLICT (repository_id) DO UPDATE
SET last_index_scan_at = %s
RETURNING repository_id
`

const getRepositoriesForIndexScanGlobalRepositoriesQuery = `
SELECT
	r.id,
	CASE
		-- Return non-NULL last_changed only for policies that are attached to a HEAD commit.
		-- We don't want to superfluously return the same repos because they had an update, but
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
		-- We don't want to superfluously return the same repos because they had an update, but
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
		-- We don't want to superfluously return the same repos because they had an update, but
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

//
//

func (s *store) GetQueuedRepoRev(ctx context.Context, batchSize int) (_ []RepoRev, err error) {
	ctx, _, endObservation := s.operations.getQueuedRepoRev.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("batchSize", batchSize),
	}})
	defer endObservation(1, observation.Args{})

	return scanRepoRevs(s.db.Query(ctx, sqlf.Sprintf(getQueuedRepoRevQuery, batchSize)))
}

const getQueuedRepoRevQuery = `
SELECT
	id,
	repository_id,
	rev
FROM codeintel_autoindex_queue
WHERE processed_at IS NULL
ORDER BY queued_at ASC
FOR UPDATE SKIP LOCKED
LIMIT %s
`

func (s *store) MarkRepoRevsAsProcessed(ctx context.Context, ids []int) (err error) {
	ctx, _, endObservation := s.operations.markRepoRevsAsProcessed.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("numIDs", len(ids)),
	}})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(markRepoRevsAsProcessedQuery, pq.Array(ids)))
}

const markRepoRevsAsProcessedQuery = `
UPDATE codeintel_autoindex_queue
SET processed_at = NOW()
WHERE id = ANY(%s)
`

//
//

func scanRepoRev(s dbutil.Scanner) (rr RepoRev, err error) {
	err = s.Scan(&rr.ID, &rr.RepositoryID, &rr.Rev)
	return rr, err
}

var scanRepoRevs = basestore.NewSliceScanner(scanRepoRev)
