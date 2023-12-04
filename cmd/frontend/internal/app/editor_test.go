package app

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestEditorRev(t *testing.T) {
	repoName := api.RepoName("myRepo")
	logger := logtest.Scoped(t)
	backend.Mocks.Repos.ResolveRev = func(_ context.Context, _ *types.Repo, rev string) (api.CommitID, error) {
		if rev == "branch" {
			return api.CommitID(strings.Repeat("b", 40)), nil
		}
		if rev == "" || rev == "defaultBranch" {
			return api.CommitID(strings.Repeat("d", 40)), nil
		}
		if len(rev) == 40 {
			return api.CommitID(rev), nil
		}
		t.Fatalf("unexpected RepoRev request rev: %q", rev)
		return "", nil
	}
	backend.Mocks.Repos.GetByName = func(v0 context.Context, name api.RepoName) (*types.Repo, error) {
		return &types.Repo{ID: api.RepoID(1), Name: name},

			nil
	}
	ctx := context.Background()

	cases := []struct {
		inputRev     string
		expEditorRev string
		beExplicit   bool
	}{
		{strings.Repeat("a", 40), "@" + strings.Repeat("a", 40), false},
		{"branch", "@branch", false},
		{"", "", false},
		{"defaultBranch", "", false},
		{strings.Repeat("d", 40), "", false},                           // default revision
		{strings.Repeat("d", 40), "@" + strings.Repeat("d", 40), true}, // default revision, explicit
	}
	for _, c := range cases {
		got := editorRev(ctx, logger, dbmocks.NewMockDB(), repoName, c.inputRev, c.beExplicit)
		if got != c.expEditorRev {
			t.Errorf("On input rev %q: got %q, want %q", c.inputRev, got, c.expEditorRev)
		}
	}
}

func TestEditorRedirect(t *testing.T) {
	repos := dbmocks.NewMockRepoStore()
	repos.GetFirstRepoNameByCloneURLFunc.SetDefaultReturn("", nil)

	externalServices := dbmocks.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultReturn(
		[]*types.ExternalService{
			{
				ID:          1,
				Kind:        extsvc.KindGitHub,
				DisplayName: "GITHUB #1",
				Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.example.com", "repositoryQuery": ["none"], "token": "abc"}`),
			},
			{
				ID:          2,
				Kind:        extsvc.KindOther,
				DisplayName: "OtherPretty",
				Config:      extsvc.NewUnencryptedConfig(`{"url": "https://somecodehost.com/bar", "repositoryPathPattern": "pretty/{repo}"}`),
			},
			{
				ID:          3,
				Kind:        extsvc.KindOther,
				DisplayName: "OtherDefault",
				Config:      extsvc.NewUnencryptedConfig(`{"url": "https://default.com"}`),
			},
			// This service won't be used, but is included to prevent regression where ReposourceCloneURLToRepoName returned an error when
			// Phabricator was iterated over before the actual code host (e.g. The clone URL is handled by reposource.GitLab).
			{
				ID:          4,
				Kind:        extsvc.KindPhabricator,
				DisplayName: "PHABRICATOR #1",
				Config:      extsvc.NewUnencryptedConfig(`{"repos": [{"path": "default.com/foo/bar", "callsign": "BAR"}], "token": "abc", "url": "https://phabricator.example.com"}`),
			},
			// Code host with SCP-style remote URLs
			{
				ID:          5,
				Kind:        extsvc.KindOther,
				DisplayName: "OtherSCP",
				Config:      extsvc.NewUnencryptedConfig(`{"url":"ssh://git@git.codehost.com"}`),
			},
		},
		nil,
	)

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefaultReturn(repos)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	cases := []struct {
		name            string
		q               url.Values
		wantRedirectURL string
		wantParseErr    string
		wantRedirectErr string
	}{
		{
			name: "open file",
			q: url.Values{
				"remote_url": []string{"git@github.com:a/b"},
				"branch":     []string{"dev"},
				"revision":   []string{"0ad12f"},
				"file":       []string{"mux.go"},
				"start_row":  []string{"123"},
				"start_col":  []string{"1"},
				"end_row":    []string{"123"},
				"end_col":    []string{"10"},
			},
			wantRedirectURL: "/github.com/a/b@0ad12f/-/blob/mux.go?L124%3A2-124%3A11",
		},
		{
			name: "open file no selection",
			q: url.Values{
				"editor":     []string{"Atom"},
				"version":    []string{"v1.2.1"},
				"remote_url": []string{"git@github.com:a/b"},
				"branch":     []string{"dev"},
				"revision":   []string{"0ad12f"},
				"file":       []string{"mux.go"},
			},
			wantRedirectURL: "/github.com/a/b@0ad12f/-/blob/mux.go?L1",
		},
		{
			name: "open file in repository (Phabricator mirrored)",
			q: url.Values{
				"remote_url": []string{"https://default.com/foo/bar"},
				"branch":     []string{"dev"},
				"revision":   []string{"0ad12f"},
				"file":       []string{"mux.go"},
				"start_row":  []string{"123"},
				"start_col":  []string{"1"},
				"end_row":    []string{"123"},
				"end_col":    []string{"10"},
			},
			wantRedirectURL: "/default.com/foo/bar@0ad12f/-/blob/mux.go?L124%3A2-124%3A11",
		},
		{
			name: "open file (generic code host with repositoryPathPattern)",
			q: url.Values{
				"remote_url": []string{"https://somecodehost.com/bar/a/b"},
				"branch":     []string{"dev"},
				"revision":   []string{"0ad12f"},
				"file":       []string{"mux.go"},
				"start_row":  []string{"123"},
				"start_col":  []string{"1"},
				"end_row":    []string{"123"},
				"end_col":    []string{"10"},
			},
			wantRedirectURL: "/pretty/a/b@0ad12f/-/blob/mux.go?L124%3A2-124%3A11",
		},
		{
			name: "open file (generic code host without repositoryPathPattern)",
			q: url.Values{
				"remote_url": []string{"https://default.com/a/b"},
				"branch":     []string{"dev"},
				"revision":   []string{"0ad12f"},
				"file":       []string{"mux.go"},
				"start_row":  []string{"123"},
				"start_col":  []string{"1"},
				"end_row":    []string{"123"},
				"end_col":    []string{"10"},
			},
			wantRedirectURL: "/default.com/a/b@0ad12f/-/blob/mux.go?L124%3A2-124%3A11",
		},
		{
			name: "open file (generic git host with slash prefix in path)",
			q: url.Values{
				"remote_url": []string{"git@git.codehost.com:/owner/repo"},
				"branch":     []string{"dev"},
				"revision":   []string{"0ad12f"},
				"file":       []string{"mux.go"},
				"start_row":  []string{"123"},
				"start_col":  []string{"1"},
				"end_row":    []string{"123"},
				"end_col":    []string{"10"},
			},
			wantRedirectURL: "/git.codehost.com/owner/repo@0ad12f/-/blob/mux.go?L124%3A2-124%3A11",
		},
		{
			name: "open file (generic git host without slash prefix in path)",
			q: url.Values{
				"remote_url": []string{"git@git.codehost.com:owner/repo"},
				"branch":     []string{"dev"},
				"revision":   []string{"0ad12f"},
				"file":       []string{"mux.go"},
				"start_row":  []string{"123"},
				"start_col":  []string{"1"},
				"end_row":    []string{"123"},
				"end_col":    []string{"10"},
			},
			wantRedirectURL: "/git.codehost.com/owner/repo@0ad12f/-/blob/mux.go?L124%3A2-124%3A11",
		},
		{
			name: "search",
			q: url.Values{
				"search": []string{"foobar"},

				// Editor extensions specify these when trying to perform a global search,
				// so we cannot treat these as "search in repo/branch/file". When these are
				// present, a global search must be performed:
				"remote_url": []string{"git@github.com:a/b"},
				"branch":     []string{"dev"},
				"file":       []string{"mux.go"},
			},
			wantRedirectURL: "/search?patternType=literal&q=foobar",
		},
		{
			name: "search in repository",
			q: url.Values{
				"search":            []string{"foobar"},
				"search_remote_url": []string{"git@github.com:a/b"},
			},
			wantRedirectURL: "/search?patternType=literal&q=repo%3Agithub%5C.com%2Fa%2Fb%24+foobar",
		},
		{
			name: "search in repository branch",
			q: url.Values{
				"search":            []string{"foobar"},
				"search_remote_url": []string{"git@github.com:a/b"},
				"search_branch":     []string{"dev"},
			},
			wantRedirectURL: "/search?patternType=literal&q=repo%3Agithub%5C.com%2Fa%2Fb%24%40dev+foobar",
		},
		{
			name: "search in repository revision",
			q: url.Values{
				"search":            []string{"foobar"},
				"search_remote_url": []string{"git@github.com:a/b"},
				"search_branch":     []string{"dev"},
				"search_revision":   []string{"0ad12f"},
			},
			wantRedirectURL: "/search?patternType=literal&q=repo%3Agithub%5C.com%2Fa%2Fb%24%400ad12f+foobar",
		},
		{
			name: "search in repository with generic code host (with repositoryPathPattern)",
			q: url.Values{
				"editor":            []string{"Atom"},
				"version":           []string{"v1.2.1"},
				"search":            []string{"foobar"},
				"search_remote_url": []string{"https://somecodehost.com/bar/a/b"},
			},
			wantRedirectURL: "/search?patternType=literal&q=repo%3Apretty%2Fa%2Fb%24+foobar",
		},
		{
			name: "search in repository file",
			q: url.Values{
				"search":            []string{"foobar"},
				"search_remote_url": []string{"git@github.com:a/b"},
				"search_file":       []string{"baz"},
			},
			wantRedirectURL: "/search?patternType=literal&q=repo%3Agithub%5C.com%2Fa%2Fb%24+file%3A%5Ebaz%24+foobar",
		},
		{
			name: "search in file",
			q: url.Values{
				"search":      []string{"foobar"},
				"search_file": []string{"baz"},
			},
			wantRedirectURL: "/search?patternType=literal&q=file%3A%5Ebaz%24+foobar",
		},
		{
			name:         "empty request",
			wantParseErr: "could not determine query string",
		},
		{
			name:            "unknown request",
			q:               url.Values{},
			wantRedirectErr: "could not determine request type, missing ?search or ?remote_url",
		},
		{
			name: "editor and version is optional",
			q: url.Values{
				"editor":     []string{"Atom"},
				"version":    []string{"v1.2.1"},
				"remote_url": []string{"git@github.com:a/b"},
				"branch":     []string{"dev"},
				"revision":   []string{"0ad12f"},
				"file":       []string{"mux.go"},
				"start_row":  []string{"123"},
				"start_col":  []string{"1"},
				"end_row":    []string{"123"},
				"end_col":    []string{"10"},
			},
			wantRedirectURL: "/github.com/a/b@0ad12f/-/blob/mux.go?L124%3A2-124%3A11",
		},
	}
	errStr := func(e error) string {
		if e == nil {
			return ""
		}
		return e.Error()
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			editorRequest, parseErr := parseEditorRequest(db, c.q)
			if errStr(parseErr) != c.wantParseErr {
				t.Fatalf("got parseErr %q want %q", parseErr, c.wantParseErr)
			}
			if parseErr == nil {
				redirectURL, redirectErr := editorRequest.redirectURL(context.TODO())
				if errStr(redirectErr) != c.wantRedirectErr {
					t.Fatalf("got redirectErr %q want %q", redirectErr, c.wantRedirectErr)
				}
				if redirectURL != c.wantRedirectURL {
					t.Fatalf("got redirectURL %q want %q", redirectURL, c.wantRedirectURL)
				}
			}
		})
	}
}
