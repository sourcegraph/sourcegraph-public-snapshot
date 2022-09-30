package dbstore

import (
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// StalledDependencyIndexingJobMaxAge is the maximum allowable duration between updating
// the state of a dependency indexing queueing job as "processing" and locking the job row during
// processing. An unlocked row that is marked as processing likely indicates that the worker
// that dequeued the job has died. There should be a nearly-zero delay between these states
// during normal operation.
const StalledDependencyIndexingJobMaxAge = time.Second * 25

// DependencyIndexingJobMaxNumResets is the maximum number of times a dependency indexing
// job can be reset. If an job's failed attempts counter reaches this threshold, it will be
// moved into "errored" rather than "queued" on its next reset.
const DependencyIndexingJobMaxNumResets = 3

var dependencyIndexingJobWorkerStoreOptions = dbworkerstore.Options{
	Name:              "codeintel_dependency_indexing",
	TableName:         "lsif_dependency_indexing_jobs",
	ColumnExpressions: dependencyIndexingJobColumns,
	Scan:              dbworkerstore.BuildWorkerScan(scanDependencyIndexingJob),
	OrderByExpression: sqlf.Sprintf("lsif_dependency_indexing_jobs.queued_at, lsif_dependency_indexing_jobs.upload_id"),
	StalledMaxAge:     StalledDependencyIndexingJobMaxAge,
	MaxNumResets:      DependencyIndexingJobMaxNumResets,
}

func WorkerutilDependencyIndexStore(s basestore.ShareableStore, observationContext *observation.Context) dbworkerstore.Store {
	return dbworkerstore.NewWithMetrics(s.Handle(), dependencyIndexingJobWorkerStoreOptions, observationContext)
}

var dependencyIndexingJobColumns = []*sqlf.Query{
	sqlf.Sprintf("lsif_dependency_indexing_jobs.id"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.state"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.failure_message"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.started_at"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.finished_at"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.process_after"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.num_resets"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.num_failures"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.upload_id"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.external_service_kind"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.external_service_sync"),
}
