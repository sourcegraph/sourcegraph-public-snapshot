package codycontext

import (
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestFileMatchToContextMatches(t *testing.T) {
	cases := []struct {
		fileMatch *result.FileMatch
		want      FileChunkContext
	}{
		{
			// No chunk matches returns first 20 lines
			fileMatch: &result.FileMatch{
				File: result.File{
					Path:     "main.go",
					CommitID: "abc123",
					Repo: types.MinimalRepo{
						Name: "repo",
						ID:   1,
					},
				},
				ChunkMatches: nil,
			},
			want: FileChunkContext{
				RepoName:  "repo",
				RepoID:    1,
				CommitID:  "abc123",
				Path:      "main.go",
				StartLine: 0,
			},
		},
		{
			// With chunk match returns context around first chunk
			fileMatch: &result.FileMatch{
				File: result.File{
					Path:     "main.go",
					CommitID: "abc123",
					Repo: types.MinimalRepo{
						Name: "repo",
						ID:   1,
					},
				},
				ChunkMatches: []result.ChunkMatch{{
					Content:      "first chunk of content",
					ContentStart: result.Location{Line: 90, Column: 2},
				}, {
					Content:      "second chunk of content",
					ContentStart: result.Location{Line: 37, Column: 10},
				}},
			},
			want: FileChunkContext{
				RepoName:  "repo",
				RepoID:    1,
				CommitID:  "abc123",
				Path:      "main.go",
				StartLine: 85,
			},
		},
		{
			// With symbol match returns context around first symbol
			fileMatch: &result.FileMatch{
				File: result.File{
					Path:     "main.go",
					CommitID: "abc123",
					Repo: types.MinimalRepo{
						Name: "repo",
						ID:   1,
					},
				},
				Symbols: []*result.SymbolMatch{
					{
						Symbol: result.Symbol{
							Line: 23,
							Name: "symbol",
						},
					},
					{
						Symbol: result.Symbol{
							Line: 37,
							Name: "symbol",
						},
					},
				},
			},
			want: FileChunkContext{
				RepoName:  "repo",
				RepoID:    1,
				CommitID:  "abc123",
				Path:      "main.go",
				StartLine: 18,
			},
		},
	}

	for _, tc := range cases {
		got := fileMatchToContextMatch(tc.fileMatch)
		if diff := cmp.Diff(tc.want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestReposAsRegexp(t *testing.T) {
	t.Run("SRCH-658", func(t *testing.T) {
		repos := []types.RepoIDName{{Name: "github.com/sourcegraph/docs"}, {Name: "github.com/sourcegraph/docs"}}
		pattern := reposAsRegexp(repos)
		re := regexp.MustCompile(pattern)
		require.True(t, re.MatchString("github.com/sourcegraph/docs"))
		require.False(t, re.MatchString("github.com/sourcegraph/docsite"))
	})
}

func TestBuildKeywordQuery(t *testing.T) {
	cases := []struct {
		args GetContextArgs
		want autogold.Value
	}{{
		args: GetContextArgs{
			Repos: []types.RepoIDName{{ID: 0, Name: "testrepo"}},
			Query: "test query",
		},
		want: autogold.Expect("repo:^testrepo$  -file:\\.min\\.js$ -file:\\.map$ -file:\\.tsbuildinfo$ -file:(\\/|^)umd\\/ -file:(\\/|^)amd\\/ -file:(\\/|^)cjs\\/ test query"),
	}, {
		args: GetContextArgs{
			Repos: []types.RepoIDName{{ID: 0, Name: "testrepo"}, {ID: 1, Name: "special*Chars$"}},
			Query: "test query",
		},
		want: autogold.Expect("repo:(?:^testrepo$)|(?:^special\\*Chars\\$$)  -file:\\.min\\.js$ -file:\\.map$ -file:\\.tsbuildinfo$ -file:(\\/|^)umd\\/ -file:(\\/|^)amd\\/ -file:(\\/|^)cjs\\/ test query"),
	}, {
		args: GetContextArgs{
			Repos:        []types.RepoIDName{{ID: 0, Name: "testrepo"}},
			FilePatterns: []types.RegexpPattern{"file1", "dir/.*"},
			Query:        "test query",
		},
		want: autogold.Expect("repo:^testrepo$ file:(?:file1|dir/.*) -file:\\.min\\.js$ -file:\\.map$ -file:\\.tsbuildinfo$ -file:(\\/|^)umd\\/ -file:(\\/|^)amd\\/ -file:(\\/|^)cjs\\/ test query"),
	}}

	for _, tc := range cases {
		got := buildKeywordQuery(tc.args)
		tc.want.Equal(t, got)
	}
}
