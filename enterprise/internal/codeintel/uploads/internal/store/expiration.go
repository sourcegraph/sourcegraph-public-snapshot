package store

import (
	"context"
	"os"
	"sort"
	"time"

	"github.com/keegancsmith/sqlf"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

// GetLastUploadRetentionScanForRepository returns the last timestamp, if any, that the repository with the
// given identifier was considered for upload expiration checks.
func (s *store) GetLastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error) {
	ctx, _, endObservation := s.operations.getLastUploadRetentionScanForRepository.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	t, ok, err := basestore.ScanFirstTime(s.db.Query(ctx, sqlf.Sprintf(lastUploadRetentionScanForRepositoryQuery, repositoryID)))
	if !ok {
		return nil, err
	}

	return &t, nil
}

const lastUploadRetentionScanForRepositoryQuery = `
SELECT last_retention_scan_at FROM lsif_last_retention_scan WHERE repository_id = %s
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

// UpdateUploadRetention updates the last data retention scan timestamp on the upload
// records with the given protected identifiers and sets the expired field on the upload
// records with the given expired identifiers.
func (s *store) UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) (err error) {
	ctx, _, endObservation := s.operations.updateUploadRetention.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("numProtectedIDs", len(protectedIDs)),
		attribute.IntSlice("protectedIDs", protectedIDs),
		attribute.Int("numExpiredIDs", len(expiredIDs)),
		attribute.IntSlice("expiredIDs", expiredIDs),
	}})
	defer endObservation(1, observation.Args{})

	// Ensure ids are sorted so that we take row locks during the UPDATE
	// query in a determinstic order. This should prevent deadlocks with
	// other queries that mass update lsif_uploads.
	sort.Ints(protectedIDs)
	sort.Ints(expiredIDs)

	return s.withTransaction(ctx, func(tx *store) error {
		now := time.Now()
		if len(protectedIDs) > 0 {
			queries := make([]*sqlf.Query, 0, len(protectedIDs))
			for _, id := range protectedIDs {
				queries = append(queries, sqlf.Sprintf("%s", id))
			}

			if err := tx.db.Exec(ctx, sqlf.Sprintf(updateUploadRetentionQuery, sqlf.Sprintf("last_retention_scan_at = %s", now), sqlf.Join(queries, ","))); err != nil {
				return err
			}
		}

		if len(expiredIDs) > 0 {
			queries := make([]*sqlf.Query, 0, len(expiredIDs))
			for _, id := range expiredIDs {
				queries = append(queries, sqlf.Sprintf("%s", id))
			}

			if err := tx.db.Exec(ctx, sqlf.Sprintf(updateUploadRetentionQuery, sqlf.Sprintf("expired = TRUE"), sqlf.Join(queries, ","))); err != nil {
				return err
			}
		}

		return nil
	})
}

const updateUploadRetentionQuery = `
UPDATE lsif_uploads SET %s WHERE id IN (%s)
`

// SoftDeleteExpiredUploads marks upload records that are both expired and have no references
// as deleted. The associated repositories will be marked as dirty so that their commit graphs
// are updated in the near future.
func (s *store) SoftDeleteExpiredUploads(ctx context.Context, batchSize int) (_, _ int, err error) {
	ctx, trace, endObservation := s.operations.softDeleteExpiredUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	var a, b int
	err = s.withTransaction(ctx, func(tx *store) error {
		// Just in case
		if os.Getenv("DEBUG_PRECISE_CODE_INTEL_SOFT_DELETE_BAIL_OUT") != "" {
			s.logger.Warn("Soft deletion is currently disabled")
			return nil
		}

		unset, _ := tx.db.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "soft-deleting expired uploads")
		defer unset(ctx)
		scannedCount, repositories, err := scanCountsWithTotalCount(tx.db.Query(ctx, sqlf.Sprintf(softDeleteExpiredUploadsQuery, batchSize)))
		if err != nil {
			return err
		}

		count := 0
		for _, numUpdated := range repositories {
			count += numUpdated
		}
		trace.AddEvent("TODO Domain Owner",
			attribute.Int("count", count),
			attribute.Int("numRepositories", len(repositories)))

		for repositoryID := range repositories {
			if err := s.setRepositoryAsDirtyWithTx(ctx, repositoryID, tx.db); err != nil {
				return err
			}
		}

		a = scannedCount
		b = count
		return nil
	})
	return a, b, err
}

const softDeleteExpiredUploadsQuery = `
WITH

-- First, select the set of uploads that are not protected by any policy. This will
-- be the set that we _may_ soft-delete due to age, as long as it's unreferenced by
-- any other upload that canonically provides some package. The following CTES will
-- handle the "unreferenced" part of that condition.
expired_uploads AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE u.state = 'completed' AND u.expired
	ORDER BY u.last_referenced_scan_at NULLS FIRST, u.finished_at, u.id
	LIMIT %s
),

-- From the set of unprotected uploads, find the set of packages they provide.
packages_defined_by_target_uploads AS (
	SELECT p.scheme, p.manager, p.name, p.version
	FROM lsif_packages p
	WHERE p.dump_id IN (SELECT id FROM expired_uploads)
),

-- From the set of provided packages, find the entire set of uploads that provide those
-- packages. This will necessarily include the set of target uploads above, as well as
-- any other uploads that happen to define the same package (including version). This
-- result set also includes a _rank_ column, where rank = 1 indicates that the upload
-- canonically provides that package and will be visible in cross-index navigation for
-- that package.
ranked_uploads_providing_packages AS (
	SELECT
		u.id,
		p.scheme,
		p.manager,
		p.name,
		p.version,
		-- Rank each upload providing the same package from the same directory
		-- within a repository by commit date. We'll choose the oldest commit
		-- date as the canonical choice, and set the reference counts to all
		-- of the duplicate commits to zero.
		` + packageRankingQueryFragment + ` AS rank
	FROM lsif_uploads u
	LEFT JOIN lsif_packages p ON p.dump_id = u.id
	WHERE
		(
			-- Select our target uploads
			u.id = ANY (SELECT id FROM expired_uploads) OR

			-- Also select uploads that provide the same package as a target upload.
			(p.scheme, p.manager, p.name, p.version) IN (
				SELECT p.scheme, p.manager, p.name, p.version
				FROM packages_defined_by_target_uploads p
			)
		) AND

		-- Don't match deleted uploads
		u.state = 'completed'
),

-- Filter the set of our original (expired) candidate uploads so that it includes only
-- uploads that canonically provide a referenced package. In the candidate set below,
-- we will select all of the expired uploads that do NOT appear in this result set.
referenced_uploads_providing_package_canonically AS (
	SELECT ru.id
	FROM ranked_uploads_providing_packages ru
	WHERE
		-- Only select from our original set (not the larger intermediate ones)
		ru.id IN (SELECT id FROM expired_uploads) AND

		-- Only select canonical package providers
		ru.rank = 1 AND

		-- Only select packages with non-zero references
		EXISTS (
			SELECT 1
			FROM lsif_references r
			WHERE
				r.scheme = ru.scheme AND
				r.manager = ru.manager AND
				r.name = ru.name AND
				r.version = ru.version AND
				r.dump_id != ru.id
			)
),

-- Filter the set of our original candidate uploads to exclude the "safe" uploads found
-- above. This should include uploads that are expired and either not a canonical provider
-- of their package, or their package is unreferenced by any other upload. We can then lock
-- the uploads in a deterministic order and update the state of each upload to 'deleting'.
-- Before hard-deletion, we will clear all associated data for this upload in the codeintel-db.
candidates AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE
		u.id IN (SELECT id FROM expired_uploads) AND
		NOT EXISTS (
			SELECT 1
			FROM referenced_uploads_providing_package_canonically pkg_refcount
			WHERE pkg_refcount.id = u.id
		)
),
locked_uploads AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE u.id IN (SELECT id FROM expired_uploads)
	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_uploads table.
	ORDER BY u.id FOR UPDATE
),
updated AS (
	UPDATE lsif_uploads u

	SET
		-- Update this value unconditionally
		last_referenced_scan_at = NOW(),

		-- Delete the candidates we've identified, but keep the state the same for all other uploads
		state = CASE WHEN u.id IN (SELECT id FROM candidates) THEN 'deleting' ELSE 'completed' END
	WHERE u.id IN (SELECT id FROM locked_uploads)
	RETURNING u.id, u.repository_id, u.state
)

-- Return the repositories which were affected so we can recalculate the commit graph
SELECT (SELECT COUNT(*) FROM expired_uploads), u.repository_id, COUNT(*) FROM updated u WHERE u.state = 'deleting' GROUP BY u.repository_id
`

// SoftDeleteExpiredUploadsViaTraversal selects an expired upload and uses that as the starting
// point for a backwards traversal through the reference graph. If all reachable uploads are expired,
// then the entire set of reachable uploads can be soft-deleted. Otherwise, each of the uploads we
// found during the traversal are accessible by some "live" upload and must be retained.
//
// We set a last-checked timestamp to attempt to round-robin this graph traversal.
func (s *store) SoftDeleteExpiredUploadsViaTraversal(ctx context.Context, traversalLimit int) (_, _ int, err error) {
	ctx, trace, endObservation := s.operations.softDeleteExpiredUploadsViaTraversal.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	var a, b int
	err = s.withTransaction(ctx, func(tx *store) error {
		unset, _ := tx.db.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "soft-deleting expired uploads (via graph traversal)")
		defer unset(ctx)
		scannedCount, repositories, err := scanCountsWithTotalCount(tx.db.Query(ctx, sqlf.Sprintf(
			softDeleteExpiredUploadsViaTraversalQuery,
			traversalLimit,
			traversalLimit,
		)))
		if err != nil {
			return err
		}

		count := 0
		for _, numUpdated := range repositories {
			count += numUpdated
		}
		trace.AddEvent("TODO Domain Owner",
			attribute.Int("count", count),
			attribute.Int("numRepositories", len(repositories)))

		for repositoryID := range repositories {
			if err := s.setRepositoryAsDirtyWithTx(ctx, repositoryID, tx.db); err != nil {
				return err
			}
		}

		a = scannedCount
		b = count
		return nil
	})
	return a, b, err

}

const softDeleteExpiredUploadsViaTraversalQuery = `
WITH RECURSIVE

-- First, select a single root upload from which we will perform a traversal through
-- its dependents. Our goal is to find the set of transitive dependents that terminate
-- at our chosen root. If all the uploads reached on this traversal are expired, we can
-- remove the entire en masse. Otherwise, there is a non-expired upload that can reach
-- each of the traversed uploads, and we have to keep them as-is until the next check.
--
-- We choose an upload that is completed, expired, canonically provides some package.
-- If there is more than one such candidate, we choose the one that we've seen in this
-- traversal least recently.
root_upload_and_packages AS (
	SELECT * FROM (
		SELECT
			u.id,
			u.expired,
			u.last_traversal_scan_at,
			u.finished_at,
			p.scheme,
			p.manager,
			p.name,
			p.version,
			` + packageRankingQueryFragment + ` AS rank
		FROM lsif_uploads u
		LEFT JOIN lsif_packages p ON p.dump_id = u.id
		WHERE u.state = 'completed' AND u.expired
	) s

	WHERE s.rank = 1 AND EXISTS (
		SELECT 1
		FROM lsif_references r
		WHERE
			r.scheme = s.scheme AND
			r.manager = s.manager AND
			r.name = s.name AND
			r.version = s.version AND
			r.dump_id != s.id
		)
	ORDER BY s.last_traversal_scan_at NULLS FIRST, s.finished_at, s.id
	LIMIT 1
),

-- Traverse the dependency graph backwards starting from our chosen root upload. The result
-- set will include all (canonical) id and expiration status of uploads that transitively
-- depend on chosen our root.
transitive_dependents(id, expired, scheme, manager, name, version) AS MATERIALIZED (
	(
		-- Base case: select our root upload and its canonical packages
		SELECT up.id, up.expired, up.scheme, up.manager, up.name, up.version FROM root_upload_and_packages up
	) UNION (
		-- Iterative case: select new (canonical) uploads that have a direct dependency of
		-- some upload in our working set. This condition will continue to be evaluated until
		-- it reaches a fixed point, giving us the complete connected component containing our
		-- root upload.

		SELECT s.id, s.expired, s.scheme, s.manager, s.name, s.version
		FROM (
			SELECT
				u.id,
				u.expired,
				p.scheme,
				p.manager,
				p.name,
				p.version,
				` + packageRankingQueryFragment + ` AS rank
			FROM transitive_dependents d
			JOIN lsif_references r ON
				r.scheme = d.scheme AND
				r.manager = d.manager AND
				r.name = d.name AND
				r.version = d.version AND
				r.dump_id != d.id
			JOIN lsif_uploads u ON u.id = r.dump_id
			JOIN lsif_packages p ON p.dump_id = u.id
			WHERE
				u.state = 'completed' AND
				-- We don't need to continue to traverse paths that already have a non-expired
				-- upload. We can cut the search short here. Unfortuantely I don't know a good
				-- way to express that the ENTIRE traversal should stop. My attempts so far
				-- have all required an (illegal) reference to the working table in a subquery
				-- or aggregate.
				d.expired
		) s

		-- Keep only canonical package providers from the iterative step
		WHERE s.rank = 1
	)
),

-- Force evaluation of the traversal defined above, but stop searching after we've seen a given
-- number of nodes (our traversal limit). We don't want to spend unbounded time traversing a large
-- subgraph, so we cap the number of rows we'll pull from that result set. We'll handle the case
-- where we hit this limit in the update below as it would be unsafe to delete an upload based on
-- an incomplete view of its dependency graph.
candidates AS (
	SELECT * FROM transitive_dependents d
	LIMIT (%s + 1)
),
locked_uploads AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE u.id IN (SELECT id FROM candidates)
	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_uploads table.
	ORDER BY u.id FOR UPDATE
),
updated AS (
	UPDATE lsif_uploads u

	SET
		-- Update this value unconditionally
		last_traversal_scan_at = NOW(),

		-- Delete all of the upload we've traversed if and only if we've identified the entire
		-- relevant subgraph (we didn't hit our LIMIT above) and every upload of the subgraph is
		-- expired. If this is not the case, we leave the state the same for all uploads.
		state = CASE
			WHEN (SELECT bool_and(d.expired) AND COUNT(*) <= %s FROM candidates d) THEN 'deleting'
			ELSE 'completed'
		END
	WHERE u.id IN (SELECT id FROM locked_uploads)
	RETURNING u.id, u.repository_id, u.state
)

-- Return the repositories which were affected so we can recalculate the commit graph
SELECT (SELECT COUNT(*) FROM candidates), u.repository_id, COUNT(*) FROM updated u WHERE u.state = 'deleting' GROUP BY u.repository_id
`

//
//

// SetRepositoryAsDirtyWithTx marks the given repository's commit graph as out of date.
func (s *store) setRepositoryAsDirtyWithTx(ctx context.Context, repositoryID int, tx *basestore.Store) (err error) {
	ctx, _, endObservation := s.operations.setRepositoryAsDirty.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
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
