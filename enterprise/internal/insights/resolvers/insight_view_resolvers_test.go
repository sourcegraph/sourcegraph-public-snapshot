package resolvers

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func TestValidateLineChartSearchInsightInput(t *testing.T) {
	btrue := true
	bfalse := false
	t.Run("should pass validation with capture groups", func(t *testing.T) {
		input := graphqlbackend.LineChartSearchInsightDataSeriesInput{
			Query: "",
			TimeScope: graphqlbackend.TimeScopeInput{StepInterval: &graphqlbackend.TimeIntervalStepInput{
				Unit:  "MONTH",
				Value: 1,
			}},
			RepositoryScope: graphqlbackend.RepositoryScopeInput{Repositories: []string{"github.com/mycompany/myrepo"}},
			Options: graphqlbackend.LineChartDataSeriesOptionsInput{
				Label:     addrStr("mylabel"),
				LineColor: addrStr("red"),
			},
			GeneratedFromCaptureGroups: &btrue,
		}

		err := validateLineChartSearchInsightInput(input)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("should pass validation without capture groups", func(t *testing.T) {
		input := graphqlbackend.LineChartSearchInsightDataSeriesInput{
			Query: "",
			TimeScope: graphqlbackend.TimeScopeInput{StepInterval: &graphqlbackend.TimeIntervalStepInput{
				Unit:  "MONTH",
				Value: 1,
			}},
			RepositoryScope: graphqlbackend.RepositoryScopeInput{Repositories: []string{}},
			Options: graphqlbackend.LineChartDataSeriesOptionsInput{
				Label:     addrStr("mylabel"),
				LineColor: addrStr("red"),
			},
			GeneratedFromCaptureGroups: &bfalse,
		}

		err := validateLineChartSearchInsightInput(input)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("fails because global capture groups", func(t *testing.T) {
		input := graphqlbackend.LineChartSearchInsightDataSeriesInput{
			Query: "",
			TimeScope: graphqlbackend.TimeScopeInput{StepInterval: &graphqlbackend.TimeIntervalStepInput{
				Unit:  "MONTH",
				Value: 1,
			}},
			RepositoryScope: graphqlbackend.RepositoryScopeInput{Repositories: []string{}},
			Options: graphqlbackend.LineChartDataSeriesOptionsInput{
				Label:     addrStr("mylabel"),
				LineColor: addrStr("red"),
			},
			GeneratedFromCaptureGroups: &btrue,
		}

		err := validateLineChartSearchInsightInput(input)
		if err == nil {
			t.Error(err)
		}
	})
}

func addrStr(input string) *string {
	return &input
}

func TestFilterRepositories(t *testing.T) {
	tests := []struct {
		name         string
		repositories []string
		filters      types.InsightViewFilters
		want         []string
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := filterRepositories(test.filters, test.repositories)
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
