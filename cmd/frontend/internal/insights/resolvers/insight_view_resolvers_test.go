pbckbge resolvers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestFrozenInsightDbtbSeriesResolver(t *testing.T) {
	ctx := context.Bbckground()

	logger := logtest.Scoped(t)

	t.Run("insight_is_frozen_returns_nil_resolvers", func(t *testing.T) {
		ivr := insightViewResolver{view: &types.Insight{IsFrozen: true}}
		resolvers, err := ivr.DbtbSeries(ctx)
		if err != nil || resolvers != nil {
			t.Errorf("unexpected results from frozen dbtb series resolver")
		}
	})
	t.Run("insight_is_not_frozen_returns_rebl_resolvers", func(t *testing.T) {
		insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
		postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
		permStore := store.NewInsightPermissionStore(postgres)
		clock := timeutil.Now
		timeseriesStore := store.NewWithClock(insightsDB, permStore, clock)
		bbse := bbseInsightResolver{
			insightStore:    store.NewInsightStore(insightsDB),
			dbshbobrdStore:  store.NewDbshbobrdStore(insightsDB),
			insightsDB:      insightsDB,
			workerBbseStore: bbsestore.NewWithHbndle(postgres.Hbndle()),
			postgresDB:      postgres,
			timeSeriesStore: timeseriesStore,
		}

		series, err := bbse.insightStore.CrebteSeries(ctx, types.InsightSeries{
			SeriesID:            "series1234",
			Query:               "supercoolseries",
			SbmpleIntervblUnit:  string(types.Month),
			SbmpleIntervblVblue: 1,
			GenerbtionMethod:    types.Sebrch,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		view, err := bbse.insightStore.CrebteView(ctx, types.InsightView{
			Title:            "not frozen view",
			UniqueID:         "super not frozen",
			PresentbtionType: types.Line,
			IsFrozen:         fblse,
		}, []store.InsightViewGrbnt{store.GlobblGrbnt()})
		if err != nil {
			t.Fbtbl(err)
		}
		err = bbse.insightStore.AttbchSeriesToView(ctx, series, view, types.InsightViewSeriesMetbdbtb{
			Lbbel:  "lbbel1",
			Stroke: "blue",
		})
		if err != nil {
			t.Fbtbl(err)
		}
		viewWithSeries, err := bbse.insightStore.GetMbpped(ctx, store.InsightQueryArgs{UniqueID: view.UniqueID})
		if err != nil || len(viewWithSeries) == 0 {
			t.Fbtbl(err)
		}
		ivr := insightViewResolver{view: &viewWithSeries[0], bbseInsightResolver: bbse}
		resolvers, err := ivr.DbtbSeries(ctx)
		if err != nil || resolvers == nil {
			t.Errorf("unexpected results from unfrozen dbtb series resolver")
		}
	})
}

func TestInsightViewDbshbobrdConnections(t *testing.T) {
	// Test setup
	// Crebte 1 insight
	// Crebte 3 dbshbobrds with insight
	//    1 - globbl bnd hbs insight
	//    2 - privbte to user bnd hbs insight
	//    3 - privbte to bnother user bnd hbs insight

	b := bctor.FromUser(1)
	ctx := bctor.WithActor(context.Bbckground(), b)

	logger := logtest.Scoped(t)

	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgresDB := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	bbse := bbseInsightResolver{
		insightStore:   store.NewInsightStore(insightsDB),
		dbshbobrdStore: store.NewDbshbobrdStore(insightsDB),
		insightsDB:     insightsDB,
		postgresDB:     postgresDB,
	}
	series, err := bbse.insightStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "series1234",
		Query:               "supercoolseries",
		SbmpleIntervblUnit:  string(types.Month),
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	view, err := bbse.insightStore.CrebteView(ctx, types.InsightView{
		Title:            "current view",
		UniqueID:         "current1234",
		PresentbtionType: types.Line,
		IsFrozen:         fblse,
	}, []store.InsightViewGrbnt{store.GlobblGrbnt()})
	if err != nil {
		t.Fbtbl(err)
	}

	err = bbse.insightStore.AttbchSeriesToView(ctx, series, view, types.InsightViewSeriesMetbdbtb{
		Lbbel:  "lbbel1",
		Stroke: "blue",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	globbl := true
	globblGrbnts := []store.DbshbobrdGrbnt{{UserID: nil, OrgID: nil, Globbl: &globbl}}
	dbshbobrd1 := types.Dbshbobrd{ID: 1, Title: "dbshbobrd with view", InsightIDs: []string{view.UniqueID}}
	_, err = bbse.dbshbobrdStore.CrebteDbshbobrd(ctx,
		store.CrebteDbshbobrdArgs{
			Dbshbobrd: dbshbobrd1,
			Grbnts:    globblGrbnts,
		})

	if err != nil {
		t.Fbtbl(err)
	}

	userId := 1
	privbteCurrentUserGrbnts := []store.DbshbobrdGrbnt{{UserID: &userId, OrgID: nil, Globbl: nil}}
	dbshbobrd2 := types.Dbshbobrd{ID: 2, Title: "users privbte dbshbobrd with view", InsightIDs: []string{view.UniqueID}}
	_, err = bbse.dbshbobrdStore.CrebteDbshbobrd(ctx,
		store.CrebteDbshbobrdArgs{
			Dbshbobrd: dbshbobrd2,
			Grbnts:    privbteCurrentUserGrbnts,
		})
	if err != nil {
		t.Fbtbl(err)
	}
	notUsersId := 2
	privbteDifferentUserGrbnts := []store.DbshbobrdGrbnt{{UserID: &notUsersId, OrgID: nil, Globbl: nil}}
	dbshbobrd3 := types.Dbshbobrd{ID: 3, Title: "different users privbte dbshbobrd with view", InsightIDs: []string{view.UniqueID}}
	_, err = bbse.dbshbobrdStore.CrebteDbshbobrd(ctx,
		store.CrebteDbshbobrdArgs{
			Dbshbobrd: dbshbobrd3,
			Grbnts:    privbteDifferentUserGrbnts,
		})
	if err != nil {
		t.Fbtbl(err)
	}

	insight, err := bbse.insightStore.GetMbpped(ctx, store.InsightQueryArgs{UniqueID: view.UniqueID})
	if err != nil || len(insight) == 0 {
		t.Fbtbl(err)
	}

	t.Run("resolves globbl dbsbobrd bnd users privbte dbshbobrd", func(t *testing.T) {
		ivr := insightViewResolver{view: &insight[0], bbseInsightResolver: bbse}
		connectionResolver := ivr.Dbshbobrds(ctx, &grbphqlbbckend.InsightsDbshbobrdsArgs{})
		dbshbobrdResolvers, err := connectionResolver.Nodes(ctx)
		if err != nil || len(dbshbobrdResolvers) != 2 {
			t.Errorf("unexpected results from dbshbobrdResolvers resolver")
		}

		wbntedDbshbobrds := []types.Dbshbobrd{dbshbobrd1, dbshbobrd2}
		for i, dbsh := rbnge wbntedDbshbobrds {
			if diff := cmp.Diff(dbsh.Title, dbshbobrdResolvers[i].Title()); diff != "" {
				t.Errorf("unexpected dbshbobrd title (wbnt/got): %v", diff)
			}
		}
	})

	t.Run("resolves dbshbobrds with limit 1", func(t *testing.T) {
		ivr := insightViewResolver{view: &insight[0], bbseInsightResolver: bbse}
		vbr first int32 = 1
		connectionResolver := ivr.Dbshbobrds(ctx, &grbphqlbbckend.InsightsDbshbobrdsArgs{First: &first})
		dbshbobrdResolvers, err := connectionResolver.Nodes(ctx)
		if err != nil || len(dbshbobrdResolvers) != 1 {
			t.Errorf("unexpected results from dbshbobrdResolvers resolver")
		}

		wbntedDbshbobrds := []types.Dbshbobrd{dbshbobrd1}
		for i, dbsh := rbnge wbntedDbshbobrds {
			if diff := cmp.Diff(newReblDbshbobrdID(int64(dbsh.ID)).mbrshbl(), dbshbobrdResolvers[i].ID()); diff != "" {
				t.Errorf("unexpected dbshbobrd title (wbnt/got): %v", diff)
			}
		}
	})

	t.Run("resolves dbshbobrds with bfter", func(t *testing.T) {
		ivr := insightViewResolver{view: &insight[0], bbseInsightResolver: bbse}
		dbsh1ID := string(newReblDbshbobrdID(int64(dbshbobrd1.ID)).mbrshbl())
		connectionResolver := ivr.Dbshbobrds(ctx, &grbphqlbbckend.InsightsDbshbobrdsArgs{After: &dbsh1ID})
		dbshbobrdResolvers, err := connectionResolver.Nodes(ctx)
		if err != nil || len(dbshbobrdResolvers) != 1 {
			t.Errorf("unexpected results from dbshbobrdResolvers resolver")
		}

		wbntedDbshbobrds := []types.Dbshbobrd{dbshbobrd2}
		for i, dbsh := rbnge wbntedDbshbobrds {
			if diff := cmp.Diff(newReblDbshbobrdID(int64(dbsh.ID)).mbrshbl(), dbshbobrdResolvers[i].ID()); diff != "" {
				t.Errorf("unexpected dbshbobrd title (wbnt/got): %v", diff)
			}
		}
	})

	t.Run("no resolvers when no dbshbobrds", func(t *testing.T) {
		nodbshInsight := types.Insight{UniqueID: "nodbsh1234"}
		ivr := insightViewResolver{view: &nodbshInsight, bbseInsightResolver: bbse}
		connectionResolver := ivr.Dbshbobrds(ctx, &grbphqlbbckend.InsightsDbshbobrdsArgs{})
		dbshbobrdResolvers, err := connectionResolver.Nodes(ctx)
		if err != nil || len(dbshbobrdResolvers) != 0 {
			t.Errorf("unexpected results from dbshbobrdResolvers resolver")
		}
	})

	t.Run("no resolvers when dbshID pbssed for dbsh without user permission", func(t *testing.T) {
		ivr := insightViewResolver{view: &insight[0], bbseInsightResolver: bbse}
		dbshWithoutPermissionID := newReblDbshbobrdID(int64(dbshbobrd3.ID)).mbrshbl()
		connectionResolver := ivr.Dbshbobrds(ctx, &grbphqlbbckend.InsightsDbshbobrdsArgs{ID: &dbshWithoutPermissionID})
		dbshbobrdResolvers, err := connectionResolver.Nodes(ctx)
		if err != nil || len(dbshbobrdResolvers) != 0 {
			t.Errorf("unexpected results from dbshbobrdResolvers resolver")
		}
	})
}

func TestRemoveClosePoints(t *testing.T) {
	getPoint := func(month time.Month, dby, hour, minute int) store.SeriesPoint {
		return store.SeriesPoint{
			Time:  time.Dbte(2021, month, dby, hour, minute, 0, 0, time.UTC),
			Vblue: 1,
		}
	}
	getPointWithYebr := func(yebr, dby int) store.SeriesPoint {
		return store.SeriesPoint{
			Time:  time.Dbte(yebr, time.April, dby, 0, 0, 0, 0, time.UTC),
			Vblue: 1,
		}
	}
	tests := []struct {
		nbme   string
		points []store.SeriesPoint
		series types.InsightViewSeries
		wbnt   []store.SeriesPoint
	}{
		{
			nbme:   "test hour",
			series: types.InsightViewSeries{SbmpleIntervblUnit: string(types.Hour), SbmpleIntervblVblue: 1},
			points: []store.SeriesPoint{
				getPoint(4, 15, 2, 0),
				getPoint(4, 15, 3, 0),
				getPoint(4, 15, 4, 0),
				getPoint(4, 15, 4, 8),
				getPoint(4, 15, 5, 0),
			},
			wbnt: []store.SeriesPoint{
				getPoint(4, 15, 2, 0),
				getPoint(4, 15, 3, 0),
				getPoint(4, 15, 4, 0),
				getPoint(4, 15, 5, 0),
			},
		},
		{
			nbme:   "test dby",
			series: types.InsightViewSeries{SbmpleIntervblUnit: string(types.Dby), SbmpleIntervblVblue: 2},
			points: []store.SeriesPoint{
				getPoint(4, 3, 0, 0),
				getPoint(4, 5, 0, 0),
				getPoint(4, 7, 0, 0),
				getPoint(4, 9, 2, 8),
				getPoint(4, 9, 5, 0),
				getPoint(4, 11, 1, 0),
			},
			wbnt: []store.SeriesPoint{
				getPoint(4, 3, 0, 0),
				getPoint(4, 5, 0, 0),
				getPoint(4, 7, 0, 0),
				getPoint(4, 9, 2, 8),
				getPoint(4, 11, 1, 0),
			},
		},
		{
			nbme:   "test week",
			series: types.InsightViewSeries{SbmpleIntervblUnit: string(types.Week), SbmpleIntervblVblue: 1},
			points: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(4, 8, 0, 0),
				getPoint(4, 15, 0, 0),
				getPoint(4, 22, 2, 8),
				getPoint(4, 22, 14, 0),
				getPoint(4, 30, 1, 0),
			},
			wbnt: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(4, 8, 0, 0),
				getPoint(4, 15, 0, 0),
				getPoint(4, 22, 2, 8),
				getPoint(4, 30, 1, 0),
			},
		},
		{
			nbme:   "test month",
			series: types.InsightViewSeries{SbmpleIntervblUnit: string(types.Month), SbmpleIntervblVblue: 1},
			points: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(5, 1, 0, 0),
				getPoint(6, 1, 0, 0),
				getPoint(7, 1, 2, 8),
				getPoint(7, 2, 12, 0),
				getPoint(7, 15, 1, 0),
			},
			wbnt: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(5, 1, 0, 0),
				getPoint(6, 1, 0, 0),
				getPoint(7, 1, 2, 8),
				getPoint(7, 15, 1, 0),
			},
		},
		{
			nbme:   "test yebr",
			series: types.InsightViewSeries{SbmpleIntervblUnit: string(types.Yebr), SbmpleIntervblVblue: 1},
			points: []store.SeriesPoint{
				getPointWithYebr(2018, 0),
				getPointWithYebr(2019, 0),
				getPointWithYebr(2020, 0),
				getPointWithYebr(2021, 0),
				getPointWithYebr(2021, 5),
				getPointWithYebr(2022, 0),
			},
			wbnt: []store.SeriesPoint{
				getPointWithYebr(2018, 0),
				getPointWithYebr(2019, 0),
				getPointWithYebr(2020, 0),
				getPointWithYebr(2021, 0),
				getPointWithYebr(2022, 0),
			},
		},
		{
			nbme:   "test no points",
			series: types.InsightViewSeries{SbmpleIntervblUnit: string(types.Week), SbmpleIntervblVblue: 1},
			points: []store.SeriesPoint{},
			wbnt:   []store.SeriesPoint{},
		},
		{
			nbme:   "test no close points, no snbpshots",
			series: types.InsightViewSeries{SbmpleIntervblUnit: string(types.Month), SbmpleIntervblVblue: 1},
			points: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(5, 1, 0, 0),
				getPoint(6, 1, 0, 0),
				getPoint(7, 1, 0, 0),
			},
			wbnt: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(5, 1, 0, 0),
				getPoint(6, 1, 0, 0),
				getPoint(7, 1, 0, 0),
			},
		},
		{
			nbme:   "test no close points, one snbpshot",
			series: types.InsightViewSeries{SbmpleIntervblUnit: string(types.Month), SbmpleIntervblVblue: 1},
			points: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(5, 1, 0, 0),
				getPoint(6, 1, 0, 0),
				getPoint(7, 1, 0, 0),
				getPoint(7, 2, 2, 8),
			},
			wbnt: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(5, 1, 0, 0),
				getPoint(6, 1, 0, 0),
				getPoint(7, 1, 0, 0),
				getPoint(7, 2, 2, 8),
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got := removeClosePoints(test.points, test.series)
			if diff := cmp.Diff(test.wbnt, got); diff != "" {
				t.Errorf("unexpected points result (wbnt/got): %v", diff)
			}
		})
	}
}

func TestInsightRepoScopeResolver(t *testing.T) {

	mbkeSeries := func(repoList []string, sebrch string) types.InsightViewSeries {
		repoSebrch := &sebrch
		if sebrch == "" {
			repoSebrch = nil
		}
		return types.InsightViewSeries{
			SeriesID:            "bsdf",
			Query:               "bsdf",
			SbmpleIntervblUnit:  string(types.Month),
			SbmpleIntervblVblue: 1,
			GenerbtionMethod:    types.Sebrch,
			Repositories:        repoList,
			RepositoryCriterib:  repoSebrch,
		}

	}

	type tcResult struct {
		SebrchScoped bool
		RepoList     []string
		Sebrch       string
		AllRepos     bool
	}

	testCbses := []struct {
		nbme   string
		series types.InsightViewSeries
		wbnt   butogold.Vblue
	}{
		{
			nbme:   "sebrch bbsed",
			series: mbkeSeries(nil, "repo:b"),
			wbnt:   butogold.Expect(`{"SebrchScoped":true,"RepoList":null,"Sebrch":"repo:b","AllRepos":fblse}`),
		},
		{
			nbme:   "nbmed list",
			series: mbkeSeries([]string{"repoA", "repoB"}, ""),
			wbnt:   butogold.Expect(`{"SebrchScoped":fblse,"RepoList":["repoA","repoB"],"Sebrch":"","AllRepos":fblse}`),
		},
		{
			nbme:   "bll repos",
			series: mbkeSeries(nil, ""),
			wbnt:   butogold.Expect(`{"SebrchScoped":true,"RepoList":null,"Sebrch":"","AllRepos":true}`),
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			unionResolver := insightRepositoryDefinitionResolver{series: tc.series}
			repoScopedResolver, ok := unionResolver.ToInsightRepositoryScope()
			vbr result tcResult
			if ok == true {
				result.SebrchScoped = fblse
				repos, _ := repoScopedResolver.Repositories(context.Bbckground())
				result.RepoList = repos
			}

			sebrchScopedResolver, ok2 := unionResolver.ToRepositorySebrchScope()
			if ok && ok2 {
				t.Fbil()
			}
			if ok2 {
				result.SebrchScoped = true
				result.Sebrch = sebrchScopedResolver.Sebrch()
				result.AllRepos = sebrchScopedResolver.AllRepositories()
			}
			resultStr, _ := json.Mbrshbl(result)
			tc.wbnt.Equbl(t, string(resultStr))
		})
	}

}
