package dbstore

import (
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// StalledDependencySyncingJobMaxAge is the maximum allowable duration between updating
// the state of a dependency indexing job as "processing" and locking the job row during
// processing. An unlocked row that is marked as processing likely indicates that the worker
// that dequeued the job has died. There should be a nearly-zero delay between these states
// during normal operation.
const StalledDependencySyncingJobMaxAge = time.Second * 25

// DependencySyncingJobMaxNumResets is the maximum number of times a dependency indexing
// job can be reset. If an job's failed attempts counter reaches this threshold, it will be
// moved into "errored" rather than "queued" on its next reset.
const DependencySyncingJobMaxNumResets = 3

var dependencySyncingJobWorkerStoreOptions = dbworkerstore.Options{
	Name:              "codeintel_dependency_syncing",
	TableName:         "lsif_dependency_syncing_jobs",
	ColumnExpressions: dependencySyncingJobColumns,
	Scan:              dbworkerstore.BuildWorkerScan(scanDependencySyncingJob),
	OrderByExpression: sqlf.Sprintf("lsif_dependency_syncing_jobs.queued_at, lsif_dependency_syncing_jobs.upload_id"),
	StalledMaxAge:     StalledDependencySyncingJobMaxAge,
	MaxNumResets:      DependencySyncingJobMaxNumResets,
}

func WorkerutilDependencySyncStore(s basestore.ShareableStore, observationContext *observation.Context) dbworkerstore.Store {
	return dbworkerstore.NewWithMetrics(s.Handle(), dependencySyncingJobWorkerStoreOptions, observationContext)
}

var dependencySyncingJobColumns = []*sqlf.Query{
	sqlf.Sprintf("lsif_dependency_syncing_jobs.id"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.state"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.failure_message"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.started_at"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.finished_at"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.process_after"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.num_resets"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.num_failures"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.upload_id"),
}

func scanDependencySyncingJob(s dbutil.Scanner) (job DependencySyncingJob, err error) {
	return job, s.Scan(
		&job.ID,
		&job.State,
		&job.FailureMessage,
		&job.StartedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFailures,
		&job.UploadID,
	)
}
