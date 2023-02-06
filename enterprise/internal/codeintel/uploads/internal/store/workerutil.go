package store

import (
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// StalledUploadMaxAge is the maximum allowable duration between updating the state of an
// upload as "processing" and locking the upload row during processing. An unlocked row that
// is marked as processing likely indicates that the worker that dequeued the upload has died.
// There should be a nearly-zero delay between these states during normal operation.
const StalledUploadMaxAge = time.Second * 25

// UploadMaxNumResets is the maximum number of times an upload can be reset. If an upload's
// failed attempts counter reaches this threshold, it will be moved into "errored" rather than
// "queued" on its next reset.
const UploadMaxNumResets = 3

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
	sqlf.Sprintf("u.content_type"),
	sqlf.Sprintf("u.should_reindex"),
	sqlf.Sprintf("NULL"),
	sqlf.Sprintf("u.uncompressed_size"),
}

var UploadWorkerStoreOptions = dbworkerstore.Options[types.Upload]{
	Name:              "codeintel_upload",
	TableName:         "lsif_uploads",
	ViewName:          "lsif_uploads_with_repository_name u",
	ColumnExpressions: uploadColumnsWithNullRank,
	Scan:              dbworkerstore.BuildWorkerScan(scanCompleteUpload),
	OrderByExpression: sqlf.Sprintf(`
		u.associated_index_id IS NULL DESC,
		COALESCE(u.process_after, u.uploaded_at),
		u.id
	`),
	StalledMaxAge: StalledUploadMaxAge,
	MaxNumResets:  UploadMaxNumResets,
}

func (s *store) WorkerutilStore(observationCtx *observation.Context) dbworkerstore.Store[types.Upload] {
	return dbworkerstore.New(observationCtx, s.db.Handle(), UploadWorkerStoreOptions)
}
