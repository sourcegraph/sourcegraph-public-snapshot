package graphqlbackend

import (
	"path"
	"sort"
	"strings"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func TestSearchSorting(t *testing.T) {
	tests := []struct {
		name, query string
		expect      []searchTestItem
	}{
		{
			name:  "basic repo match",
			query: "mux",
			expect: []searchTestItem{
				{repoURI: "github.com/gorilla/mux"},
				{repoURI: "github.com/docker/docker", removed: true},
				{repoURI: "github.com/kubernetes/kubernetes", removed: true},
			},
		},
		{
			name:  "basic file match",
			query: "bar",
			expect: []searchTestItem{
				{filePath: "pkg/test/foo/bar.go"},
				{filePath: "pkg/test/foo/baz.go", removed: true},
				{filePath: "pkg/main.go", removed: true},
			},
		},
		{
			name:  "basic profile match",
			query: "vscode",
			expect: []searchTestItem{
				{searchProfile: "vscode (sample)"},
				{searchProfile: "golang", removed: true},
				{searchProfile: "kubernetes", removed: true},
			},
		},
		{
			// See https://github.com/sourcegraph/sourcegraph/issues/6418
			name:  "fork ranking",
			query: "mux",
			expect: []searchTestItem{
				{repoURI: "github.com/gorilla/mux"},
				{repoURI: "github.com/another2/mux"},
				{repoURI: "github.com/slimsag/mux", fork: true},
			},
		},
		{
			name:  "same score same length",
			query: "mux",
			expect: []searchTestItem{
				{repoURI: "github.com/another/mux"},
				{repoURI: "github.com/gorilla/mux"},
			},
		},
		{
			// See https://github.com/sourcegraph/sourcegraph/issues/7233#issue-258643923
			name:  "issue 7233",
			query: "backend",
			expect: []searchTestItem{
				// TODO: dir path is not matched so all items get removed. That's not good!
				{filePath: "pkg/backend/trace.go", removed: true},
				{filePath: "vendor/github.com/dgrijalva/jwt-go/token.go", removed: true},
				{filePath: "vendor/github.com/emicklei/go-restful/swagger/CHANGES.md", removed: true},
				{filePath: "vendor/github.com/coreos/go-oidc/key/manager.go", removed: true},
				{filePath: "vendor/github.com/emicklei/go-restful/CHANGES.md", removed: true},
				{filePath: "pkg/backend/repo_vcs.go", removed: true},
				{filePath: "cmd/frontend/internal/cli/middleware/blackhole.go", removed: true},
				{filePath: "vendor/github.com/beorn7/perks/quantile/exampledata.txt", removed: true},
				{filePath: "pkg/backend/repos_vcs_test.go", removed: true},
				{filePath: "cmd/frontend/internal/app/tracking/slack/sourcegraph_slack_bot.go", removed: true},
			},
		},
		{
			// See https://github.com/sourcegraph/sourcegraph/issues/7233#issuecomment-330449691
			name:  "issue 7233 2",
			query: "backend",
			expect: []searchTestItem{
				// TODO: dir path is not matched so all subsequent items get removed. That's not good!
				{filePath: "cmd/frontend/internal/graphqlbackend/graphqlbackend.go"},
				{filePath: "pkg/backend/pkgs.go", removed: true},
				{filePath: "pkg/backend/repos.go", removed: true},
				{filePath: "pkg/backend/trace.go", removed: true},
				{filePath: "pkg/backend/mocks.go", removed: true},
				{filePath: "pkg/backend/defs_refs.go", removed: true},
				{filePath: "pkg/backend/repos_vcs.go", removed: true},
				{filePath: "pkg/backend/repos_mock.go", removed: true},
				{filePath: "pkg/backend/repos_test.go", removed: true},
				{filePath: "pkg/backend/repos_vcs_test.go", removed: true},
			},
		},
		{
			name:  "profiles then repos then files",
			query: "code",
			expect: []searchTestItem{
				{searchProfile: "code"},
				{filePath: "some/src/code"},
				{repoURI: "github.com/muxuezi/code"},
			},
		},
		{
			// See first image of https://github.com/sourcegraph/sourcegraph/issues/7233#issuecomment-330976509
			name:  "issue 7233 files above repos",
			query: "mux.g",
			expect: []searchTestItem{
				{filePath: "mux.go"},
				{filePath: "mux_test.go"},
				{repoURI: "github.com/muxuezi/muxuezi.github.io"},
				{repoURI: "github.com/kcparashar/tmux.github.io"},
				{repoURI: "github.com/tmux/tmux.github.io"},
				{repoURI: "github.com/termux/termux.github.io"},
				{repoURI: "github.com/nacimux/vuejs.org"},
			},
		},
		{
			// See second image of https://github.com/sourcegraph/sourcegraph/issues/7233#issuecomment-330976509
			name:  "issue 7233 file name above dir",
			query: "backend",
			expect: []searchTestItem{
				// TODO: dir path is not matched so all subsequent items get removed. That's not good!
				{filePath: "web/src/backend.tsx"},
				{filePath: "cmd/frontend/internal/graphqlbackend/graphqlbackend.go"},
				{filePath: "cmd/frontend/internal/graphqlbackend/graphqlbackend_test.go"},
				{filePath: "pkg/backend/pkgs.go", removed: true},
				{filePath: "pkg/backend/trace.go", removed: true},
				{filePath: "pkg/backend/mocks.go", removed: true},
				{filePath: "pkg/backend/repos.go", removed: true},
				{filePath: "web/src/backend/lsp.tsx", removed: true},
				{filePath: "pkg/backend/defs_refs.go", removed: true},
				{filePath: "pkg/backend/repos_vcs.go", removed: true},
				{filePath: "pkg/backend/repos_test.go", removed: true},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Convert expected []searchTestItem -> []*searchResultResolver for sorting purposes.
			// Then invert expected items to prevent e.g. no sorting at all from passing tests.
			scorer := newScorer(test.query)
			expectResolvers := itemsToSearchResultResolvers(scorer, test.expect)
			var got []*searchResultResolver
			for i := len(expectResolvers) - 1; i >= 0; i-- {
				got = append(got, expectResolvers[i])
			}

			// Now actually sort our unsorted results.
			sort.Sort(searchResultSorter(got))

			// Compare results against what we expect.
			printDiff := func() {
				// Determine optimal string padding.
				padding := 0
				for _, e := range test.expect {
					if v := len(e.String()); v > padding {
						padding = v
					}
				}
				padding += 4

				// Print diff.
				for i, e := range test.expect {
					got := tStringResult(got[i])
					want := e.String()
					eq := "=="
					if got != want {
						eq = "!="
					}
					t.Logf("%d. %s got %+v want %+v", i, eq, padString(got, padding), padString(want, padding))
				}
			}

			for i, e := range test.expect {
				if e.String() != tStringResult(got[i]) {
					// test failure.
					printDiff()
					t.FailNow()
					return
				}
			}
		})
	}
}

type searchTestItem struct {
	repoURI, filePath, searchProfile string
	removed                          bool // whether or not item was removed due to final score == 0
	fork                             bool
}

func (i searchTestItem) String() string {
	switch {
	case i.removed:
		return "<removed>"
	case i.repoURI != "":
		return "repo:" + i.repoURI
	case i.filePath != "":
		return "file:" + i.filePath
	case i.searchProfile != "":
		return "profile:" + i.searchProfile
	default:
		panic("never here")
	}
}

func (i searchTestItem) toSearchResultResolver(scorer *scorer) *searchResultResolver {
	var result interface{}
	switch {
	case i.repoURI != "":
		result = &repositoryResolver{repo: &sourcegraph.Repo{URI: i.repoURI, Fork: i.fork}}
	case i.filePath != "":
		result = &fileResolver{name: path.Base(i.filePath), path: i.filePath}
	case i.searchProfile != "":
		result = &searchProfile{name: i.searchProfile}
	default:
		panic("bad test")
	}
	return newSearchResultResolver(result, scorer.calcScore(result))
}

func itemsToSearchResultResolvers(scorer *scorer, items []searchTestItem) []*searchResultResolver {
	var res []*searchResultResolver
	for _, i := range items {
		res = append(res, i.toSearchResultResolver(scorer))
	}
	return res
}

func tStringResult(result *searchResultResolver) string {
	var name string
	switch r := result.result.(type) {
	case *repositoryResolver:
		name = "repo:" + r.repo.URI
	case *fileResolver:
		name = "file:" + r.path
	case *searchProfile:
		name = "profile:" + r.name
	default:
		panic("never here")
	}
	if result.score == 0 {
		return "<removed>"
	}
	return name
}

func padString(s string, n int) string {
	if len(s) > n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
