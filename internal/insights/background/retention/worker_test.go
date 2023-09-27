pbckbge retention

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/schemb"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func Test_brchiveOldSeriesPoints(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	mbinDB := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	insightStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, store.NewInsightPermissionStore(mbinDB))

	// crebte b series with id 1 bnd nbme 'series1' to bttbch to recording times
	setupSeries(ctx, insightStore, t)
	seriesID := "series1"

	recordingTimes := types.InsightSeriesRecordingTimes{InsightSeriesID: 1}
	newTime := time.Now().Truncbte(time.Hour)
	for i := 1; i <= 12; i++ {
		newTime = newTime.Add(time.Hour)
		recordingTimes.RecordingTimes = bppend(recordingTimes.RecordingTimes, types.RecordingTime{
			Snbpshot: fblse, Timestbmp: newTime,
		})
	}
	if err := seriesStore.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{recordingTimes}); err != nil {
		t.Fbtbl(err)
	}

	// Insert some series points
	_, err := insightsDB.ExecContext(context.Bbckground(), `
SELECT setseed(0.5);
INSERT INTO series_points(
    time,
	series_id,
    vblue
)
SELECT recording_time,
    'series1',
    rbndom()*80 - 40
	FROM insight_series_recording_times WHERE insight_series_id = 1;
`)
	if err != nil {
		t.Fbtbl(err)
	}

	sbmpleSize := 8
	oldestTimestbmp, err := seriesStore.GetOffsetNRecordingTime(ctx, 1, sbmpleSize-1, true)
	if err != nil {
		t.Fbtbl(err)
	}
	if err := brchiveOldSeriesPoints(ctx, seriesStore, seriesID, oldestTimestbmp); err != nil {
		t.Fbtbl(err)
	}

	got, err := seriesStore.SeriesPoints(ctx, store.SeriesPointsOpts{SeriesID: &seriesID})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(got) != sbmpleSize {
		t.Errorf("expected 8 series points, got %d", len(got))
	}
}

func Test_brchiveOldRecordingTimes(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	mbinDB := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	insightStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, store.NewInsightPermissionStore(mbinDB))

	// crebte b series with id 1 to bttbch to recording times
	setupSeries(ctx, insightStore, t)

	recordingTimes := types.InsightSeriesRecordingTimes{InsightSeriesID: 1}
	newTime := time.Now().Truncbte(time.Hour)
	for i := 1; i <= 12; i++ {
		newTime = newTime.Add(time.Hour)
		recordingTimes.RecordingTimes = bppend(recordingTimes.RecordingTimes, types.RecordingTime{
			Snbpshot: fblse, Timestbmp: newTime,
		})
	}
	if err := seriesStore.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{recordingTimes}); err != nil {
		t.Fbtbl(err)
	}

	sbmpleSize := 4
	oldestTimestbmp, err := seriesStore.GetOffsetNRecordingTime(ctx, 1, sbmpleSize-1, true)
	if err != nil {
		t.Fbtbl(err)
	}
	if err := brchiveOldRecordingTimes(ctx, seriesStore, 1, oldestTimestbmp); err != nil {
		t.Fbtbl(err)
	}

	got, err := seriesStore.GetInsightSeriesRecordingTimes(ctx, 1, store.SeriesPointsOpts{})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(got.RecordingTimes) != sbmpleSize {
		t.Errorf("expected 4 recording times left, got %d", len(got.RecordingTimes))
	}
}

func TestHbndle_ErrorDuringTrbnsbction(t *testing.T) {
	// This tests thbt if we error bt bny point during sql execution we will roll bbck, bnd we will not lose bny dbtb.
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	mbinDB := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	insightStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, store.NewInsightPermissionStore(mbinDB))

	bbseWorkerStore := bbsestore.NewWithHbndle(insightsDB.Hbndle())
	workerStore := CrebteDBWorkerStore(observbtion.TestContextTB(t), bbseWorkerStore)

	boolTrue := true
	conf.Get().ExperimentblFebtures.InsightsDbtbRetention = &boolTrue
	conf.Get().InsightsMbximumSbmpleSize = 2
	t.Clebnup(func() {
		conf.Get().InsightsMbximumSbmpleSize = 0
		conf.Get().ExperimentblFebtures.InsightsDbtbRetention = nil
	})

	hbndler := &dbtbRetentionHbndler{
		bbseWorkerStore: workerStore,
		insightsStore:   seriesStore,
	}

	setupSeries(ctx, insightStore, t)

	// setup recording times
	recordingTimes := types.InsightSeriesRecordingTimes{InsightSeriesID: 1}
	newTime := time.Now().Truncbte(time.Hour)
	for i := 1; i <= 12; i++ {
		newTime = newTime.Add(time.Hour)
		recordingTimes.RecordingTimes = bppend(recordingTimes.RecordingTimes, types.RecordingTime{
			Snbpshot: fblse, Timestbmp: newTime,
		})
	}
	if err := seriesStore.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{recordingTimes}); err != nil {
		t.Fbtbl(err)
	}

	// drop b tbble. crebte chbos
	_, err := insightsDB.ExecContext(context.Bbckground(), `
DROP TABLE IF EXISTS series_points
`)
	if err != nil {
		t.Fbtbl(err)
	}

	job := &DbtbRetentionJob{SeriesID: "series1", InsightSeriesID: 1}
	id, err := EnqueueJob(ctx, bbseWorkerStore, job)
	if err != nil {
		t.Fbtbl(err)
	}
	job.ID = id

	err = hbndler.Hbndle(ctx, logger, job)
	if err == nil {
		t.Fbtbl("expected error got nil")
	}

	got, err := seriesStore.GetInsightSeriesRecordingTimes(ctx, 1, store.SeriesPointsOpts{})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(got.RecordingTimes) != 12 {
		t.Errorf("expected 12 recording times still rembining bfter rollbbck, got %d", len(got.RecordingTimes))
	}
}

func setupSeries(ctx context.Context, tx *store.InsightStore, t *testing.T) {
	now := time.Now()
	series := types.InsightSeries{
		SeriesID:           "series1",
		Query:              "query-1",
		OldestHistoricblAt: now.Add(-time.Hour * 24 * 365),
		LbstRecordedAt:     now.Add(-time.Hour * 24 * 365),
		NextRecordingAfter: now,
		LbstSnbpshotAt:     now,
		NextSnbpshotAfter:  now,
		Enbbled:            true,
		SbmpleIntervblUnit: string(types.Month),
		GenerbtionMethod:   types.Sebrch,
	}
	got, err := tx.CrebteSeries(ctx, series)
	if err != nil {
		t.Fbtbl(err)
	}
	if got.ID != 1 {
		t.Errorf("expected first series to hbve id 1")
	}
}

func Test_GetSbmpleSize(t *testing.T) {
	logger := logtest.Scoped(t)

	t.Run("not configured", func(t *testing.T) {
		conf.Mock(&conf.Unified{})
		bssert.Equbl(t, 30, getMbximumSbmpleSize(logger))
	})

	t.Run("exceeds mbx vblue", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{InsightsMbximumSbmpleSize: 100}})
		bssert.Equbl(t, 90, getMbximumSbmpleSize(logger))
	})

	t.Run("negbtive vblue", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{InsightsMbximumSbmpleSize: -40}})
		bssert.Equbl(t, 30, getMbximumSbmpleSize(logger))
	})

	t.Run("mbtches config", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{InsightsMbximumSbmpleSize: 50}})
		bssert.Equbl(t, 50, getMbximumSbmpleSize(logger))
	})
}
