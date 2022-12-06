package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGithubSource_GetRepo(t *testing.T) {
	testCases := []struct {
		name          string
		nameWithOwner string
		assert        func(*testing.T, *types.Repo)
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
			assert: func(t *testing.T, have *types.Repo) {
				t.Helper()

				want := &types.Repo{
					Name:        "github.com/sourcegraph/sourcegraph",
					Description: "Code search and navigation tool (self-hosted)",
					URI:         "github.com/sourcegraph/sourcegraph",
					Stars:       2220,
					ExternalRepo: api.ExternalRepoSpec{
						ID:          "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==",
						ServiceType: "github",
						ServiceID:   "https://github.com/",
					},
					Sources: map[string]*types.SourceInfo{
						"extsvc:github:0": {
							ID:       "extsvc:github:0",
							CloneURL: "https://github.com/sourcegraph/sourcegraph",
						},
					},
					Metadata: &github.Repository{
						ID:             "MDEwOlJlcG9zaXRvcnk0MTI4ODcwOA==",
						DatabaseID:     41288708,
						NameWithOwner:  "sourcegraph/sourcegraph",
						Description:    "Code search and navigation tool (self-hosted)",
						URL:            "https://github.com/sourcegraph/sourcegraph",
						StargazerCount: 2220,
						ForkCount:      164,
						// We're hitting github.com here, so visibility will be empty irrespective
						// of repository type. This is a GitHub enterprise only feature.
						Visibility: "",
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
			setUpRcache(t)

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			svc := &types.ExternalService{
				Kind: extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.GitHubConnection{
					Url: "https://github.com",
				})),
			}

			ctx := context.Background()
			githubSrc, err := NewGithubSource(ctx, logtest.Scoped(t), database.NewMockExternalServiceStore(), svc, cf)
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

func setUpRcache(t *testing.T) {
	// The GitHubSource uses the github.Client under the hood, which
	// uses rcache, a caching layer that uses Redis.
	// We need to clear the cache before we run the tests
	rcache.SetupForTest(t)
}

func TestGithubSource_GetRepo_Enterprise(t *testing.T) {
	testCases := []struct {
		name          string
		nameWithOwner string
		assert        func(*testing.T, *types.Repo)
		err           string
	}{
		{
			name:          "internal repo in github enterprise",
			nameWithOwner: "admiring-austin-120/fluffy-enigma",
			assert: func(t *testing.T, have *types.Repo) {
				t.Helper()

				want := &types.Repo{
					Name:        "ghe.sgdev.org/admiring-austin-120/fluffy-enigma",
					Description: "Internal repo used in tests in sourcegraph code.",
					URI:         "ghe.sgdev.org/admiring-austin-120/fluffy-enigma",
					Stars:       0,
					Private:     true,
					ExternalRepo: api.ExternalRepoSpec{
						ID:          "MDEwOlJlcG9zaXRvcnk0NDIyODU=",
						ServiceType: "github",
						ServiceID:   "https://ghe.sgdev.org/",
					},
					Sources: map[string]*types.SourceInfo{
						"extsvc:github:0": {
							ID:       "extsvc:github:0",
							CloneURL: "https://ghe.sgdev.org/admiring-austin-120/fluffy-enigma",
						},
					},
					Metadata: &github.Repository{
						ID:             "MDEwOlJlcG9zaXRvcnk0NDIyODU=",
						DatabaseID:     442285,
						NameWithOwner:  "admiring-austin-120/fluffy-enigma",
						Description:    "Internal repo used in tests in sourcegraph code.",
						URL:            "https://ghe.sgdev.org/admiring-austin-120/fluffy-enigma",
						StargazerCount: 0,
						ForkCount:      0,
						IsPrivate:      true,
						Visibility:     github.VisibilityInternal,
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
		tc.name = "GITHUB-ENTERPRISE/" + tc.name

		t.Run(tc.name, func(t *testing.T) {
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					ExperimentalFeatures: &schema.ExperimentalFeatures{
						EnableGithubInternalRepoVisibility: true,
					},
				},
			})

			setUpRcache(t)
			fixtureName := "githubenterprise-getrepo"
			gheToken := os.Getenv("GHE_TOKEN")
			fmt.Println(gheToken)

			if update(fixtureName) && gheToken == "" {
				t.Fatalf("GHE_TOKEN needs to be set to a token that can access ghe.sgdev.org to update this test fixture")
			}

			svc := &types.ExternalService{
				Kind: extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.GitHubConnection{
					Url:   "https://ghe.sgdev.org",
					Token: gheToken,
				})),
			}

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			ctx := context.Background()
			githubSrc, err := NewGithubSource(ctx, logtest.Scoped(t), database.NewMockExternalServiceStore(), svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			repo, err := githubSrc.GetRepo(context.Background(), tc.nameWithOwner)
			if err != nil {
				t.Fatalf("GetRepo failed: %v", err)
			}

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if tc.assert != nil {
				tc.assert(t, repo)
			}
		})
	}
}

func TestGithubSource_makeRepo(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "github-repos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var repos []*github.Repository
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	svc := types.ExternalService{
		ID:     1,
		Kind:   extsvc.KindGitHub,
		Config: extsvc.NewEmptyConfig(),
	}

	tests := []struct {
		name   string
		schema *schema.GitHubConnection
	}{
		{
			name: "simple",
			schema: &schema.GitHubConnection{
				Url: "https://github.com",
			},
		}, {
			name: "ssh",
			schema: &schema.GitHubConnection{
				Url:        "https://github.com",
				GitURLType: "ssh",
			},
		}, {
			name: "path-pattern",
			schema: &schema.GitHubConnection{
				Url:                   "https://github.com",
				RepositoryPathPattern: "gh/{nameWithOwner}",
			},
		}, {
			name: "name-with-owner",
			schema: &schema.GitHubConnection{
				Url:                   "https://github.com",
				RepositoryPathPattern: "{nameWithOwner}",
			},
		},
	}
	for _, test := range tests {
		test.name = "GithubSource_makeRepo_" + test.name
		t.Run(test.name, func(t *testing.T) {

			s, err := newGithubSource(logtest.Scoped(t), database.NewMockExternalServiceStore(), &svc, test.schema, nil)
			if err != nil {
				t.Fatal(err)
			}

			var got []*types.Repo
			for _, r := range repos {
				got = append(got, s.makeRepo(r))
			}

			testutil.AssertGolden(t, "testdata/golden/"+test.name, update(test.name), got)
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
	assertAllReposListed := func(want []string) typestest.ReposAssertion {
		return func(t testing.TB, rs types.Repos) {
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
		assert typestest.ReposAssertion
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
			setUpRcache(t)

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

			svc := &types.ExternalService{
				Kind:   extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, tc.conf)),
			}

			ctx := context.Background()
			githubSrc, err := NewGithubSource(ctx, logtest.Scoped(t), database.NewMockExternalServiceStore(), svc, cf)
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

func TestGithubSource_WithAuthenticator(t *testing.T) {
	svc := &types.ExternalService{
		Kind: extsvc.KindGitHub,
		Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.GitHubConnection{
			Url:   "https://github.com",
			Token: os.Getenv("GITHUB_TOKEN"),
		})),
	}

	ctx := context.Background()
	githubSrc, err := NewGithubSource(ctx, logtest.Scoped(t), database.NewMockExternalServiceStore(), svc, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("supported", func(t *testing.T) {
		src, err := githubSrc.WithAuthenticator(&auth.OAuthBearerToken{})
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if gs, ok := src.(*GitHubSource); !ok {
			t.Error("cannot coerce Source into GitHubSource")
		} else if gs == nil {
			t.Error("unexpected nil Source")
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		for name, tc := range map[string]auth.Authenticator{
			"nil":         nil,
			"BasicAuth":   &auth.BasicAuth{},
			"OAuthClient": &auth.OAuthClient{},
		} {
			t.Run(name, func(t *testing.T) {
				src, err := githubSrc.WithAuthenticator(tc)
				if err == nil {
					t.Error("unexpected nil error")
				} else if !errors.HasType(err, UnsupportedAuthenticatorError{}) {
					t.Errorf("unexpected error of type %T: %v", err, err)
				}
				if src != nil {
					t.Errorf("expected non-nil Source: %v", src)
				}
			})
		}
	})
}

func TestGithubSource_excludes_disabledAndLocked(t *testing.T) {
	svc := &types.ExternalService{
		Kind: extsvc.KindGitHub,
		Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.GitHubConnection{
			Url:   "https://github.com",
			Token: os.Getenv("GITHUB_TOKEN"),
		})),
	}

	ctx := context.Background()
	githubSrc, err := NewGithubSource(ctx, logtest.Scoped(t), database.NewMockExternalServiceStore(), svc, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range []*github.Repository{
		{IsDisabled: true},
		{IsLocked: true},
		{IsDisabled: true, IsLocked: true},
	} {
		if !githubSrc.excludes(r) {
			t.Errorf("GitHubSource should exclude %+v", r)
		}
	}
}

func TestGithubSource_GetVersion(t *testing.T) {
	logger := logtest.Scoped(t)
	t.Run("github.com", func(t *testing.T) {
		svc := &types.ExternalService{
			Kind: extsvc.KindGitHub,
			Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.GitHubConnection{
				Url: "https://github.com",
			})),
		}

		ctx := context.Background()
		githubSrc, err := NewGithubSource(ctx, logger, database.NewMockExternalServiceStore(), svc, nil)
		if err != nil {
			t.Fatal(err)
		}

		have, err := githubSrc.Version(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if want := "unknown"; have != want {
			t.Fatalf("wrong version returned. want=%s, have=%s", want, have)
		}
	})

	t.Run("github enterprise", func(t *testing.T) {
		setUpRcache(t)

		fixtureName := "githubenterprise-version"
		gheToken := os.Getenv("GHE_TOKEN")
		if update(fixtureName) && gheToken == "" {
			t.Fatalf("GHE_TOKEN needs to be set to a token that can access ghe.sgdev.org to update this test fixture")
		}

		cf, save := newClientFactory(t, fixtureName)
		defer save(t)

		svc := &types.ExternalService{
			Kind: extsvc.KindGitHub,
			Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.GitHubConnection{
				Url:   "https://ghe.sgdev.org",
				Token: gheToken,
			})),
		}

		ctx := context.Background()
		githubSrc, err := NewGithubSource(ctx, logger, database.NewMockExternalServiceStore(), svc, cf)
		if err != nil {
			t.Fatal(err)
		}

		have, err := githubSrc.Version(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if want := "2.22.6"; have != want {
			t.Fatalf("wrong version returned. want=%s, have=%s", want, have)
		}
	})
}

func TestRepositoryQuery_Do(t *testing.T) {
	for _, tc := range []struct {
		name  string
		query string
		first int
		limit int
		now   time.Time
	}{
		{
			name:  "exceeds-limit",
			query: "stars:10000..10100",
			first: 10,
			limit: 20, // We simulate a lower limit that the 1000 limit on github.com
		},
		{
			name:  "doesnt-exceed-limit",
			query: "repo:tsenart/vegeta stars:>=14000",
			first: 10,
			limit: 20,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cf, save := httptestutil.NewGitHubRecorderFactory(t, update(t.Name()), t.Name())
			t.Cleanup(save)

			cli, err := cf.Doer()
			if err != nil {
				t.Fatal(err)
			}

			apiURL, _ := url.Parse("https://api.github.com")
			token := &auth.OAuthBearerToken{Token: os.Getenv("GITHUB_TOKEN")}

			q := repositoryQuery{
				Logger:   logtest.Scoped(t),
				Query:    tc.query,
				First:    tc.first,
				Limit:    tc.limit,
				Searcher: github.NewV4Client("Test", apiURL, token, cli),
			}

			results := make(chan *githubResult)
			go func() {
				q.Do(context.Background(), results)
				close(results)
			}()

			type result struct {
				Repo  *github.Repository
				Error string
			}

			var have []result
			for r := range results {
				res := result{Repo: r.repo}
				if r.err != nil {
					res.Error = r.err.Error()
				}
				have = append(have, res)
			}

			testutil.AssertGolden(t, "testdata/golden/"+t.Name(), update(t.Name()), have)
		})
	}
}

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}
