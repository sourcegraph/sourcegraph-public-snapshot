package dbstore

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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
