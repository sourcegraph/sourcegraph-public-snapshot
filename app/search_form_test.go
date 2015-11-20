package app

import (
	"net/url"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestSearchFormInfo(t *testing.T) {
	sampleRepo := &sourcegraph.Repo{URI: "x/r", DefaultBranch: "b"}
	sampleCommit := &vcs.Commit{ID: "c"}

	revRouteVar := func(rev, commitID string) map[string]string {
		s := sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "dummy/repo"}, Rev: rev, CommitID: commitID}
		return s.RouteVars()
	}

	tests := map[string]struct {
		tmplData interface{}
		want     *searchForm
	}{
		"global search (non-results page)": {
			tmplData: struct{}{},
			want: &searchForm{
				ActionURL:     &url.URL{Path: "/search"},
				PJAXContainer: "#pjax-container",
			},
		},
		"global search (results page)": {
			tmplData: struct {
				SearchOptions *sourcegraph.SearchOptions
				SearchResults *sourcegraph.SearchResults
			}{
				SearchOptions: &sourcegraph.SearchOptions{Query: "myquery"},
				SearchResults: &sourcegraph.SearchResults{
					ResolvedTokens: sourcegraph.PBTokensWrap(sourcegraph.Tokens{sourcegraph.Term("myquery")}),
				},
			},
			want: &searchForm{
				ActionURL:      &url.URL{Path: "/search"},
				InputValue:     "myquery",
				ResolvedTokens: sourcegraph.Tokens{sourcegraph.Term("myquery")},
				PJAXContainer:  "#pjax-container",
			},
		},
		"repo search (non-results page)": {
			tmplData: struct {
				Repo             *sourcegraph.Repo
				CurrentRouteVars map[string]string
			}{
				Repo:             sampleRepo,
				CurrentRouteVars: revRouteVar("", ""),
			},
			want: &searchForm{
				ActionURL: &url.URL{Path: "/x/r/.search"},
				ActionImplicitQueryPrefix: sourcegraph.Tokens{
					sourcegraph.RepoToken{URI: "x/r", Repo: sampleRepo},
				},
				InputValue: "x/r",
				ResolvedTokens: sourcegraph.Tokens{
					sourcegraph.RepoToken{URI: "x/r", Repo: sampleRepo},
				},
				PJAXContainer: "#repo-pjax-container",
			},
		},
		"repo search (results page)": {
			tmplData: struct {
				Repo             *sourcegraph.Repo
				CurrentRouteVars map[string]string
				SearchOptions    *sourcegraph.SearchOptions
				SearchResults    *sourcegraph.SearchResults
			}{
				Repo:             sampleRepo,
				CurrentRouteVars: revRouteVar("", ""),
				SearchOptions:    &sourcegraph.SearchOptions{Query: "myquery"},
				SearchResults: &sourcegraph.SearchResults{
					ResolvedTokens: sourcegraph.PBTokensWrap(sourcegraph.Tokens{
						sourcegraph.RepoToken{URI: "x/r", Repo: sampleRepo},
						sourcegraph.Term("myquery"),
					}),
				},
			},
			want: &searchForm{
				ActionURL: &url.URL{Path: "/x/r/.search"},
				ActionImplicitQueryPrefix: sourcegraph.Tokens{
					sourcegraph.RepoToken{URI: "x/r", Repo: sampleRepo},
				},
				InputValue: "x/r myquery",
				ResolvedTokens: sourcegraph.Tokens{
					sourcegraph.RepoToken{URI: "x/r", Repo: sampleRepo},
					sourcegraph.Term("myquery"),
				},
				PJAXContainer: "#repo-pjax-container",
			},
		},
		"repo specific-commit search (non-results page)": {
			tmplData: struct {
				Repo             *sourcegraph.Repo
				CurrentRouteVars map[string]string
				RepoBuildInfo    *sourcegraph.RepoBuildInfo
			}{
				Repo:             sampleRepo,
				CurrentRouteVars: revRouteVar("v", "c"),
				RepoBuildInfo: &sourcegraph.RepoBuildInfo{
					LastSuccessful:       &sourcegraph.Build{CommitID: "c"},
					LastSuccessfulCommit: sampleCommit,
				},
			},
			want: &searchForm{
				ActionURL: &url.URL{Path: "/x/r@v/.search"},
				ActionImplicitQueryPrefix: sourcegraph.Tokens{
					sourcegraph.RepoToken{URI: "x/r", Repo: sampleRepo},
					sourcegraph.RevToken{Rev: "v", Commit: sampleCommit},
				},
				InputValue: "x/r :v",
				ResolvedTokens: sourcegraph.Tokens{
					sourcegraph.RepoToken{URI: "x/r", Repo: sampleRepo},
					sourcegraph.RevToken{Rev: "v", Commit: sampleCommit},
				},
				PJAXContainer: "#repo-pjax-container",
			},
		},
		"repo specific-commit search (results page)": {
			tmplData: struct {
				Repo             *sourcegraph.Repo
				CurrentRouteVars map[string]string
				RepoBuildInfo    *sourcegraph.RepoBuildInfo
				SearchOptions    *sourcegraph.SearchOptions
				SearchResults    *sourcegraph.SearchResults
			}{
				Repo:             sampleRepo,
				CurrentRouteVars: revRouteVar("v", "c"),
				RepoBuildInfo: &sourcegraph.RepoBuildInfo{
					LastSuccessful:       &sourcegraph.Build{CommitID: "c"},
					LastSuccessfulCommit: sampleCommit,
				},
				SearchOptions: &sourcegraph.SearchOptions{Query: "myquery"},
				SearchResults: &sourcegraph.SearchResults{
					ResolvedTokens: sourcegraph.PBTokensWrap(sourcegraph.Tokens{
						sourcegraph.RepoToken{URI: "x/r", Repo: sampleRepo},
						sourcegraph.RevToken{Rev: "v", Commit: sampleCommit},
						sourcegraph.Term("myquery"),
					}),
				},
			},
			want: &searchForm{
				ActionURL: &url.URL{Path: "/x/r@v/.search"},
				ActionImplicitQueryPrefix: sourcegraph.Tokens{
					sourcegraph.RepoToken{URI: "x/r", Repo: sampleRepo},
					sourcegraph.RevToken{Rev: "v", Commit: sampleCommit},
				},
				InputValue: "x/r :v myquery",
				ResolvedTokens: sourcegraph.Tokens{
					sourcegraph.RepoToken{URI: "x/r", Repo: sampleRepo},
					sourcegraph.RevToken{Rev: "v", Commit: sampleCommit},
					sourcegraph.Term("myquery"),
				},
				PJAXContainer: "#repo-pjax-container",
			},
		},
	}
	for label, test := range tests {
		fi, err := searchFormInfo(test.tmplData)
		if err != nil {
			t.Errorf("%s: error: %s", label, err)
			continue
		}
		if !reflect.DeepEqual(fi, test.want) {
			t.Errorf("%s: got != want\n\ngot:\n%+v\n\nwant:\n%+v", label, fi, test.want)
		}
	}
}
