package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// HardDeleteUploadsByIDs deletes the upload record with the given identifier.
func (s *store) HardDeleteUploadsByIDs(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservation := s.operations.hardDeleteUploadsByIDs.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("numIDs", len(ids)),
		attribute.IntSlice("ids", ids),
	}})
	defer endObservation(1, observation.Args{})

	if len(ids) == 0 {
		return nil
	}

	var idQueries []*sqlf.Query
	for _, id := range ids {
		idQueries = append(idQueries, sqlf.Sprintf("%s", id))
	}

	return s.db.Exec(ctx, sqlf.Sprintf(hardDeleteUploadsByIDsQuery, sqlf.Join(idQueries, ", ")))
}

const hardDeleteUploadsByIDsQuery = `
WITH
locked_uploads AS (
	SELECT u.id, u.associated_index_id
	FROM lsif_uploads u
	WHERE u.id IN (%s)
	ORDER BY u.id FOR UPDATE
),
delete_uploads AS (
	DELETE FROM lsif_uploads WHERE id IN (SELECT id FROM locked_uploads)
),
locked_indexes AS (
	SELECT u.id
	FROM lsif_indexes U
	WHERE u.id IN (SELECT associated_index_id FROM locked_uploads)
	ORDER BY u.id FOR UPDATE
)
DELETE FROM lsif_indexes WHERE id IN (SELECT id FROM locked_indexes)
`

// DeleteUploadsStuckUploading soft deletes any upload record that has been uploading since the given time.
func (s *store) DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (_, _ int, err error) {
	ctx, trace, endObservation := s.operations.deleteUploadsStuckUploading.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("uploadedBefore", uploadedBefore.Format(time.RFC3339)), // TODO - should be a duration
	}})
	defer endObservation(1, observation.Args{})

	unset, _ := s.db.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "stuck in uploading state")
	defer unset(ctx)

	query := sqlf.Sprintf(deleteUploadsStuckUploadingQuery, uploadedBefore)
	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, query))
	if err != nil {
		return 0, 0, err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("count", count))

	return count, count, nil
}

const deleteUploadsStuckUploadingQuery = `
WITH
candidates AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE u.state = 'uploading' AND u.uploaded_at < %s

	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_uploads table.
	ORDER BY u.id FOR UPDATE
),
deleted AS (
	UPDATE lsif_uploads u
	SET state = 'deleted'
	WHERE id IN (SELECT id FROM candidates)
	RETURNING u.repository_id
)
SELECT COUNT(*) FROM deleted
`

// deletedRepositoryGracePeriod is the minimum allowable duration between a repo deletion
// and the upload and index records for that repository being deleted.
const deletedRepositoryGracePeriod = time.Minute * 30

// DeleteUploadsWithoutRepository deletes uploads associated with repositories that were deleted at least
// DeletedRepositoryGracePeriod ago. This returns the repository identifier mapped to the number of uploads
// that were removed for that repository.
func (s *store) DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (_, _ int, err error) {
	ctx, trace, endObservation := s.operations.deleteUploadsWithoutRepository.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	var a, b int
	err = s.withTransaction(ctx, func(tx *store) error {
		unset, _ := tx.db.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "upload associated with repository not known to this instance")
		defer unset(ctx)

		query := sqlf.Sprintf(deleteUploadsWithoutRepositoryQuery, now.UTC(), deletedRepositoryGracePeriod/time.Second)
		totalCount, repositories, err := scanCountsWithTotalCount(tx.db.Query(ctx, query))
		if err != nil {
			return err
		}

		count := 0
		for _, numDeleted := range repositories {
			count += numDeleted
		}
		trace.AddEvent("TODO Domain Owner",
			attribute.Int("count", count),
			attribute.Int("numRepositories", len(repositories)))

		a = totalCount
		b = count
		return nil
	})
	return a, b, err
}

const deleteUploadsWithoutRepositoryQuery = `
WITH
candidates AS (
	SELECT u.id
	FROM repo r
	JOIN lsif_uploads u ON u.repository_id = r.id
	WHERE
		%s - r.deleted_at >= %s * interval '1 second' OR
		r.blocked IS NOT NULL

	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_uploads table.
	ORDER BY u.id FOR UPDATE
),
deleted AS (
	-- Note: we can go straight from completed -> deleted here as we
	-- do not need to preserve the deleted repository's current commit
	-- graph (the API cannot resolve any queries for this repository).

	UPDATE lsif_uploads u
	SET state = 'deleted'
	WHERE u.id IN (SELECT id FROM candidates)
	RETURNING u.id, u.repository_id
)
SELECT (SELECT COUNT(*) FROM candidates), d.repository_id, COUNT(*) FROM deleted d GROUP BY d.repository_id
`

// DeleteOldAuditLogs removes lsif_upload audit log records older than the given max age.
func (s *store) DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (_, _ int, err error) {
	ctx, _, endObservation := s.operations.deleteOldAuditLogs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	query := sqlf.Sprintf(deleteOldAuditLogsQuery, now, int(maxAge/time.Second))
	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, query))
	return count, count, err
}

const deleteOldAuditLogsQuery = `
WITH deleted AS (
	DELETE FROM lsif_uploads_audit_logs
	WHERE %s - log_timestamp > (%s * '1 second'::interval)
	RETURNING upload_id
)
SELECT count(*) FROM deleted
`

func (s *store) ReconcileCandidates(ctx context.Context, batchSize int) (_ []int, err error) {
	return basestore.ScanInts(s.db.Query(ctx, sqlf.Sprintf(reconcileQuery, batchSize)))
}

const reconcileQuery = `
WITH
candidates AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE u.state = 'completed'
	ORDER BY u.last_reconcile_at DESC NULLS FIRST, u.id
	LIMIT %s
),
locked_candidates AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE id = ANY(SELECT id FROM candidates)
	ORDER BY u.id
	FOR UPDATE
)
UPDATE lsif_uploads
SET last_reconcile_at = NOW()
WHERE id = ANY(SELECT id FROM locked_candidates)
RETURNING id
`

func (s *store) ProcessStaleSourcedCommits(
	ctx context.Context,
	minimumTimeSinceLastCheck time.Duration,
	commitResolverBatchSize int,
	_ time.Duration,
	shouldDelete func(ctx context.Context, repositoryID int, repositoryName, commit string) (bool, error),
) (int, int, error) {
	return s.processStaleSourcedCommits(ctx, minimumTimeSinceLastCheck, commitResolverBatchSize, shouldDelete, time.Now())
}

func (s *store) processStaleSourcedCommits(
	ctx context.Context,
	minimumTimeSinceLastCheck time.Duration,
	commitResolverBatchSize int,
	shouldDelete func(ctx context.Context, repositoryID int, repositoryName, commit string) (bool, error),
	now time.Time,
) (totalScanned, totalDeleted int, err error) {
	ctx, _, endObservation := s.operations.processStaleSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	var a, b int
	err = s.withTransaction(ctx, func(tx *store) error {
		now = now.UTC()
		interval := int(minimumTimeSinceLastCheck / time.Second)

		staleIndexes, err := scanSourcedCommits(tx.db.Query(ctx, sqlf.Sprintf(
			staleIndexSourcedCommitsQuery,
			now,
			interval,
			commitResolverBatchSize,
		)))
		if err != nil {
			return err
		}

		for _, sc := range staleIndexes {
			var (
				keep   []string
				remove []string
			)

			for _, commit := range sc.Commits {
				if ok, err := shouldDelete(ctx, sc.RepositoryID, sc.RepositoryName, commit); err != nil {
					return err
				} else if ok {
					remove = append(remove, commit)
				} else {
					keep = append(keep, commit)
				}
			}

			unset, _ := tx.db.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "upload associated with unknown commit")
			defer unset(ctx)

			indexesDeleted, _, err := basestore.ScanFirstInt(tx.db.Query(ctx, sqlf.Sprintf(
				updateSourcedCommitsQuery2,
				sc.RepositoryID,
				pq.Array(keep),
				pq.Array(remove),
				now,
				pq.Array(keep),
				pq.Array(remove),
			)))
			if err != nil {
				return err
			}

			totalDeleted += indexesDeleted
		}

		a = len(staleIndexes)
		b = totalDeleted
		return nil
	})
	return a, b, err
}

const staleIndexSourcedCommitsQuery = `
WITH candidates AS (
	SELECT
		repository_id,
		commit,
		-- Keep track of the most recent update of this commit that we know about
		-- as any earlier dates for the same repository and commit pair carry no
		-- useful information.
		MAX(commit_last_checked_at) as max_last_checked_at
	FROM lsif_indexes
	WHERE
		-- Ignore records already marked as deleted
		state NOT IN ('deleted', 'deleting') AND
		-- Ignore records that have been checked recently. Note this condition is
		-- true for a null commit_last_checked_at (which has never been checked).
		(%s - commit_last_checked_at > (%s * '1 second'::interval)) IS DISTINCT FROM FALSE
	GROUP BY repository_id, commit
)
SELECT r.id, r.name, c.commit
FROM candidates c
JOIN repo r ON r.id = c.repository_id
-- Order results so that the repositories with the commits that have been updated
-- the least frequently come first. Once a number of commits are processed from a
-- given repository the ordering may change.
ORDER BY MIN(c.max_last_checked_at) OVER (PARTITION BY c.repository_id), c.commit
LIMIT %s
`

const updateSourcedCommitsQuery2 = `
WITH
candidate_indexes AS (
	SELECT u.id
	FROM lsif_indexes u
	WHERE
		u.repository_id = %s AND
		(
			u.commit = ANY(%s) OR
			u.commit = ANY(%s)
		)

	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_indexes table.
	ORDER BY u.id FOR UPDATE
),
update_indexes AS (
	UPDATE lsif_indexes
	SET commit_last_checked_at = %s
	WHERE id IN (SELECT id FROM candidate_indexes WHERE commit = ANY(%s))
	RETURNING 1
),
delete_indexes AS (
	DELETE FROM lsif_indexes
	WHERE id IN (SELECT id FROM candidate_indexes WHERE commit = ANY(%s))
	RETURNING 1
)
SELECT COUNT(*) FROM delete_indexes
`

// DeleteIndexesWithoutRepository deletes indexes associated with repositories that were deleted at least
// DeletedRepositoryGracePeriod ago. This returns the repository identifier mapped to the number of indexes
// that were removed for that repository.
func (s *store) DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (totalCount int, deletedCount int, err error) {
	ctx, trace, endObservation := s.operations.deleteIndexesWithoutRepository.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	var a, b int
	err = s.withTransaction(ctx, func(tx *store) error {
		// TODO(efritz) - this would benefit from an index on repository_id. We currently have
		// a similar one on this index, but only for uploads that are completed or visible at tip.
		totalCount, repositories, err := scanCountsAndTotalCount(tx.db.Query(ctx, sqlf.Sprintf(deleteIndexesWithoutRepositoryQuery, now.UTC(), deletedRepositoryGracePeriod/time.Second)))
		if err != nil {
			return err
		}

		count := 0
		for _, numDeleted := range repositories {
			count += numDeleted
		}
		trace.AddEvent("scanCounts",
			attribute.Int("count", count),
			attribute.Int("numRepositories", len(repositories)))

		a = totalCount
		b = count
		return nil
	})
	return a, b, err
}

const deleteIndexesWithoutRepositoryQuery = `
WITH
candidates AS (
	SELECT u.id
	FROM repo r
	JOIN lsif_indexes u ON u.repository_id = r.id
	WHERE
		%s - r.deleted_at >= %s * interval '1 second' OR
		r.blocked IS NOT NULL

	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_indexes table.
	ORDER BY u.id FOR UPDATE
),
deleted AS (
	DELETE FROM lsif_indexes u
	WHERE id IN (SELECT id FROM candidates)
	RETURNING u.id, u.repository_id
)
SELECT (SELECT COUNT(*) FROM candidates), d.repository_id, COUNT(*) FROM deleted d GROUP BY d.repository_id
`

// ExpireFailedRecords removes autoindexing job records that meet the following conditions:
//
//   - The record is in the "failed" state
//   - The time between the job finishing and the current timestamp exceeds the given max age
//   - It is not the most recent-to-finish failure for the same repo, root, and indexer values
//     **unless** there is a more recent success.
func (s *store) ExpireFailedRecords(ctx context.Context, batchSize int, failedIndexMaxAge time.Duration, now time.Time) (_, _ int, err error) {
	ctx, _, endObservation := s.operations.expireFailedRecords.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(expireFailedRecordsQuery, now, int(failedIndexMaxAge/time.Second), batchSize))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var c1, c2 int
	for rows.Next() {
		if err := rows.Scan(&c1, &c2); err != nil {
			return 0, 0, err
		}
	}

	return c1, c2, nil
}

const expireFailedRecordsQuery = `
WITH
ranked_indexes AS (
	SELECT
		u.*,
		RANK() OVER (
			PARTITION BY
				repository_id,
				root,
				indexer
			ORDER BY
				finished_at DESC
		) AS rank
	FROM lsif_indexes u
	WHERE
		u.state = 'failed' AND
		%s - u.finished_at >= %s * interval '1 second'
),
locked_indexes AS (
	SELECT i.id
	FROM lsif_indexes i
	JOIN ranked_indexes ri ON ri.id = i.id

	-- We either select ranked indexes that have a rank > 1, meaning
	-- there's another more recent failure in this "pipeline" that has
	-- relevant information to debug the failure.
	--
	-- If we have rank = 1, but there's a newer SUCCESSFUL record for
	-- the same "pipeline", then we can say that this failure information
	-- is no longer relevant.

	WHERE ri.rank != 1 OR EXISTS (
		SELECT 1
		FROM lsif_indexes i2
		WHERE
			i2.state = 'completed' AND
			i2.finished_at > i.finished_at AND
			i2.repository_id = i.repository_id AND
			i2.root = i.root AND
			i2.indexer = i.indexer
	)
	ORDER BY i.id
	FOR UPDATE SKIP LOCKED
	LIMIT %d
),
del AS (
	DELETE FROM lsif_indexes
	WHERE id IN (SELECT id FROM locked_indexes)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM ranked_indexes),
	(SELECT COUNT(*) FROM del)
`

func (s *store) ProcessSourcedCommits(
	ctx context.Context,
	minimumTimeSinceLastCheck time.Duration,
	commitResolverMaximumCommitLag time.Duration,
	limit int,
	f func(ctx context.Context, repositoryID int, repositoryName, commit string) (bool, error),
	now time.Time,
) (_, _ int, err error) {
	sourcedUploads, err := s.GetStaleSourcedCommits(ctx, minimumTimeSinceLastCheck, limit, now)
	if err != nil {
		return 0, 0, err
	}

	numDeleted := 0
	numCommits := 0
	for _, sc := range sourcedUploads {
		for _, commit := range sc.Commits {
			numCommits++

			shouldDelete, err := f(ctx, sc.RepositoryID, sc.RepositoryName, commit)
			if err != nil {
				return 0, 0, err
			}

			if shouldDelete {
				_, uploadsDeleted, err := s.DeleteSourcedCommits(ctx, sc.RepositoryID, commit, commitResolverMaximumCommitLag, now)
				if err != nil {
					return 0, 0, err
				}

				numDeleted += uploadsDeleted
			}

			if _, err := s.UpdateSourcedCommits(ctx, sc.RepositoryID, commit, now); err != nil {
				return 0, 0, err
			}
		}
	}

	return numCommits, numDeleted, nil
}

//
//

// GetStaleSourcedCommits returns a set of commits attached to repositories that have been
// least recently checked for resolvability via gitserver. We do this periodically in
// order to determine which records in the database are unreachable by normal query
// paths and clean up that occupied (but useless) space. The output is of this method is
// ordered by repository ID then by commit.
func (s *store) GetStaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []SourcedCommits, err error) {
	ctx, trace, endObservation := s.operations.getStaleSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	var a []SourcedCommits
	err = s.withTransaction(ctx, func(tx *store) error {
		now = now.UTC()
		interval := int(minimumTimeSinceLastCheck / time.Second)
		uploadSubquery := sqlf.Sprintf(staleSourcedCommitsSubquery, now, interval)
		query := sqlf.Sprintf(staleSourcedCommitsQuery, uploadSubquery, limit)

		sourcedCommits, err := scanSourcedCommits(tx.db.Query(ctx, query))
		if err != nil {
			return err
		}

		numCommits := 0
		for _, commits := range sourcedCommits {
			numCommits += len(commits.Commits)
		}
		trace.AddEvent("TODO Domain Owner",
			attribute.Int("numRepositories", len(sourcedCommits)),
			attribute.Int("numCommits", numCommits))

		a = sourcedCommits
		return nil
	})
	return a, err
}

const staleSourcedCommitsQuery = `
WITH
	candidates AS (%s)
SELECT r.id, r.name, c.commit
FROM candidates c
JOIN repo r ON r.id = c.repository_id
-- Order results so that the repositories with the commits that have been updated
-- the least frequently come first. Once a number of commits are processed from a
-- given repository the ordering may change.
ORDER BY MIN(c.max_last_checked_at) OVER (PARTITION BY c.repository_id), c.commit
LIMIT %s
`

const staleSourcedCommitsSubquery = `
SELECT
	repository_id,
	commit,
	-- Keep track of the most recent update of this commit that we know about
	-- as any earlier dates for the same repository and commit pair carry no
	-- useful information.
	MAX(commit_last_checked_at) as max_last_checked_at
FROM lsif_uploads
WHERE
	-- Ignore records already marked as deleted
	state NOT IN ('deleted', 'deleting') AND
	-- Ignore records that have been checked recently. Note this condition is
	-- true for a null commit_last_checked_at (which has never been checked).
	(%s - commit_last_checked_at > (%s * '1 second'::interval)) IS DISTINCT FROM FALSE
GROUP BY repository_id, commit
`

// UpdateSourcedCommits updates the commit_last_checked_at field of each upload records belonging to
// the given repository identifier and commit. This method returns the count of upload records modified
func (s *store) UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (uploadsUpdated int, err error) {
	ctx, trace, endObservation := s.operations.updateSourcedCommits.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
		attribute.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	candidateUploadsSubquery := sqlf.Sprintf(candidateUploadsCTE, repositoryID, commit)
	updateSourcedCommitQuery := sqlf.Sprintf(updateSourcedCommitsQuery, candidateUploadsSubquery, now)

	uploadsUpdated, err = scanCount(s.db.Query(ctx, updateSourcedCommitQuery))
	if err != nil {
		return 0, err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("uploadsUpdated", uploadsUpdated))

	return uploadsUpdated, nil
}

const updateSourcedCommitsQuery = `
WITH
candidate_uploads AS (%s),
update_uploads AS (
	UPDATE lsif_uploads u
	SET commit_last_checked_at = %s
	WHERE id IN (SELECT id FROM candidate_uploads)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM update_uploads) AS num_uploads
`

const candidateUploadsCTE = `
SELECT u.id, u.state, u.uploaded_at
FROM lsif_uploads u
WHERE u.repository_id = %s AND u.commit = %s

-- Lock these rows in a deterministic order so that we don't
-- deadlock with other processes updating the lsif_uploads table.
ORDER BY u.id FOR UPDATE
`

func scanCount(rows *sql.Rows, queryErr error) (value int, err error) {
	if queryErr != nil {
		return 0, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scan(&value); err != nil {
			return 0, err
		}
	}

	return value, nil
}

// DeleteSourcedCommits deletes each upload record belonging to the given repository identifier
// and commit. Uploads are soft deleted. This method returns the count of upload modified.
//
// If a maximum commit lag is supplied, then any upload records in the uploading, queued, or processing states
// younger than the provided lag will not be deleted, but its timestamp will be modified as if the sibling method
// UpdateSourcedCommits was called instead. This configurable parameter enables support for remote code hosts
// that are not the source of truth; if we deleted all pending records without resolvable commits introduce races
// between the customer's Sourcegraph instance and their CI (and their CI will usually win).
func (s *store) DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (
	uploadsUpdated, uploadsDeleted int,
	err error,
) {
	ctx, trace, endObservation := s.operations.deleteSourcedCommits.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
		attribute.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	unset, _ := s.db.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "upload associated with unknown commit")
	defer unset(ctx)

	now = now.UTC()
	interval := int(maximumCommitLag / time.Second)

	candidateUploadsSubquery := sqlf.Sprintf(candidateUploadsCTE, repositoryID, commit)
	taggedCandidateUploadsSubquery := sqlf.Sprintf(taggedCandidateUploadsCTE, now, interval)
	deleteSourcedCommitsQuery := sqlf.Sprintf(deleteSourcedCommitsQuery, candidateUploadsSubquery, taggedCandidateUploadsSubquery, now)

	uploadsUpdated, uploadsDeleted, err = scanPairOfCounts(s.db.Query(ctx, deleteSourcedCommitsQuery))
	if err != nil {
		return 0, 0, err
	}
	trace.AddEvent("TODO Domain Owner",
		attribute.Int("uploadsUpdated", uploadsUpdated),
		attribute.Int("uploadsDeleted", uploadsDeleted))

	return uploadsUpdated, uploadsDeleted, nil
}

const deleteSourcedCommitsQuery = `
WITH
candidate_uploads AS (%s),
tagged_candidate_uploads AS (%s),
update_uploads AS (
	UPDATE lsif_uploads u
	SET commit_last_checked_at = %s
	WHERE EXISTS (SELECT 1 FROM tagged_candidate_uploads tu WHERE tu.id = u.id AND tu.protected)
	RETURNING 1
),
delete_uploads AS (
	UPDATE lsif_uploads u
	SET state = CASE WHEN u.state = 'completed' THEN 'deleting' ELSE 'deleted' END
	WHERE EXISTS (SELECT 1 FROM tagged_candidate_uploads tu WHERE tu.id = u.id AND NOT tu.protected)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM update_uploads) AS num_uploads_updated,
	(SELECT COUNT(*) FROM delete_uploads) AS num_uploads_deleted
`

const taggedCandidateUploadsCTE = `
SELECT
	u.*,
	(u.state IN ('uploading', 'queued', 'processing') AND %s - u.uploaded_at <= (%s * '1 second'::interval)) AS protected
FROM candidate_uploads u
`

func scanPairOfCounts(rows *sql.Rows, queryErr error) (value1, value2 int, err error) {
	if queryErr != nil {
		return 0, 0, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scan(&value1, &value2); err != nil {
			return 0, 0, err
		}
	}

	return value1, value2, nil
}

//
//

func scanCountsAndTotalCount(rows *sql.Rows, queryErr error) (totalCount int, _ map[int]int, err error) {
	if queryErr != nil {
		return 0, nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	visibilities := map[int]int{}
	for rows.Next() {
		var id int
		var count int
		if err := rows.Scan(&totalCount, &id, &count); err != nil {
			return 0, nil, err
		}

		visibilities[id] = count
	}

	return totalCount, visibilities, nil
}
