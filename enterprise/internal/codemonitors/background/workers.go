package background

import (
	"context"
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
	return goroutine.NewPeriodicGoroutine(ctx, 10*time.Second, enqueueActive)
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

func createDBWorkerStore(s *cm.Store) dbworkerstore.Store {
	return dbworkerstore.New(s.Handle(), dbworkerstore.Options{
		TableName:         "cm_trigger_jobs",
		ColumnExpressions: cm.TriggerJobsColumns,
		Scan:              cm.ScanTriggerJobs,
		StalledMaxAge:     60 * time.Second,
		RetryAfter:        10 * time.Second,
		MaxNumRetries:     3,
		OrderByExpression: sqlf.Sprintf("cm_trigger_jobs DESC"),
	})
}

type queryRunner struct {
	*cm.Store
}

func (r *queryRunner) Handle(ctx context.Context, workerStore dbworkerstore.Store, record workerutil.Record) error {
	s := r.Store.With(workerStore)
	q, err := s.GetQueryByRecordID(ctx, record.RecordID())
	if err != nil {
		return err
	}

	// TODO (stefan): run the query

	// Update next_run.
	err = s.SetTriggerQueryNextRun(ctx, q.Id, s.Clock()().Add(5*time.Minute))
	if err != nil {
		return err
	}
	return nil
}
