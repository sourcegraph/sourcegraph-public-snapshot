package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// InsertUpload inserts a new upload and returns its identifier.
func (s *store) InsertUpload(ctx context.Context, upload shared.Upload) (id int, err error) {
	ctx, _, endObservation := s.operations.insertUpload.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("id", id),
		}})
	}()

	if upload.UploadedParts == nil {
		upload.UploadedParts = []int{}
	}

	id, _, err = basestore.ScanFirstInt(s.db.Query(
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
			upload.ContentType,
			upload.UncompressedSize,
		),
	))

	return id, err
}

const insertUploadQuery = `
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
	content_type,
	uncompressed_size
) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING id
`

// AddUploadPart adds the part index to the given upload's uploaded parts array. This method is idempotent
// (the resulting array is deduplicated on update).
func (s *store) AddUploadPart(ctx context.Context, uploadID, partIndex int) (err error) {
	ctx, _, endObservation := s.operations.addUploadPart.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("uploadID", uploadID),
		log.Int("partIndex", partIndex),
	}})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(addUploadPartQuery, partIndex, uploadID))
}

const addUploadPartQuery = `
UPDATE lsif_uploads SET uploaded_parts = array(SELECT DISTINCT * FROM unnest(array_append(uploaded_parts, %s))) WHERE id = %s
`

// MarkQueued updates the state of the upload to queued and updates the upload size.
func (s *store) MarkQueued(ctx context.Context, id int, uploadSize *int64) (err error) {
	ctx, _, endObservation := s.operations.markQueued.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(markQueuedQuery, dbutil.NullInt64{N: uploadSize}, id))
}

const markQueuedQuery = `
UPDATE lsif_uploads
SET
	state = 'queued',
	queued_at = clock_timestamp(),
	upload_size = %s
WHERE id = %s
`

// MarkFailed updates the state of the upload to failed, increments the num_failures column and sets the finished_at time
func (s *store) MarkFailed(ctx context.Context, id int, reason string) (err error) {
	ctx, _, endObservation := s.operations.markFailed.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(markFailedQuery, reason, id))
}

const markFailedQuery = `
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

// DeleteOverlapapingDumps deletes all completed uploads for the given repository with the same
// commit, root, and indexer. This is necessary to perform during conversions before changing
// the state of a processing upload to completed as there is a unique index on these four columns.
func (s *store) DeleteOverlappingDumps(ctx context.Context, repositoryID int, commit, root, indexer string) (err error) {
	ctx, trace, endObservation := s.operations.deleteOverlappingDumps.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
		log.String("root", root),
		log.String("indexer", indexer),
	}})
	defer endObservation(1, observation.Args{})

	unset, _ := s.db.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "upload overlapping with a newer upload")
	defer unset(ctx)
	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(deleteOverlappingDumpsQuery, repositoryID, commit, root, indexer)))
	if err != nil {
		return err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("count", count))

	return nil
}

const deleteOverlappingDumpsQuery = `
WITH
candidates AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE
		u.state = 'completed' AND
		u.repository_id = %s AND
		u.commit = %s AND
		u.root = %s AND
		u.indexer = %s

	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_uploads table.
	ORDER BY u.id FOR UPDATE
),
updated AS (
	UPDATE lsif_uploads
	SET state = 'deleting'
	WHERE id IN (SELECT id FROM candidates)
	RETURNING 1
)
SELECT COUNT(*) FROM updated
`

func (s *store) WorkerutilStore(observationCtx *observation.Context) dbworkerstore.Store[shared.Upload] {
	return dbworkerstore.New(observationCtx, s.db.Handle(), UploadWorkerStoreOptions)
}
