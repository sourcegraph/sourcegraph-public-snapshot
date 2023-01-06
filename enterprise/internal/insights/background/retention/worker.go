package retention

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ workerutil.Handler[*DataRetentionJob] = &dataRetentionHandler{}

type dataRetentionHandler struct {
	baseWorkerStore dbworkerstore.Store[*DataRetentionJob]
	insightsStore   *store.Store
}

func (h *dataRetentionHandler) Handle(ctx context.Context, logger log.Logger, record *DataRetentionJob) (err error) {
	// Default should match what is shown in the schema not to be confusing
	maximumSampleSize := 90
	if configured := conf.Get().InsightsMaximumSampleSize; configured >= 0 {
		maximumSampleSize = configured
	}

	// All the retention operations need to be completed in the same transaction
	tx, err := h.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	oldestRecordingTime, err := selectOldestRecordingTimeBeforeMax(ctx, tx, record.InsightSeriesID, maximumSampleSize)
	if err != nil {
		return errors.Wrap(err, "selectOldestRecordingTimeBeforeMax")
	}

	if oldestRecordingTime == nil {
		// this series does not have any data beyond the max sample size
		logger.Debug("data retention procedure not needed", log.Int("seriesID", record.InsightSeriesID), log.Int("maxSampleSize", maximumSampleSize))
		return nil
	}

	if err := archiveOldRecordingTimes(ctx, tx, record.InsightSeriesID, *oldestRecordingTime); err != nil {
		return err
	}

	if err := archiveOldSeriesPoints(ctx, tx, record.SeriesID, *oldestRecordingTime); err != nil {
		return err
	}

	return nil
}

// NewWorker returns a worker that will find what data to prune and separate for a series.
func NewWorker(ctx context.Context, logger log.Logger, workerStore dbworkerstore.Store[*DataRetentionJob], insightsStore *store.Store, metrics workerutil.WorkerObservability) *workerutil.Worker[*DataRetentionJob] {
	options := workerutil.WorkerOptions{
		Name:              "insights_data_retention_worker",
		NumHandlers:       5,
		Interval:          12 * time.Hour,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           metrics,
	}

	return dbworker.NewWorker[*DataRetentionJob](ctx, workerStore, &dataRetentionHandler{
		baseWorkerStore: workerStore,
		insightsStore:   insightsStore,
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
		RetryAfter:        15 * time.Minute,
		MaxNumRetries:     5,
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
			job.InsightSeriesID,
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
INSERT INTO insights_data_retention_jobs (series_id, series_id_string) VALUES (%s, %s)
RETURNING id
`

func selectOldestRecordingTimeBeforeMax(ctx context.Context, tx *store.Store, seriesID int, maxSampleSize int) (_ *time.Time, err error) {
	oldestTime, got, err := basestore.ScanFirstTime(tx.Query(ctx, sqlf.Sprintf(selectOldestRecordingTimeSql, seriesID, maxSampleSize-1)))
	if err != nil {
		return nil, err
	}
	if !got || oldestTime.IsZero() {
		return nil, nil
	}
	return &oldestTime, nil
}

const selectOldestRecordingTimeSql = `
SELECT recording_time FROM insight_series_recording_times
WHERE insight_series_id = %s AND snapshot IS FALSE
ORDER BY recording_time DESC OFFSET %s LIMIT 1;`

// archiveOldSeriesPoints will insert old series points in a separate table and then delete them from the main table.
func archiveOldSeriesPoints(ctx context.Context, tx *store.Store, seriesID string, oldestTimestamp time.Time) error {
	if err := tx.Exec(ctx, sqlf.Sprintf(insertSeriesPointsSql, seriesID, oldestTimestamp)); err != nil {
		return errors.Wrap(err, "insertSeriesPoints")
	}
	if err := tx.Exec(ctx, sqlf.Sprintf(deleteSeriesPointsSql, seriesID, oldestTimestamp)); err != nil {
		return errors.Wrap(err, "deleteSeriesPoints")
	}
	return nil
}

const insertSeriesPointsSql = `
INSERT INTO archived_series_points
(SELECT * FROM series_points WHERE series_id = %s AND time < %s)
`

const deleteSeriesPointsSql = `
DELETE FROM series_points 
WHERE series_id = %s AND time < %s
`

// archiveOldRecordingTimes will insert old recording times in a separate table and then delete them from the main table.
func archiveOldRecordingTimes(ctx context.Context, tx *store.Store, seriesID int, oldestTimestamp time.Time) error {
	if err := tx.Exec(ctx, sqlf.Sprintf(insertRecordingTimesSql, seriesID, oldestTimestamp)); err != nil {
		return errors.Wrap(err, "insertRecordingTimes")
	}
	if err := tx.Exec(ctx, sqlf.Sprintf(deleteRecordingTimesSql, seriesID, oldestTimestamp)); err != nil {
		return errors.Wrap(err, "deleteRecordingTimes")
	}
	return nil
}

const insertRecordingTimesSql = `
INSERT INTO archived_insight_series_recording_times 
(SELECT * FROM insight_series_recording_times WHERE insight_series_id = %s AND snapshot IS FALSE AND recording_time < %s)
ON CONFLICT DO NOTHING
`

const deleteRecordingTimesSql = `
DELETE FROM insight_series_recording_times 
WHERE insight_series_id = %s AND snapshot IS FALSE and recording_time < %s
`
