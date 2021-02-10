package background

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// newInsightEnqueuer returns a background goroutine which will periodically find all of the search
// and webhook insights across all user settings, and enqueue work for the query runner and webhook
// runner workers to perform.
func newInsightEnqueuer(ctx context.Context, store *store.Store) goroutine.BackgroundRoutine {
	// TODO: 1 minute may be too slow? hmm
	return goroutine.NewPeriodicGoroutine(ctx, 1*time.Minute, goroutine.NewHandlerWithErrorMessage(
		"insights_enqueuer",
		func(ctx context.Context) error {
			// TODO: needs metrics
			// TODO: similar to EnqueueTriggerQueries, actually enqueue work
			return nil
		},
	))
}

// newQueryRunner returns a worker that will execute search queries and insert information about
// the results into the code insights database.
//
// TODO(slimsag): needs main app DB for settings discovery
func newQueryRunner(ctx context.Context, insightsStore *store.Store, metrics *metrics) *workerutil.Worker {
	store := createDBWorkerStoreForInsightsJobs(insightsStore) // TODO(slimsag): should not create in TimescaleDB
	options := workerutil.WorkerOptions{
		Name:        "insights_query_runner_worker",
		NumHandlers: 1,
		Interval:    5 * time.Second,
		Metrics:     metrics.workerMetrics,
	}
	worker := dbworker.NewWorker(ctx, store, &queryRunner{
		workerStore:   insightsStore, // TODO(slimsag): should not create in TimescaleDB
		insightsStore: insightsStore,
	}, options)
	return worker
}

// newQueryRunnerResetter returns a worker that will reset pending query runner jobs if they take
// too long to complete.
func newQueryRunnerResetter(ctx context.Context, s *store.Store, metrics *metrics) *dbworker.Resetter {
	store := createDBWorkerStoreForInsightsJobs(s) // TODO(slimsag): should not create in TimescaleDB
	options := dbworker.ResetterOptions{
		Name:     "code_insights_trigger_jobs_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.resetterMetrics,
	}
	return dbworker.NewResetter(store, options)
}

func createDBWorkerStoreForInsightsJobs(s *store.Store) dbworkerstore.Store {
	return dbworkerstore.New(s.Handle(), dbworkerstore.Options{
		Name:      "insights_trigger_jobs_worker_store",
		TableName: "insights_trigger_jobs",
		// TODO(slimsag): table names
		ColumnExpressions: InsightsJobsColumns,
		Scan:              ScanInsightsJobs,

		// We will let a search query or webhook run for up to 60s. After that, it times out and
		// retries in 10s. If 3 timeouts occur, it is not retried.
		StalledMaxAge:     60 * time.Second,
		RetryAfter:        10 * time.Second,
		MaxNumRetries:     3,
		OrderByExpression: sqlf.Sprintf("id"),
	})
}

// TODO(slimsag): move to a insights/dbworkerstore package?

type InsightsJobs struct {
	// TODO(slimsag): all these columns are wrong.
	Id    int
	Query int64

	// The query we ran including after: filter.
	QueryString *string

	// Whether we got any results.
	Results    *bool
	NumResults *int

	// Fields demanded for any dbworker.
	State          string
	FailureMessage *string
	StartedAt      *time.Time
	FinishedAt     *time.Time
	ProcessAfter   *time.Time
	NumResets      int32
	NumFailures    int32
	LogContents    *string
}

func (r *InsightsJobs) RecordID() int {
	return r.Id
}

func ScanInsightsJobs(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	records, err := scanInsightsJobs(rows, err)
	if err != nil {
		return &InsightsJobs{}, false, err
	}
	return records[0], true, nil
}

func scanInsightsJobs(rows *sql.Rows, err error) ([]*InsightsJobs, error) {
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()
	var ms []*InsightsJobs
	for rows.Next() {
		m := &InsightsJobs{}
		if err := rows.Scan(
			// TODO(slimsag): all these columns are wrong.
			&m.Id,
			&m.Query,
			&m.QueryString,
			&m.Results,
			&m.NumResults,
			&m.State,
			&m.FailureMessage,
			&m.StartedAt,
			&m.FinishedAt,
			&m.ProcessAfter,
			&m.NumResets,
			&m.NumFailures,
			&m.LogContents,
		); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	if err != nil {
		return nil, err
	}
	// Rows.Err will report the last error encountered by Rows.Scan.
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ms, nil
}

var InsightsJobsColumns = []*sqlf.Query{
	// TODO(slimsag): all these columns are wrong.
	sqlf.Sprintf("cm_trigger_jobs.id"),
	sqlf.Sprintf("cm_trigger_jobs.query"),
	sqlf.Sprintf("cm_trigger_jobs.query_string"),
	sqlf.Sprintf("cm_trigger_jobs.results"),
	sqlf.Sprintf("cm_trigger_jobs.num_results"),
	sqlf.Sprintf("cm_trigger_jobs.state"),
	sqlf.Sprintf("cm_trigger_jobs.failure_message"),
	sqlf.Sprintf("cm_trigger_jobs.started_at"),
	sqlf.Sprintf("cm_trigger_jobs.finished_at"),
	sqlf.Sprintf("cm_trigger_jobs.process_after"),
	sqlf.Sprintf("cm_trigger_jobs.num_resets"),
	sqlf.Sprintf("cm_trigger_jobs.num_failures"),
	sqlf.Sprintf("cm_trigger_jobs.log_contents"),
}
