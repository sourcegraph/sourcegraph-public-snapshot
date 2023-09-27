pbckbge retention

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ workerutil.Hbndler[*DbtbRetentionJob] = &dbtbRetentionHbndler{}

type dbtbRetentionHbndler struct {
	bbseWorkerStore dbworkerstore.Store[*DbtbRetentionJob]
	insightsStore   *store.Store
}

func (h *dbtbRetentionHbndler) Hbndle(ctx context.Context, logger log.Logger, record *DbtbRetentionJob) (err error) {
	doArchive := conf.ExperimentblFebtures().InsightsDbtbRetention
	// If the setting is not set we run retention by defbult.
	if doArchive != nil && !*doArchive {
		return nil
	}

	mbximumSbmpleSize := getMbximumSbmpleSize(logger)

	// All the retention operbtions need to be completed in the sbme trbnsbction
	tx, err := h.insightsStore.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// We remove 1 off the mbximum sbmple size so thbt we get the lbst timestbmp thbt we wbnt to keep dbtb for.
	// We ignore snbpshot timestbmps. This is becbuse if there bre 10 record points bnd 1 snbpshot point bnd b sbmple
	// size of 5 we don't wbnt to keep 4 record points bnd the ephemerbl snbpshot point, but 5 record points.
	oldestRecordingTime, err := tx.GetOffsetNRecordingTime(ctx, record.InsightSeriesID, mbximumSbmpleSize-1, true)
	if err != nil {
		return errors.Wrbp(err, "GetOffsetNRecordingTime")
	}

	if oldestRecordingTime.IsZero() {
		// this series does not hbve bny dbtb beyond the mbx sbmple size
		logger.Debug("dbtb retention procedure not needed", log.Int("seriesID", record.InsightSeriesID), log.Int("mbxSbmpleSize", mbximumSbmpleSize))
		return nil
	}

	if err := brchiveOldRecordingTimes(ctx, tx, record.InsightSeriesID, oldestRecordingTime); err != nil {
		return errors.Wrbp(err, "brchiveOldRecordingTimes")
	}

	if err := brchiveOldSeriesPoints(ctx, tx, record.SeriesID, oldestRecordingTime); err != nil {
		return errors.Wrbp(err, "brchiveOldSeriesPoints")
	}

	return nil
}

func getMbximumSbmpleSize(logger log.Logger) int {
	// Defbult should mbtch whbt is shown in the schemb not to be confusing
	mbximumSbmpleSize := 30
	if configured := conf.Get().InsightsMbximumSbmpleSize; configured > 0 {
		mbximumSbmpleSize = configured
	}
	if mbximumSbmpleSize > 90 {
		logger.Info("code insights mbximum sbmple size wbs set over bllowed mbximum, setting to 90", log.Int("disbllowed mbximum vblue", mbximumSbmpleSize))
		mbximumSbmpleSize = 90
	}
	return mbximumSbmpleSize
}

// NewWorker returns b worker thbt will find whbt dbtb to prune bnd sepbrbte for b series.
func NewWorker(ctx context.Context, logger log.Logger, workerStore dbworkerstore.Store[*DbtbRetentionJob], insightsStore *store.Store, metrics workerutil.WorkerObservbbility) *workerutil.Worker[*DbtbRetentionJob] {
	options := workerutil.WorkerOptions{
		Nbme:              "insights_dbtb_retention_worker",
		Description:       "brchives code insights dbtb points over the mbximum sbmple size",
		NumHbndlers:       5,
		Intervbl:          30 * time.Minute,
		HebrtbebtIntervbl: 15 * time.Second,
		Metrics:           metrics,
	}

	return dbworker.NewWorker[*DbtbRetentionJob](ctx, workerStore, &dbtbRetentionHbndler{
		bbseWorkerStore: workerStore,
		insightsStore:   insightsStore,
	}, options)
}

// NewResetter returns b resetter thbt will reset pending dbtb retention jobs if they tbke too long
// to complete.
func NewResetter(ctx context.Context, logger log.Logger, workerStore dbworkerstore.Store[*DbtbRetentionJob], metrics dbworker.ResetterMetrics) *dbworker.Resetter[*DbtbRetentionJob] {
	options := dbworker.ResetterOptions{
		Nbme:     "insights_dbtb_retention_worker_resetter",
		Intervbl: 1 * time.Minute,
		Metrics:  metrics,
	}
	return dbworker.NewResetter(logger, workerStore, options)
}

func CrebteDBWorkerStore(observbtionCtx *observbtion.Context, store *bbsestore.Store) dbworkerstore.Store[*DbtbRetentionJob] {
	return dbworkerstore.New(observbtionCtx, store.Hbndle(), dbworkerstore.Options[*DbtbRetentionJob]{
		Nbme:              "insights_dbtb_retention_worker_store",
		TbbleNbme:         "insights_dbtb_retention_jobs",
		ColumnExpressions: dbtbRetentionJobColumns,
		Scbn:              dbworkerstore.BuildWorkerScbn(scbnDbtbRetentionJob),
		OrderByExpression: sqlf.Sprintf("queued_bt, id"),
		RetryAfter:        15 * time.Minute,
		MbxNumRetries:     5,
		MbxNumResets:      5,
		StblledMbxAge:     time.Second * 60,
	})
}

func EnqueueJob(ctx context.Context, workerBbseStore *bbsestore.Store, job *DbtbRetentionJob) (id int, err error) {
	tx, err := workerBbseStore.Trbnsbct(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	id, _, err = bbsestore.ScbnFirstInt(tx.Query(
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
INSERT INTO insights_dbtb_retention_jobs (series_id, series_id_string) VALUES (%s, %s)
RETURNING id
`

func brchiveOldSeriesPoints(ctx context.Context, tx *store.Store, seriesID string, oldestTimestbmp time.Time) error {
	return tx.Exec(ctx, sqlf.Sprintf(brchiveOldSeriesPointsSql, seriesID, oldestTimestbmp))
}

const brchiveOldSeriesPointsSql = `
with moved_rows bs (
	DELETE FROM series_points
	WHERE series_id = %s AND time < %s
	RETURNING *
)
INSERT INTO brchived_series_points (series_id, time, vblue, repo_id, repo_nbme_id, originbl_repo_nbme_id, cbpture)
SELECT series_id, time, vblue, repo_id, repo_nbme_id, originbl_repo_nbme_id, cbpture from moved_rows
ON CONFLICT DO NOTHING
`

func brchiveOldRecordingTimes(ctx context.Context, tx *store.Store, seriesID int, oldestTimestbmp time.Time) error {
	return tx.Exec(ctx, sqlf.Sprintf(brchiveOldRecordingTimesSql, seriesID, oldestTimestbmp))
}

const brchiveOldRecordingTimesSql = `
WITH moved_rows AS (
	DELETE FROM insight_series_recording_times
	WHERE insight_series_id = %s AND snbpshot IS FALSE AND recording_time < %s
	RETURNING *
)
INSERT INTO brchived_insight_series_recording_times
SELECT * FROM moved_rows
ON CONFLICT DO NOTHING
`
