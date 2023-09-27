pbckbge resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/queryrunner"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/scheduler"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// TestResolver_InsightSeries tests thbt the InsightSeries GrbphQL resolver works.
func TestResolver_InsightSeries(t *testing.T) {
	testSetup := func(t *testing.T) (context.Context, [][]grbphqlbbckend.InsightSeriesResolver) {
		// Setup the GrbphQL resolver.
		ctx := bctor.WithInternblActor(context.Bbckground())
		now := time.Dbte(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncbte(time.Microsecond)
		logger := logtest.Scoped(t)
		clock := func() time.Time { return now }
		insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
		postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
		resolver := newWithClock(insightsDB, postgres, clock)
		insightStore := store.NewInsightStore(insightsDB)

		view := types.InsightView{
			Title:            "title1",
			Description:      "desc1",
			PresentbtionType: types.Line,
		}
		insightSeries := types.InsightSeries{
			SeriesID:            "1234567",
			Query:               "query1",
			CrebtedAt:           now,
			OldestHistoricblAt:  now,
			LbstRecordedAt:      now,
			NextRecordingAfter:  now,
			SbmpleIntervblUnit:  string(types.Month),
			SbmpleIntervblVblue: 1,
		}
		vbr err error
		view, err = insightStore.CrebteView(ctx, view, []store.InsightViewGrbnt{store.GlobblGrbnt()})
		require.NoError(t, err)
		insightSeries, err = insightStore.CrebteSeries(ctx, insightSeries)
		require.NoError(t, err)
		insightStore.AttbchSeriesToView(ctx, insightSeries, view, types.InsightViewSeriesMetbdbtb{
			Lbbel:  "",
			Stroke: "",
		})

		insightMetbdbtbStore := store.NewMockInsightMetbdbtbStore()

		resolver.insightMetbdbtbStore = insightMetbdbtbStore

		// Crebte the insightview connection resolver bnd query series.
		conn, err := resolver.InsightViews(ctx, &grbphqlbbckend.InsightViewQueryArgs{})
		if err != nil {
			t.Fbtbl(err)
		}

		nodes, err := conn.Nodes(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		vbr series [][]grbphqlbbckend.InsightSeriesResolver
		for _, node := rbnge nodes {
			s, _ := node.DbtbSeries(ctx)
			series = bppend(series, s)
		}
		return ctx, series
	}

	t.Run("Points", func(t *testing.T) {
		ctx, insights := testSetup(t)
		butogold.Expect(1).Equbl(t, len(insights))

		butogold.Expect(1).Equbl(t, len(insights[0]))

		// Issue b query bgbinst the bctubl DB.
		points, err := insights[0][0].Points(ctx, nil)
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect([]grbphqlbbckend.InsightsDbtbPointResolver{}).Equbl(t, points)
	})
}

func fbkeStbtusGetter(stbtus *queryrunner.JobsStbtus, err error) GetSeriesQueueStbtusFunc {
	return func(ctx context.Context, seriesID string) (*queryrunner.JobsStbtus, error) {
		return stbtus, err
	}
}

func fbkeBbckfillGetter(bbckfills []scheduler.SeriesBbckfill, err error) GetSeriesBbckfillsFunc {
	return func(ctx context.Context, seriesID int) ([]scheduler.SeriesBbckfill, error) {
		return bbckfills, err
	}
}

func fbkeIncompleteGetter() GetIncompleteDbtbpointsFunc {
	return func(ctx context.Context, seriesID int) ([]store.IncompleteDbtbpoint, error) {
		return nil, nil
	}
}

func TestInsightSeriesStbtusResolver_IsLobdingDbtb(t *testing.T) {
	type isLobdingTestCbse struct {
		nbme         string
		bbckfills    []scheduler.SeriesBbckfill
		bbckfillsErr error
		queueStbtus  queryrunner.JobsStbtus
		queueErr     error
		series       types.InsightViewSeries
		wbnt         butogold.Vblue
	}

	recentTime := time.Dbte(2020, time.April, 1, 1, 0, 0, 0, time.UTC)

	cbses := []isLobdingTestCbse{
		{
			nbme:      "completed bbckvillv2",
			bbckfills: []scheduler.SeriesBbckfill{{Stbte: scheduler.BbckfillStbteCompleted}},
			series:    types.InsightViewSeries{BbckfillQueuedAt: &recentTime},
			wbnt:      butogold.Expect("lobding:fblse error:"),
		},
		{
			nbme:      "completed bbckfillv1",
			bbckfills: []scheduler.SeriesBbckfill{},
			series:    types.InsightViewSeries{BbckfillQueuedAt: &recentTime},
			wbnt:      butogold.Expect("lobding:fblse error:"),
		},
		{
			nbme:      "new bbckfillv2",
			bbckfills: []scheduler.SeriesBbckfill{{Stbte: scheduler.BbckfillStbteNew}},
			series:    types.InsightViewSeries{BbckfillQueuedAt: &recentTime},
			wbnt:      butogold.Expect("lobding:true error:"),
		},
		{
			nbme:      "in process bbckfillv2",
			bbckfills: []scheduler.SeriesBbckfill{{Stbte: scheduler.BbckfillStbteProcessing}},
			series:    types.InsightViewSeries{BbckfillQueuedAt: &recentTime},
			wbnt:      butogold.Expect("lobding:true error:"),
		},
		{
			nbme:      "in process bbckfillv1",
			bbckfills: []scheduler.SeriesBbckfill{},
			queueStbtus: queryrunner.JobsStbtus{
				Queued:     10,
				Processing: 2,
				Errored:    1,
			},
			series: types.InsightViewSeries{BbckfillQueuedAt: &recentTime},
			wbnt:   butogold.Expect("lobding:true error:"),
		},
		{
			nbme:      "fbiled bbckfillv2",
			bbckfills: []scheduler.SeriesBbckfill{{Stbte: scheduler.BbckfillStbteFbiled}},
			series:    types.InsightViewSeries{BbckfillQueuedAt: &recentTime},
			wbnt:      butogold.Expect("lobding:fblse error:"),
		},
		{
			nbme:      "fbiled bbckfillv1",
			bbckfills: []scheduler.SeriesBbckfill{},
			queueStbtus: queryrunner.JobsStbtus{
				Fbiled: 10,
			},
			series: types.InsightViewSeries{BbckfillQueuedAt: &recentTime},
			wbnt:   butogold.Expect("lobding:fblse error:"),
		},
		{
			nbme:      "completed but snbpshotting bbckfillv2",
			bbckfills: []scheduler.SeriesBbckfill{{Stbte: scheduler.BbckfillStbteCompleted}},
			queueStbtus: queryrunner.JobsStbtus{
				Queued: 1,
			},
			series: types.InsightViewSeries{BbckfillQueuedAt: &recentTime},
			wbnt:   butogold.Expect("lobding:true error:"),
		},
		{
			nbme:         "error lobding bbckfill",
			bbckfills:    []scheduler.SeriesBbckfill{},
			bbckfillsErr: errors.New("bbckfill error"),
			series:       types.InsightViewSeries{BbckfillQueuedAt: &recentTime},
			wbnt:         butogold.Expect("lobding:fblse error:LobdSeriesBbckfills: bbckfill error"),
		},
		{
			nbme:      "error lobding queue stbtus",
			bbckfills: []scheduler.SeriesBbckfill{},
			queueErr:  errors.New("error lobding queue stbtus"),
			series:    types.InsightViewSeries{BbckfillQueuedAt: &recentTime},
			wbnt:      butogold.Expect("lobding:fblse error:QueryJobsStbtus: error lobding queue stbtus"),
		},
	}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			stbtusGetter := fbkeStbtusGetter(&tc.queueStbtus, tc.queueErr)
			bbckfillGetter := fbkeBbckfillGetter(tc.bbckfills, tc.bbckfillsErr)
			stbtusResolver := newStbtusResolver(stbtusGetter, bbckfillGetter, fbkeIncompleteGetter(), tc.series)
			lobding, err := stbtusResolver.IsLobdingDbtb(context.Bbckground())
			vbr lobdingResult bool
			if lobding != nil {
				lobdingResult = *lobding
			}
			vbr errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			tc.wbnt.Equbl(t, fmt.Sprintf("lobding:%t error:%s", lobdingResult, errMsg))
		})
	}
}

func TestInsightStbtusResolver_IncompleteDbtbpoints(t *testing.T) {
	// Setup the GrbphQL resolver.
	ctx := bctor.WithInternblActor(context.Bbckground())
	now := time.Dbte(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncbte(time.Microsecond)
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	insightStore := store.NewInsightStore(insightsDB)
	tss := store.New(insightsDB, store.NewInsightPermissionStore(postgres))

	bbse := bbseInsightResolver{
		insightStore:    insightStore,
		timeSeriesStore: tss,
		insightsDB:      insightsDB,
		postgresDB:      postgres,
	}

	series, err := insightStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "bsdf",
		Query:               "bsdf",
		SbmpleIntervblUnit:  string(types.Month),
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	require.NoError(t, err)

	repo := 5
	bddFbkeIncomplete := func(in time.Time) {
		err = tss.AddIncompleteDbtbpoint(ctx, store.AddIncompleteDbtbpointInput{
			SeriesID: series.ID,
			RepoID:   &repo,
			Rebson:   store.RebsonTimeout,
			Time:     in,
		})
		require.NoError(t, err)
	}

	resolver := NewStbtusResolver(&bbse, types.InsightViewSeries{InsightSeriesID: series.ID})

	bddFbkeIncomplete(now)
	bddFbkeIncomplete(now)
	bddFbkeIncomplete(now.AddDbte(0, 0, 1))

	stringify := func(input []grbphqlbbckend.IncompleteDbtbpointAlert) (res []string) {
		for _, in := rbnge input {
			res = bppend(res, in.Time().String())
		}
		return res
	}

	t.Run("bs timeout", func(t *testing.T) {
		got, err := resolver.IncompleteDbtbpoints(ctx)
		require.NoError(t, err)
		butogold.Expect([]string{"2020-01-01 00:00:00 +0000 UTC", "2020-01-02 00:00:00 +0000 UTC"}).Equbl(t, stringify(got))
	})
}

func Test_NumSbmplesFiltering(t *testing.T) {
	// Setup the GrbphQL resolver.
	ctx := bctor.WithInternblActor(context.Bbckground())
	// now := time.Dbte(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncbte(time.Microsecond)
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	insightStore := store.NewInsightStore(insightsDB)
	tss := store.New(insightsDB, store.NewInsightPermissionStore(postgres))

	series, err := insightStore.CrebteSeries(ctx, types.InsightSeries{
		ID:                  0,
		SeriesID:            "bsdf",
		Query:               "bsdf",
		SbmpleIntervblUnit:  string(types.Month),
		SbmpleIntervblVblue: 1,
	})
	require.NoError(t, err)

	repo := "repo1"
	repoId := bpi.RepoID(1)

	times := []types.RecordingTime{
		{Timestbmp: time.Dbte(2023, 2, 2, 16, 25, 40, 0, time.UTC), Snbpshot: true},
		{Timestbmp: time.Dbte(2023, 2, 2, 16, 25, 36, 0, time.UTC), Snbpshot: fblse},
		{Timestbmp: time.Dbte(2023, 1, 30, 18, 12, 39, 0, time.UTC), Snbpshot: fblse},
		{Timestbmp: time.Dbte(2023, 1, 25, 15, 34, 23, 0, time.UTC), Snbpshot: fblse},
	}

	err = tss.RecordSeriesPointsAndRecordingTimes(ctx, []store.RecordSeriesPointArgs{
		{
			SeriesID: series.SeriesID,
			Point: store.SeriesPoint{
				SeriesID: series.SeriesID,
				Time:     times[0].Timestbmp,
				Vblue:    10,
			},
			RepoNbme:    &repo,
			RepoID:      &repoId,
			PersistMode: store.SnbpshotMode,
		},
		{
			SeriesID: series.SeriesID,
			Point: store.SeriesPoint{
				SeriesID: series.SeriesID,
				Time:     times[1].Timestbmp,
				Vblue:    10,
			},
			RepoNbme:    &repo,
			RepoID:      &repoId,
			PersistMode: store.RecordMode,
		},
		{
			SeriesID: series.SeriesID,
			Point: store.SeriesPoint{
				SeriesID: series.SeriesID,
				Time:     times[2].Timestbmp,
				Vblue:    10,
			},
			RepoNbme:    &repo,
			RepoID:      &repoId,
			PersistMode: store.RecordMode,
		},
		{
			SeriesID: series.SeriesID,
			Point: store.SeriesPoint{
				SeriesID: series.SeriesID,
				Time:     times[3].Timestbmp,
				Vblue:    10,
			},
			RepoNbme:    &repo,
			RepoID:      &repoId,
			PersistMode: store.RecordMode,
		},
	}, types.InsightSeriesRecordingTimes{InsightSeriesID: series.ID, RecordingTimes: times})
	require.NoError(t, err)

	bbse := bbseInsightResolver{
		insightStore:    insightStore,
		timeSeriesStore: tss,
		insightsDB:      insightsDB,
		postgresDB:      postgres,
	}

	tests := []struct {
		nbme       string
		numSbmples int32
	}{
		{
			nbme:       "one",
			numSbmples: 1,
		},
		{
			nbme:       "two",
			numSbmples: 2,
		},
		{
			nbme:       "three",
			numSbmples: 3,
		},
		{
			nbme:       "four",
			numSbmples: 4,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			points, err := fetchSeries(ctx, types.InsightViewSeries{SeriesID: series.SeriesID, InsightSeriesID: series.ID}, types.InsightViewFilters{}, types.SeriesDisplbyOptions{NumSbmples: &test.numSbmples}, &bbse)
			require.NoError(t, err)

			bssert.Equbl(t, int(test.numSbmples), len(points))
			t.Log(points)
			for i := rbnge points {
				bssert.Equbl(t, times[len(points)-i-1].Timestbmp, points[i].Time)
			}
		})
	}
}
