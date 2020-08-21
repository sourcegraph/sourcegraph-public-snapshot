package repos

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// NewSyncWorker creates a new external service sync worker.
func NewSyncWorker(ctx context.Context, db dbutil.DB, handler dbworker.Handler, numHandlers int) *workerutil.Worker {
	dbHandle := basestore.NewHandleWithDB(db)

	syncJobColumns := append(store.DefaultColumnExpressions(), []*sqlf.Query{
		sqlf.Sprintf("external_service_id"),
		sqlf.Sprintf("next_sync_at"),
	}...)

	store := store.NewStore(dbHandle, store.StoreOptions{
		TableName:         "external_service_sync_jobs",
		ViewName:          "external_service_sync_jobs_with_next_sync_at",
		Scan:              scanSyncJob,
		OrderByExpression: sqlf.Sprintf("next_sync_at"),
		ColumnExpressions: syncJobColumns,
		StalledMaxAge:     30 * time.Second,
		// Zero for now as we expect errors to be transient
		MaxNumResets: 0,
	})

	return dbworker.NewWorker(ctx, store, dbworker.WorkerOptions{
		Name:        "sync_worker",
		Handler:     handler,
		NumHandlers: numHandlers,
		Interval:    1 * time.Minute,
		Metrics: workerutil.WorkerMetrics{
			HandleOperation: newObservationOperation(),
		},
	})
}

func newObservationOperation() *observation.Operation {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	m := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"repo_updater_external_service_syncer",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of results returned"),
	)

	return observationContext.Operation(observation.Op{
		Name:         "Syncer.Process",
		MetricLabels: []string{"process"},
		Metrics:      m,
	})
}

func scanSyncJob(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	var job SyncJob

	for rows.Next() {
		if err := rows.Scan(
			&job.ID,
			&job.State,
			&job.FailureMessage,
			&job.StartedAt,
			&job.FinishedAt,
			&job.ProcessAfter,
			&job.NumResets,
			&job.ExternalServiceID,
			&job.NextSyncAt,
		); err != nil {
			return nil, false, err
		}
	}

	return &job, true, nil
}

// SyncJob represents an external service that needs to be synced
type SyncJob struct {
	ID                int
	State             string
	FailureMessage    sql.NullString
	StartedAt         sql.NullTime
	FinishedAt        sql.NullTime
	ProcessAfter      sql.NullTime
	NumResets         int
	ExternalServiceID int64
	NextSyncAt        sql.NullTime
}

// RecordID implements workerutil.Record and indicates the queued item id
func (s *SyncJob) RecordID() int {
	return s.ID
}

type syncHandler struct{}

func (h *syncHandler) Handle(ctx context.Context, tx workerutil.Store, record workerutil.Record) error {
	// myStore := h.myStore.With(tx) // combine store handles
	// myRecord := record.(MyType)   // convert type of record
	// do processing ...
	return errors.New("TODO")
}
