package retention

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

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
