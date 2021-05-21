package background

import (
	"context"
	"database/sql"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type pendingBatchSpecHandler struct {
	s *store.Store
}

func (h *pendingBatchSpecHandler) HandlerFunc() dbworker.HandlerFunc {
	return func(ctx context.Context, tx dbworkerstore.Store, record workerutil.Record) error {
		s := h.s.With(tx)
		log15.Info("processing pending batch spec", "record", record, "s", s)

		// TODO: transform the pending batch spec into a set of executor jobs.
		return nil
	}
}

type pendingBatchSpecRecord struct {
	*btypes.PendingBatchSpec
}

var _ workerutil.Record = pendingBatchSpecRecord{}

func (r pendingBatchSpecRecord) RecordID() int {
	return int(r.ID)
}

func newPendingBatchSpecWorker(
	ctx context.Context,
	s *store.Store,
	metrics batchChangesMetrics,
) *workerutil.Worker {
	handler := &pendingBatchSpecHandler{s}
	workerStore := newPendingBatchSpecWorkerStore(s)
	options := workerutil.WorkerOptions{
		Name:        "batches_pending_batch_spec_worker",
		NumHandlers: 5,
		Interval:    5 * time.Second,
		Metrics:     metrics.pendingBatchSpecWorkerMetrics,
	}

	return dbworker.NewWorker(ctx, workerStore, handler.HandlerFunc(), options)
}

func newPendingBatchSpecWorkerResetter(s *store.Store, metrics batchChangesMetrics) *dbworker.Resetter {
	workerStore := newPendingBatchSpecWorkerStore(s)
	options := dbworker.ResetterOptions{
		Name:     "batches_pending_batch_spec_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.pendingBatchSpecWorkerResetterMetrics,
	}

	return dbworker.NewResetter(workerStore, options)
}

func newPendingBatchSpecWorkerStore(s *store.Store) dbworkerstore.Store {
	return dbworkerstore.New(s.Handle(), dbworkerstore.Options{
		Name:              "batches_pending_batch_spec_store",
		TableName:         "pending_batch_specs",
		ColumnExpressions: store.PendingBatchSpecColumns,
		Scan: func(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
			pbs, exists, err := store.ScanFirstPendingBatchSpec(rows, err)
			return pendingBatchSpecRecord{pbs}, exists, err
		},
		OrderByExpression: sqlf.Sprintf("pending_batch_specs.updated_at ASC"),
		StalledMaxAge:     60 * time.Second,
		MaxNumResets:      60,
		RetryAfter:        5 * time.Second,
		MaxNumRetries:     60,
	})
}
