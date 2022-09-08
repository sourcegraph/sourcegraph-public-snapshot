package dbstore

import (
	"context"
	"database/sql"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

// GetUploadByID returns an upload by its identifier and boolean flag indicating its existence.
func (s *Store) GetUploadByID(ctx context.Context, id int) (_ Upload, _ bool, err error) {
	ctx, _, endObservation := s.operations.getUploadByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.Store))
	if err != nil {
		return Upload{}, false, err
	}

	return scanFirstUpload(s.Store.Query(ctx, sqlf.Sprintf(getUploadByIDQuery, id, authzConds)))
}

const getUploadByIDQuery = `
-- source: internal/codeintel/uploads/internal/stores/store_uploads.go:GetUploadByID
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
	u.uncompressed_size
FROM lsif_uploads u
LEFT JOIN (` + uploadRankQueryFragment + `) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE repo.deleted_at IS NULL AND u.state != 'deleted' AND u.id = %s AND %s
`

const visibleAtTipSubselectQuery = `SELECT 1 FROM lsif_uploads_visible_at_tip uvt WHERE uvt.repository_id = u.repository_id AND uvt.upload_id = u.id`

const uploadRankQueryFragment = `
SELECT
	r.id,
	ROW_NUMBER() OVER (ORDER BY COALESCE(r.process_after, r.uploaded_at), r.id) as rank
FROM lsif_uploads_with_repository_name r
WHERE r.state = 'queued'
`

// Upload is a subset of the lsif_uploads table and stores both processed and unprocessed
// records.
type Upload struct {
	ID                int
	Commit            string
	Root              string
	VisibleAtTip      bool
	UploadedAt        time.Time
	State             string
	FailureMessage    *string
	StartedAt         *time.Time
	FinishedAt        *time.Time
	ProcessAfter      *time.Time
	NumResets         int
	NumFailures       int
	RepositoryID      int
	RepositoryName    string
	Indexer           string
	IndexerVersion    string
	NumParts          int
	UploadedParts     []int
	UploadSize        *int64
	UncompressedSize  *int64
	Rank              *int
	AssociatedIndexID *int
}

func (u Upload) RecordID() int {
	return u.ID
}

func scanUpload(s dbutil.Scanner) (upload Upload, _ error) {
	var rawUploadedParts []sql.NullInt32
	if err := s.Scan(
		&upload.ID,
		&upload.Commit,
		&upload.Root,
		&upload.VisibleAtTip,
		&upload.UploadedAt,
		&upload.State,
		&upload.FailureMessage,
		&upload.StartedAt,
		&upload.FinishedAt,
		&upload.ProcessAfter,
		&upload.NumResets,
		&upload.NumFailures,
		&upload.RepositoryID,
		&upload.RepositoryName,
		&upload.Indexer,
		&dbutil.NullString{S: &upload.IndexerVersion},
		&upload.NumParts,
		pq.Array(&rawUploadedParts),
		&upload.UploadSize,
		&upload.AssociatedIndexID,
		&upload.Rank,
		&upload.UncompressedSize,
	); err != nil {
		return upload, err
	}

	upload.UploadedParts = make([]int, 0, len(rawUploadedParts))
	for _, uploadedPart := range rawUploadedParts {
		upload.UploadedParts = append(upload.UploadedParts, int(uploadedPart.Int32))
	}

	return upload, nil
}

// scanFirstUpload scans a slice of uploads from the return value of `*Store.query` and returns the first.
var scanFirstUpload = basestore.NewFirstScanner(scanUpload)

type GetUploadsOptions struct {
	RepositoryID            int
	State                   string
	Term                    string
	VisibleAtTip            bool
	DependencyOf            int
	DependentOf             int
	UploadedBefore          *time.Time
	UploadedAfter           *time.Time
	LastRetentionScanBefore *time.Time
	AllowExpired            bool
	AllowDeletedRepo        bool
	AllowDeletedUpload      bool
	OldestFirst             bool
	Limit                   int
	Offset                  int

	// InCommitGraph ensures that the repository commit graph was updated strictly
	// after this upload was processed. This condition helps us filter out new uploads
	// that we might later mistake for unreachable.
	InCommitGraph bool
}

// InsertUpload inserts a new upload and returns its identifier.
func (s *Store) InsertUpload(ctx context.Context, upload Upload) (id int, err error) {
	ctx, _, endObservation := s.operations.insertUpload.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("id", id),
		}})
	}()

	if upload.UploadedParts == nil {
		upload.UploadedParts = []int{}
	}

	id, _, err = basestore.ScanFirstInt(s.Store.Query(
		ctx,
		sqlf.Sprintf(
			insertUploadQuery,
			upload.Commit,
			upload.Root,
			upload.RepositoryID,
			upload.Indexer,
			upload.IndexerVersion,
			upload.State,
			upload.NumParts,
			pq.Array(upload.UploadedParts),
			upload.UploadSize,
			upload.AssociatedIndexID,
			upload.UncompressedSize,
		),
	))

	return id, err
}

const insertUploadQuery = `
-- source: internal/codeintel/stores/dbstore/uploads.go:InsertUpload
INSERT INTO lsif_uploads (
	commit,
	root,
	repository_id,
	indexer,
	indexer_version,
	state,
	num_parts,
	uploaded_parts,
	upload_size,
	associated_index_id,
	uncompressed_size
) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING id
`

// AddUploadPart adds the part index to the given upload's uploaded parts array. This method is idempotent
// (the resulting array is deduplicated on update).
func (s *Store) AddUploadPart(ctx context.Context, uploadID, partIndex int) (err error) {
	ctx, _, endObservation := s.operations.addUploadPart.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("uploadID", uploadID),
		log.Int("partIndex", partIndex),
	}})
	defer endObservation(1, observation.Args{})

	return s.Store.Exec(ctx, sqlf.Sprintf(addUploadPartQuery, partIndex, uploadID))
}

const addUploadPartQuery = `
-- source: internal/codeintel/stores/dbstore/uploads.go:AddUploadPart
UPDATE lsif_uploads SET uploaded_parts = array(SELECT DISTINCT * FROM unnest(array_append(uploaded_parts, %s))) WHERE id = %s
`

// MarkQueued updates the state of the upload to queued and updates the upload size.
func (s *Store) MarkQueued(ctx context.Context, id int, uploadSize *int64) (err error) {
	ctx, _, endObservation := s.operations.markQueued.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	return s.Store.Exec(ctx, sqlf.Sprintf(markQueuedQuery, uploadSize, id))
}

const markQueuedQuery = `
-- source: internal/codeintel/stores/dbstore/uploads.go:MarkQueued
UPDATE lsif_uploads
SET
	state = 'queued',
	queued_at = clock_timestamp(),
	upload_size = %s
WHERE id = %s
`

// MarkFailed updates the state of the upload to failed, increments the num_failures column and sets the finished_at time
func (s *Store) MarkFailed(ctx context.Context, id int, reason string) (err error) {
	ctx, _, endObservation := s.operations.markFailed.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	return s.Store.Exec(ctx, sqlf.Sprintf(markFailedQuery, reason, id))
}

const markFailedQuery = `
-- source: internal/codeintel/stores/dbstore/uploads.go:MarkFailed
UPDATE
	lsif_uploads
SET
	state = 'failed',
	finished_at = clock_timestamp(),
	failure_message = %s,
	num_failures = num_failures + 1
WHERE
	id = %s
`

var uploadColumnsWithNullRank = []*sqlf.Query{
	sqlf.Sprintf("u.id"),
	sqlf.Sprintf("u.commit"),
	sqlf.Sprintf("u.root"),
	sqlf.Sprintf("EXISTS (" + visibleAtTipSubselectQuery + ") AS visible_at_tip"),
	sqlf.Sprintf("u.uploaded_at"),
	sqlf.Sprintf("u.state"),
	sqlf.Sprintf("u.failure_message"),
	sqlf.Sprintf("u.started_at"),
	sqlf.Sprintf("u.finished_at"),
	sqlf.Sprintf("u.process_after"),
	sqlf.Sprintf("u.num_resets"),
	sqlf.Sprintf("u.num_failures"),
	sqlf.Sprintf("u.repository_id"),
	sqlf.Sprintf("u.repository_name"),
	sqlf.Sprintf("u.indexer"),
	sqlf.Sprintf("u.indexer_version"),
	sqlf.Sprintf("u.num_parts"),
	sqlf.Sprintf("u.uploaded_parts"),
	sqlf.Sprintf("u.upload_size"),
	sqlf.Sprintf("u.associated_index_id"),
	sqlf.Sprintf("NULL"),
	sqlf.Sprintf("u.uncompressed_size"),
}

// HardDeleteUploadByID deletes the upload record with the given identifier.
func (s *Store) HardDeleteUploadByID(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservation := s.operations.hardDeleteUploadByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
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

	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Before deleting the record, ensure that we decrease the number of existant references
	// to all of this upload's dependencies. This also selects a new upload to canonically provide
	// the same package as the deleted upload, if such an upload exists.
	if _, err := tx.UpdateReferenceCounts(ctx, ids, DependencyReferenceCountUpdateTypeRemove); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(hardDeleteUploadByIDQuery, sqlf.Join(idQueries, ", "))); err != nil {
		return err
	}

	return nil
}

const hardDeleteUploadByIDQuery = `
-- source: internal/codeintel/stores/dbstore/uploads.go:HardDeleteUploadByID
WITH locked_uploads AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE u.id IN (%s)
	ORDER BY u.id FOR UPDATE
)
DELETE FROM lsif_uploads WHERE id IN (SELECT id FROM locked_uploads)
`

// SelectRepositoriesForRetentionScan returns a set of repository identifiers with live code intelligence
// data and a fresh associated commit graph. Repositories that were returned previously from this call
// within the  given process delay are not returned.
func (s *Store) SelectRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) (_ []int, err error) {
	return s.selectRepositoriesForRetentionScan(ctx, processDelay, limit, timeutil.Now())
}

func (s *Store) selectRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int, now time.Time) (_ []int, err error) {
	ctx, _, endObservation := s.operations.selectRepositoriesForRetentionScan.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return basestore.ScanInts(s.Query(ctx, sqlf.Sprintf(
		repositoryIDsForRetentionScanQuery,
		now,
		int(processDelay/time.Second),
		limit,
		now,
		now,
	)))
}

const repositoryIDsForRetentionScanQuery = `
-- source: internal/codeintel/stores/dbstore/uploads.go:selectRepositoriesForRetentionScan
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
func (s *Store) UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) error {
	return s.updateUploadRetention(ctx, protectedIDs, expiredIDs, time.Now())
}

func (s *Store) updateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int, now time.Time) (err error) {
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

	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

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
-- source: internal/codeintel/stores/dbstore/uploads.go:UpdateUploadRetention
UPDATE lsif_uploads SET %s WHERE id IN (%s)
`

type DependencyReferenceCountUpdateType int

const (
	DependencyReferenceCountUpdateTypeNone DependencyReferenceCountUpdateType = iota
	DependencyReferenceCountUpdateTypeAdd
	DependencyReferenceCountUpdateTypeRemove
)

var deltaMap = map[DependencyReferenceCountUpdateType]int{
	DependencyReferenceCountUpdateTypeNone:   +0,
	DependencyReferenceCountUpdateTypeAdd:    +1,
	DependencyReferenceCountUpdateTypeRemove: -1,
}

// UpdateReferenceCounts updates the reference counts of uploads indicated by the given identifiers
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
func (s *Store) UpdateReferenceCounts(ctx context.Context, ids []int, dependencyUpdateType DependencyReferenceCountUpdateType) (updated int, err error) {
	ctx, _, endObservation := s.operations.updateReferenceCounts.With(ctx, &err, observation.Args{LogFields: []log.Field{
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

	tx, err := s.transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	idArray := pq.Array(ids)

	excludeCondition := sqlf.Sprintf("TRUE")
	if dependencyUpdateType == DependencyReferenceCountUpdateTypeRemove {
		excludeCondition = sqlf.Sprintf("NOT (u.id = ANY (%s))", idArray)
	}

	result, err := tx.ExecResult(ctx, sqlf.Sprintf(
		updateReferenceCountsQuery,
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

const updateReferenceCountsQuery = `
-- source: internal/codeintel/stores/dbstore/uploads.go:UpdateReferenceCounts
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

// UpdateCommitedAt updates the commit date for the given repository.
func (s *Store) UpdateCommitedAt(ctx context.Context, uploadID int, committedAt time.Time) (err error) {
	ctx, _, endObservation := s.operations.updateCommitedAt.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("uploadID", uploadID),
	}})
	defer endObservation(1, observation.Args{})

	return s.Exec(ctx, sqlf.Sprintf(updateCommitedAtQuery, committedAt, uploadID))
}

const updateCommitedAtQuery = `
-- source: internal/codeintel/stores/dbstore/uploads.go:UpdateCommitedAt
UPDATE lsif_uploads SET committed_at = %s WHERE id = %s
`

func intsToString(vs []int) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.Itoa(v))
	}

	return strings.Join(strs, ", ")
}

// LastUploadRetentionScanForRepository returns the last timestamp, if any, that the repository with the
// given identifier was considered for upload expiration checks.
func (s *Store) LastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error) {
	ctx, _, endObservation := s.operations.lastUploadRetentionScanForRepository.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	t, ok, err := basestore.ScanFirstTime(s.Query(ctx, sqlf.Sprintf(lastUploadRetentionScanForRepositoryQuery, repositoryID)))
	if !ok {
		return nil, err
	}

	return &t, nil
}

const lastUploadRetentionScanForRepositoryQuery = `
-- source: internal/codeintel/stores/dbstore/uploads.go:LastUploadRetentionScanForRepository
SELECT last_retention_scan_at FROM lsif_last_retention_scan WHERE repository_id = %s
`
