package resolvers

import (
	"context"
	"database/sql"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	internalTypes "github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
)

func addrStr(input string) *string {
	return &input
}

func TestFilterRepositories(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name           string
		repositories   []string
		filters        types.InsightViewFilters
		want           []string
		searchContexts []struct {
			name  string
			query string
		}
	}{
		{name: "test one exclude",
			repositories: []string{"github.com/sourcegraph/sourcegraph", "gitlab.com/myrepo/repo"},
			filters:      types.InsightViewFilters{ExcludeRepoRegex: addrStr("gitlab.com")},
			want:         []string{"github.com/sourcegraph/sourcegraph"},
		},
		{name: "test one include",
			repositories: []string{"github.com/sourcegraph/sourcegraph", "gitlab.com/myrepo/repo"},
			filters:      types.InsightViewFilters{IncludeRepoRegex: addrStr("gitlab.com")},
			want:         []string{"gitlab.com/myrepo/repo"},
		},
		{name: "test no filters",
			repositories: []string{"github.com/sourcegraph/sourcegraph", "gitlab.com/myrepo/repo"},
			filters:      types.InsightViewFilters{},
			want:         []string{"github.com/sourcegraph/sourcegraph", "gitlab.com/myrepo/repo"},
		},
		{name: "test exclude and include",
			repositories: []string{"github.com/sourcegraph/sourcegraph", "gitlab.com/myrepo/repo", "gitlab.com/yourrepo/yourrepo"},
			filters:      types.InsightViewFilters{ExcludeRepoRegex: addrStr("github.*"), IncludeRepoRegex: addrStr("myrepo")},
			want:         []string{"gitlab.com/myrepo/repo"},
		},
		{name: "test exclude all",
			repositories: []string{"github.com/sourcegraph/sourcegraph", "gitlab.com/myrepo/repo", "gitlab.com/yourrepo/yourrepo"},
			filters:      types.InsightViewFilters{ExcludeRepoRegex: addrStr(".*")},
			want:         []string{},
		},
		{name: "test include all",
			repositories: []string{"github.com/sourcegraph/sourcegraph", "gitlab.com/myrepo/repo", "gitlab.com/yourrepo/yourrepo"},
			filters:      types.InsightViewFilters{IncludeRepoRegex: addrStr(".*")},
			want:         []string{"github.com/sourcegraph/sourcegraph", "gitlab.com/myrepo/repo", "gitlab.com/yourrepo/yourrepo"},
		},
		{name: "test context include",
			repositories: []string{"github.com/sourcegraph/sourcegraph", "gitlab.com/myrepo/repo", "gitlab.com/yourrepo/yourrepo"},
			filters:      types.InsightViewFilters{SearchContexts: []string{"@dev/mycontext123"}},
			searchContexts: []struct {
				name  string
				query string
			}{
				{name: "@dev/mycontext123", query: "repo:^github\\.com/sourcegraph/.*"},
			},
			want: []string{"github.com/sourcegraph/sourcegraph"},
		},
		{name: "test context exclude",
			repositories: []string{"github.com/sourcegraph/sourcegraph", "gitlab.com/myrepo/repo", "gitlab.com/yourrepo/yourrepo"},
			filters:      types.InsightViewFilters{SearchContexts: []string{"@dev/mycontext123"}},
			searchContexts: []struct {
				name  string
				query string
			}{
				{name: "@dev/mycontext123", query: "-repo:^github\\.com/sourcegraph/.*"},
			},
			want: []string{"gitlab.com/myrepo/repo", "gitlab.com/yourrepo/yourrepo"},
		},
		{name: "test context exclude include",
			repositories: []string{"github.com/sourcegraph/sourcegraph", "gitlab.com/myrepo/repo", "gitlab.com/yourrepo/yourrepo"},
			filters:      types.InsightViewFilters{SearchContexts: []string{"@dev/mycontext123"}},
			searchContexts: []struct {
				name  string
				query string
			}{
				{name: "@dev/mycontext123", query: "-repo:^github.* repo:myrepo"},
			},
			want: []string{"gitlab.com/myrepo/repo"},
		},
		{name: "test context exclude regex include",
			repositories: []string{"github.com/sourcegraph/sourcegraph", "gitlab.com/myrepo/repo", "gitlab.com/yourrepo/yourrepo"},
			filters:      types.InsightViewFilters{SearchContexts: []string{"@dev/mycontext123"}, IncludeRepoRegex: addrStr("myrepo")},
			searchContexts: []struct {
				name  string
				query string
			}{
				{name: "@dev/mycontext123", query: "-repo:^github.*"},
			},
			want: []string{"gitlab.com/myrepo/repo"},
		},
		{name: "test context include regex exclude",
			repositories: []string{"github.com/sourcegraph/sourcegraph", "gitlab.com/myrepo/repo", "gitlab.com/yourrepo/yourrepo"},
			filters:      types.InsightViewFilters{SearchContexts: []string{"@dev/mycontext123"}, ExcludeRepoRegex: addrStr("^github.*")},
			searchContexts: []struct {
				name  string
				query string
			}{
				{name: "@dev/mycontext123", query: "repo:myrepo"},
			},
			want: []string{"gitlab.com/myrepo/repo"},
		},
		{name: "test context and regex include",
			repositories: []string{"github.com/sourcegraph/sourcegraph", "gitlab.com/myrepo/repo", "gitlab.com/yourrepo/yourrepo"},
			filters:      types.InsightViewFilters{SearchContexts: []string{"@dev/mycontext123"}, IncludeRepoRegex: addrStr("myrepo")},
			searchContexts: []struct {
				name  string
				query string
			}{
				{name: "@dev/mycontext123", query: "repo:gitlab"},
			},
			want: []string{"gitlab.com/myrepo/repo"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mocks := make(map[string]*internalTypes.SearchContext)
			for _, searchContextDef := range test.searchContexts {
				mocks[searchContextDef.name] = &internalTypes.SearchContext{Name: searchContextDef.name, Query: searchContextDef.query}
			}

			got, err := filterRepositories(ctx, test.filters, test.repositories, &fakeSearchContextLoader{mocks: mocks})
			if err != nil {
				t.Error(err)
			}
			// sort for test determinism
			sort.Slice(got, func(i, j int) bool {
				return got[i] < got[j]
			})
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("unexpected repository result (want/got): %v", diff)
			}
		})
	}
}

func TestFrozenInsightDataSeriesResolver(t *testing.T) {
	ctx := context.Background()

	t.Run("insight_is_frozen_returns_nil_resolvers", func(t *testing.T) {
		ivr := insightViewResolver{view: &types.Insight{IsFrozen: true}}
		resolvers, err := ivr.DataSeries(ctx)
		if err != nil || resolvers != nil {
			t.Errorf("unexpected results from frozen data series resolver")
		}
	})
	t.Run("insight_is_not_frozen_returns_real_resolvers", func(t *testing.T) {
		insightsDB := dbtest.NewInsightsDB(t)
		postgres := dbtest.NewDB(t)
		permStore := store.NewInsightPermissionStore(postgres)
		clock := timeutil.Now
		timeseriesStore := store.NewWithClock(insightsDB, permStore, clock)
		base := baseInsightResolver{
			insightStore:    store.NewInsightStore(insightsDB),
			dashboardStore:  store.NewDashboardStore(insightsDB),
			insightsDB:      insightsDB,
			workerBaseStore: basestore.NewWithDB(postgres, sql.TxOptions{}),
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

	insightsDB := dbtest.NewInsightsDB(t)
	postgresDB := dbtest.NewDB(t)
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
	globalGrants := []store.DashboardGrant{{nil, nil, &global}}
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
	privateCurrentUserGrants := []store.DashboardGrant{{&userId, nil, nil}}
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
	privateDifferentUserGrants := []store.DashboardGrant{{&notUsersId, nil, nil}}
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

type fakeSearchContextLoader struct {
	mocks map[string]*internalTypes.SearchContext
}

func (f *fakeSearchContextLoader) GetByName(ctx context.Context, name string) (*internalTypes.SearchContext, error) {
	return f.mocks[name], nil
}
