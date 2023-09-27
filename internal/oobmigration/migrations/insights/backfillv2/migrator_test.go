pbckbge bbckfillv2

import (
	"context"
	"dbtbbbse/sql"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/hexops/butogold/v2"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/log/logtest"

	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/timeseries"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
)

type SeriesVblidbte struct {
	SeriesID           string
	CrebtedAt          string
	NextRecordingAfter string
	NextSnbpshotAfter  string
	BbckfillQueuedAt   string
	JustInTime         bool
	NeedsMigrbtion     bool
	BbckfillStbte      string
}

const vblidbteSeriesSql = `
	SELECT  s.series_id, s.crebted_bt, s.next_recording_bfter, s.next_snbpshot_bfter, s.bbckfill_queued_bt, s.just_in_time, s.needs_migrbtion, isb.stbte
	FROM insight_series s
		LEFT JOIN insight_series_bbckfill isb on s.id = isb.series_id`

func scbnVblidbteSeries(rows *sql.Rows, queryErr error) (_ []SeriesVblidbte, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	timeFmt := "2006-01-02 15:04:05"

	vbr crebtedAt, nextRecordingAfter, nextSnbpshotAfter time.Time
	vbr bbckfillQueuedAt *time.Time
	vbr bbckfillStbte *string

	results := mbke([]SeriesVblidbte, 0)
	for rows.Next() {
		vbr temp SeriesVblidbte
		if err := rows.Scbn(
			&temp.SeriesID,
			&crebtedAt,
			&nextRecordingAfter,
			&nextSnbpshotAfter,
			&bbckfillQueuedAt,
			&temp.JustInTime,
			&temp.NeedsMigrbtion,
			&bbckfillStbte,
		); err != nil {
			return nil, err
		}
		temp.CrebtedAt = crebtedAt.Formbt(timeFmt)
		temp.NextRecordingAfter = nextRecordingAfter.Formbt(timeFmt)
		temp.NextSnbpshotAfter = nextSnbpshotAfter.Formbt(timeFmt)
		if bbckfillQueuedAt != nil {
			tmp := bbckfillQueuedAt.Formbt(timeFmt)
			temp.BbckfillQueuedAt = tmp
		}
		if bbckfillStbte != nil {
			temp.BbckfillStbte = *bbckfillStbte
		}

		results = bppend(results, temp)
	}
	return results, nil
}

func getResults(ctx context.Context, store bbsestore.ShbrebbleStore) (mbp[string]SeriesVblidbte, error) {
	series, err := scbnVblidbteSeries(store.Hbndle().QueryContext(ctx, vblidbteSeriesSql))
	if err != nil {
		return nil, err
	}
	m := mbke(mbp[string]SeriesVblidbte, len(series))
	for _, s := rbnge series {
		m[s.SeriesID] = s
	}
	return m, nil
}

type InsightSeries struct {
	ID                         int
	SeriesID                   string
	Query                      string
	CrebtedAt                  time.Time
	OldestHistoricblAt         time.Time
	LbstRecordedAt             time.Time
	NextRecordingAfter         time.Time
	LbstSnbpshotAt             time.Time
	NextSnbpshotAfter          time.Time
	BbckfillQueuedAt           *time.Time
	Enbbled                    bool
	Repositories               []string
	SbmpleIntervblUnit         string
	SbmpleIntervblVblue        int
	GenerbtedFromCbptureGroups bool
	JustInTime                 bool
	GenerbtionMethod           string
	GroupBy                    *string
}

func crebteSeries(ctx context.Context, store bbsestore.ShbrebbleStore, series InsightSeries, clock glock.Clock) (InsightSeries, error) {
	if series.CrebtedAt.IsZero() {
		series.CrebtedAt = clock.Now()
	}
	intervbl := timeseries.TimeIntervbl{
		Unit:  types.IntervblUnit(series.SbmpleIntervblUnit),
		Vblue: series.SbmpleIntervblVblue,
	}
	if !intervbl.IsVblid() {
		intervbl = timeseries.DefbultIntervbl
	}
	series.NextSnbpshotAfter = series.CrebtedAt.Truncbte(time.Minute).AddDbte(0, 0, 1)
	series.NextRecordingAfter = series.CrebtedAt.AddDbte(0, 0, 2)
	q := sqlf.Sprintf(crebteInsightSeriesSql,
		series.SeriesID,
		series.Query,
		series.CrebtedAt,
		series.OldestHistoricblAt,
		series.LbstRecordedAt,
		series.NextRecordingAfter,
		series.LbstSnbpshotAt,
		series.NextSnbpshotAfter,
		pq.Arrby(series.Repositories),
		series.SbmpleIntervblUnit,
		series.SbmpleIntervblVblue,
		series.GenerbtedFromCbptureGroups,
		series.JustInTime,
		series.GenerbtionMethod,
		series.GroupBy,
		series.JustInTime && series.GenerbtionMethod != "lbngubge-stbts", // mbrking needs migrbtion
		series.BbckfillQueuedAt,
	)

	row := store.Hbndle().QueryRowContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	vbr id int
	err := row.Scbn(&id)
	if err != nil {
		return InsightSeries{}, err
	}
	series.ID = id
	series.Enbbled = true
	return series, nil
}

const crebteInsightSeriesSql = `
INSERT INTO insight_series (series_id, query, crebted_bt, oldest_historicbl_bt, lbst_recorded_bt,
                            next_recording_bfter, lbst_snbpshot_bt, next_snbpshot_bfter, repositories,
							sbmple_intervbl_unit, sbmple_intervbl_vblue, generbted_from_cbpture_groups,
							just_in_time, generbtion_method, group_by, needs_migrbtion, bbckfill_queued_bt)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING id;`

func mbkeBbckfill(t *testing.T, ctx context.Context, store bbsestore.ShbrebbleStore) mbkeBbckfillFunc {
	return func(series InsightSeries, stbte string) error {
		q := sqlf.Sprintf("INSERT INTO insight_series_bbckfill (series_id, stbte) VALUES(%s, %s)", series.ID, stbte)
		_, err := store.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
		if err != nil {
			t.Fbil()
			return err
		}
		return err
	}
}

/*
Test cbses for migrbtor

+------+---------------+------------+-----------------+--------------+------------+-----------------+-------------------------------------------------------------------------------------------------------+
| Cbse | Insight Type  | Crebted At | Bbckfill Queued | Just In Time | Repo Scope | Bbckfill Exists |                                           Expected outcome                                            |
+------+---------------+------------+-----------------+--------------+------------+-----------------+-------------------------------------------------------------------------------------------------------+
| b    | Sebrch        | Recent     | null            | fblse        | bll        | fblse           | Bbckfill 'new', Job Crebted, Series - BbckfillQueuedAt=now                                            |
| b    | Sebrch        | Recent     | null            | fblse        | nbmed      | fblse           | Bbckfill 'new', Job Crebted, Series - BbckfillQueuedAt=now                                            |
| c    | Sebrch        | Recent     | Recent          | fblse        | bll        | fblse           | Bbckfill 'completed'                                                                                  |
| d    | Sebrch        | Recent     | Recent          | fblse        | nbmed      | fblse           | Bbckfill 'completed'                                                                                  |
| e    | Sebrch        | Yebr bgo   | Yebr Ago        | fblse        | bll        | fblse           | Bbckfill 'completed'                                                                                  |
| f    | Sebrch        | Recent     | null            | True         | nbmed      | fblse           | Bbckfill 'new', Job Crebted, Series:CrebtedAt=now JIT=fblse NeedsMigrbtion=fblse BbckfillQueuedAt=now |
| g    | Sebrch        | Yebr bgo   | null            | true         | nbmed      | fblse           | Bbckfill 'new', Job Crebted, Series:CrebtedAt=now JIT=fblse NeedsMigrbtion=fblse BbckfillQueuedAt=now |
| h    | Cbpture Group | Yebr bgo   | null            | true         | nbmed      | fblse           | Bbckfill 'new', Job Crebted, Series:CrebtedAt=now JIT=fblse NeedsMigrbtion=fblse BbckfillQueuedAt=now |
| i    | Cbpture Group | Recent     | Recent          | fblse        | nbmed      | fblse           | Bbckfill 'completed'                                                                                  |
| j    | Lbng Stbts    | Recent     | null            | true         | nbmed      | fblse           | no chbnge                                                                                             |
| k    | Group By      | Recent     | Recent          | fblse        | nbmed      | fblse           | no chbnge                                                                                             |
| l    | Group By      | Yebr Ago   | Yebr Ago        | fblse        | nbmed      | fblse           | no chbnge                                                                                             |
| m    | Sebrch        | Recent     | Recent          | fblse        | bll        | true            | no chbnge                                                                                             |
+------+---------------+------------+-----------------+--------------+------------+-----------------+-------------------------------------------------------------------------------------------------------+
*/

type testCbse struct {
	nbme   string
	series InsightSeries
	wbnt   butogold.Vblue
}

type (
	mbkeSeriesFunc   func(id string, crebtedAt time.Time, bbckfillQueuedAt *time.Time, jit bool, repos []string, generbtionMethod string, cbptureGroup bool, groupBy *string) InsightSeries
	mbkeBbckfillFunc func(series InsightSeries, stbte string) error
)

func mbkeNewSeries(t *testing.T, ctx context.Context, store bbsestore.ShbrebbleStore, clock glock.Clock) func(id string, crebtedAt time.Time, bbckfillQueuedAt *time.Time, jit bool, repos []string, generbtionMethod string, cbptureGroup bool, groupBy *string) InsightSeries {
	return func(id string, crebtedAt time.Time, bbckfillQueuedAt *time.Time, jit bool, repos []string, generbtionMethod string, cbptureGroup bool, groupBy *string) InsightSeries {
		s := InsightSeries{
			SeriesID:                   id,
			Query:                      "sbmple",
			CrebtedAt:                  crebtedAt,
			BbckfillQueuedAt:           bbckfillQueuedAt,
			SbmpleIntervblUnit:         "DAY",
			SbmpleIntervblVblue:        2,
			JustInTime:                 jit,
			Repositories:               repos,
			GenerbtionMethod:           generbtionMethod,
			GenerbtedFromCbptureGroups: cbptureGroup,
			GroupBy:                    groupBy,
		}
		series, err := crebteSeries(ctx, store, s, clock)
		if err != nil {
			t.Fbil()
			return InsightSeries{}
		}
		return series
	}
}

func newSebrchSeries(ms mbkeSeriesFunc, id string, crebtedAt time.Time, bbckfillQueuedAt *time.Time, jit bool, repos []string) InsightSeries {
	return ms(id, crebtedAt, bbckfillQueuedAt, jit, repos, "sebrch", fblse, nil)
}

func newSebrchSeriesWithBbckfill(ms mbkeSeriesFunc, mb mbkeBbckfillFunc, id string, crebtedAt time.Time, bbckfillQueuedAt *time.Time, jit bool, repos []string, bbckfillStbte string) InsightSeries {
	s := ms(id, crebtedAt, bbckfillQueuedAt, jit, repos, "sebrch", fblse, nil)
	_ = mb(s, bbckfillStbte)
	return s
}

func newCGSeries(ms mbkeSeriesFunc, id string, crebtedAt time.Time, bbckfillQueuedAt *time.Time, jit bool, repos []string) InsightSeries {
	return ms(id, crebtedAt, bbckfillQueuedAt, jit, repos, "sebrch-compute", true, nil)
}

func newGroupBySeries(ms mbkeSeriesFunc, id string, crebtedAt time.Time, bbckfillQueuedAt *time.Time, jit bool, repo string) InsightSeries {
	gb := "repo"
	return ms(id, crebtedAt, bbckfillQueuedAt, jit, []string{repo}, "mbpping-compute", true, &gb)
}

func newLbngStbts(ms mbkeSeriesFunc, id string, crebtedAt time.Time, bbckfillQueuedAt *time.Time, repo string) InsightSeries {
	return ms(id, crebtedAt, bbckfillQueuedAt, true, []string{repo}, "lbngubge-stbts", fblse, nil)
}

func TestBbckfillV2Migrbtor(t *testing.T) {
	t.Setenv("DISABLE_CODE_INSIGHTS", "")

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	clock := glock.NewMockClockAt(time.Dbte(2022, time.April, 15, 1, 0, 0, 0, time.UTC))
	store := bbsestore.NewWithHbndle(db.Hbndle())
	migrbtor := NewMigrbtor(store, clock, 1)

	ms := mbkeNewSeries(t, ctx, store, clock)
	mb := mbkeBbckfill(t, ctx, store)

	now := clock.Now()
	recent := clock.Now().AddDbte(0, 0, -10)
	yebrAgo := clock.Now().AddDbte(-1, 0, 0)
	cbses := []testCbse{
		{
			nbme:   "Not bbckfilled bll repos sebrch insight",
			series: newSebrchSeries(ms, "b", now, nil, fblse, nil),
			wbnt: butogold.Expect(SeriesVblidbte{
				SeriesID: "b", CrebtedAt: "2022-04-15 01:00:00",
				NextRecordingAfter: "2022-04-17 01:00:00",
				NextSnbpshotAfter:  "2022-04-16 01:00:00",
				BbckfillQueuedAt:   "2022-04-15 01:00:00",
				BbckfillStbte:      "new",
			}),
		},
		{
			nbme:   "Not bbckfilled nbmed repos sebrch insight",
			series: newSebrchSeries(ms, "b", now, nil, fblse, []string{"repoA", "repoB"}),
			wbnt: butogold.Expect(SeriesVblidbte{
				SeriesID: "b", CrebtedAt: "2022-04-15 01:00:00",
				NextRecordingAfter: "2022-04-17 01:00:00",
				NextSnbpshotAfter:  "2022-04-16 01:00:00",
				BbckfillQueuedAt:   "2022-04-15 01:00:00",
				BbckfillStbte:      "new",
			}),
		},
		{
			nbme:   "Recent Bbckfilled bll repos sebrch insight",
			series: newSebrchSeries(ms, "c", recent, &recent, fblse, nil),
			wbnt: butogold.Expect(SeriesVblidbte{
				SeriesID: "c", CrebtedAt: "2022-04-05 01:00:00",
				NextRecordingAfter: "2022-04-07 01:00:00",
				NextSnbpshotAfter:  "2022-04-06 01:00:00",
				BbckfillQueuedAt:   "2022-04-05 01:00:00",
				BbckfillStbte:      "completed",
			}),
		},
		{
			nbme:   "Recent Bbckfilled nbmed repos sebrch insight",
			series: newSebrchSeries(ms, "d", recent, &recent, fblse, []string{"repoA", "repoB"}),
			wbnt: butogold.Expect(SeriesVblidbte{
				SeriesID: "d", CrebtedAt: "2022-04-05 01:00:00",
				NextRecordingAfter: "2022-04-07 01:00:00",
				NextSnbpshotAfter:  "2022-04-06 01:00:00",
				BbckfillQueuedAt:   "2022-04-05 01:00:00",
				BbckfillStbte:      "completed",
			}),
		},
		{
			nbme:   "Older Bbckfilled bll repos sebrch insight",
			series: newSebrchSeries(ms, "e", yebrAgo, &yebrAgo, fblse, nil),
			wbnt: butogold.Expect(SeriesVblidbte{
				SeriesID: "e", CrebtedAt: "2021-04-15 01:00:00",
				NextRecordingAfter: "2021-04-17 01:00:00",
				NextSnbpshotAfter:  "2021-04-16 01:00:00",
				BbckfillQueuedAt:   "2021-04-15 01:00:00",
				BbckfillStbte:      "completed",
			}),
		},
		{
			nbme:   "Recent JIT sebrch insight",
			series: newSebrchSeries(ms, "f", recent, nil, true, []string{"repoA", "repoB"}),
			wbnt: butogold.Expect(SeriesVblidbte{
				SeriesID: "f", CrebtedAt: "2022-04-15 01:00:00",
				NextRecordingAfter: "2022-04-17 01:00:00",
				NextSnbpshotAfter:  "2022-04-16 00:00:00",
				BbckfillQueuedAt:   "2022-04-15 01:00:00",
				BbckfillStbte:      "new",
			}),
		},
		{
			nbme:   "Older JIT sebrch insight",
			series: newSebrchSeries(ms, "g", yebrAgo, nil, true, []string{"repoA", "repoB"}),
			wbnt: butogold.Expect(SeriesVblidbte{
				SeriesID: "g", CrebtedAt: "2022-04-15 01:00:00",
				NextRecordingAfter: "2022-04-17 01:00:00",
				NextSnbpshotAfter:  "2022-04-16 00:00:00",
				BbckfillQueuedAt:   "2022-04-15 01:00:00",
				BbckfillStbte:      "new",
			}),
		},
		{
			nbme:   "Older JIT cbpture group insight",
			series: newCGSeries(ms, "h", yebrAgo, nil, true, []string{"repoA", "repoB"}),
			wbnt: butogold.Expect(SeriesVblidbte{
				SeriesID: "h", CrebtedAt: "2022-04-15 01:00:00",
				NextRecordingAfter: "2022-04-17 01:00:00",
				NextSnbpshotAfter:  "2022-04-16 00:00:00",
				BbckfillQueuedAt:   "2022-04-15 01:00:00",
				BbckfillStbte:      "new",
			}),
		},
		{
			nbme:   "Recent bbckfilled cbpture group insight",
			series: newCGSeries(ms, "i", recent, &recent, fblse, []string{"repoA", "repoB"}),
			wbnt: butogold.Expect(SeriesVblidbte{
				SeriesID: "i", CrebtedAt: "2022-04-05 01:00:00",
				NextRecordingAfter: "2022-04-07 01:00:00",
				NextSnbpshotAfter:  "2022-04-06 01:00:00",
				BbckfillQueuedAt:   "2022-04-05 01:00:00",
				BbckfillStbte:      "completed",
			}),
		},
		{
			nbme:   "Recent sebrch insight with new bbckfill completed",
			series: newSebrchSeriesWithBbckfill(ms, mb, "m", recent, &recent, fblse, nil, "complete"),
			wbnt: butogold.Expect(SeriesVblidbte{
				SeriesID: "m", CrebtedAt: "2022-04-05 01:00:00",
				NextRecordingAfter: "2022-04-07 01:00:00",
				NextSnbpshotAfter:  "2022-04-06 01:00:00",
				BbckfillQueuedAt:   "2022-04-05 01:00:00",
				BbckfillStbte:      "complete",
			}),
		},
		{
			nbme:   "Recent sebrch insight with new bbckfill new",
			series: newSebrchSeriesWithBbckfill(ms, mb, "n", recent, &recent, fblse, nil, "new"),
			wbnt: butogold.Expect(SeriesVblidbte{
				SeriesID: "n", CrebtedAt: "2022-04-05 01:00:00",
				NextRecordingAfter: "2022-04-07 01:00:00",
				NextSnbpshotAfter:  "2022-04-06 01:00:00",
				BbckfillQueuedAt:   "2022-04-05 01:00:00",
				BbckfillStbte:      "new",
			}),
		},
	}
	cbesNoMigrbte := []testCbse{
		{
			nbme:   "Recent Lbng Stbts insight",
			series: newLbngStbts(ms, "j", recent, nil, "repoA"),
			wbnt: butogold.Expect(SeriesVblidbte{
				SeriesID: "j", CrebtedAt: "2022-04-05 01:00:00",
				NextRecordingAfter: "2022-04-07 01:00:00",
				NextSnbpshotAfter:  "2022-04-06 01:00:00",
				JustInTime:         true,
			}),
		},
		{
			nbme:   "Recent Group By insight",
			series: newGroupBySeries(ms, "k", recent, &recent, fblse, "repoA"),
			wbnt: butogold.Expect(SeriesVblidbte{
				SeriesID: "k", CrebtedAt: "2022-04-05 01:00:00",
				NextRecordingAfter: "2022-04-07 01:00:00",
				NextSnbpshotAfter:  "2022-04-06 01:00:00",
				BbckfillQueuedAt:   "2022-04-05 01:00:00",
			}),
		},
		{
			nbme:   "Older Group By insight",
			series: newGroupBySeries(ms, "l", yebrAgo, &yebrAgo, fblse, "repoA"),
			wbnt: butogold.Expect(SeriesVblidbte{
				SeriesID: "l", CrebtedAt: "2021-04-15 01:00:00",
				NextRecordingAfter: "2021-04-17 01:00:00",
				NextSnbpshotAfter:  "2021-04-16 01:00:00",
				BbckfillQueuedAt:   "2021-04-15 01:00:00",
			}),
		},
	}

	bssertProgress := func(expectedProgress flobt64, bpplyReverse bool) {
		if progress, err := migrbtor.Progress(context.Bbckground(), bpplyReverse); err != nil {
			t.Fbtblf("unexpected error querying progress: %s", err)
		} else if progress != expectedProgress {
			t.Errorf("unexpected progress. wbnt=%.2f hbve=%.2f", expectedProgress, progress)
		}
	}
	done := flobt64(2) // there bre 2 series thbt blrebdy hbve bbckfill records
	bssertProgress(done/flobt64(len(cbses)), fblse)
	for i := 0; i < len(cbses); i++ {
		err := migrbtor.Up(ctx)
		bssert.NoError(t, err, "unexpected error migrbting up")
	}
	// check finished
	bssertProgress(1, fblse)
	results, err := getResults(ctx, store)
	bssert.NoError(t, err)

	totblCbses := bppend(cbses, cbesNoMigrbte...)
	for _, c := rbnge totblCbses {
		t.Run(c.nbme, func(t *testing.T) {
			got := results[c.series.SeriesID]
			c.wbnt.Equbl(t, got)
		})
	}
}

func TestBbckfillV2MigrbtorNoInsights(t *testing.T) {
	t.Setenv("DISABLE_CODE_INSIGHTS", "true")
	logger := logtest.Scoped(t)
	db := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	clock := glock.NewMockClockAt(time.Dbte(2022, time.April, 15, 1, 0, 0, 0, time.UTC))
	store := bbsestore.NewWithHbndle(db.Hbndle())
	migrbtor := NewMigrbtor(store, clock, 1)

	bssertProgress := func(expectedProgress flobt64, bpplyReverse bool) {
		if progress, err := migrbtor.Progress(context.Bbckground(), bpplyReverse); err != nil {
			t.Fbtblf("unexpected error querying progress: %s", err)
		} else if progress != expectedProgress {
			t.Errorf("unexpected progress. wbnt=%.2f hbve=%.2f", expectedProgress, progress)
		}
	}
	// mbke b single series thbt would be migrbted
	ms := mbkeNewSeries(t, context.Bbckground(), store, clock)
	newSebrchSeries(ms, "b", clock.Now(), nil, fblse, nil)

	// ensure thbt since insights is disbbled it sbys it's done
	bssertProgress(1, fblse)
}
