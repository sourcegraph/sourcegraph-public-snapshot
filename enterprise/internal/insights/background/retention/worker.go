package retention

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

var _ workerutil.Handler[*DataRetentionJob] = &dataRetentionHandler{}

type dataRetentionHandler struct {
	baseWorkerStore dbworkerstore.Store[*DataRetentionJob]
}

func (h *dataRetentionHandler) Handle(ctx context.Context, logger log.Logger, record *DataRetentionJob) error {
	logger.Debug("data retention handler called", log.Int("seriesID", record.SeriesID))
	return nil
}

// NewWorker returns a worker that will find what data to prune and separate for a series.
func NewWorker(ctx context.Context, logger log.Logger, workerStore dbworkerstore.Store[*DataRetentionJob], metrics workerutil.WorkerObservability) *workerutil.Worker[*DataRetentionJob] {
	options := workerutil.WorkerOptions{
		Name:              "insights_data_retention_worker",
		NumHandlers:       5,
		Interval:          5 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           metrics,
	}

	return dbworker.NewWorker[*DataRetentionJob](ctx, workerStore, &dataRetentionHandler{
		baseWorkerStore: workerStore,
	}, options)
}

// NewResetter returns a resetter that will reset pending data retention jobs if they take too long
// to complete.
func NewResetter(ctx context.Context, logger log.Logger, workerStore dbworkerstore.Store[*DataRetentionJob], metrics dbworker.ResetterMetrics) *dbworker.Resetter[*DataRetentionJob] {
	options := dbworker.ResetterOptions{
		Name:     "insights_data_retention_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics,
	}
	return dbworker.NewResetter(logger, workerStore, options)
}

func CreateDBWorkerStore(observationCtx *observation.Context, store *basestore.Store) dbworkerstore.Store[*DataRetentionJob] {
	return dbworkerstore.New(observationCtx, store.Handle(), dbworkerstore.Options[*DataRetentionJob]{
		Name:              "insights_data_retention_worker_store",
		TableName:         "insights_data_retention_jobs",
		ColumnExpressions: dataRetentionJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanDataRetentionJob),
		OrderByExpression: sqlf.Sprintf("queued_at, id"),
		MaxNumResets:      5,
		StalledMaxAge:     time.Second * 5,
	})
}

func EnqueueJob(ctx context.Context, workerBaseStore *basestore.Store, job *DataRetentionJob) (id int, err error) {
	tx, err := workerBaseStore.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	id, _, err = basestore.ScanFirstInt(tx.Query(
		ctx,
		sqlf.Sprintf(
			enqueueJobFmtStr,
			job.SeriesID,
		),
	))
	if err != nil {
		return 0, err
	}
	job.ID = id
	return id, nil
}

const enqueueJobFmtStr = `
INSERT INTO insights_data_retention_jobs (series_id) VALUES (%s)
RETURNING id
`
