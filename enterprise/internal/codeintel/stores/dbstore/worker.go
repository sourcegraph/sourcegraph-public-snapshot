package dbstore

import (
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// StalledUploadMaxAge is the maximum allowable duration between updating the state of an
// upload as "processing" and locking the upload row during processing. An unlocked row that
// is marked as processing likely indicates that the worker that dequeued the upload has died.
// There should be a nearly-zero delay between these states during normal operation.
const StalledUploadMaxAge = time.Second * 5

// UploadMaxNumResets is the maximum number of times an upload can be reset. If an upload's
// failed attempts counter reaches this threshold, it will be moved into "errored" rather than
// "queued" on its next reset.
const UploadMaxNumResets = 3

var uploadWorkerStoreOptions = dbworkerstore.Options{
	Name:              "precise_code_intel_upload_worker_store",
	TableName:         "lsif_uploads",
	ViewName:          "lsif_uploads_with_repository_name u",
	ColumnExpressions: uploadColumnsWithNullRank,
	Scan:              scanFirstUploadRecord,
	OrderByExpression: sqlf.Sprintf("uploaded_at"),
	StalledMaxAge:     StalledUploadMaxAge,
	MaxNumResets:      UploadMaxNumResets,
}

func WorkerutilUploadStore(s basestore.ShareableStore, observationContext *observation.Context) dbworkerstore.Store {
	return dbworkerstore.NewWithMetrics(s.Handle(), uploadWorkerStoreOptions, observationContext)
}

// StalledIndexMaxAge is the maximum allowable duration between updating the state of an
// index as "processing" and locking the index row during processing. An unlocked row that
// is marked as processing likely indicates that the indexer that dequeued the index has
// died. There should be a nearly-zero delay between these states during normal operation.
const StalledIndexMaxAge = time.Second * 5

// IndexMaxNumResets is the maximum number of times an index can be reset. If an index's
// failed attempts counter reaches this threshold, it will be moved into "errored" rather than
// "queued" on its next reset.
const IndexMaxNumResets = 3

var indexWorkerStoreOptions = dbworkerstore.Options{
	Name:              "precise_code_intel_index_worker_store",
	TableName:         "lsif_indexes",
	ViewName:          "lsif_indexes_with_repository_name u",
	ColumnExpressions: indexColumnsWithNullRank,
	Scan:              scanFirstIndexRecord,
	OrderByExpression: sqlf.Sprintf("queued_at"),
	StalledMaxAge:     StalledIndexMaxAge,
	MaxNumResets:      IndexMaxNumResets,
}

func WorkerutilIndexStore(s basestore.ShareableStore, observationContext *observation.Context) dbworkerstore.Store {
	return dbworkerstore.NewWithMetrics(s.Handle(), indexWorkerStoreOptions, observationContext)
}
