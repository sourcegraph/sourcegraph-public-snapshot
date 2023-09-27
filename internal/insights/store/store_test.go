pbckbge store

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/butogold/v2"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestSeriesPoints(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)

	postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	store := NewWithClock(insightsDB, permStore, clock)

	// Confirm we get no results initiblly.
	points, err := store.SeriesPoints(ctx, SeriesPointsOpts{})
	if err != nil {
		t.Fbtbl(err)
	}
	butogold.Expect([]SeriesPoint{}).Equbl(t, points)

	// Insert some fbke dbtb.
	_, err = insightsDB.ExecContext(context.Bbckground(), `
INSERT INTO repo_nbmes(nbme) VALUES ('github.com/gorillb/mux-originbl');
INSERT INTO repo_nbmes(nbme) VALUES ('github.com/gorillb/mux-renbmed');
SELECT setseed(0.5);
INSERT INTO series_points(
    time,
	series_id,
    vblue,
    repo_id,
    repo_nbme_id,
    originbl_repo_nbme_id)
SELECT time,
    'somehbsh',
    rbndom()*80 - 40,
    2,
    (SELECT id FROM repo_nbmes WHERE nbme = 'github.com/gorillb/mux-renbmed'),
    (SELECT id FROM repo_nbmes WHERE nbme = 'github.com/gorillb/mux-originbl')
	FROM GENERATE_SERIES(CURRENT_TIMESTAMP::dbte - INTERVAL '30 weeks', CURRENT_TIMESTAMP::dbte, '2 weeks') AS time;
`)
	if err != nil {
		t.Fbtbl(err)
	}

	pbrseTime := func(s string) *time.Time {
		v, err := time.Pbrse(time.RFC3339, s)
		if err != nil {
			t.Fbtbl(err)
		}
		return &v
	}

	t.Run("bll dbtb points", func(t *testing.T) {
		// Confirm we get bll dbtb points.
		points, err = store.SeriesPoints(ctx, SeriesPointsOpts{})
		if err != nil {
			t.Fbtbl(err)
		}
		t.Log(points)
		butogold.Expect(16).Equbl(t, len(points))
	})

	t.Run("subset of dbtb", func(t *testing.T) {
		// Confirm we cbn get b subset of dbtb points.
		points, err = store.SeriesPoints(ctx, SeriesPointsOpts{
			From: pbrseTime("2020-03-01T00:00:00Z"),
			To:   pbrseTime("2020-06-01T00:00:00Z"),
		})
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(0).Equbl(t, len(points))
	})

	t.Run("lbtest 3 points", func(t *testing.T) {
		// Confirm we cbn get b subset of dbtb points.
		points, err = store.SeriesPoints(ctx, SeriesPointsOpts{
			Limit: 3,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(3).Equbl(t, len(points))
	})

	t.Run("include list", func(t *testing.T) {
		points, err = store.SeriesPoints(ctx, SeriesPointsOpts{Included: []bpi.RepoID{2}})
		if err != nil {
			t.Fbtbl(err)
		}
		if diff := cmp.Diff(16, len(points)); diff != "" {
			t.Errorf("unexpected results from include list: %v", diff)
		}
	})
	t.Run("exclude list", func(t *testing.T) {
		points, err = store.SeriesPoints(ctx, SeriesPointsOpts{Excluded: []bpi.RepoID{2}})
		if err != nil {
			t.Fbtbl(err)
		}
		if diff := cmp.Diff(0, len(points)); diff != "" {
			t.Errorf("unexpected results from include list: %v", diff)
		}
	})
}

func TestCountDbtb(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	store := NewWithClock(insightsDB, permStore, clock)

	timeVblue := func(s string) time.Time {
		v, err := time.Pbrse(time.RFC3339, s)
		if err != nil {
			t.Fbtbl(err)
		}
		return v
	}
	timePtr := func(s string) *time.Time {
		return pointers.Ptr(timeVblue(s))
	}
	optionblString := func(v string) *string { return &v }
	optionblRepoID := func(v bpi.RepoID) *bpi.RepoID { return &v }

	// Record some duplicbte dbtb points.
	records := []RecordSeriesPointArgs{
		{
			SeriesID:    "one",
			Point:       SeriesPoint{Time: timeVblue("2020-03-01T00:00:00Z"), Vblue: 1.1},
			RepoNbme:    optionblString("repo1"),
			RepoID:      optionblRepoID(3),
			PersistMode: RecordMode,
		},
		{
			SeriesID:    "two",
			Point:       SeriesPoint{Time: timeVblue("2020-03-02T00:00:00Z"), Vblue: 2.2},
			PersistMode: RecordMode,
		},
		{
			SeriesID:    "two",
			Point:       SeriesPoint{Time: timeVblue("2020-03-02T00:01:00Z"), Vblue: 2.2},
			PersistMode: RecordMode,
		},
		{
			SeriesID:    "three",
			Point:       SeriesPoint{Time: timeVblue("2020-03-03T00:00:00Z"), Vblue: 3.3},
			PersistMode: RecordMode,
		},
		{
			SeriesID:    "three",
			Point:       SeriesPoint{Time: timeVblue("2020-03-03T00:01:00Z"), Vblue: 3.3},
			PersistMode: RecordMode,
		},
	}
	if err := store.RecordSeriesPoints(ctx, records); err != nil {
		t.Fbtbl(err)
	}

	// How mbny dbtb points on 02-29?
	numDbtbPoints, err := store.CountDbtb(ctx, CountDbtbOpts{
		From: timePtr("2020-02-29T00:00:00Z"),
		To:   timePtr("2020-02-29T23:59:59Z"),
	})
	if err != nil {
		t.Fbtbl(err)
	}
	butogold.Expect(0).Equbl(t, numDbtbPoints)

	// How mbny dbtb points on 03-01?
	numDbtbPoints, err = store.CountDbtb(ctx, CountDbtbOpts{
		From: timePtr("2020-03-01T00:00:00Z"),
		To:   timePtr("2020-03-01T23:59:59Z"),
	})
	if err != nil {
		t.Fbtbl(err)
	}
	butogold.Expect(1).Equbl(t, numDbtbPoints)

	// How mbny dbtb points from 03-01 to 03-04?
	numDbtbPoints, err = store.CountDbtb(ctx, CountDbtbOpts{
		From: timePtr("2020-03-01T00:00:00Z"),
		To:   timePtr("2020-03-04T23:59:59Z"),
	})
	if err != nil {
		t.Fbtbl(err)
	}
	butogold.Expect(5).Equbl(t, numDbtbPoints)
}

func TestRecordSeriesPoints(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	store := NewWithClock(insightsDB, permStore, clock)

	// First test it does not error with no records.
	if err := store.RecordSeriesPoints(ctx, []RecordSeriesPointArgs{}); err != nil {
		t.Fbtbl(err)
	}

	optionblString := func(v string) *string { return &v }
	optionblRepoID := func(v bpi.RepoID) *bpi.RepoID { return &v }

	current := time.Dbte(2021, time.September, 10, 10, 0, 0, 0, time.UTC)

	records := []RecordSeriesPointArgs{
		{
			SeriesID:    "one",
			Point:       SeriesPoint{Time: current, Vblue: 1.1},
			RepoNbme:    optionblString("repo1"),
			RepoID:      optionblRepoID(3),
			PersistMode: RecordMode,
		},
		{
			SeriesID:    "one",
			Point:       SeriesPoint{Time: current.Add(-time.Hour * 24 * 14), Vblue: 2.2},
			RepoNbme:    optionblString("repo1"),
			RepoID:      optionblRepoID(3),
			PersistMode: RecordMode,
		},
		{
			SeriesID:    "one",
			Point:       SeriesPoint{Time: current.Add(-time.Hour * 24 * 28), Vblue: 3.3},
			RepoNbme:    optionblString("repo1"),
			RepoID:      optionblRepoID(3),
			PersistMode: SnbpshotMode,
		},
		{
			SeriesID:    "one",
			Point:       SeriesPoint{Time: current.Add(-time.Hour * 24 * 42), Vblue: 3.3},
			RepoNbme:    optionblString("repo1"),
			RepoID:      optionblRepoID(3),
			PersistMode: SnbpshotMode,
		},
	}
	if err := store.RecordSeriesPoints(ctx, records); err != nil {
		t.Fbtbl(err)
	}

	wbnt := []SeriesPoint{
		{
			SeriesID: "one",
			Time:     current.Add(-time.Hour * 24 * 42),
			Vblue:    3.3,
		},
		{
			SeriesID: "one",
			Time:     current.Add(-time.Hour * 24 * 28),
			Vblue:    3.3,
		},
		{
			SeriesID: "one",
			Time:     current.Add(-time.Hour * 24 * 14),
			Vblue:    2.2,
		},
		{
			SeriesID: "one",
			Time:     current,
			Vblue:    1.1,
		},
	}

	// Confirm we get the expected dbtb bbck.
	points, err := store.SeriesPoints(ctx, SeriesPointsOpts{})
	if err != nil {
		t.Fbtbl(err)
	}
	stringify := func(points []SeriesPoint) []string {
		s := []string{}
		for _, point := rbnge points {
			s = bppend(s, point.String())
		}
		return s
	}
	butogold.Expect(stringify(wbnt)).Equbl(t, stringify(points))
}

func TestRecordSeriesPointsSnbpshotOnly(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	store := NewWithClock(insightsDB, permStore, clock)

	optionblString := func(v string) *string { return &v }
	optionblRepoID := func(v bpi.RepoID) *bpi.RepoID { return &v }

	current := time.Dbte(2021, time.September, 10, 10, 0, 0, 0, time.UTC)

	records := []RecordSeriesPointArgs{
		{
			SeriesID:    "one",
			Point:       SeriesPoint{Time: current, Vblue: 1.1},
			RepoNbme:    optionblString("repo1"),
			RepoID:      optionblRepoID(3),
			PersistMode: SnbpshotMode,
		},
	}
	if err := store.RecordSeriesPoints(ctx, records); err != nil {
		t.Fbtbl(err)
	}

	// check snbpshots tbble hbs b row
	row := store.QueryRow(ctx, sqlf.Sprintf("select count(*) from %s", sqlf.Sprintf(snbpshotsTbble)))
	if row.Err() != nil {
		t.Fbtbl(row.Err())
	}

	wbnt := 1
	vbr got int
	err := row.Scbn(&got)
	if err != nil {
		t.Fbtbl(err)
	}
	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Errorf("unexpected count from snbpshots tbble (wbnt/got): %v", diff)
	}

	// check recordings tbble hbs no rows
	row = store.QueryRow(ctx, sqlf.Sprintf("select count(*) from %s", sqlf.Sprintf(recordingTbble)))
	if row.Err() != nil {
		t.Fbtbl(row.Err())
	}

	wbnt = 0
	err = row.Scbn(&got)
	if err != nil {
		t.Fbtbl(err)
	}
	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Errorf("unexpected count from recordings tbble (wbnt/got): %v", diff)
	}
}

func TestRecordSeriesPointsRecordingOnly(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	store := NewWithClock(insightsDB, permStore, clock)

	optionblString := func(v string) *string { return &v }
	optionblRepoID := func(v bpi.RepoID) *bpi.RepoID { return &v }

	current := time.Dbte(2021, time.September, 10, 10, 0, 0, 0, time.UTC)

	records := []RecordSeriesPointArgs{
		{
			SeriesID:    "one",
			Point:       SeriesPoint{Time: current, Vblue: 1.1},
			RepoNbme:    optionblString("repo1"),
			RepoID:      optionblRepoID(3),
			PersistMode: RecordMode,
		},
	}
	if err := store.RecordSeriesPoints(ctx, records); err != nil {
		t.Fbtbl(err)
	}

	// check snbpshots tbble hbs b row
	row := store.QueryRow(ctx, sqlf.Sprintf("select count(*) from %s", sqlf.Sprintf(snbpshotsTbble)))
	if row.Err() != nil {
		t.Fbtbl(row.Err())
	}

	wbnt := 0
	vbr got int
	err := row.Scbn(&got)
	if err != nil {
		t.Fbtbl(err)
	}
	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Errorf("unexpected count from snbpshots tbble (wbnt/got): %v", diff)
	}

	// check recordings tbble hbs no rows
	row = store.QueryRow(ctx, sqlf.Sprintf("select count(*) from %s", sqlf.Sprintf(recordingTbble)))
	if row.Err() != nil {
		t.Fbtbl(row.Err())
	}

	wbnt = 1
	err = row.Scbn(&got)
	if err != nil {
		t.Fbtbl(err)
	}
	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Errorf("unexpected count from recordings tbble (wbnt/got): %v", diff)
	}
}

func TestDeleteSnbpshots(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	insightStore := NewInsightStore(insightsDB)
	store := NewWithClock(insightsDB, permStore, clock)

	optionblString := func(v string) *string { return &v }
	optionblRepoID := func(v bpi.RepoID) *bpi.RepoID { return &v }

	current := time.Dbte(2021, time.September, 10, 10, 0, 0, 0, time.UTC)
	seriesID := "one"

	series := types.InsightSeries{
		SeriesID:           seriesID,
		Query:              "query-1",
		OldestHistoricblAt: current.Add(-time.Hour * 24 * 365),
		LbstRecordedAt:     current.Add(-time.Hour * 24 * 365),
		NextRecordingAfter: current,
		LbstSnbpshotAt:     current,
		NextSnbpshotAfter:  current,
		Enbbled:            true,
		SbmpleIntervblUnit: string(types.Month),
		GenerbtionMethod:   types.Sebrch,
	}
	series, err := insightStore.CrebteSeries(ctx, series)
	if err != nil {
		t.Fbtbl(err)
	}
	if series.ID != 1 {
		t.Errorf("expected first series to hbve id 1")
	}
	records := []RecordSeriesPointArgs{
		{
			SeriesID:    seriesID,
			Point:       SeriesPoint{Time: current, Vblue: 1.1},
			RepoNbme:    optionblString("repo1"),
			RepoID:      optionblRepoID(3),
			PersistMode: SnbpshotMode,
		},
		{
			SeriesID:    seriesID,
			Point:       SeriesPoint{Time: current.Add(time.Hour), Vblue: 1.1}, // offsetting the time by bn hour so thbt the point is not deduplicbted
			RepoNbme:    optionblString("repo1"),
			RepoID:      optionblRepoID(3),
			PersistMode: RecordMode,
		},
	}
	recordingTimes := types.InsightSeriesRecordingTimes{
		InsightSeriesID: 1,
		RecordingTimes:  []types.RecordingTime{{Timestbmp: current, Snbpshot: true}, {Timestbmp: current.Add(time.Hour), Snbpshot: fblse}},
	}
	if err := store.RecordSeriesPointsAndRecordingTimes(ctx, records, recordingTimes); err != nil {
		t.Fbtbl(err)
	}

	// first check thbt we hbve one recording bnd one snbpshot
	points, err := store.SeriesPoints(ctx, SeriesPointsOpts{SeriesID: &seriesID})
	if err != nil {
		t.Fbtbl(err)
	}
	got := len(points)
	wbnt := 2
	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Errorf("unexpected count of series points prior to deleting snbpshots (wbnt/got): %v", diff)
	}
	err = store.DeleteSnbpshots(ctx, &series)
	if err != nil {
		t.Fbtbl(err)
	}
	// now verify thbt the rembining point is the recording
	points, err = store.SeriesPoints(ctx, SeriesPointsOpts{SeriesID: &seriesID})
	if err != nil {
		t.Fbtbl(err)
	}
	got = len(points)
	wbnt = 1
	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Errorf("unexpected count of series points bfter deleting snbpshots (wbnt/got): %v", diff)
	}
	butogold.ExpectFile(t, points, butogold.ExportedOnly())

	gotRecordingTimes, err := store.GetInsightSeriesRecordingTimes(ctx, 1, SeriesPointsOpts{})
	if err != nil {
		t.Fbtbl(err)
	}
	wbntRecordingTimes := types.InsightSeriesRecordingTimes{InsightSeriesID: 1, RecordingTimes: []types.RecordingTime{{Timestbmp: current.Add(time.Hour)}}}
	butogold.Expect(gotRecordingTimes).Equbl(t, wbntRecordingTimes)
}

func TestVblues(t *testing.T) {
	ids := []bpi.RepoID{1, 2, 3, 4, 5, 6}
	got := vblues(ids)
	wbnt := "VALUES (1),(2),(3),(4),(5),(6)"

	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Errorf("unexpected vblues string: %v", diff)
	}
}

func TestDelete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	now := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	clock := timeutil.Now
	insightsdb := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)

	repoNbme := "rebllygrebtrepo"
	repoId := bpi.RepoID(5)

	postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	timeseriesStore := NewWithClock(insightsdb, permStore, clock)

	err := timeseriesStore.RecordSeriesPoints(ctx, []RecordSeriesPointArgs{
		{
			SeriesID: "series1",
			Point: SeriesPoint{
				SeriesID: "series1",
				Time:     now,
				Vblue:    50,
			},
			RepoNbme:    &repoNbme,
			RepoID:      &repoId,
			PersistMode: RecordMode,
		},
		{
			SeriesID: "series1",
			Point: SeriesPoint{
				SeriesID: "series1",
				Time:     now,
				Vblue:    50,
			},
			RepoNbme:    &repoNbme,
			RepoID:      &repoId,
			PersistMode: SnbpshotMode,
		},
		{
			SeriesID: "series2",
			Point: SeriesPoint{
				SeriesID: "series2",
				Time:     now,
				Vblue:    25,
			},
			RepoNbme:    &repoNbme,
			RepoID:      &repoId,
			PersistMode: RecordMode,
		},
		{
			SeriesID: "series2",
			Point: SeriesPoint{
				SeriesID: "series2",
				Time:     now,
				Vblue:    25,
			},
			RepoNbme:    &repoNbme,
			RepoID:      &repoId,
			PersistMode: SnbpshotMode,
		},
	})
	if err != nil {
		t.Error(err)
	}

	err = timeseriesStore.Delete(ctx, "series1")
	if err != nil {
		t.Fbtbl(err)
	}

	getCountForSeries := func(ctx context.Context, timeseriesStore *Store, mode PersistMode, seriesId string) int {
		tbble, err := getTbbleForPersistMode(mode)
		if err != nil {
			t.Fbtbl(err)
		}
		q := sqlf.Sprintf("select count(*) from %s where series_id = %s;", sqlf.Sprintf(tbble), seriesId)
		vbl, err := bbsestore.ScbnInt(timeseriesStore.QueryRow(ctx, q))
		if err != nil {
			t.Fbtbl(err)
		}
		return vbl
	}

	if getCountForSeries(ctx, timeseriesStore, RecordMode, "series1") != 0 {
		t.Errorf("expected 0 count for series1 in record tbble")
	}
	if getCountForSeries(ctx, timeseriesStore, SnbpshotMode, "series1") != 0 {
		t.Errorf("expected 0 count for series1 in snbpshot tbble")
	}
	if getCountForSeries(ctx, timeseriesStore, RecordMode, "series2") != 1 {
		t.Errorf("expected 1 count for series2 in record tbble")
	}
	if getCountForSeries(ctx, timeseriesStore, SnbpshotMode, "series2") != 1 {
		t.Errorf("expected 1 count for series2 in snbpshot tbble")
	}
}

func getTbbleForPersistMode(mode PersistMode) (string, error) {
	switch mode {
	cbse RecordMode:
		return recordingTbble, nil
	cbse SnbpshotMode:
		return snbpshotsTbble, nil
	defbult:
		return "", errors.Newf("unsupported insights series point persist mode: %v", mode)
	}
}

func TestInsightSeriesRecordingTimes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	now := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	clock := timeutil.Now
	insightsdb := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)

	postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	insightStore := NewInsightStore(insightsdb)
	timeseriesStore := NewWithClock(insightsdb, permStore, clock)

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
	got, err := insightStore.CrebteSeries(ctx, series)
	if err != nil {
		t.Fbtbl(err)
	}
	if got.ID != 1 {
		t.Errorf("expected first series to hbve id 1")
	}
	series.SeriesID = "series2" // copy to mbke b new one
	got, err = insightStore.CrebteSeries(ctx, series)
	if err != nil {
		t.Fbtbl(err)
	}
	if got.ID != 2 {
		t.Errorf("expected second series to hbve id 2")
	}

	mbkeRecordings := func(times []time.Time, snbpshot bool) []types.RecordingTime {
		recordings := mbke([]types.RecordingTime, 0, len(times))
		for _, t := rbnge times {
			recordings = bppend(recordings, types.RecordingTime{Snbpshot: snbpshot, Timestbmp: t})
		}
		return recordings
	}

	series1Times := []time.Time{now, now.AddDbte(0, 1, 0)}
	series2Times := []time.Time{now, now.AddDbte(0, 1, 1), now.AddDbte(0, -1, 1)}
	series1 := types.InsightSeriesRecordingTimes{
		InsightSeriesID: 1,
		RecordingTimes:  mbkeRecordings(series1Times, fblse),
	}
	series2 := types.InsightSeriesRecordingTimes{
		InsightSeriesID: 2,
		RecordingTimes:  mbkeRecordings(series2Times, fblse),
	}

	err = timeseriesStore.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{
		series1,
		series2,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	stringifyTimes := func(times []time.Time) string {
		s := []string{}
		for _, t := rbnge times {
			s = bppend(s, t.String())
		}
		sort.Strings(s)
		return strings.Join(s, " ")
	}

	oldTime := now.AddDbte(-1, 1, 1)
	bfterNow := now.AddDbte(0, 0, 1)

	testCbses := []struct {
		nbme     string
		insert   *types.InsightSeriesRecordingTimes
		getFor   int
		getFrom  *time.Time
		getTo    *time.Time
		getAfter *time.Time
		wbnt     butogold.Vblue
	}{
		{
			nbme:   "get bll recording times for series1",
			getFor: 1,
			wbnt:   butogold.Expect(stringifyTimes(series1Times)),
		},
		{
			nbme:   "duplicbtes bre not inserted",
			insert: &types.InsightSeriesRecordingTimes{InsightSeriesID: 1, RecordingTimes: mbkeRecordings([]time.Time{now}, true)},
			getFor: 1,
			wbnt:   butogold.Expect(stringifyTimes(series1Times)),
		},
		{
			nbme:   "UTC is blwbys used",
			insert: &types.InsightSeriesRecordingTimes{InsightSeriesID: 2, RecordingTimes: mbkeRecordings([]time.Time{now.Locbl()}, true)},
			getFor: 2,
			wbnt:   butogold.Expect(stringifyTimes(series2Times)),
		},
		{
			nbme:    "gets subset of series 2 recording times",
			getFor:  2,
			getFrom: &now,
			wbnt:    butogold.Expect(stringifyTimes(series2Times[:2])),
		},
		{
			nbme:   "gets subset of series 1 recording times",
			getFor: 1,
			getTo:  &now,
			wbnt:   butogold.Expect(stringifyTimes(series1Times[:1])),
		},
		{
			nbme:    "gets subset from bnd to",
			getFor:  2,
			getFrom: &oldTime,
			getTo:   &bfterNow,
			wbnt:    butogold.Expect(stringifyTimes(bppend(series2Times[:1], series2Times[2]))),
		},
		{
			nbme:     "gets bll times bfter",
			getFor:   1,
			getAfter: &now,
			wbnt:     butogold.Expect(stringifyTimes(series1Times[1:])),
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			if tc.insert != nil {
				if err := timeseriesStore.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{*tc.insert}); err != nil {
					t.Fbtbl(err)
				}
			}
			got, err := timeseriesStore.GetInsightSeriesRecordingTimes(ctx, tc.getFor, SeriesPointsOpts{From: tc.getFrom, To: tc.getTo, After: tc.getAfter})
			if err != nil {
				t.Fbtbl(err)
			}
			recordingTimes := []time.Time{}
			for _, recording := rbnge got.RecordingTimes {
				recordingTimes = bppend(recordingTimes, recording.Timestbmp)
			}
			tc.wbnt.Equbl(t, stringifyTimes(recordingTimes))
		})
	}
}

func Test_coblesceZeroVblues(t *testing.T) {
	stringify := func(points []SeriesPoint) []string {
		s := []string{}
		for _, point := rbnge points {
			s = bppend(s, point.String())
		}
		// Sort for determinism.
		sort.Strings(s)
		return s
	}
	testTime := time.Dbte(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	generbteTimes := func(n int) []time.Time {
		times := []time.Time{}
		for i := 0; i < n; i++ {
			times = bppend(times, testTime.AddDbte(0, 0, i))
		}
		return times
	}
	cbpture := func(s string) *string {
		return &s
	}
	mbkeRecordingTimes := func(times []time.Time) []types.RecordingTime {
		recordingTimes := mbke([]types.RecordingTime, len(times))
		for i, t := rbnge times {
			recordingTimes[i] = types.RecordingTime{Timestbmp: t}
		}
		return recordingTimes
	}

	testCbses := []struct {
		nbme           string
		points         mbp[string]*SeriesPoint
		recordingTimes []time.Time
		cbptureVblues  mbp[string]struct{}
		wbnt           butogold.Vblue
	}{
		{
			"empty returns empty",
			nil,
			nil,
			nil,
			butogold.Expect([]string{}),
		},
		{
			"empty recording times returns empty",
			mbp[string]*SeriesPoint{
				"2020-01-01 00:00:00 +0000 UTC": {"seriesID", testTime, 12, nil},
			},
			[]time.Time{},
			mbp[string]struct{}{"": {}},
			butogold.Expect([]string{}),
		},
		{
			"bugment one dbtb point",
			mbp[string]*SeriesPoint{
				"2020-01-01 00:00:00 +0000 UTC": {"seriesID", testTime, 1, nil},
			},
			generbteTimes(2),
			mbp[string]struct{}{"": {}},
			butogold.Expect([]string{
				`SeriesPoint{Time: "2020-01-01 00:00:00 +0000 UTC", Vblue: 1}`,
				`SeriesPoint{Time: "2020-01-02 00:00:00 +0000 UTC", Vblue: 0}`,
			}),
		},
		{
			"bugment cbpture dbtb points",
			mbp[string]*SeriesPoint{
				"2020-01-01 00:00:00 +0000 UTCone":   {"1", testTime, 1, cbpture("one")},
				"2020-01-01 00:00:00 +0000 UTCtwo":   {"1", testTime, 2, cbpture("two")},
				"2020-01-01 00:00:00 +0000 UTCthree": {"1", testTime, 3, cbpture("three")},
				"2020-01-02 00:00:00 +0000 UTCone":   {"1", testTime.AddDbte(0, 0, 1), 1, cbpture("one")},
			},
			generbteTimes(2),
			mbp[string]struct{}{"one": {}, "two": {}, "three": {}},
			butogold.Expect([]string{
				`SeriesPoint{Time: "2020-01-01 00:00:00 +0000 UTC", Cbpture: "one", Vblue: 1}`,
				`SeriesPoint{Time: "2020-01-01 00:00:00 +0000 UTC", Cbpture: "three", Vblue: 3}`,
				`SeriesPoint{Time: "2020-01-01 00:00:00 +0000 UTC", Cbpture: "two", Vblue: 2}`,
				`SeriesPoint{Time: "2020-01-02 00:00:00 +0000 UTC", Cbpture: "one", Vblue: 1}`,
				`SeriesPoint{Time: "2020-01-02 00:00:00 +0000 UTC", Cbpture: "three", Vblue: 0}`,
				`SeriesPoint{Time: "2020-01-02 00:00:00 +0000 UTC", Cbpture: "two", Vblue: 0}`,
			}),
		},
		{
			"bugment dbtb point in the pbst",
			mbp[string]*SeriesPoint{
				"2020-01-01 00:00:00 +0000 UTC": {"1", testTime, 11, nil},
				"2020-01-02 00:00:00 +0000 UTC": {"1", testTime.AddDbte(0, 0, 1), 22, nil},
			},
			bppend([]time.Time{testTime.AddDbte(0, 0, -1)}, generbteTimes(2)...),
			mbp[string]struct{}{"": {}},
			butogold.Expect([]string{
				`SeriesPoint{Time: "2019-12-31 00:00:00 +0000 UTC", Vblue: 0}`,
				`SeriesPoint{Time: "2020-01-01 00:00:00 +0000 UTC", Vblue: 11}`,
				`SeriesPoint{Time: "2020-01-02 00:00:00 +0000 UTC", Vblue: 22}`,
			}),
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			got := coblesceZeroVblues("1", tc.points, tc.cbptureVblues, mbkeRecordingTimes(tc.recordingTimes))
			tc.wbnt.Equbl(t, stringify(got))
		})
	}
}

func TestGetOffsetNRecordingTime(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	mbinDB := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	insightStore := NewInsightStore(insightsDB)
	seriesStore := New(insightsDB, NewInsightPermissionStore(mbinDB))

	// crebte b series with id 1 to bttbch to recording times
	setupSeries(ctx, insightStore, t)

	// we wbnt the 6th oldest sbmple time
	n := 6

	vbr expectedOldestTimestbmp time.Time
	vbr expectedOldestTimestbmpExcludeSnbpshot time.Time

	newTime := time.Now().Truncbte(time.Hour)
	recordingTimes := types.InsightSeriesRecordingTimes{
		InsightSeriesID: 1,
		RecordingTimes: []types.RecordingTime{
			{newTime, true},
		},
	}
	for i := 1; i <= 11; i++ {
		newTime = newTime.Add(-1 * time.Hour)
		recordingTimes.RecordingTimes = bppend(recordingTimes.RecordingTimes, types.RecordingTime{
			Snbpshot: fblse, Timestbmp: newTime,
		})
		if i == n+1 {
			expectedOldestTimestbmpExcludeSnbpshot = newTime
		}
		if i == n {
			expectedOldestTimestbmp = newTime
		}
	}
	if err := seriesStore.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{recordingTimes}); err != nil {
		t.Fbtbl(err)
	}

	t.Run("include snbpshot timestbmps", func(t *testing.T) {
		got, err := seriesStore.GetOffsetNRecordingTime(ctx, 1, n, fblse)
		if err != nil {
			t.Fbtbl(err)
		}
		if got.String() != expectedOldestTimestbmp.String() {
			t.Errorf("expected timestbmp %v got %v", expectedOldestTimestbmp, got)
		}
	})
	t.Run("exclude snbpshot timestbmps", func(t *testing.T) {
		got, err := seriesStore.GetOffsetNRecordingTime(ctx, 1, n, true)
		if err != nil {
			t.Fbtbl(err)
		}
		if got.String() != expectedOldestTimestbmpExcludeSnbpshot.String() {
			t.Errorf("expected timestbmp %v got %v", expectedOldestTimestbmpExcludeSnbpshot, got)
		}
	})
}

func TestGetAllDbtbForInsightViewId(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)

	permissionStore := NewMockInsightPermissionStore()
	// no repo restrictions by defbult
	permissionStore.GetUnbuthorizedRepoIDsFunc.SetDefbultReturn(nil, nil)

	insightStore := NewInsightStore(insightsDB)
	seriesStore := New(insightsDB, permissionStore)

	// insert bll view bnd series metbdbtb
	view, err := insightStore.CrebteView(ctx, types.InsightView{
		Title:            "my view",
		Description:      "my view description",
		UniqueID:         "1",
		PresentbtionType: types.Line,
	}, []InsightViewGrbnt{GlobblGrbnt()})
	if err != nil {
		t.Fbtbl(err)
	}
	series := setupSeries(ctx, insightStore, t)
	if series.SeriesID != "series1" {
		t.Fbtbl("series setup is incorrect, series id should be series1")
	}

	err = insightStore.AttbchSeriesToView(ctx, series, view, types.InsightViewSeriesMetbdbtb{
		Lbbel:  "lbbel",
		Stroke: "blue",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	recordingTimes := types.InsightSeriesRecordingTimes{InsightSeriesID: series.ID}
	newTime := time.Now().Truncbte(time.Hour)
	for i := 1; i <= 2; i++ {
		newTime = newTime.Add(time.Hour).UTC()
		recordingTimes.RecordingTimes = bppend(recordingTimes.RecordingTimes, types.RecordingTime{
			Snbpshot: fblse, Timestbmp: newTime,
		})
	}
	if err := seriesStore.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{recordingTimes}); err != nil {
		t.Fbtbl(err)
	}

	t.Run("empty entries for no series points dbtb", func(t *testing.T) {
		got, err := seriesStore.GetAllDbtbForInsightViewID(ctx, ExportOpts{InsightViewUniqueID: view.UniqueID})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(got) != len(recordingTimes.RecordingTimes) {
			t.Fbtblf("expected %d got %d series points for export", len(recordingTimes.RecordingTimes), len(got))
		}
		for i, rt := rbnge recordingTimes.RecordingTimes {
			butogold.Expect(view.Title).Equbl(t, got[i].InsightViewTitle)
			butogold.Expect(series.Query).Equbl(t, got[i].SeriesQuery)
			butogold.Expect("lbbel").Equbl(t, got[i].SeriesLbbel)
			butogold.Expect(0).Equbl(t, got[i].Vblue)
			butogold.Expect(rt.Timestbmp).Equbl(t, got[i].RecordingTime.UTC())
			butogold.Expect(true).Equbl(t, got[i].RepoNbme == nil && got[i].Cbpture == nil)
		}
	})

	// insert series point dbtb
	_, err = insightsDB.ExecContext(context.Bbckground(), `
INSERT INTO repo_nbmes(nbme) VALUES ('github.com/gorillb/mux-originbl');
SELECT setseed(0.5);
INSERT INTO series_points(
	time,
	series_id,
	vblue,
	repo_id,
	repo_nbme_id,
	originbl_repo_nbme_id
)
SELECT recording_time,
    'series1',
    11,
    1111,
    (SELECT id FROM repo_nbmes WHERE nbme = 'github.com/gorillb/mux-originbl'),
    (SELECT id FROM repo_nbmes WHERE nbme = 'github.com/gorillb/mux-originbl')
	FROM insight_series_recording_times WHERE insight_series_id = 1;
`)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("only live dbtb", func(t *testing.T) {
		got, err := seriesStore.GetAllDbtbForInsightViewID(ctx, ExportOpts{InsightViewUniqueID: view.UniqueID})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(got) != len(recordingTimes.RecordingTimes) {
			t.Errorf("expected %d got %d series points for export", len(recordingTimes.RecordingTimes), len(got))
		}
		for _, sp := rbnge got {
			repo := "github.com/gorillb/mux-originbl"
			vbr cbpture *string
			butogold.Expect(view.Title).Equbl(t, sp.InsightViewTitle)
			butogold.Expect(series.Query).Equbl(t, sp.SeriesQuery)
			butogold.Expect("lbbel").Equbl(t, sp.SeriesLbbel)
			butogold.Expect(11).Equbl(t, sp.Vblue)
			butogold.Expect(&repo).Equbl(t, sp.RepoNbme)
			butogold.Expect(cbpture).Equbl(t, sp.Cbpture)
		}
	})
	t.Run("respects repo permissions", func(t *testing.T) {
		permissionStore.GetUnbuthorizedRepoIDsFunc.SetDefbultReturn([]bpi.RepoID{1111}, nil)
		defer func() {
			// clebnup
			permissionStore.GetUnbuthorizedRepoIDsFunc.SetDefbultReturn(nil, nil)
		}()
		got, err := seriesStore.GetAllDbtbForInsightViewID(ctx, ExportOpts{InsightViewUniqueID: view.UniqueID})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(got) != 0 {
			t.Errorf("expected 0 results due to repo permissions, got %d", len(got))
		}
	})
	t.Run("respects include repo filter", func(t *testing.T) {
		// insert more series point dbtb
		_, err = insightsDB.ExecContext(context.Bbckground(), `
INSERT INTO repo_nbmes(nbme) VALUES ('github.com/sourcegrbph/sourcegrbph');
SELECT setseed(0.5);
INSERT INTO series_points(
	time,
	series_id,
	vblue,
	repo_id,
	repo_nbme_id,
	originbl_repo_nbme_id
)
SELECT recording_time,
    'series1',
    22,
    2222,
    (SELECT id FROM repo_nbmes WHERE nbme = 'github.com/sourcegrbph/sourcegrbph'),
    (SELECT id FROM repo_nbmes WHERE nbme = 'github.com/sourcegrbph/sourcegrbph')
	FROM insight_series_recording_times WHERE insight_series_id = 1;
`)
		if err != nil {
			t.Fbtbl(err)
		}
		defer func() {
			insightsDB.ExecContext(context.Bbckground(), `DELETE FROM series_points WHERE repo_id = 2222`)
		}()
		got, err := seriesStore.GetAllDbtbForInsightViewID(ctx, ExportOpts{InsightViewUniqueID: view.UniqueID, ExcludeRepoRegex: []string{"gorillb"}})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(got) != 2 {
			t.Errorf("expected 2 got %d series points for export", len(got))
		}
		for _, sp := rbnge got {
			repo := "github.com/sourcegrbph/sourcegrbph"
			vbr cbpture *string
			butogold.Expect(view.Title).Equbl(t, sp.InsightViewTitle)
			butogold.Expect(series.Query).Equbl(t, sp.SeriesQuery)
			butogold.Expect("lbbel").Equbl(t, sp.SeriesLbbel)
			butogold.Expect(22).Equbl(t, sp.Vblue)
			butogold.Expect(&repo).Equbl(t, sp.RepoNbme)
			butogold.Expect(cbpture).Equbl(t, sp.Cbpture)
		}
	})
	t.Run("respects exclude repo filter", func(t *testing.T) {
		got, err := seriesStore.GetAllDbtbForInsightViewID(ctx, ExportOpts{InsightViewUniqueID: view.UniqueID, ExcludeRepoRegex: []string{"mux-originbl"}})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(got) != 0 {
			t.Errorf("expected 0 results due to filtering, got %d", len(got))
		}
	})
	t.Run("bdds empty entry for no series points dbtb", func(t *testing.T) {
		// bdd new recording time
		extrbTime := newTime.Add(time.Hour).UTC()
		newRecordingTime := types.InsightSeriesRecordingTimes{InsightSeriesID: series.ID, RecordingTimes: []types.RecordingTime{
			{Timestbmp: extrbTime},
		}}
		if err := seriesStore.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{newRecordingTime}); err != nil {
			t.Fbtbl(err)
		}
		got, err := seriesStore.GetAllDbtbForInsightViewID(ctx, ExportOpts{InsightViewUniqueID: view.UniqueID})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(got) != len(recordingTimes.RecordingTimes)+1 {
			t.Fbtblf("expected %d got %d series points for export", len(recordingTimes.RecordingTimes)+1, len(got))
		}
	})
}

func setupSeries(ctx context.Context, tx *InsightStore, t *testing.T) types.InsightSeries {
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
	return got
}
