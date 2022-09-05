package store

import (
	"context"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/commitgraph"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GetUploads returns a list of uploads and the total count of records matching the given conditions.
func (s *store) GetUploads(ctx context.Context, opts shared.GetUploadsOptions) (uploads []shared.Upload, totalCount int, err error) {
	ctx, trace, endObservation := s.operations.getUploads.With(ctx, &err, observation.Args{LogFields: buildLogFields(opts)})
	defer endObservation(1, observation.Args{})

	tableExpr, conds, cte := buildConditionsAndCte(opts)
	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
	if err != nil {
		return nil, 0, err
	}
	conds = append(conds, authzConds)

	var orderExpression *sqlf.Query
	if opts.OldestFirst {
		orderExpression = sqlf.Sprintf("uploaded_at, id DESC")
	} else {
		orderExpression = sqlf.Sprintf("uploaded_at DESC, id")
	}

	query := sqlf.Sprintf(
		getUploadsQuery,
		buildCTEPrefix(cte),
		tableExpr,
		sqlf.Join(conds, " AND "),
		orderExpression,
		opts.Limit,
		opts.Offset,
	)
	uploads, totalCount, err = scanUploadsWithCount(s.db.Query(ctx, query))
	if err != nil {
		return nil, 0, err
	}
	trace.Log(
		log.Int("totalCount", totalCount),
		log.Int("numUploads", len(uploads)),
	)

	return uploads, totalCount, nil
}

const getUploadsQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:GetUploads
%s -- Dynamic CTE definitions for use in the WHERE clause
SELECT
	u.id,
	u.commit,
	u.root,
	EXISTS (` + visibleAtTipSubselectQuery + `) AS visible_at_tip,
	u.uploaded_at,
	u.state,
	u.failure_message,
	u.started_at,
	u.finished_at,
	u.process_after,
	u.num_resets,
	u.num_failures,
	u.repository_id,
	repo.name,
	u.indexer,
	u.indexer_version,
	u.num_parts,
	u.uploaded_parts,
	u.upload_size,
	u.associated_index_id,
	s.rank,
	u.uncompressed_size,
	COUNT(*) OVER() AS count
FROM %s
LEFT JOIN (` + uploadRankQueryFragment + `) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE %s ORDER BY %s LIMIT %d OFFSET %d
`

const uploadRankQueryFragment = `
SELECT
	r.id,
	ROW_NUMBER() OVER (ORDER BY COALESCE(r.process_after, r.uploaded_at), r.id) as rank
FROM lsif_uploads_with_repository_name r
WHERE r.state = 'queued'
`

const visibleAtTipSubselectQuery = `SELECT 1 FROM lsif_uploads_visible_at_tip uvt WHERE uvt.repository_id = u.repository_id AND uvt.upload_id = u.id`

const deletedUploadsFromAuditLogsCTEQuery = `
SELECT
	DISTINCT ON(s.upload_id) s.upload_id AS id, au.commit, au.root,
	au.uploaded_at, 'deleted' AS state,
	snapshot->'failure_message' AS failure_message,
	(snapshot->'started_at')::timestamptz AS started_at,
	(snapshot->'finished_at')::timestamptz AS finished_at,
	(snapshot->'process_after')::timestamptz AS process_after,
	COALESCE((snapshot->'num_resets')::integer, -1) AS num_resets,
	COALESCE((snapshot->'num_failures')::integer, -1) AS num_failures,
	au.repository_id,
	au.indexer, au.indexer_version,
	COALESCE((snapshot->'num_parts')::integer, -1) AS num_parts,
	NULL::integer[] as uploaded_parts,
	au.upload_size, au.associated_index_id,
	COALESCE((snapshot->'expired')::boolean, false) AS expired,
	NULL::bigint AS uncompressed_size
FROM (
	SELECT upload_id, snapshot_transition_columns(transition_columns ORDER BY sequence ASC) AS snapshot
	FROM lsif_uploads_audit_logs
	WHERE record_deleted_at IS NOT NULL
	GROUP BY upload_id
) AS s
JOIN lsif_uploads_audit_logs au ON au.upload_id = s.upload_id
`

const rankedDependencyCandidateCTEQuery = `
SELECT
	p.dump_id as pkg_id,
	r.dump_id as ref_id,
	-- Rank each upload providing the same package from the same directory
	-- within a repository by commit date. We'll choose the oldest commit
	-- date as the canonical choice and ignore the uploads for younger
	-- commits providing the same package.
	` + packageRankingQueryFragment + ` AS rank
FROM lsif_uploads u
JOIN lsif_packages p ON p.dump_id = u.id
JOIN lsif_references r ON r.scheme = p.scheme
	AND r.name = p.name
	AND r.version = p.version
	AND r.dump_id != p.dump_id
WHERE
	-- Don't match deleted uploads
	u.state = 'completed' AND
	%s
`

// packageRankingQueryFragment uses `lsif_uploads u` JOIN `lsif_packages p` to return a rank
// for each row grouped by package and source code location and ordered by the associated Git
// commit date.
const packageRankingQueryFragment = `
rank() OVER (
	PARTITION BY
		-- Group providers of the same package together
		p.scheme, p.name, p.version,
		-- Defined by the same directory within a repository
		u.repository_id, u.indexer, u.root
	ORDER BY
		-- Rank each grouped upload by the associated commit date
		u.committed_at,
		-- Break ties via the unique identifier
		u.id
)
`

const rankedDependentCandidateCTEQuery = `
SELECT
	p.dump_id as pkg_id,
	p.scheme as scheme,
	p.name as name,
	p.version as version,
	-- Rank each upload providing the same package from the same directory
	-- within a repository by commit date. We'll choose the oldest commit
	-- date as the canonical choice and ignore the uploads for younger
	-- commits providing the same package.
	` + packageRankingQueryFragment + ` AS rank
FROM lsif_uploads u
JOIN lsif_packages p ON p.dump_id = u.id
WHERE
	-- Don't match deleted uploads
	u.state = 'completed' AND
	%s
`

// DeletedRepositoryGracePeriod is the minimum allowable duration between a repo deletion
// and the upload and index records for that repository being deleted.
const DeletedRepositoryGracePeriod = time.Minute * 30

// DeleteUploadsWithoutRepository deletes uploads associated with repositories that were deleted at least
// DeletedRepositoryGracePeriod ago. This returns the repository identifier mapped to the number of uploads
// that were removed for that repository.
func (s *store) DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error) {
	ctx, trace, endObservation := s.operations.deleteUploadsWithoutRepository.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	unset, _ := tx.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "upload associated with repository not known to this instance")
	defer unset(ctx)

	query := sqlf.Sprintf(deleteUploadsWithoutRepositoryQuery, now.UTC(), DeletedRepositoryGracePeriod/time.Second)
	repositories, err := scanCounts(tx.Query(ctx, query))
	if err != nil {
		return nil, err
	}

	count := 0
	for _, numDeleted := range repositories {
		count += numDeleted
	}
	trace.Log(
		log.Int("count", count),
		log.Int("numRepositories", len(repositories)),
	)

	return repositories, nil
}

const deleteUploadsWithoutRepositoryQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:DeleteUploadsWithoutRepository
WITH
candidates AS (
	SELECT u.id
	FROM repo r
	JOIN lsif_uploads u ON u.repository_id = r.id
	WHERE %s - r.deleted_at >= %s * interval '1 second'

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
SELECT d.repository_id, COUNT(*) FROM deleted d GROUP BY d.repository_id
`

// DeleteUploadsStuckUploading soft deletes any upload record that has been uploading since the given time.
func (s *store) DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (_ int, err error) {
	ctx, trace, endObservation := s.operations.deleteUploadsStuckUploading.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("uploadedBefore", uploadedBefore.Format(time.RFC3339)), // TODO - should be a duration
	}})
	defer endObservation(1, observation.Args{})

	unset, _ := s.db.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "stuck in uploading state")
	defer unset(ctx)

	query := sqlf.Sprintf(deleteUploadsStuckUploadingQuery, uploadedBefore)
	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, query))
	if err != nil {
		return 0, err
	}
	trace.Log(log.Int("count", count))

	return count, nil
}

const deleteUploadsStuckUploadingQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:DeleteUploadsStuckUploading
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
SELECT count(*) FROM deleted
`

// SoftDeleteExpiredUploads marks upload records that are both expired and have no references
// as deleted. The associated repositories will be marked as dirty so that their commit graphs
// are updated in the near future.
func (s *store) SoftDeleteExpiredUploads(ctx context.Context) (count int, err error) {
	ctx, trace, endObservation := s.operations.softDeleteExpiredUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	// Just in case
	if os.Getenv("DEBUG_PRECISE_CODE_INTEL_REFERENCE_COUNTS_BAIL_OUT") != "" {
		s.logger.Warn("Reference count operations are currently disabled")
		return 0, nil
	}

	unset, _ := tx.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "soft-deleting expired uploads")
	defer unset(ctx)
	repositories, err := scanCounts(tx.Query(ctx, sqlf.Sprintf(softDeleteExpiredUploadsQuery)))
	if err != nil {
		return 0, err
	}

	for _, numUpdated := range repositories {
		count += numUpdated
	}
	trace.Log(
		log.Int("count", count),
		log.Int("numRepositories", len(repositories)),
	)

	for repositoryID := range repositories {
		if err := s.setRepositoryAsDirtyWithTx(ctx, repositoryID, tx); err != nil {
			return 0, err
		}
	}

	return count, nil
}

const softDeleteExpiredUploadsQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:SoftDeleteExpiredUploads
WITH candidates AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE u.state = 'completed' AND u.expired AND u.reference_count = 0
	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_uploads table.
	ORDER BY u.id FOR UPDATE
),
updated AS (
	UPDATE lsif_uploads u
	SET state = 'deleting'
	WHERE u.id IN (SELECT id FROM candidates)
	RETURNING u.id, u.repository_id
)
SELECT u.repository_id, count(*) FROM updated u GROUP BY u.repository_id
`

// HardDeleteUploadsByIDs deletes the upload record with the given identifier.
func (s *store) HardDeleteUploadsByIDs(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservation := s.operations.hardDeleteUploadsByIDs.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numIDs", len(ids)),
		log.String("ids", intsToString(ids)),
	}})
	defer endObservation(1, observation.Args{})

	if len(ids) == 0 {
		return nil
	}

	var idQueries []*sqlf.Query
	for _, id := range ids {
		idQueries = append(idQueries, sqlf.Sprintf("%s", id))
	}

	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.db.Done(err) }()

	// Before deleting the record, ensure that we decrease the number of existant references
	// to all of this upload's dependencies. This also selects a new upload to canonically provide
	// the same package as the deleted upload, if such an upload exists.
	if _, err := tx.UpdateUploadsReferenceCounts(ctx, ids, shared.DependencyReferenceCountUpdateTypeRemove); err != nil {
		return err
	}

	if err := tx.db.Exec(ctx, sqlf.Sprintf(hardDeleteUploadsByIDsQuery, sqlf.Join(idQueries, ", "))); err != nil {
		return err
	}

	return nil
}

const hardDeleteUploadsByIDsQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:HardDeleteUploadsByIDs
WITH locked_uploads AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE u.id IN (%s)
	ORDER BY u.id FOR UPDATE
)
DELETE FROM lsif_uploads WHERE id IN (SELECT id FROM locked_uploads)
`

// UpdateUploadRetention updates the last data retention scan timestamp on the upload
// records with the given protected identifiers and sets the expired field on the upload
// records with the given expired identifiers.
func (s *store) UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) (err error) {
	ctx, _, endObservation := s.operations.updateUploadRetention.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numProtectedIDs", len(protectedIDs)),
		log.String("protectedIDs", intsToString(protectedIDs)),
		log.Int("numExpiredIDs", len(expiredIDs)),
		log.String("expiredIDs", intsToString(expiredIDs)),
	}})
	defer endObservation(1, observation.Args{})

	// Ensure ids are sorted so that we take row locks during the UPDATE
	// query in a determinstic order. This should prevent deadlocks with
	// other queries that mass update lsif_uploads.
	sort.Ints(protectedIDs)
	sort.Ints(expiredIDs)

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	now := time.Now()
	if len(protectedIDs) > 0 {
		queries := make([]*sqlf.Query, 0, len(protectedIDs))
		for _, id := range protectedIDs {
			queries = append(queries, sqlf.Sprintf("%s", id))
		}

		if err := tx.Exec(ctx, sqlf.Sprintf(updateUploadRetentionQuery, sqlf.Sprintf("last_retention_scan_at = %s", now), sqlf.Join(queries, ","))); err != nil {
			return err
		}
	}

	if len(expiredIDs) > 0 {
		queries := make([]*sqlf.Query, 0, len(expiredIDs))
		for _, id := range expiredIDs {
			queries = append(queries, sqlf.Sprintf("%s", id))
		}

		if err := tx.Exec(ctx, sqlf.Sprintf(updateUploadRetentionQuery, sqlf.Sprintf("expired = TRUE"), sqlf.Join(queries, ","))); err != nil {
			return err
		}
	}

	return nil
}

const updateUploadRetentionQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:UpdateUploadRetention
UPDATE lsif_uploads SET %s WHERE id IN (%s)
`

// BackfillReferenceCountBatch calculates the reference count for a batch of upload records that do not
// have a set value. This method is used to backfill old upload records prior to reference counting-based
// expiration, or records that have been re-set to NULL and re-calculated (e.g., emergency resets).
func (s *store) BackfillReferenceCountBatch(ctx context.Context, batchSize int) (err error) {
	ctx, _, endObservation := s.operations.backfillReferenceCountBatch.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("batchSize", batchSize),
	}})
	defer func() { endObservation(1, observation.Args{}) }()

	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	ids, err := basestore.ScanInts(tx.db.Query(ctx, sqlf.Sprintf(backfillReferenceCountBatchQuery, batchSize)))
	if err != nil {
		return err
	}

	_, err = tx.UpdateUploadsReferenceCounts(ctx, ids, shared.DependencyReferenceCountUpdateTypeNone)
	return err
}

const backfillReferenceCountBatchQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:BackfillReferenceCountBatch
SELECT u.id
FROM lsif_uploads u
WHERE u.state = 'completed' AND u.reference_count IS NULL
ORDER BY u.id
FOR UPDATE SKIP LOCKED
LIMIT %s
`

// SourcedCommitsWithoutCommittedAt returns the repository and commits of uploads that do not have an
// associated committed_at value.
func (s *store) SourcedCommitsWithoutCommittedAt(ctx context.Context, batchSize int) (_ []shared.SourcedCommits, err error) {
	ctx, _, endObservation := s.operations.sourcedCommitsWithoutCommittedAt.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("batchSize", batchSize),
	}})
	defer func() { endObservation(1, observation.Args{}) }()

	batch, err := scanSourcedCommits(s.db.Query(ctx, sqlf.Sprintf(sourcedCommitsWithoutCommittedAtQuery, batchSize)))
	if err != nil {
		return nil, err
	}

	return batch, nil
}

const sourcedCommitsWithoutCommittedAtQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:SourcedCommitsWithoutCommittedAt
SELECT u.repository_id, r.name, u.commit
FROM lsif_uploads u
JOIN repo r ON r.id = u.repository_id
WHERE u.state = 'completed' AND u.committed_at IS NULL
GROUP BY u.repository_id, r.name, u.commit
ORDER BY repository_id, commit
LIMIT %s
`

// UpdateCommittedAt tupdates the committed_at column for upload matching the given repository and commit.
func (s *store) UpdateCommittedAt(ctx context.Context, repositoryID int, commit, commitDateString string) (err error) {
	ctx, _, endObservation := s.operations.updateCommittedAt.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
	}})
	defer func() { endObservation(1, observation.Args{}) }()

	return s.db.Exec(ctx, sqlf.Sprintf(updateCommittedAtQuery, commitDateString, repositoryID, commit))
}

const updateCommittedAtQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:UpdateCommittedAt
UPDATE lsif_uploads SET committed_at = %s WHERE state = 'completed' AND repository_id = %s AND commit = %s AND committed_at IS NULL
`

var deltaMap = map[shared.DependencyReferenceCountUpdateType]int{
	shared.DependencyReferenceCountUpdateTypeNone:   +0,
	shared.DependencyReferenceCountUpdateTypeAdd:    +1,
	shared.DependencyReferenceCountUpdateTypeRemove: -1,
}

// UpdateUploadsReferenceCounts updates the reference counts of uploads indicated by the given identifiers
// as well as the set of uploads that would be affected by one of the upload's insertion or removal.
// The behavior of this method is determined by the dependencyUpdateType value.
//
//   - Use DependencyReferenceCountUpdateTypeNone to calculate the reference count of each of the given
//     uploads without considering dependency upload counts.
//   - Use DependencyReferenceCountUpdateTypeAdd to calculate the reference count of each of the given
//     uploads while adding one to each direct dependency's reference count.
//   - Use DependencyReferenceCountUpdateTypeRemove to calculate the reference count of each of the given
//     uploads while removing one from each direct dependency's reference count.
//
// To keep reference counts consistent, this method should be called directly after insertion and directly
// before deletion of each upload record.
func (s *store) UpdateUploadsReferenceCounts(ctx context.Context, ids []int, dependencyUpdateType shared.DependencyReferenceCountUpdateType) (updated int, err error) {
	ctx, _, endObservation := s.operations.updateUploadsReferenceCounts.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numIDs", len(ids)),
		log.String("ids", intsToString(ids)),
		log.Int("dependencyUpdateType", int(dependencyUpdateType)),
	}})
	defer func() { endObservation(1, observation.Args{}) }()

	if len(ids) == 0 {
		return 0, nil
	}

	// Just in case
	if os.Getenv("DEBUG_PRECISE_CODE_INTEL_REFERENCE_COUNTS_BAIL_OUT") != "" {
		s.logger.Warn("Reference count operations are currently disabled")
		return 0, nil
	}

	idArray := pq.Array(ids)

	excludeCondition := sqlf.Sprintf("TRUE")
	if dependencyUpdateType == shared.DependencyReferenceCountUpdateTypeRemove {
		excludeCondition = sqlf.Sprintf("NOT (u.id = ANY (%s))", idArray)
	}

	result, err := s.db.ExecResult(ctx, sqlf.Sprintf(
		updateUploadsReferenceCountsQuery,
		idArray,
		idArray,
		excludeCondition,
		idArray,
		deltaMap[dependencyUpdateType],
	))
	if err != nil {
		return 0, err
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}

const updateUploadsReferenceCountsQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:UpdateReferenceCounts
WITH
-- Select the set of package identifiers provided by the target upload list. This
-- result set includes non-canonical results.
packages_defined_by_target_uploads AS (
	SELECT p.scheme, p.name, p.version
	FROM lsif_packages p
	WHERE p.dump_id = ANY (%s)
),

-- Select the ranked set of uploads that provide a package that is also provided
-- by the target upload list. This over-selects the set of uploads that visibly
-- provide a package so that we can re-rank the canonical uploads for a package
-- on the fly.
ranked_uploads_providing_packages AS (
	SELECT
		u.id,
		p.scheme,
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
			u.id = ANY (%s) OR

			-- Also select uploads that provide the same package as a target upload.
			--
			-- It is necessary to select these extra records as the insertion or
			-- deletion of an upload record can change the rank of uploads/packages.
			-- We need to ensure that we update the reference counts of every upload
			-- in this set, not just the ones that were recently inserted or deleted.
			(p.scheme, p.name, p.version) IN (
				SELECT p.scheme, p.name, p.version
				FROM packages_defined_by_target_uploads p
			)
		) AND

		-- Don't match deleted uploads. We may be dealing with uploads still in the
		-- processing state, though, so we allow those here.
		u.state NOT IN ('deleted', 'deleting') AND

		-- If we are deleting uploads that provide intelligence for a package, we need
		-- to ensure that we calculate the correct dependencies as if the records have
		-- been deleted. This condition throws out exact target uploads while keeping
		-- the (newly adjusted) ranked set of uploads providing the same package.
		(%s)
),

-- Calculate the number of references to each upload represented by the CTE
-- ranked_uploads_providing_packages. Those that are not the canonical upload
-- providing their package will have ref count of zero, by having no associated
-- row in this intermediate result set. The canonical uploads will have their
-- reference count re-calculated based on the current set of dependencies known
-- to Postgres.
canonical_package_reference_counts AS (
	SELECT
		ru.id,
		rc.count
	FROM ranked_uploads_providing_packages ru,
	LATERAL (
		SELECT
			COUNT(*) AS count
		FROM lsif_references r
		WHERE
			r.scheme = ru.scheme AND
			r.name = ru.name AND
			r.version = ru.version AND
			r.dump_id != ru.id
	) rc
	WHERE ru.rank = 1
),

-- Count (and ranks) the set of edges that cross over from the target list of uploads
-- to existing uploads that provide a dependent package. This is the modifier by which
-- dependency reference counts must be altered in order for existing package reference
-- counts to remain up-to-date.
dependency_reference_counts AS (
	SELECT
		u.id,
		` + packageRankingQueryFragment + ` AS rank,
		count(*) AS count
	FROM lsif_uploads u
	JOIN lsif_packages p ON p.dump_id = u.id
	JOIN lsif_references r
	ON
		r.scheme = p.scheme AND
		r.name = p.name AND
		r.version = p.version AND
		r.dump_id != p.dump_id
	WHERE
		-- Here we want the set of actually reachable uploads
		u.state = 'completed' AND
		r.dump_id = ANY (%s)
	GROUP BY u.id, p.scheme, p.name, p.version
),

-- Discard dependency edges to non-canonical uploads. Sum the remaining edge counts
-- to find the amount by which we need to update the reference count for the remaining
-- dependent uploads.
canonical_dependency_reference_counts AS (
	SELECT rc.id, SUM(rc.count) AS count
	FROM dependency_reference_counts rc
	WHERE rc.rank = 1
	GROUP BY rc.id
),

-- Determine the set of reference count values to write to the lsif_uploads table, then
-- lock all of the affected rows in a deterministic order. This should prevent hitting
-- deadlock conditions when multiple bulk operations are happening over intersecting
-- rows of the same table.
locked_uploads AS (
	SELECT
		u.id,

		-- If ru.id IS NOT NULL, then we have recalculated the reference count for this
		-- row in the CTE canonical_package_reference_counts. Use this value. Otherwise,
		-- this row is a dependency of the target upload list and we only be incrementally
		-- modifying the row's reference count.
		--
		CASE WHEN ru.id IS NOT NULL THEN COALESCE(pkg_refcount.count, 0) ELSE u.reference_count END +

		-- If ru.id IN canonical_dependency_reference_counts, then we incrementally modify
		-- the row's reference count proportional the number of additional dependent edges
		-- counted in the CTE. The placeholder here is an integer in the range [-1, 1] used
		-- to specify if we are adding or removing a set of upload records.
		COALESCE(dep_refcount.count, 0) * %s AS reference_count
	FROM lsif_uploads u
	LEFT JOIN ranked_uploads_providing_packages ru ON ru.id = u.id
	LEFT JOIN canonical_package_reference_counts pkg_refcount ON pkg_refcount.id = u.id
	LEFT JOIN canonical_dependency_reference_counts dep_refcount ON dep_refcount.id = u.id
	-- Prevent creating no-op updates for every row in the table
	WHERE ru.id IS NOT NULL OR dep_refcount.id IS NOT NULL
	ORDER BY u.id FOR UPDATE
)

-- Perform deterministically ordered update
UPDATE lsif_uploads u
SET reference_count = lu.reference_count
FROM locked_uploads lu WHERE lu.id = u.id
`

// UpdateUploadsVisibleToCommits uses the given commit graph and the tip of non-stale branches and tags to determine the
// set of LSIF uploads that are visible for each commit, and the set of uploads which are visible at the tip of a
// non-stale branch or tag. The decorated commit graph is serialized to Postgres for use by find closest dumps
// queries.
//
// If dirtyToken is supplied, the repository will be unmarked when the supplied token does matches the most recent
// token stored in the database, the flag will not be cleared as another request for update has come in since this
// token has been read.
func (s *store) UpdateUploadsVisibleToCommits(
	ctx context.Context,
	repositoryID int,
	commitGraph *gitdomain.CommitGraph,
	refDescriptions map[string][]gitdomain.RefDescription,
	maxAgeForNonStaleBranches time.Duration,
	maxAgeForNonStaleTags time.Duration,
	dirtyToken int,
	now time.Time,
) (err error) {
	ctx, trace, endObservation := s.operations.updateUploadsVisibleToCommits.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repositoryID),
			log.Int("numCommitGraphKeys", len(commitGraph.Order())),
			log.Int("numRefDescriptions", len(refDescriptions)),
			log.Int("dirtyToken", dirtyToken),
		},
	})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Determine the retention policy for this repository
	maxAgeForNonStaleBranches, maxAgeForNonStaleTags, err = refineRetentionConfiguration(ctx, tx, repositoryID, maxAgeForNonStaleBranches, maxAgeForNonStaleTags)
	if err != nil {
		return err
	}
	trace.Log(
		log.String("maxAgeForNonStaleBranches", maxAgeForNonStaleBranches.String()),
		log.String("maxAgeForNonStaleTags", maxAgeForNonStaleTags.String()),
	)

	// Pull all queryable upload metadata known to this repository so we can correlate
	// it with the current  commit graph.
	commitGraphView, err := scanCommitGraphView(tx.Query(ctx, sqlf.Sprintf(calculateVisibleUploadsCommitGraphQuery, repositoryID)))
	if err != nil {
		return err
	}
	trace.Log(
		log.Int("numCommitGraphViewMetaKeys", len(commitGraphView.Meta)),
		log.Int("numCommitGraphViewTokenKeys", len(commitGraphView.Tokens)),
	)

	// Determine which uploads are visible to which commits for this repository
	graph := commitgraph.NewGraph(commitGraph, commitGraphView)

	pctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Return a structure holding several channels that are populated by a background goroutine.
	// When we write this data to temporary tables, we have three consumers pulling values from
	// these channels in parallel. We need to make sure that once we return from this function that
	// the producer routine shuts down. This prevents the producer from leaking if there is an
	// error in one of the consumers before all values have been emitted.
	sanitizedInput := sanitizeCommitInput(pctx, graph, refDescriptions, maxAgeForNonStaleBranches, maxAgeForNonStaleTags)

	// Write the graph into temporary tables in Postgres
	if err := s.writeVisibleUploads(ctx, sanitizedInput, tx); err != nil {
		return err
	}

	// Persist data to permenant table: t_lsif_nearest_uploads -> lsif_nearest_uploads
	if err := s.persistNearestUploads(ctx, repositoryID, tx); err != nil {
		return err
	}

	// Persist data to permenant table: t_lsif_nearest_uploads_links -> lsif_nearest_uploads_links
	if err := s.persistNearestUploadsLinks(ctx, repositoryID, tx); err != nil {
		return err
	}

	// Persist data to permenant table: t_lsif_uploads_visible_at_tip -> lsif_uploads_visible_at_tip
	if err := s.persistUploadsVisibleAtTip(ctx, repositoryID, tx); err != nil {
		return err
	}

	if dirtyToken != 0 {
		// If the user requests us to clear a dirty token, set the updated_token value to
		// the dirty token if it wouldn't decrease the value. Dirty repositories are determined
		// by having a non-equal dirty and update token, and we want the most recent upload
		// token to win this write.
		nowTimestamp := sqlf.Sprintf("transaction_timestamp()")
		if !now.IsZero() {
			nowTimestamp = sqlf.Sprintf("%s", now)
		}
		if err := tx.Exec(ctx, sqlf.Sprintf(calculateVisibleUploadsDirtyRepositoryQuery, dirtyToken, nowTimestamp, repositoryID)); err != nil {
			return err
		}
	}

	// All completed uploads are now visible. Mark any uploads queued for deletion as deleted as
	// they are no longer reachable from the commit graph and cannot be used to fulfill any API
	// requests.
	unset, _ := tx.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "upload not reachable within the commit graph")
	defer unset(ctx)
	if err := tx.Exec(ctx, sqlf.Sprintf(calculateVisibleUploadsDeleteUploadsQueuedForDeletionQuery, repositoryID)); err != nil {
		return err
	}

	return nil
}

const calculateVisibleUploadsCommitGraphQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:UpdateUploadsVisibleToCommits
SELECT id, commit, md5(root || ':' || indexer) as token, 0 as distance FROM lsif_uploads WHERE state = 'completed' AND repository_id = %s
`

const calculateVisibleUploadsDirtyRepositoryQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:UpdateUploadsVisibleToCommits
UPDATE lsif_dirty_repositories SET update_token = GREATEST(update_token, %s), updated_at = %s WHERE repository_id = %s
`

const calculateVisibleUploadsDeleteUploadsQueuedForDeletionQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:UpdateUploadsVisibleToCommits
WITH
candidates AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE u.state = 'deleting' AND u.repository_id = %s

	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_uploads table.
	ORDER BY u.id FOR UPDATE
)
UPDATE lsif_uploads
SET state = 'deleted'
WHERE id IN (SELECT id FROM candidates)
`

// GetUploadIDsWithReferences returns uploads that probably contain an import
// or implementation moniker whose identifier matches any of the given monikers' identifiers. This method
// will not return uploads for commits which are unknown to gitserver, nor will it return uploads which
// are listed in the given ignored identifier slice. This method also returns the number of records
// scanned (but possibly filtered out from the return slice) from the database (the offset for the
// subsequent request) and the total number of records in the database.
func (s *store) GetUploadIDsWithReferences(
	ctx context.Context,
	orderedMonikers []precise.QualifiedMonikerData,
	ignoreIDs []int,
	repositoryID int,
	commit string,
	limit int,
	offset int,
	trace observation.TraceLogger,
) (ids []int, recordsScanned int, totalCount int, err error) {
	scanner, totalCount, err := s.GetVisibleUploadsMatchingMonikers(ctx, repositoryID, commit, orderedMonikers, limit, offset)
	if err != nil {
		return nil, 0, 0, errors.Wrap(err, "dbstore.ReferenceIDs")
	}

	defer func() {
		if closeErr := scanner.Close(); closeErr != nil {
			err = errors.Append(err, errors.Wrap(closeErr, "dbstore.ReferenceIDs.Close"))
		}
	}()

	ignoreIDsMap := map[int]struct{}{}
	for _, id := range ignoreIDs {
		ignoreIDsMap[id] = struct{}{}
	}

	filtered := map[int]struct{}{}

	for len(filtered) < limit {
		packageReference, exists, err := scanner.Next()
		if err != nil {
			return nil, 0, 0, errors.Wrap(err, "dbstore.GetUploadIDsWithReferences.Next")
		}
		if !exists {
			break
		}
		recordsScanned++

		if _, ok := filtered[packageReference.DumpID]; ok {
			// This index includes a definition so we can skip testing the filters here. The index
			// will be included in the moniker search regardless if it contains additional references.
			continue
		}

		if _, ok := ignoreIDsMap[packageReference.DumpID]; ok {
			// Ignore this dump
			continue
		}

		filtered[packageReference.DumpID] = struct{}{}
	}

	trace.Log(
		log.Int("uploadIDsWithReferences.numFiltered", len(filtered)),
		log.Int("uploadIDsWithReferences.numRecordsScanned", recordsScanned),
	)

	flattened := make([]int, 0, len(filtered))
	for k := range filtered {
		flattened = append(flattened, k)
	}
	sort.Ints(flattened)

	return flattened, recordsScanned, totalCount, nil
}

// GetVisibleUploadsMatchingMonikers returns visible uploads that refer (via package information) to any of the
// given monikers' packages.
//
// Visibility is determined in two parts: if the index belongs to the given repository, it is visible if
// it can be seen from the given index; otherwise, an index is visible if it can be seen from the tip of
// the default branch of its own repository.
// ReferenceIDs
func (s *store) GetVisibleUploadsMatchingMonikers(ctx context.Context, repositoryID int, commit string, monikers []precise.QualifiedMonikerData, limit, offset int) (_ shared.PackageReferenceScanner, _ int, err error) {
	ctx, trace, endObservation := s.operations.getVisibleUploadsMatchingMonikers.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
		log.Int("numMonikers", len(monikers)),
		log.String("monikers", monikersToString(monikers)),
		log.Int("limit", limit),
		log.Int("offset", offset),
	}})
	defer endObservation(1, observation.Args{})

	if len(monikers) == 0 {
		return shared.PackageReferenceScannerFromSlice(), 0, nil
	}

	qs := make([]*sqlf.Query, 0, len(monikers))
	for _, moniker := range monikers {
		qs = append(qs, sqlf.Sprintf("(%s, %s, %s)", moniker.Scheme, moniker.Name, moniker.Version))
	}

	visibleUploadsQuery := makeVisibleUploadsQuery(repositoryID, commit)

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
	if err != nil {
		return nil, 0, err
	}

	countQuery := sqlf.Sprintf(referenceIDsCountQuery, visibleUploadsQuery, repositoryID, sqlf.Join(qs, ", "), authzConds)
	totalCount, _, err := basestore.ScanFirstInt(s.db.Query(ctx, countQuery))
	if err != nil {
		return nil, 0, err
	}
	trace.Log(log.Int("totalCount", totalCount))

	query := sqlf.Sprintf(referenceIDsQuery, visibleUploadsQuery, repositoryID, sqlf.Join(qs, ", "), authzConds, limit, offset)
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	return shared.PackageReferenceScannerFromRows(rows), totalCount, nil
}

const referenceIDsCTEDefinitions = `
-- source: internal/codeintel/stores/dbstore/xrepo.go:ReferenceIDs
WITH
visible_uploads AS (
	(%s)
	UNION
	(SELECT uvt.upload_id FROM lsif_uploads_visible_at_tip uvt WHERE uvt.repository_id != %s AND uvt.is_default_branch)
)
`

const referenceIDsBaseQuery = `
FROM lsif_references r
LEFT JOIN lsif_dumps u ON u.id = r.dump_id
JOIN repo ON repo.id = u.repository_id
WHERE
	(r.scheme, r.name, r.version) IN (%s) AND
	r.dump_id IN (SELECT * FROM visible_uploads) AND
	%s -- authz conds
`

const referenceIDsQuery = referenceIDsCTEDefinitions + `
SELECT r.dump_id, r.scheme, r.name, r.version
` + referenceIDsBaseQuery + `
ORDER BY dump_id
LIMIT %s OFFSET %s
`

const referenceIDsCountQuery = referenceIDsCTEDefinitions + `
SELECT COUNT(distinct r.dump_id)
` + referenceIDsBaseQuery

// refineRetentionConfiguration returns the maximum age for no-stale branches and tags, effectively, as configured
// for the given repository. If there is no retention configuration for the given repository, the given default
// values are returned unchanged.
func refineRetentionConfiguration(ctx context.Context, store *basestore.Store, repositoryID int, maxAgeForNonStaleBranches, maxAgeForNonStaleTags time.Duration) (_, _ time.Duration, err error) {
	rows, err := store.Query(ctx, sqlf.Sprintf(retentionConfigurationQuery, repositoryID))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var v1, v2 int
		if err := rows.Scan(&v1, &v2); err != nil {
			return 0, 0, err
		}

		maxAgeForNonStaleBranches = time.Second * time.Duration(v1)
		maxAgeForNonStaleTags = time.Second * time.Duration(v2)
	}

	return maxAgeForNonStaleBranches, maxAgeForNonStaleTags, nil
}

const retentionConfigurationQuery = `
SELECT max_age_for_non_stale_branches_seconds, max_age_for_non_stale_tags_seconds
FROM lsif_retention_configuration
WHERE repository_id = %s
`

// sanitizeCommitInput reads the data that needs to be persisted from the given graph and writes the
// sanitized values (ensures values match the column types) into channels for insertion into a particular
// table.
func sanitizeCommitInput(
	ctx context.Context,
	graph *commitgraph.Graph,
	refDescriptions map[string][]gitdomain.RefDescription,
	maxAgeForNonStaleBranches time.Duration,
	maxAgeForNonStaleTags time.Duration,
) *sanitizedCommitInput {
	maxAges := map[gitdomain.RefType]time.Duration{
		gitdomain.RefTypeBranch: maxAgeForNonStaleBranches,
		gitdomain.RefTypeTag:    maxAgeForNonStaleTags,
	}

	nearestUploadsRowValues := make(chan []any)
	nearestUploadsLinksRowValues := make(chan []any)
	uploadsVisibleAtTipRowValues := make(chan []any)

	sanitized := &sanitizedCommitInput{
		nearestUploadsRowValues:      nearestUploadsRowValues,
		nearestUploadsLinksRowValues: nearestUploadsLinksRowValues,
		uploadsVisibleAtTipRowValues: uploadsVisibleAtTipRowValues,
	}

	go func() {
		defer close(nearestUploadsRowValues)
		defer close(nearestUploadsLinksRowValues)
		defer close(uploadsVisibleAtTipRowValues)

		listSerializer := newUploadMetaListSerializer()

		for envelope := range graph.Stream() {
			if envelope.Uploads != nil {
				if !countingWrite(
					ctx,
					nearestUploadsRowValues,
					&sanitized.numNearestUploadsRecords,
					// row values
					dbutil.CommitBytea(envelope.Uploads.Commit),
					listSerializer.Serialize(envelope.Uploads.Uploads),
				) {
					return
				}
			}

			if envelope.Links != nil {
				if !countingWrite(
					ctx,
					nearestUploadsLinksRowValues,
					&sanitized.numNearestUploadsLinksRecords,
					// row values
					dbutil.CommitBytea(envelope.Links.Commit),
					dbutil.CommitBytea(envelope.Links.AncestorCommit),
					envelope.Links.Distance,
				) {
					return
				}
			}
		}

		for commit, refDescriptions := range refDescriptions {
			isDefaultBranch := false
			names := make([]string, 0, len(refDescriptions))

			for _, refDescription := range refDescriptions {
				if refDescription.IsDefaultBranch {
					isDefaultBranch = true
				} else {
					maxAge, ok := maxAges[refDescription.Type]
					if !ok || refDescription.CreatedDate == nil || time.Since(*refDescription.CreatedDate) > maxAge {
						continue
					}
				}

				names = append(names, refDescription.Name)
			}
			sort.Strings(names)

			if len(names) == 0 {
				continue
			}

			for _, uploadMeta := range graph.UploadsVisibleAtCommit(commit) {
				if !countingWrite(
					ctx,
					uploadsVisibleAtTipRowValues,
					&sanitized.numUploadsVisibleAtTipRecords,
					// row values
					uploadMeta.UploadID,
					strings.Join(names, ","),
					isDefaultBranch,
				) {
					return
				}
			}
		}
	}()

	return sanitized
}

// writeVisibleUploads serializes the given input into a the following set of temporary tables in the database.
//
//   - t_lsif_nearest_uploads        (mirroring lsif_nearest_uploads)
//   - t_lsif_nearest_uploads_links  (mirroring lsif_nearest_uploads_links)
//   - t_lsif_uploads_visible_at_tip (mirroring lsif_uploads_visible_at_tip)
//
// The data in these temporary tables can then be moved into a persisted/permanent table. We previously would perform a
// bulk delete of the records associated with a repository, then reinsert all of the data needed to be persisted. This
// caused massive table bloat on some instances. Storing into a temporary table and then inserting/updating/deleting
// records into the persisted table minimizes the number of tuples we need to touch and drastically reduces table bloat.
func (s *store) writeVisibleUploads(ctx context.Context, sanitizedInput *sanitizedCommitInput, tx *basestore.Store) (err error) {
	ctx, trace, endObservation := s.operations.writeVisibleUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if err := s.createTemporaryNearestUploadsTables(ctx, tx); err != nil {
		return err
	}

	g, gctx := errgroup.WithContext(ctx)

	// Insert the set of uploads that are visible from each commit for a given repository into a temporary table.
	nearestUploadsWriter := func() error {
		return batch.InsertValues(
			gctx,
			tx.Handle(),
			"t_lsif_nearest_uploads",
			batch.MaxNumPostgresParameters,
			[]string{"commit_bytea", "uploads"},
			sanitizedInput.nearestUploadsRowValues,
		)
	}

	// Insert the commits not inserted into the table above by adding links to a unique ancestor and their relative
	// distance in the graph into another temporary table. We use this as a cheap way to reconstruct the full data
	// set, which is multiplicative in the size of the commit graph AND the number of unique roots.
	nearestUploadsLinksWriter := func() error {
		return batch.InsertValues(
			gctx,
			tx.Handle(),
			"t_lsif_nearest_uploads_links",
			batch.MaxNumPostgresParameters,
			[]string{"commit_bytea", "ancestor_commit_bytea", "distance"},
			sanitizedInput.nearestUploadsLinksRowValues,
		)
	}

	// Insert the set of uploads visible from the tip of the default branch into a temporary table. These values are
	// used to determine which bundles for a repository we open during a global find references query.
	uploadsVisibleAtTipWriter := func() error {
		return batch.InsertValues(
			gctx,
			tx.Handle(),
			"t_lsif_uploads_visible_at_tip",
			batch.MaxNumPostgresParameters,
			[]string{"upload_id", "branch_or_tag_name", "is_default_branch"},
			sanitizedInput.uploadsVisibleAtTipRowValues,
		)
	}

	g.Go(nearestUploadsWriter)
	g.Go(nearestUploadsLinksWriter)
	g.Go(uploadsVisibleAtTipWriter)

	if err := g.Wait(); err != nil {
		return err
	}
	trace.Log(
		log.Int("numNearestUploadsRecords", int(sanitizedInput.numNearestUploadsRecords)),
		log.Int("numNearestUploadsLinksRecords", int(sanitizedInput.numNearestUploadsLinksRecords)),
		log.Int("numUploadsVisibleAtTipRecords", int(sanitizedInput.numUploadsVisibleAtTipRecords)),
	)

	return nil
}

// persistNearestUploads modifies the lsif_nearest_uploads table so that it has same data
// as t_lsif_nearest_uploads for the given repository.
func (s *store) persistNearestUploads(ctx context.Context, repositoryID int, tx *basestore.Store) (err error) {
	ctx, trace, endObservation := s.operations.persistNearestUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	rowsInserted, rowsUpdated, rowsDeleted, err := s.bulkTransfer(
		ctx,
		sqlf.Sprintf(nearestUploadsInsertQuery, repositoryID, repositoryID),
		sqlf.Sprintf(nearestUploadsUpdateQuery, repositoryID),
		sqlf.Sprintf(nearestUploadsDeleteQuery, repositoryID),
		tx,
	)
	if err != nil {
		return err
	}
	trace.Log(
		log.Int("lsif_nearest_uploads.ins", rowsInserted),
		log.Int("lsif_nearest_uploads.upd", rowsUpdated),
		log.Int("lsif_nearest_uploads.del", rowsDeleted),
	)

	return nil
}

const nearestUploadsInsertQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:persistNearestUploads
INSERT INTO lsif_nearest_uploads
SELECT %s, source.commit_bytea, source.uploads
FROM t_lsif_nearest_uploads source
WHERE source.commit_bytea NOT IN (SELECT nu.commit_bytea FROM lsif_nearest_uploads nu WHERE nu.repository_id = %s)
`

const nearestUploadsUpdateQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:persistNearestUploads
UPDATE lsif_nearest_uploads nu
SET uploads = source.uploads
FROM t_lsif_nearest_uploads source
WHERE
	nu.repository_id = %s AND
	nu.commit_bytea = source.commit_bytea AND
	nu.uploads != source.uploads
`

const nearestUploadsDeleteQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:persistNearestUploads
DELETE FROM lsif_nearest_uploads nu
WHERE
	nu.repository_id = %s AND
	nu.commit_bytea NOT IN (SELECT source.commit_bytea FROM t_lsif_nearest_uploads source)
`

// persistNearestUploadsLinks modifies the lsif_nearest_uploads_links table so that it has same
// data as t_lsif_nearest_uploads_links for the given repository.
func (s *store) persistNearestUploadsLinks(ctx context.Context, repositoryID int, tx *basestore.Store) (err error) {
	ctx, trace, endObservation := s.operations.persistNearestUploadsLinks.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	rowsInserted, rowsUpdated, rowsDeleted, err := s.bulkTransfer(
		ctx,
		sqlf.Sprintf(nearestUploadsLinksInsertQuery, repositoryID, repositoryID),
		sqlf.Sprintf(nearestUploadsLinksUpdateQuery, repositoryID),
		sqlf.Sprintf(nearestUploadsLinksDeleteQuery, repositoryID),
		tx,
	)
	if err != nil {
		return err
	}
	trace.Log(
		log.Int("lsif_nearest_uploads_links.ins", rowsInserted),
		log.Int("lsif_nearest_uploads_links.upd", rowsUpdated),
		log.Int("lsif_nearest_uploads_links.del", rowsDeleted),
	)

	return nil
}

const nearestUploadsLinksInsertQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:persistNearestUploadsLinks
INSERT INTO lsif_nearest_uploads_links
SELECT %s, source.commit_bytea, source.ancestor_commit_bytea, source.distance
FROM t_lsif_nearest_uploads_links source
WHERE source.commit_bytea NOT IN (SELECT nul.commit_bytea FROM lsif_nearest_uploads_links nul WHERE nul.repository_id = %s)
`

const nearestUploadsLinksUpdateQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:persistNearestUploadsLinks
UPDATE lsif_nearest_uploads_links nul
SET ancestor_commit_bytea = source.ancestor_commit_bytea, distance = source.distance
FROM t_lsif_nearest_uploads_links source
WHERE
	nul.repository_id = %s AND
	nul.commit_bytea = source.commit_bytea AND
	nul.ancestor_commit_bytea != source.ancestor_commit_bytea AND
	nul.distance != source.distance
`

const nearestUploadsLinksDeleteQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:persistNearestUploadsLinks
DELETE FROM lsif_nearest_uploads_links nul
WHERE
	nul.repository_id = %s AND
	nul.commit_bytea NOT IN (SELECT source.commit_bytea FROM t_lsif_nearest_uploads_links source)
`

// persistUploadsVisibleAtTip modifies the lsif_uploads_visible_at_tip table so that it has same
// data as t_lsif_uploads_visible_at_tip for the given repository.
func (s *store) persistUploadsVisibleAtTip(ctx context.Context, repositoryID int, tx *basestore.Store) (err error) {
	ctx, trace, endObservation := s.operations.persistUploadsVisibleAtTip.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	insertQuery := sqlf.Sprintf(uploadsVisibleAtTipInsertQuery, repositoryID, repositoryID)
	deleteQuery := sqlf.Sprintf(uploadsVisibleAtTipDeleteQuery, repositoryID)

	rowsInserted, rowsUpdated, rowsDeleted, err := s.bulkTransfer(ctx, insertQuery, nil, deleteQuery, tx)
	if err != nil {
		return err
	}
	trace.Log(
		log.Int("lsif_uploads_visible_at_tip.ins", rowsInserted),
		log.Int("lsif_uploads_visible_at_tip.upd", rowsUpdated),
		log.Int("lsif_uploads_visible_at_tip.del", rowsDeleted),
	)

	return nil
}

const uploadsVisibleAtTipInsertQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:persistUploadsVisibleAtTip
INSERT INTO lsif_uploads_visible_at_tip
SELECT %s, source.upload_id, source.branch_or_tag_name, source.is_default_branch
FROM t_lsif_uploads_visible_at_tip source
WHERE NOT EXISTS (
	SELECT 1
	FROM lsif_uploads_visible_at_tip vat
	WHERE
		vat.repository_id = %s AND
		vat.upload_id = source.upload_id AND
		vat.branch_or_tag_name = source.branch_or_tag_name AND
		vat.is_default_branch = source.is_default_branch
)
`

const uploadsVisibleAtTipDeleteQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:persistUploadsVisibleAtTip
DELETE FROM lsif_uploads_visible_at_tip vat
WHERE
	vat.repository_id = %s AND
	NOT EXISTS (
		SELECT 1
		FROM t_lsif_uploads_visible_at_tip source
		WHERE
			source.upload_id = vat.upload_id AND
			source.branch_or_tag_name = vat.branch_or_tag_name AND
			source.is_default_branch = vat.is_default_branch
	)
`

// bulkTransfer performs the given insert, update, and delete queries and returns the number of records
// touched by each. If any query is nil, the returned count will be zero.
func (s *store) bulkTransfer(ctx context.Context, insertQuery, updateQuery, deleteQuery *sqlf.Query, tx *basestore.Store) (rowsInserted int, rowsUpdated int, rowsDeleted int, err error) {
	prepareQuery := func(query *sqlf.Query) *sqlf.Query {
		if query == nil {
			return sqlf.Sprintf("SELECT 0")
		}

		return sqlf.Sprintf("%s RETURNING 1", query)
	}

	rows, err := tx.Query(ctx, sqlf.Sprintf(bulkTransferQuery, prepareQuery(insertQuery), prepareQuery(updateQuery), prepareQuery(deleteQuery)))
	if err != nil {
		return 0, 0, 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if rows.Next() {
		if err := rows.Scan(&rowsInserted, &rowsUpdated, &rowsDeleted); err != nil {
			return 0, 0, 0, err
		}

		return rowsInserted, rowsUpdated, rowsDeleted, nil
	}

	return 0, 0, 0, nil
}

const bulkTransferQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:bulkTransfer
WITH
	ins AS (%s),
	upd AS (%s),
	del AS (%s)
SELECT
	(SELECT COUNT(*) FROM ins) AS num_ins,
	(SELECT COUNT(*) FROM upd) AS num_upd,
	(SELECT COUNT(*) FROM del) AS num_del
`

func (s *store) createTemporaryNearestUploadsTables(ctx context.Context, tx *basestore.Store) error {
	temporaryTableQueries := []string{
		temporaryNearestUploadsTableQuery,
		temporaryNearestUploadsLinksTableQuery,
		temporaryUploadsVisibleAtTipTableQuery,
	}

	for _, temporaryTableQuery := range temporaryTableQueries {
		if err := tx.Exec(ctx, sqlf.Sprintf(temporaryTableQuery)); err != nil {
			return err
		}
	}

	return nil
}

const temporaryNearestUploadsTableQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:createTemporaryNearestUploadsTables
CREATE TEMPORARY TABLE t_lsif_nearest_uploads (
	commit_bytea bytea NOT NULL,
	uploads      jsonb NOT NULL
) ON COMMIT DROP
`

const temporaryNearestUploadsLinksTableQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:createTemporaryNearestUploadsTables
CREATE TEMPORARY TABLE t_lsif_nearest_uploads_links (
	commit_bytea          bytea NOT NULL,
	ancestor_commit_bytea bytea NOT NULL,
	distance              integer NOT NULL
) ON COMMIT DROP
`

const temporaryUploadsVisibleAtTipTableQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:createTemporaryNearestUploadsTables
CREATE TEMPORARY TABLE t_lsif_uploads_visible_at_tip (
	upload_id integer NOT NULL,
	branch_or_tag_name text NOT NULL,
	is_default_branch boolean NOT NULL
) ON COMMIT DROP
`

// countingWrite writes the given slice of interfaces to the given channel. This function returns true
// if the write succeeded and false if the context was canceled. On success, the counter's underlying
// value will be incremented (non-atomically).
func countingWrite(ctx context.Context, ch chan<- []any, counter *uint32, values ...any) bool {
	select {
	case ch <- values:
		*counter++
		return true

	case <-ctx.Done():
		return false
	}
}

func intsToString(vs []int) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.Itoa(v))
	}

	return strings.Join(strs, ", ")
}

func nilTimeToString(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.String()
}

func buildConditionsAndCte(opts shared.GetUploadsOptions) (*sqlf.Query, []*sqlf.Query, []cteDefinition) {
	conds := make([]*sqlf.Query, 0, 12)

	allowDeletedUploads := (opts.AllowDeletedUpload && opts.State == "") || opts.State == "deleted"

	if opts.RepositoryID != 0 {
		conds = append(conds, sqlf.Sprintf("u.repository_id = %s", opts.RepositoryID))
	}
	if opts.Term != "" {
		conds = append(conds, makeSearchCondition(opts.Term))
	}
	if opts.State != "" {
		conds = append(conds, makeStateCondition(opts.State))
	} else if !allowDeletedUploads {
		conds = append(conds, sqlf.Sprintf("u.state != 'deleted'"))
	}
	if opts.VisibleAtTip {
		conds = append(conds, sqlf.Sprintf("EXISTS ("+visibleAtTipSubselectQuery+")"))
	}

	cteDefinitions := make([]cteDefinition, 0, 2)
	if opts.DependencyOf != 0 {
		cteDefinitions = append(cteDefinitions, cteDefinition{
			name:       "ranked_dependencies",
			definition: sqlf.Sprintf(rankedDependencyCandidateCTEQuery, sqlf.Sprintf("r.dump_id = %s", opts.DependencyOf)),
		})

		// Limit results to the set of uploads canonically providing packages referenced by the given upload identifier
		// (opts.DependencyOf). We do this by selecting the top ranked values in the CTE defined above, which are the
		// referenced package providers grouped by package name, version, repository, and root.
		conds = append(conds, sqlf.Sprintf(`u.id IN (SELECT rd.pkg_id FROM ranked_dependencies rd WHERE rd.rank = 1)`))
	}
	if opts.DependentOf != 0 {
		cteCondition := sqlf.Sprintf(`(p.scheme, p.name, p.version) IN (
			SELECT p.scheme, p.name, p.version
			FROM lsif_packages p
			WHERE p.dump_id = %s
		)`, opts.DependentOf)

		cteDefinitions = append(cteDefinitions, cteDefinition{
			name:       "ranked_dependents",
			definition: sqlf.Sprintf(rankedDependentCandidateCTEQuery, cteCondition),
		})

		// Limit results to the set of uploads that reference the target upload if it canonically provides the
		// matching package. If the target upload does not canonically provide a package, the results will contain
		// no dependent uploads.
		conds = append(conds, sqlf.Sprintf(`u.id IN (
			SELECT r.dump_id
			FROM ranked_dependents rd
			JOIN lsif_references r ON r.scheme = rd.scheme
				AND r.name = rd.name
				AND r.version = rd.version
				AND r.dump_id != rd.pkg_id
			WHERE rd.pkg_id = %s AND rd.rank = 1
		)`, opts.DependentOf))
	}

	sourceTableExpr := sqlf.Sprintf("lsif_uploads u")
	if allowDeletedUploads {
		cteDefinitions = append(cteDefinitions, cteDefinition{
			name:       "deleted_uploads",
			definition: sqlf.Sprintf(deletedUploadsFromAuditLogsCTEQuery),
		})

		sourceTableExpr = sqlf.Sprintf(`(
			SELECT
				id,
				commit,
				root,
				uploaded_at,
				state,
				failure_message,
				started_at,
				finished_at,
				process_after,
				num_resets,
				num_failures,
				repository_id,
				indexer,
				indexer_version,
				num_parts,
				uploaded_parts,
				upload_size,
				associated_index_id,
				expired,
				uncompressed_size
			FROM lsif_uploads
			UNION ALL
			SELECT *
			FROM deleted_uploads
		) AS u`)
	}

	if opts.UploadedBefore != nil {
		conds = append(conds, sqlf.Sprintf("u.uploaded_at < %s", *opts.UploadedBefore))
	}
	if opts.UploadedAfter != nil {
		conds = append(conds, sqlf.Sprintf("u.uploaded_at > %s", *opts.UploadedAfter))
	}
	if opts.InCommitGraph {
		conds = append(conds, sqlf.Sprintf("u.finished_at < (SELECT updated_at FROM lsif_dirty_repositories ldr WHERE ldr.repository_id = u.repository_id)"))
	}
	if opts.LastRetentionScanBefore != nil {
		conds = append(conds, sqlf.Sprintf("(u.last_retention_scan_at IS NULL OR u.last_retention_scan_at < %s)", *opts.LastRetentionScanBefore))
	}
	if !opts.AllowExpired {
		conds = append(conds, sqlf.Sprintf("NOT u.expired"))
	}
	if !opts.AllowDeletedRepo {
		conds = append(conds, sqlf.Sprintf("repo.deleted_at IS NULL"))
	}

	return sourceTableExpr, conds, cteDefinitions
}

// makeSearchCondition returns a disjunction of LIKE clauses against all searchable columns of an upload.
func makeSearchCondition(term string) *sqlf.Query {
	searchableColumns := []string{
		"u.commit",
		"u.root",
		"(u.state)::text",
		"u.failure_message",
		"repo.name",
		"u.indexer",
		"u.indexer_version",
	}

	var termConds []*sqlf.Query
	for _, column := range searchableColumns {
		termConds = append(termConds, sqlf.Sprintf(column+" ILIKE %s", "%"+term+"%"))
	}

	return sqlf.Sprintf("(%s)", sqlf.Join(termConds, " OR "))
}

// makeStateCondition returns a disjunction of clauses comparing the upload against the target state.
func makeStateCondition(state string) *sqlf.Query {
	states := make([]string, 0, 2)
	if state == "errored" || state == "failed" {
		// Treat errored and failed states as equivalent
		states = append(states, "errored", "failed")
	} else {
		states = append(states, state)
	}

	queries := make([]*sqlf.Query, 0, len(states))
	for _, state := range states {
		queries = append(queries, sqlf.Sprintf("u.state = %s", state))
	}

	return sqlf.Sprintf("(%s)", sqlf.Join(queries, " OR "))
}

func buildCTEPrefix(cteDefinitions []cteDefinition) *sqlf.Query {
	if len(cteDefinitions) == 0 {
		return sqlf.Sprintf("")
	}

	cteQueries := make([]*sqlf.Query, 0, len(cteDefinitions))
	for _, cte := range cteDefinitions {
		cteQueries = append(cteQueries, sqlf.Sprintf("%s AS (%s)", sqlf.Sprintf(cte.name), cte.definition))
	}

	return sqlf.Sprintf("WITH\n%s", sqlf.Join(cteQueries, ",\n"))
}

func buildLogFields(opts shared.GetUploadsOptions) []log.Field {
	return []log.Field{
		log.Int("repositoryID", opts.RepositoryID),
		log.String("state", opts.State),
		log.String("term", opts.Term),
		log.Bool("visibleAtTip", opts.VisibleAtTip),
		log.Int("dependencyOf", opts.DependencyOf),
		log.Int("dependentOf", opts.DependentOf),
		log.String("uploadedBefore", nilTimeToString(opts.UploadedBefore)),
		log.String("uploadedAfter", nilTimeToString(opts.UploadedAfter)),
		log.String("lastRetentionScanBefore", nilTimeToString(opts.LastRetentionScanBefore)),
		log.Bool("inCommitGraph", opts.InCommitGraph),
		log.Bool("allowExpired", opts.AllowExpired),
		log.Bool("oldestFirst", opts.OldestFirst),
		log.Int("limit", opts.Limit),
		log.Int("offset", opts.Offset),
	}
}
