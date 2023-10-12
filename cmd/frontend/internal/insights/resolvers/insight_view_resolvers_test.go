package resolvers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestFrozenInsightDataSeriesResolver(t *testing.T) {
	ctx := context.Background()

	logger := logtest.Scoped(t)

	t.Run("insight_is_frozen_returns_nil_resolvers", func(t *testing.T) {
		ivr := insightViewResolver{view: &types.Insight{IsFrozen: true}}
		resolvers, err := ivr.DataSeries(ctx)
		if err != nil || resolvers != nil {
			t.Errorf("unexpected results from frozen data series resolver")
		}
	})
	t.Run("insight_is_not_frozen_returns_real_resolvers", func(t *testing.T) {
		insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
		postgres := database.NewDB(logger, dbtest.NewDB(t))
		permStore := store.NewInsightPermissionStore(postgres)
		clock := timeutil.Now
		timeseriesStore := store.NewWithClock(insightsDB, permStore, clock)
		base := baseInsightResolver{
			insightStore:    store.NewInsightStore(insightsDB),
			dashboardStore:  store.NewDashboardStore(insightsDB),
			insightsDB:      insightsDB,
			workerBaseStore: basestore.NewWithHandle(postgres.Handle()),
			postgresDB:      postgres,
			timeSeriesStore: timeseriesStore,
		}

		series, err := base.insightStore.CreateSeries(ctx, types.InsightSeries{
			SeriesID:            "series1234",
			Query:               "supercoolseries",
			SampleIntervalUnit:  string(types.Month),
			SampleIntervalValue: 1,
			GenerationMethod:    types.Search,
		})
		if err != nil {
			t.Fatal(err)
		}
		view, err := base.insightStore.CreateView(ctx, types.InsightView{
			Title:            "not frozen view",
			UniqueID:         "super not frozen",
			PresentationType: types.Line,
			IsFrozen:         false,
		}, []store.InsightViewGrant{store.GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}
		err = base.insightStore.AttachSeriesToView(ctx, series, view, types.InsightViewSeriesMetadata{
			Label:  "label1",
			Stroke: "blue",
		})
		if err != nil {
			t.Fatal(err)
		}
		viewWithSeries, err := base.insightStore.GetMapped(ctx, store.InsightQueryArgs{UniqueID: view.UniqueID})
		if err != nil || len(viewWithSeries) == 0 {
			t.Fatal(err)
		}
		ivr := insightViewResolver{view: &viewWithSeries[0], baseInsightResolver: base}
		resolvers, err := ivr.DataSeries(ctx)
		if err != nil || resolvers == nil {
			t.Errorf("unexpected results from unfrozen data series resolver")
		}
	})
}

func TestInsightViewDashboardConnections(t *testing.T) {
	// Test setup
	// Create 1 insight
	// Create 3 dashboards with insight
	//    1 - global and has insight
	//    2 - private to user and has insight
	//    3 - private to another user and has insight

	a := actor.FromUser(1)
	ctx := actor.WithActor(context.Background(), a)

	logger := logtest.Scoped(t)

	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgresDB := database.NewDB(logger, dbtest.NewDB(t))
	base := baseInsightResolver{
		insightStore:   store.NewInsightStore(insightsDB),
		dashboardStore: store.NewDashboardStore(insightsDB),
		insightsDB:     insightsDB,
		postgresDB:     postgresDB,
	}
	series, err := base.insightStore.CreateSeries(ctx, types.InsightSeries{
		SeriesID:            "series1234",
		Query:               "supercoolseries",
		SampleIntervalUnit:  string(types.Month),
		SampleIntervalValue: 1,
		GenerationMethod:    types.Search,
	})
	if err != nil {
		t.Fatal(err)
	}
	view, err := base.insightStore.CreateView(ctx, types.InsightView{
		Title:            "current view",
		UniqueID:         "current1234",
		PresentationType: types.Line,
		IsFrozen:         false,
	}, []store.InsightViewGrant{store.GlobalGrant()})
	if err != nil {
		t.Fatal(err)
	}

	err = base.insightStore.AttachSeriesToView(ctx, series, view, types.InsightViewSeriesMetadata{
		Label:  "label1",
		Stroke: "blue",
	})
	if err != nil {
		t.Fatal(err)
	}

	global := true
	globalGrants := []store.DashboardGrant{{UserID: nil, OrgID: nil, Global: &global}}
	dashboard1 := types.Dashboard{ID: 1, Title: "dashboard with view", InsightIDs: []string{view.UniqueID}}
	_, err = base.dashboardStore.CreateDashboard(ctx,
		store.CreateDashboardArgs{
			Dashboard: dashboard1,
			Grants:    globalGrants,
		})

	if err != nil {
		t.Fatal(err)
	}

	userId := 1
	privateCurrentUserGrants := []store.DashboardGrant{{UserID: &userId, OrgID: nil, Global: nil}}
	dashboard2 := types.Dashboard{ID: 2, Title: "users private dashboard with view", InsightIDs: []string{view.UniqueID}}
	_, err = base.dashboardStore.CreateDashboard(ctx,
		store.CreateDashboardArgs{
			Dashboard: dashboard2,
			Grants:    privateCurrentUserGrants,
		})
	if err != nil {
		t.Fatal(err)
	}
	notUsersId := 2
	privateDifferentUserGrants := []store.DashboardGrant{{UserID: &notUsersId, OrgID: nil, Global: nil}}
	dashboard3 := types.Dashboard{ID: 3, Title: "different users private dashboard with view", InsightIDs: []string{view.UniqueID}}
	_, err = base.dashboardStore.CreateDashboard(ctx,
		store.CreateDashboardArgs{
			Dashboard: dashboard3,
			Grants:    privateDifferentUserGrants,
		})
	if err != nil {
		t.Fatal(err)
	}

	insight, err := base.insightStore.GetMapped(ctx, store.InsightQueryArgs{UniqueID: view.UniqueID})
	if err != nil || len(insight) == 0 {
		t.Fatal(err)
	}

	t.Run("resolves global dasboard and users private dashboard", func(t *testing.T) {
		ivr := insightViewResolver{view: &insight[0], baseInsightResolver: base}
		connectionResolver := ivr.Dashboards(ctx, &graphqlbackend.InsightsDashboardsArgs{})
		dashboardResolvers, err := connectionResolver.Nodes(ctx)
		if err != nil || len(dashboardResolvers) != 2 {
			t.Errorf("unexpected results from dashboardResolvers resolver")
		}

		wantedDashboards := []types.Dashboard{dashboard1, dashboard2}
		for i, dash := range wantedDashboards {
			if diff := cmp.Diff(dash.Title, dashboardResolvers[i].Title()); diff != "" {
				t.Errorf("unexpected dashboard title (want/got): %v", diff)
			}
		}
	})

	t.Run("resolves dashboards with limit 1", func(t *testing.T) {
		ivr := insightViewResolver{view: &insight[0], baseInsightResolver: base}
		var first int32 = 1
		connectionResolver := ivr.Dashboards(ctx, &graphqlbackend.InsightsDashboardsArgs{First: &first})
		dashboardResolvers, err := connectionResolver.Nodes(ctx)
		if err != nil || len(dashboardResolvers) != 1 {
			t.Errorf("unexpected results from dashboardResolvers resolver")
		}

		wantedDashboards := []types.Dashboard{dashboard1}
		for i, dash := range wantedDashboards {
			if diff := cmp.Diff(newRealDashboardID(int64(dash.ID)).marshal(), dashboardResolvers[i].ID()); diff != "" {
				t.Errorf("unexpected dashboard title (want/got): %v", diff)
			}
		}
	})

	t.Run("resolves dashboards with after", func(t *testing.T) {
		ivr := insightViewResolver{view: &insight[0], baseInsightResolver: base}
		dash1ID := string(newRealDashboardID(int64(dashboard1.ID)).marshal())
		connectionResolver := ivr.Dashboards(ctx, &graphqlbackend.InsightsDashboardsArgs{After: &dash1ID})
		dashboardResolvers, err := connectionResolver.Nodes(ctx)
		if err != nil || len(dashboardResolvers) != 1 {
			t.Errorf("unexpected results from dashboardResolvers resolver")
		}

		wantedDashboards := []types.Dashboard{dashboard2}
		for i, dash := range wantedDashboards {
			if diff := cmp.Diff(newRealDashboardID(int64(dash.ID)).marshal(), dashboardResolvers[i].ID()); diff != "" {
				t.Errorf("unexpected dashboard title (want/got): %v", diff)
			}
		}
	})

	t.Run("no resolvers when no dashboards", func(t *testing.T) {
		nodashInsight := types.Insight{UniqueID: "nodash1234"}
		ivr := insightViewResolver{view: &nodashInsight, baseInsightResolver: base}
		connectionResolver := ivr.Dashboards(ctx, &graphqlbackend.InsightsDashboardsArgs{})
		dashboardResolvers, err := connectionResolver.Nodes(ctx)
		if err != nil || len(dashboardResolvers) != 0 {
			t.Errorf("unexpected results from dashboardResolvers resolver")
		}
	})

	t.Run("no resolvers when dashID passed for dash without user permission", func(t *testing.T) {
		ivr := insightViewResolver{view: &insight[0], baseInsightResolver: base}
		dashWithoutPermissionID := newRealDashboardID(int64(dashboard3.ID)).marshal()
		connectionResolver := ivr.Dashboards(ctx, &graphqlbackend.InsightsDashboardsArgs{ID: &dashWithoutPermissionID})
		dashboardResolvers, err := connectionResolver.Nodes(ctx)
		if err != nil || len(dashboardResolvers) != 0 {
			t.Errorf("unexpected results from dashboardResolvers resolver")
		}
	})
}

func TestRemoveClosePoints(t *testing.T) {
	getPoint := func(month time.Month, day, hour, minute int) store.SeriesPoint {
		return store.SeriesPoint{
			Time:  time.Date(2021, month, day, hour, minute, 0, 0, time.UTC),
			Value: 1,
		}
	}
	getPointWithYear := func(year, day int) store.SeriesPoint {
		return store.SeriesPoint{
			Time:  time.Date(year, time.April, day, 0, 0, 0, 0, time.UTC),
			Value: 1,
		}
	}
	tests := []struct {
		name   string
		points []store.SeriesPoint
		series types.InsightViewSeries
		want   []store.SeriesPoint
	}{
		{
			name:   "test hour",
			series: types.InsightViewSeries{SampleIntervalUnit: string(types.Hour), SampleIntervalValue: 1},
			points: []store.SeriesPoint{
				getPoint(4, 15, 2, 0),
				getPoint(4, 15, 3, 0),
				getPoint(4, 15, 4, 0),
				getPoint(4, 15, 4, 8),
				getPoint(4, 15, 5, 0),
			},
			want: []store.SeriesPoint{
				getPoint(4, 15, 2, 0),
				getPoint(4, 15, 3, 0),
				getPoint(4, 15, 4, 0),
				getPoint(4, 15, 5, 0),
			},
		},
		{
			name:   "test day",
			series: types.InsightViewSeries{SampleIntervalUnit: string(types.Day), SampleIntervalValue: 2},
			points: []store.SeriesPoint{
				getPoint(4, 3, 0, 0),
				getPoint(4, 5, 0, 0),
				getPoint(4, 7, 0, 0),
				getPoint(4, 9, 2, 8),
				getPoint(4, 9, 5, 0),
				getPoint(4, 11, 1, 0),
			},
			want: []store.SeriesPoint{
				getPoint(4, 3, 0, 0),
				getPoint(4, 5, 0, 0),
				getPoint(4, 7, 0, 0),
				getPoint(4, 9, 2, 8),
				getPoint(4, 11, 1, 0),
			},
		},
		{
			name:   "test week",
			series: types.InsightViewSeries{SampleIntervalUnit: string(types.Week), SampleIntervalValue: 1},
			points: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(4, 8, 0, 0),
				getPoint(4, 15, 0, 0),
				getPoint(4, 22, 2, 8),
				getPoint(4, 22, 14, 0),
				getPoint(4, 30, 1, 0),
			},
			want: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(4, 8, 0, 0),
				getPoint(4, 15, 0, 0),
				getPoint(4, 22, 2, 8),
				getPoint(4, 30, 1, 0),
			},
		},
		{
			name:   "test month",
			series: types.InsightViewSeries{SampleIntervalUnit: string(types.Month), SampleIntervalValue: 1},
			points: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(5, 1, 0, 0),
				getPoint(6, 1, 0, 0),
				getPoint(7, 1, 2, 8),
				getPoint(7, 2, 12, 0),
				getPoint(7, 15, 1, 0),
			},
			want: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(5, 1, 0, 0),
				getPoint(6, 1, 0, 0),
				getPoint(7, 1, 2, 8),
				getPoint(7, 15, 1, 0),
			},
		},
		{
			name:   "test year",
			series: types.InsightViewSeries{SampleIntervalUnit: string(types.Year), SampleIntervalValue: 1},
			points: []store.SeriesPoint{
				getPointWithYear(2018, 0),
				getPointWithYear(2019, 0),
				getPointWithYear(2020, 0),
				getPointWithYear(2021, 0),
				getPointWithYear(2021, 5),
				getPointWithYear(2022, 0),
			},
			want: []store.SeriesPoint{
				getPointWithYear(2018, 0),
				getPointWithYear(2019, 0),
				getPointWithYear(2020, 0),
				getPointWithYear(2021, 0),
				getPointWithYear(2022, 0),
			},
		},
		{
			name:   "test no points",
			series: types.InsightViewSeries{SampleIntervalUnit: string(types.Week), SampleIntervalValue: 1},
			points: []store.SeriesPoint{},
			want:   []store.SeriesPoint{},
		},
		{
			name:   "test no close points, no snapshots",
			series: types.InsightViewSeries{SampleIntervalUnit: string(types.Month), SampleIntervalValue: 1},
			points: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(5, 1, 0, 0),
				getPoint(6, 1, 0, 0),
				getPoint(7, 1, 0, 0),
			},
			want: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(5, 1, 0, 0),
				getPoint(6, 1, 0, 0),
				getPoint(7, 1, 0, 0),
			},
		},
		{
			name:   "test no close points, one snapshot",
			series: types.InsightViewSeries{SampleIntervalUnit: string(types.Month), SampleIntervalValue: 1},
			points: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(5, 1, 0, 0),
				getPoint(6, 1, 0, 0),
				getPoint(7, 1, 0, 0),
				getPoint(7, 2, 2, 8),
			},
			want: []store.SeriesPoint{
				getPoint(4, 1, 0, 0),
				getPoint(5, 1, 0, 0),
				getPoint(6, 1, 0, 0),
				getPoint(7, 1, 0, 0),
				getPoint(7, 2, 2, 8),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := removeClosePoints(test.points, test.series)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("unexpected points result (want/got): %v", diff)
			}
		})
	}
}

func TestInsightRepoScopeResolver(t *testing.T) {

	makeSeries := func(repoList []string, search string) types.InsightViewSeries {
		repoSearch := &search
		if search == "" {
			repoSearch = nil
		}
		return types.InsightViewSeries{
			SeriesID:            "asdf",
			Query:               "asdf",
			SampleIntervalUnit:  string(types.Month),
			SampleIntervalValue: 1,
			GenerationMethod:    types.Search,
			Repositories:        repoList,
			RepositoryCriteria:  repoSearch,
		}

	}

	type tcResult struct {
		SearchScoped bool
		RepoList     []string
		Search       string
		AllRepos     bool
	}

	testCases := []struct {
		name   string
		series types.InsightViewSeries
		want   autogold.Value
	}{
		{
			name:   "search based",
			series: makeSeries(nil, "repo:a"),
			want:   autogold.Expect(`{"SearchScoped":true,"RepoList":null,"Search":"repo:a","AllRepos":false}`),
		},
		{
			name:   "named list",
			series: makeSeries([]string{"repoA", "repoB"}, ""),
			want:   autogold.Expect(`{"SearchScoped":false,"RepoList":["repoA","repoB"],"Search":"","AllRepos":false}`),
		},
		{
			name:   "all repos",
			series: makeSeries(nil, ""),
			want:   autogold.Expect(`{"SearchScoped":true,"RepoList":null,"Search":"","AllRepos":true}`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			unionResolver := insightRepositoryDefinitionResolver{series: tc.series}
			repoScopedResolver, ok := unionResolver.ToInsightRepositoryScope()
			var result tcResult
			if ok == true {
				result.SearchScoped = false
				repos, _ := repoScopedResolver.Repositories(context.Background())
				result.RepoList = repos
			}

			searchScopedResolver, ok2 := unionResolver.ToRepositorySearchScope()
			if ok && ok2 {
				t.Fail()
			}
			if ok2 {
				result.SearchScoped = true
				result.Search = searchScopedResolver.Search()
				result.AllRepos = searchScopedResolver.AllRepositories()
			}
			resultStr, _ := json.Marshal(result)
			tc.want.Equal(t, string(resultStr))
		})
	}

}
