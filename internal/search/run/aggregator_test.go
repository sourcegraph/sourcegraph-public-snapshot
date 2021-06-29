package run

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCheckDiffCommitSearchLimits(t *testing.T) {
	cases := []struct {
		name        string
		resultType  string
		numRepoRevs int
		fields      []query.Node
		wantError   error
	}{
		{
			name:        "diff_search_warns_on_repos_greater_than_search_limit",
			resultType:  "diff",
			numRepoRevs: 51,
			wantError:   &RepoLimitError{ResultType: "diff", Max: 50},
		},
		{
			name:        "commit_search_warns_on_repos_greater_than_search_limit",
			resultType:  "commit",
			numRepoRevs: 51,
			wantError:   &RepoLimitError{ResultType: "commit", Max: 50},
		},
		{
			name:        "commit_search_warns_on_repos_greater_than_search_limit_with_time_filter",
			fields:      []query.Node{query.Parameter{Field: "after"}},
			resultType:  "commit",
			numRepoRevs: 20000,
			wantError:   &TimeLimitError{ResultType: "commit", Max: 10000},
		},
		{
			name:        "no_warning_when_commit_search_within_search_limit",
			resultType:  "commit",
			numRepoRevs: 50,
			wantError:   nil,
		},
		{
			name:        "no_search_limit_on_queries_including_after_filter",
			fields:      []query.Node{query.Parameter{Field: "after"}},
			resultType:  "commit",
			numRepoRevs: 200,
			wantError:   nil,
		},
		{
			name:        "no_search_limit_on_queries_including_before_filter",
			fields:      []query.Node{query.Parameter{Field: "before"}},
			resultType:  "commit",
			numRepoRevs: 200,
			wantError:   nil,
		},
	}

	for _, test := range cases {
		repoRevs := make([]*search.RepositoryRevisions, test.numRepoRevs)
		for i := range repoRevs {
			repoRevs[i] = &search.RepositoryRevisions{
				Repo: types.RepoName{ID: api.RepoID(i)},
			}
		}

		haveErr := checkDiffCommitSearchLimits(
			context.Background(),
			&search.TextParameters{
				RepoPromise: (&search.RepoPromise{}).Resolve(repoRevs),
				Query:       test.fields,
			},
			test.resultType)

		if diff := cmp.Diff(test.wantError, haveErr); diff != "" {
			t.Fatalf("test %s, mismatched error (-want, +got):\n%s", test.name, diff)
		}
	}
}
