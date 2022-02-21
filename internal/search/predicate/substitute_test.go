package predicate

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func Test_searchResultsToRepoNodes(t *testing.T) {
	cases := []struct {
		matches []result.Match
		res     string
		err     string
	}{{
		matches: []result.Match{
			&result.RepoMatch{Name: "repo_a"},
		},
		res: `"repo:^repo_a$"`,
	}, {
		matches: []result.Match{
			&result.RepoMatch{Name: "repo_a", Rev: "main"},
		},
		res: `"repo:^repo_a$@main"`,
	}, {
		matches: []result.Match{
			&result.FileMatch{},
		},
		err: "expected type",
	}}

	for _, tc := range cases {
		t.Run(tc.res, func(t *testing.T) {
			nodes, err := searchResultsToRepoNodes(tc.matches)
			if err != nil {
				require.Contains(t, err.Error(), tc.err)
				return
			}
			require.Equal(t, tc.res, query.Q(nodes).String())
		})
	}
}

func Test_searchResultsToFileNodes(t *testing.T) {
	cases := []struct {
		matches []result.Match
		res     string
		err     string
	}{{
		matches: []result.Match{
			&result.FileMatch{
				File: result.File{
					Repo: types.MinimalRepo{
						Name: "repo_a",
					},
					Path: "my/file/path.txt",
				},
			},
		},
		res: `(and "repo:^repo_a$" "file:^my/file/path\\.txt$")`,
	}, {
		matches: []result.Match{
			&result.FileMatch{
				File: result.File{
					Repo: types.MinimalRepo{
						Name: "repo_a",
					},
					InputRev: func() *string { s := "main"; return &s }(),
					Path:     "my/file/path1.txt",
				},
			},
			&result.FileMatch{
				File: result.File{
					Repo: types.MinimalRepo{
						Name: "repo_b",
					},
					Path: "my/file/path2.txt",
				},
			},
		},
		res: `(and "repo:^repo_a$@main" "file:^my/file/path1\\.txt$") (and "repo:^repo_b$" "file:^my/file/path2\\.txt$")`,
	}}

	for _, tc := range cases {
		t.Run(tc.res, func(t *testing.T) {
			nodes, err := searchResultsToFileNodes(tc.matches)
			if err != nil {
				require.Contains(t, err.Error(), tc.err)
				return
			}
			require.Equal(t, tc.res, query.Q(nodes).String())
		})
	}
}
