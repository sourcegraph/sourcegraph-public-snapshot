package background

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"

	cm "github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func newTriggerQueryRunner(ctx context.Context, s *cm.Store, metrics codeMonitorsMetrics) *workerutil.Worker {
	options := workerutil.WorkerOptions{
		NumHandlers: 1,
		Interval:    5 * time.Second,
		Metrics:     workerutil.WorkerMetrics{HandleOperation: metrics.handleOperation},
	}
	worker := dbworker.NewWorker(ctx, createDBWorkerStore(s), &queryRunner{s}, options)
	return worker
}

func newTriggerQueryEnqueuer(ctx context.Context, store *cm.Store) goroutine.BackgroundRoutine {
	enqueueActive := goroutine.NewHandlerWithErrorMessage(
		"code_monitors_trigger_query_enqueuer",
		func(ctx context.Context) error {
			return store.EnqueueTriggerQueries(ctx)
		})
	return goroutine.NewPeriodicGoroutine(ctx, 1*time.Minute, enqueueActive)
}

func newTriggerQueryResetter(ctx context.Context, s *cm.Store, metrics codeMonitorsMetrics) *dbworker.Resetter {
	workerStore := createDBWorkerStore(s)

	options := dbworker.ResetterOptions{
		Name:     "code_monitors_query_resetter",
		Interval: 1 * time.Minute,
		Metrics: dbworker.ResetterMetrics{
			Errors:              metrics.errors,
			RecordResetFailures: metrics.resetFailures,
			RecordResets:        metrics.resets,
		},
	}
	return dbworker.NewResetter(workerStore, options)
}

func newTriggerJobsLogDeleter(ctx context.Context, store *cm.Store) goroutine.BackgroundRoutine {
	deleteObsoleteLogs := goroutine.NewHandlerWithErrorMessage(
		"code_monitors_trigger_jobs_log_deleter",
		func(ctx context.Context) error {
			return store.DeleteObsoleteJobLogs(ctx)
		})
	return goroutine.NewPeriodicGoroutine(ctx, 60*time.Minute, deleteObsoleteLogs)
}

func createDBWorkerStore(s *cm.Store) dbworkerstore.Store {
	return dbworkerstore.New(s.Handle(), dbworkerstore.Options{
		TableName:         "cm_trigger_jobs",
		ColumnExpressions: cm.TriggerJobsColumns,
		Scan:              cm.ScanTriggerJobs,
		StalledMaxAge:     60 * time.Second,
		RetryAfter:        10 * time.Second,
		MaxNumRetries:     3,
		OrderByExpression: sqlf.Sprintf("id"),
	})
}

type queryRunner struct {
	*cm.Store
}

func (r *queryRunner) Handle(ctx context.Context, workerStore dbworkerstore.Store, record workerutil.Record) (err error) {
	s := r.Store.With(workerStore)

	var q *cm.MonitorQuery
	q, err = s.GetQueryByRecordID(ctx, record.RecordID())
	if err != nil {
		return err
	}
	newQuery := newQueryWithAfterFilter(q)

	// Search.
	var results *gqlSearchResponse
	results, err = search(ctx, newQuery)
	if err != nil {
		return err
	}
	var numResults int
	if results != nil {
		numResults = len(results.Data.Search.Results.Results)
	}
	// Log next_run and latest_result to table cm_queries.
	newLatestResult := latestResultTime(q.LatestResult, results, err)
	err = s.SetTriggerQueryNextRun(ctx, q.Id, s.Clock()().Add(5*time.Minute), newLatestResult.UTC())
	if err != nil {
		return err
	}
	// Log the actual query we ran and whether we got any new results.
	err = s.LogSearch(ctx, newQuery, numResults > 0, record.RecordID())
	if err != nil {
		return err
	}
	return nil
}

func newQueryWithAfterFilter(q *cm.MonitorQuery) string {
	// Construct a new query which finds search results introduced after the last
	// time we queried.
	var latestResult time.Time
	if q.LatestResult != nil {
		latestResult = *q.LatestResult
	} else {
		// We've never executed this search query before, so use the current
		// time. We'll most certainly find nothing, which is okay.
		latestResult = time.Now()
	}
	afterTime := latestResult.UTC().Format(time.RFC3339)
	return strings.Join([]string{q.QueryString, fmt.Sprintf(`after:"%s"`, afterTime)}, " ")
}

func latestResultTime(previousLastResult *time.Time, v *gqlSearchResponse, searchErr error) time.Time {
	if searchErr != nil || len(v.Data.Search.Results.Results) == 0 {
		// Error performing the search, or there were no results. Assume the
		// previous info's result time.
		if previousLastResult != nil {
			return *previousLastResult
		}
		return time.Now()
	}

	// Results are ordered chronologically, so first result is the latest.
	t, err := extractTime(v.Data.Search.Results.Results[0])
	if err != nil {
		// Error already logged by extractTime.
		return time.Now()
	}
	return *t
}
