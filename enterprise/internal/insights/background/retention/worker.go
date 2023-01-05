package retention

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
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
	logger.Info("data retention handler called", log.Int("seriesID", record.SeriesID))

	// Default should match what is shown in the schema not to be confusing
	maximumSampleSize := 90
	if configured := conf.Get().InsightsMaximumSampleSize; configured != 0 {
		maximumSampleSize = configured
	}
	logger.Info("maximum sample size", log.Int("value", maximumSampleSize))

	// All the retention operations need to be completed in the same transaction
	tx, err := h.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	oldestRecordingTimes, err := selectRecordingTimesAfterMax(ctx, tx, record.SeriesID, maximumSampleSize)
	if err != nil {
		return errors.Wrap(err, "selectRecordingTimesAfterMax")
	}

	if len(oldestRecordingTimes.RecordingTimes) == 0 {
		// this series does not have any data beyond the max sample size
		logger.Info("data retention procedure not needed", log.Int("seriesID", record.SeriesID), log.Int("maxSampleSize", maximumSampleSize))
		return nil
	}

	return nil
}

// NewWorker returns a worker that will find what data to prune and separate for a series.
func NewWorker(ctx context.Context, logger log.Logger, workerStore dbworkerstore.Store[*DataRetentionJob], insightsStore *store.Store, metrics workerutil.WorkerObservability) *workerutil.Worker[*DataRetentionJob] {
	options := workerutil.WorkerOptions{
		Name:              "insights_data_retention_worker",
		NumHandlers:       5,
		Interval:          5 * time.Second,
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

func selectRecordingTimesAfterMax(ctx context.Context, tx *store.Store, seriesID int, maxSampleSize int) (_ *types.InsightSeriesRecordingTimes, err error) {
	rows, err := tx.Query(ctx, sqlf.Sprintf(selectOldestRecordingTimesSql, seriesID, maxSampleSize, seriesID))
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var oldestSamples []types.RecordingTime
	for rows.Next() {
		var recordingTime time.Time
		if err := rows.Scan(&recordingTime); err != nil {
			return nil, err
		}
		oldestSamples = append(oldestSamples, types.RecordingTime{Timestamp: recordingTime})
	}

	return &types.InsightSeriesRecordingTimes{InsightSeriesID: seriesID, RecordingTimes: oldestSamples}, nil
}

const selectOldestRecordingTimesSql = `
SELECT recording_time FROM insight_series_recording_times
WHERE insight_series_id = %s AND snapshot IS FALSE
ORDER BY recording_time ASC
LIMIT (SELECT GREATEST(count(*) - %s, 0) FROM insight_series_recording_times WHERE insight_series_id = %s);
`

// archiveOldSeriesPoints will delete old series points from the series_points table and move them to a separate
// table.
func archiveOldSeriesPoints(ctx context.Context, tx *store.Store, seriesID string, oldestTimestamp time.Time) error {
	return nil
}

// archiveOldRecordingTimes will delete old recording times from the recording times table and move them to a separate
// table.
func archiveOldRecordingTimes(ctx context.Context, tx *store.Store, seriesID int, oldestTimestamp time.Time) error {
	return nil
}
