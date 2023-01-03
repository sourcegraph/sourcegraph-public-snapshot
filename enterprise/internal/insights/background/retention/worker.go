package retention

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

var _ workerutil.Handler[*DataPruningJob] = &dataPruningHandler{}

type dataPruningHandler struct {
}

func (h *dataPruningHandler) Handle(ctx context.Context, logger log.Logger, record *DataPruningJob) error {
	logger.Debug("data pruning handler called", log.Int("seriesID", record.SeriesID))
	return nil
}

func makeStore(observationCtx *observation.Context, dbHandle basestore.TransactableHandle) dbworkerstore.Store[*DataPruningJob] {
	return dbworkerstore.New(observationCtx, dbHandle, dbworkerstore.Options[*DataPruningJob]{
		Name:              "insights_data_pruning_job_worker_store",
		TableName:         "insights_data_pruning_jobs",
		ColumnExpressions: dataPruningJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanDataPruningJob),
		OrderByExpression: sqlf.Sprintf("queued_at", "id"),
		MaxNumResets:      5,
		StalledMaxAge:     time.Second * 5,
	})
}

func EnqueueJob(ctx context.Context, workerBaseStore *basestore.Store, job *DataPruningJob) (id int, err error) {
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
INSERT INTO insights_data_pruning_jobs (series_id) VALUES (%s)
RETURNING id
`
