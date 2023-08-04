package exhaustive

import (
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// maxNumRetries sets the number of retries for exhaustive search
// resolutions to 0. We don't want to retry automatically and instead wait for
// user input.
const maxNumRetries = 0
const maxNumResets = 60

var jobWorkerOpts = dbworkerstore.Options[*Job]{
	Name:              "exhaustive_search_worker_store",
	TableName:         "exhaustive_search_jobs",
	ColumnExpressions: jobColumns,

	Scan: dbworkerstore.BuildWorkerScan(scanJob),

	OrderByExpression: sqlf.Sprintf("exhaustive_search_jobs.state = 'errored', exhaustive_search_jobs.updated_at DESC"),

	StalledMaxAge: 60 * time.Second,
	MaxNumResets:  maxNumResets,

	RetryAfter:    5 * time.Second,
	MaxNumRetries: maxNumRetries,
}

// NewJobWorkerStore returns a dbworkerstore.Store that wraps the "exhaustive_search_jobs" table.
func NewJobWorkerStore(observationCtx *observation.Context, handle basestore.TransactableHandle) dbworkerstore.Store[*Job] {
	return dbworkerstore.New(observationCtx, handle, jobWorkerOpts)
}

var jobColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("initiator_id"),
	sqlf.Sprintf("state"),
	sqlf.Sprintf("query"),
	sqlf.Sprintf("failure_message"),
	sqlf.Sprintf("started_at"),
	sqlf.Sprintf("finished_at"),
	sqlf.Sprintf("process_after"),
	sqlf.Sprintf("num_resets"),
	sqlf.Sprintf("num_failures"),
	sqlf.Sprintf("execution_logs"),
	sqlf.Sprintf("worker_hostname"),
	sqlf.Sprintf("cancel"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
}

func scanJob(sc dbutil.Scanner) (*Job, error) {
	var job Job
	// required field for the sync worker, but
	// the value is thrown out here
	var executionLogs *[]any

	return &job, sc.Scan(
		&job.ID,
		&job.InitiatorID,
		&job.State,
		&job.Query,
		&dbutil.NullString{S: job.FailureMessage},
		&dbutil.NullTime{Time: &job.StartedAt},
		&dbutil.NullTime{Time: &job.FinishedAt},
		&dbutil.NullTime{Time: &job.ProcessAfter},
		&job.NumResets,
		&job.NumFailures,
		&executionLogs,
		&job.WorkerHostname,
		&job.Cancel,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
}
