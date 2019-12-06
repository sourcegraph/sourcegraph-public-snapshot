package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func TestExampleRepositoryQuerySplit(t *testing.T) {
	q := "org:sourcegraph"
	want := `["org:sourcegraph created:>=2019","org:sourcegraph created:2018","org:sourcegraph created:2016..2017","org:sourcegraph created:<2016"]`
	have := exampleRepositoryQuerySplit(q)
	if want != have {
		t.Errorf("unexpected example query for %s:\nwant: %s\nhave: %s", q, want, have)
	}
}

func TestGithubSource_CreateChangeset(t *testing.T) {
	repo := &Repo{
		Metadata: &github.Repository{
			ID:            "MDEwOlJlcG9zaXRvcnkyMjExNDc1MTM=",
			NameWithOwner: "sourcegraph/automation-testing",
		},
	}

	testCases := []struct {
		name string
		cs   *Changeset
		err  string
	}{
		{
			name: "success",
			cs: &Changeset{
				Title:     "This is a test PR",
				Body:      "This is the description of the test PR",
				HeadRef:   "refs/heads/test-pr-6",
				BaseRef:   "refs/heads/master",
				Repo:      repo,
				Changeset: &a8n.Changeset{},
			},
		},
		{
			name: "already exists",
			cs: &Changeset{
				Title:     "This is a test PR",
				Body:      "This is the description of the test PR",
				HeadRef:   "heads/refs/always-open-pr",
				BaseRef:   "heads/refs/master",
				Repo:      repo,
				Changeset: &a8n.Changeset{},
			},
			err: github.ErrPullRequestAlreadyExists.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "GithubSource_CreateChangeset_" + tc.name

		t.Run(tc.name, func(t *testing.T) {
			// The GithubSource uses the github.Client under the hood, which
			// uses rcache, a caching layer that uses Redis.
			// We need to clear the cache before we run the tests
			rcache.SetupForTest(t)

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &ExternalService{
				Kind: "GITHUB",
				Config: marshalJSON(t, &schema.GitHubConnection{
					Url:   "https://github.com",
					Token: os.Getenv("GITHUB_TOKEN"),
				}),
			}

			githubSrc, err := NewGithubSource(svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()
			if tc.err == "" {
				tc.err = "<nil>"
			}

			err = githubSrc.CreateChangeset(ctx, tc.cs)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Changeset.Metadata.(*github.PullRequest)
			data, err := json.MarshalIndent(pr, " ", " ")
			if err != nil {
				t.Fatal(err)
			}

			path := "testdata/golden/" + tc.name
			if update(tc.name) {
				if err = ioutil.WriteFile(path, data, 0640); err != nil {
					t.Fatalf("failed to update golden file %q: %s", path, err)
				}
			}

			golden, err := ioutil.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read golden file %q: %s", path, err)
			}

			if have, want := string(data), string(golden); have != want {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(have, want, false)
				t.Error(dmp.DiffPrettyText(diffs))
			}
		})
	}
}

func TestGithubSource_CloseChangeset(t *testing.T) {
	testCases := []struct {
		name string
		cs   *Changeset
		err  string
	}{
		{
			name: "success",
			cs: &Changeset{
				Changeset: &a8n.Changeset{
					Metadata: &github.PullRequest{
						ID: "MDExOlB1bGxSZXF1ZXN0MzQ5NTIzMzE0",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "GithubSource_CloseChangeset_" + strings.Replace(tc.name, " ", "_", -1)

		t.Run(tc.name, func(t *testing.T) {
			// The GithubSource uses the github.Client under the hood, which
			// uses rcache, a caching layer that uses Redis.
			// We need to clear the cache before we run the tests
			rcache.SetupForTest(t)

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &ExternalService{
				Kind: "GITHUB",
				Config: marshalJSON(t, &schema.GitHubConnection{
					Url:   "https://github.com",
					Token: os.Getenv("GITHUB_TOKEN"),
				}),
			}

			githubSrc, err := NewGithubSource(svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()
			if tc.err == "" {
				tc.err = "<nil>"
			}

			err = githubSrc.CloseChangeset(ctx, tc.cs)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Changeset.Metadata.(*github.PullRequest)
			data, err := json.MarshalIndent(pr, " ", " ")
			if err != nil {
				t.Fatal(err)
			}

			path := "testdata/golden/" + tc.name
			if update(tc.name) {
				if err = ioutil.WriteFile(path, data, 0640); err != nil {
					t.Fatalf("failed to update golden file %q: %s", path, err)
				}
			}

			golden, err := ioutil.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read golden file %q: %s", path, err)
			}

			if have, want := string(data), string(golden); have != want {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(have, want, false)
				t.Error(dmp.DiffPrettyText(diffs))
			}
		})
	}
}

func TestGithubSource_LoadChangesets(t *testing.T) {
	testCases := []struct {
		name string
		cs   []*Changeset
		err  string
	}{
		{
			name: "found",
			cs: []*Changeset{
				{
					Repo:      &Repo{Metadata: &github.Repository{NameWithOwner: "sourcegraph/sourcegraph"}},
					Changeset: &a8n.Changeset{ExternalID: "5550"},
				},
				{
					Repo:      &Repo{Metadata: &github.Repository{NameWithOwner: "tsenart/vegeta"}},
					Changeset: &a8n.Changeset{ExternalID: "50"},
				},
				{
					Repo:      &Repo{Metadata: &github.Repository{NameWithOwner: "sourcegraph/sourcegraph"}},
					Changeset: &a8n.Changeset{ExternalID: "5834"},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "GithubSource_LoadChangesets_" + tc.name

		t.Run(tc.name, func(t *testing.T) {
			// The GithubSource uses the github.Client under the hood, which
			// uses rcache, a caching layer that uses Redis.
			// We need to clear the cache before we run the tests
			rcache.SetupForTest(t)

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &ExternalService{
				Kind: "GITHUB",
				Config: marshalJSON(t, &schema.GitHubConnection{
					Url:   "https://github.com",
					Token: os.Getenv("GITHUB_TOKEN"),
				}),
			}

			githubSrc, err := NewGithubSource(svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()
			if tc.err == "" {
				tc.err = "<nil>"
			}

			err = githubSrc.LoadChangesets(ctx, tc.cs...)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			meta := make([]*github.PullRequest, 0, len(tc.cs))
			for _, cs := range tc.cs {
				meta = append(meta, cs.Changeset.Metadata.(*github.PullRequest))
			}

			data, err := json.MarshalIndent(meta, " ", " ")
			if err != nil {
				t.Fatal(err)
			}

			path := "testdata/golden/" + tc.name
			if update(tc.name) {
				if err = ioutil.WriteFile(path, data, 0640); err != nil {
					t.Fatalf("failed to update golden file %q: %s", path, err)
				}
			}

			golden, err := ioutil.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read golden file %q: %s", path, err)
			}

			if have, want := string(data), string(golden); have != want {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(have, want, false)
				t.Error(dmp.DiffPrettyText(diffs))
			}
		})
	}
}

func TestGithubSource_GetRepo(t *testing.T) {
	testCases := []struct {
		name          string
		nameWithOwner string
		assert        func(*testing.T, *Repo)
		err           string
	}{
		{
			name:          "invalid name",
			nameWithOwner: "thisIsNotANameWithOwner",
			err:           `Invalid GitHub repository: nameWithOwner=thisIsNotANameWithOwner: invalid GitHub repository "owner/name" string: "thisIsNotANameWithOwner"`,
		},
		{
			name:          "not found",
			nameWithOwner: "foobarfoobarfoobar/please-let-this-not-exist",
			err:           `GitHub repository not found`,
		},
		{
			name:          "found",
			nameWithOwner: "sourcegraph/sourcegraph",
			assert: func(t *testing.T, have *Repo) {
				t.Helper()

				want := &Repo{
					Name:        "github.com/sourcegraph/sourcegraph",
					Description: "Code search and navigation tool (self-hosted)",
					Enabled:     true,
					URI:         "github.com/sourcegraph/sourcegraph",
					ExternalRepo: api.ExternalRepoSpec{
						ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==",
						ServiceType: "github",
						ServiceID:   "https://github.com/",
					},
					Sources: map[string]*SourceInfo{
						"extsvc:github:0": {
							ID:       "extsvc:github:0",
							CloneURL: "https://github.com/sourcegraph/sourcegraph",
						},
					},
					Metadata: &github.Repository{
						ID:            "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==",
						DatabaseID:    41288708,
						NameWithOwner: "sourcegraph/sourcegraph",
						Description:   "Code search and navigation tool (self-hosted)",
						URL:           "https://github.com/sourcegraph/sourcegraph",
					},
				}

				if !reflect.DeepEqual(have, want) {
					t.Errorf("response: %s", cmp.Diff(have, want))
				}
			},
			err: "<nil>",
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "GITHUB-DOT-COM/" + tc.name

		t.Run(tc.name, func(t *testing.T) {
			// The GithubSource uses the github.Client under the hood, which
			// uses rcache, a caching layer that uses Redis.
			// We need to clear the cache before we run the tests
			rcache.SetupForTest(t)

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &ExternalService{
				Kind: "GITHUB",
				Config: marshalJSON(t, &schema.GitHubConnection{
					Url: "https://github.com",
				}),
			}

			githubSrc, err := NewGithubSource(svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			repo, err := githubSrc.GetRepo(context.Background(), tc.nameWithOwner)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if tc.assert != nil {
				tc.assert(t, repo)
			}
		})
	}
}

func TestMatchOrg(t *testing.T) {
	testCases := map[string]string{
		"":                     "",
		"org:":                 "",
		"org:gorilla":          "gorilla",
		"org:golang-migrate":   "golang-migrate",
		"org:sourcegraph-":     "",
		"org:source--graph":    "",
		"org: sourcegraph":     "",
		"org:$ourcegr@ph":      "",
		"sourcegraph":          "",
		"org:-sourcegraph":     "",
		"org:source graph":     "",
		"org:source org:graph": "",
		"org:SOURCEGRAPH":      "SOURCEGRAPH",
		"org:Game-club-3d-game-birds-gameapp-makerCo":  "Game-club-3d-game-birds-gameapp-makerCo",
		"org:thisorgnameisfartoolongtomatchthisregexp": "",
	}

	for str, want := range testCases {
		if got := matchOrg(str); got != want {
			t.Errorf("error:\nhave: %s\nwant: %s", got, want)
		}
	}
}

func TestGithubSource_ListRepos(t *testing.T) {
	assertAllReposListed := func(want []string) ReposAssertion {
		return func(t testing.TB, rs Repos) {
			t.Helper()

			have := rs.Names()
			sort.Strings(have)
			sort.Strings(want)

			if !reflect.DeepEqual(have, want) {
				t.Error(cmp.Diff(have, want))
			}
		}
	}

	testCases := []struct {
		name   string
		assert ReposAssertion
		mw     httpcli.Middleware
		conf   *schema.GitHubConnection
		err    string
	}{
		{
			name: "found",
			assert: assertAllReposListed([]string{
				"github.com/sourcegraph/about",
				"github.com/sourcegraph/sourcegraph",
			}),
			conf: &schema.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{
					"sourcegraph/about",
					"sourcegraph/sourcegraph",
				},
			},
			err: "<nil>",
		},
		{
			name: "graphql fallback",
			mw:   githubGraphQLFailureMiddleware,
			assert: assertAllReposListed([]string{
				"github.com/sourcegraph/about",
				"github.com/sourcegraph/sourcegraph",
			}),
			conf: &schema.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{
					"sourcegraph/about",
					"sourcegraph/sourcegraph",
				},
			},
			err: "<nil>",
		},
		{
			name: "orgs",
			assert: assertAllReposListed([]string{
				"github.com/gorilla/websocket",
				"github.com/gorilla/handlers",
				"github.com/gorilla/mux",
				"github.com/gorilla/feeds",
				"github.com/gorilla/sessions",
				"github.com/gorilla/schema",
				"github.com/gorilla/csrf",
				"github.com/gorilla/rpc",
				"github.com/gorilla/pat",
				"github.com/gorilla/css",
				"github.com/gorilla/site",
				"github.com/gorilla/context",
				"github.com/gorilla/securecookie",
				"github.com/gorilla/http",
				"github.com/gorilla/reverse",
				"github.com/gorilla/muxy",
				"github.com/gorilla/i18n",
				"github.com/gorilla/template",
				"github.com/gorilla/.github",
			}),
			conf: &schema.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Orgs: []string{
					"gorilla",
				},
			},
			err: "<nil>",
		},
		{
			name: "orgs repository query",
			assert: assertAllReposListed([]string{
				"github.com/gorilla/websocket",
				"github.com/gorilla/handlers",
				"github.com/gorilla/mux",
				"github.com/gorilla/feeds",
				"github.com/gorilla/sessions",
				"github.com/gorilla/schema",
				"github.com/gorilla/csrf",
				"github.com/gorilla/rpc",
				"github.com/gorilla/pat",
				"github.com/gorilla/css",
				"github.com/gorilla/site",
				"github.com/gorilla/context",
				"github.com/gorilla/securecookie",
				"github.com/gorilla/http",
				"github.com/gorilla/reverse",
				"github.com/gorilla/muxy",
				"github.com/gorilla/i18n",
				"github.com/gorilla/template",
				"github.com/gorilla/.github",
				"github.com/golang-migrate/migrate",
				"github.com/torvalds/linux",
				"github.com/torvalds/uemacs",
				"github.com/torvalds/subsurface-for-dirk",
				"github.com/torvalds/libdc-for-dirk",
				"github.com/torvalds/test-tlb",
				"github.com/torvalds/pesconvert",
			}),
			conf: &schema.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				RepositoryQuery: []string{
					"org:gorilla",
					"org:golang-migrate",
					"org:torvalds",
				},
			},
			err: "<nil>",
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "GITHUB-LIST-REPOS/" + tc.name
		t.Run(tc.name, func(t *testing.T) {
			// The GithubSource uses the github.Client under the hood, which
			// uses rcache, a caching layer that uses Redis.
			// We need to clear the cache before we run the tests
			rcache.SetupForTest(t)

			var (
				cf   *httpcli.Factory
				save func(testing.TB)
			)
			if tc.mw != nil {
				cf, save = newClientFactory(t, tc.name, tc.mw)
			} else {
				cf, save = newClientFactory(t, tc.name)
			}

			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &ExternalService{
				Kind:   "GITHUB",
				Config: marshalJSON(t, tc.conf),
			}

			githubSrc, err := NewGithubSource(svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			repos, err := listAll(context.Background(), githubSrc)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if tc.assert != nil {
				tc.assert(t, repos)
			}
		})
	}
}

func githubGraphQLFailureMiddleware(cli httpcli.Doer) httpcli.Doer {
	return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "graphql") {
			return nil, errors.New("graphql request failed")
		}
		return cli.Do(req)
	})
}
