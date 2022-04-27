package resolvers

import (
	"context"
	"sort"
	"testing"

	internalTypes "github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
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
		base := baseInsightResolver{
			insightStore:   store.NewInsightStore(insightsDB),
			dashboardStore: store.NewDashboardStore(insightsDB),
			insightsDB:     insightsDB,
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

type fakeSearchContextLoader struct {
	mocks map[string]*internalTypes.SearchContext
}

func (f *fakeSearchContextLoader) GetByName(ctx context.Context, name string) (*internalTypes.SearchContext, error) {
	return f.mocks[name], nil
}
