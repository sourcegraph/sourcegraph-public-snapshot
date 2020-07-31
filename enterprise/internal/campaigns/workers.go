package campaigns

import (
	"context"
	"database/sql"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

// RunWorkers starts a workerutil.NewWorker that fetches enqueued changesets
// from the database and passes them to the changeset reconciler for
// processing.
func RunWorkers(
	ctx context.Context,
	s *Store,
	gitClient GitserverClient,
	sourcer repos.Sourcer,
) {
	r := &reconciler{gitserverClient: gitClient, sourcer: sourcer, store: s}

	options := workerutil.WorkerOptions{
		Handler:     r.HandlerFunc(),
		NumHandlers: 5,
		Interval:    5 * time.Second,
		Metrics: workerutil.WorkerMetrics{
			HandleOperation: newObservationOperation(),
		},
	}

	workerStore := workerutil.NewStore(s.Handle(), workerutil.StoreOptions{
		TableName:            "changesets",
		AlternateColumnNames: map[string]string{"state": "reconciler_state"},
		ColumnExpressions:    changesetColumns,
		Scan:                 scanFirstChangesetRecord,
		OrderByExpression:    sqlf.Sprintf("changesets.updated_at"),
		StalledMaxAge:        60 * time.Second,
		MaxNumResets:         5,
	})

	worker := workerutil.NewWorker(ctx, workerStore, options)
	worker.Start()
}

func scanFirstChangesetRecord(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	return scanFirstChangeset(rows, err)
}

func newObservationOperation() *observation.Operation {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"campaigns_reconciler",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of results returned"),
	)

	return observationContext.Operation(observation.Op{
		Name:         "Reconciler.Process",
		MetricLabels: []string{"process"},
		Metrics:      metrics,
	})
}
