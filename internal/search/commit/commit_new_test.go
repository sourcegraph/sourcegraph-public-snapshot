package commit

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCheckSearchLimits(t *testing.T) {
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

		haveErr := CheckSearchLimits(
			test.fields,
			len(repoRevs),
			test.resultType,
		)

		if diff := cmp.Diff(test.wantError, haveErr); diff != "" {
			t.Fatalf("test %s, mismatched error (-want, +got):\n%s", test.name, diff)
		}
	}
}

func TestQueryToGitQuery(t *testing.T) {
	type testCase struct {
		name   string
		input  query.Q
		diff   bool
		output protocol.Node
	}

	cases := []testCase{{
		name: "negated repo does not result in nil node (#26032)",
		input: []query.Node{
			query.Parameter{Field: query.FieldRepo, Negated: true},
		},
		diff:   false,
		output: &protocol.Boolean{Value: true},
	}, {
		name: "expensive nodes are placed last",
		input: []query.Node{
			query.Pattern{Value: "a"},
			query.Parameter{Field: query.FieldAuthor, Value: "b"},
		},
		diff: true,
		output: protocol.NewAnd(
			&protocol.AuthorMatches{Expr: "b", IgnoreCase: true},
			&protocol.DiffMatches{Expr: "a", IgnoreCase: true},
		),
	}, {
		name: "all supported nodes are converted",
		input: []query.Node{
			query.Parameter{Field: query.FieldAuthor, Value: "author"},
			query.Parameter{Field: query.FieldCommitter, Value: "committer"},
			query.Parameter{Field: query.FieldBefore, Value: "2021-09-10"},
			query.Parameter{Field: query.FieldAfter, Value: "2021-09-08"},
			query.Parameter{Field: query.FieldFile, Value: "file"},
			query.Parameter{Field: query.FieldMessage, Value: "message1"},
			query.Pattern{Value: "message2"},
		},
		diff: false,
		output: protocol.NewAnd(
			&protocol.CommitBefore{Time: time.Date(2021, 9, 10, 0, 0, 0, 0, time.UTC)},
			&protocol.CommitAfter{Time: time.Date(2021, 9, 8, 0, 0, 0, 0, time.UTC)},
			&protocol.AuthorMatches{Expr: "author", IgnoreCase: true},
			&protocol.CommitterMatches{Expr: "committer", IgnoreCase: true},
			&protocol.MessageMatches{Expr: "message1", IgnoreCase: true},
			&protocol.MessageMatches{Expr: "message2", IgnoreCase: true},
			&protocol.DiffModifiesFile{Expr: "file", IgnoreCase: true},
		),
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := queryToGitQuery(tc.input, tc.diff)
			require.Equal(t, tc.output, output)
		})
	}
}
