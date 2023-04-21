package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
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

func mustParse(t *testing.T, dateStr string) time.Time {
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		date, err = time.Parse("2006-01-02T15:04:05", dateStr)
		if err != nil {
			date, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				t.Fatal("Failed to parse date from", dateStr)
			}
		}
	}
	return date
}

func TestGitHub_stripDateRange(t *testing.T) {
	testCases := map[string]struct {
		query         string
		wantQuery     string
		wantDateRange *dateRange
	}{
		"from and to with ..": {
			query:     "some part of query created:2008-11-10T01:23:45+00:00..2010-01-30T23:45:59+02:00 and others",
			wantQuery: "some part of query  and others",
			wantDateRange: &dateRange{
				From: mustParse(t, "2008-11-10T01:23:45+00:00"),
				To:   mustParse(t, "2010-01-30T23:45:59+02:00"),
			},
		},
		"from with >": {
			query: "created:>2011-01-01T00:00:00+00:00 and other stuff",
			wantDateRange: &dateRange{
				From: mustParse(t, "2011-01-01T00:00:01+00:00"),
			},
		},
		"from with >=": {
			query: "created:>=2011-01-01T00:00:00+00:00 and other stuff",
			wantDateRange: &dateRange{
				From: mustParse(t, "2011-01-01T00:00:00+00:00"),
			},
		},
		"from with ..*": {
			query: "created:2010-01-01..*",
			wantDateRange: &dateRange{
				From: mustParse(t, "2010-01-01T00:00:00+00:00"),
			},
		},
		"to with <": {
			query: "created:<2015-12-12",
			wantDateRange: &dateRange{
				To: mustParse(t, "2015-12-11T23:59:59+00:00"),
			},
		},
		"to with <=": {
			query: "created:<=2015-12-12",
			wantDateRange: &dateRange{
				To: mustParse(t, "2015-12-12T23:59:59+00:00"),
			},
		},
		"to with *..": {
			query:     "created:*..2015-12-12",
			wantQuery: "",
			wantDateRange: &dateRange{
				To: mustParse(t, "2015-12-12T23:59:59"),
			},
		},
		"no date query": {
			query:         "just some random things",
			wantQuery:     "just some random things",
			wantDateRange: nil,
		},
	}

	for tname, tcase := range testCases {
		t.Run(tname, func(t *testing.T) {
			date := stripDateRange(&tcase.query)
			if tcase.wantDateRange == nil {
				assert.Nil(t, date)
			} else {
				assert.True(t, date.From.Equal(tcase.wantDateRange.From), "got %q want %q", date.From, tcase.wantDateRange.From)
				assert.True(t, date.To.Equal(tcase.wantDateRange.To), "got %q want %q", date.To, tcase.wantDateRange.To)
			}
			if tcase.wantQuery != "" {
				assert.Equal(t, tcase.wantQuery, tcase.query)
			}
		})
	}
}

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
						Visibility:       "",
						RepositoryTopics: github.RepositoryTopics{Nodes: []github.RepositoryTopic{}},
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
			githubSrc, err := NewGithubSource(ctx, logtest.Scoped(t), svc, cf)
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
						ID:               "MDEwOlJlcG9zaXRvcnk0NDIyODU=",
						DatabaseID:       442285,
						NameWithOwner:    "admiring-austin-120/fluffy-enigma",
						Description:      "Internal repo used in tests in sourcegraph code.",
						URL:              "https://ghe.sgdev.org/admiring-austin-120/fluffy-enigma",
						StargazerCount:   0,
						ForkCount:        0,
						IsPrivate:        true,
						Visibility:       github.VisibilityInternal,
						RepositoryTopics: github.RepositoryTopics{Nodes: []github.RepositoryTopic{{Topic: github.Topic{Name: "fluff"}}}},
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
			githubSrc, err := NewGithubSource(ctx, logtest.Scoped(t), svc, cf)
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

func TestMakeRepo_NullCharacter(t *testing.T) {
	r := &github.Repository{
		Description: "Fun nulls \x00\x00\x00",
	}

	svc := types.ExternalService{
		ID:     1,
		Kind:   extsvc.KindGitHub,
		Config: extsvc.NewEmptyConfig(),
	}
	schema := &schema.GitHubConnection{
		Url: "https://github.com",
	}
	s, err := newGithubSource(logtest.Scoped(t), &svc, schema, nil)
	require.NoError(t, err)
	repo := s.makeRepo(r)

	require.Equal(t, "Fun nulls ", repo.Description)
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
			s, err := newGithubSource(logtest.Scoped(t), &svc, test.schema, nil)
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

func TestGitHubSource_doRecursively(t *testing.T) {
	rcache.SetupForTest(t)
	ctx := context.Background()

	testCases := map[string]struct {
		requestsBeforeFullSet int // Number of requests before all repositories are returned
		expectedRepoCount     int
	}{
		"retries until full list of repositories": {
			requestsBeforeFullSet: 2,
			expectedRepoCount:     5,
		},
		"retries a limited amount of times": {
			requestsBeforeFullSet: 50,
			expectedRepoCount:     4,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			requestCounter := 0
			// We create a server that returns a repository count of 5, but only returns 4 repositories.
			// After the server has been hit two times, a fifth repository is added to the result set.
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					requestCounter += 1
				}()

				resp := struct {
					Data struct {
						Search struct {
							RepositoryCount int
							PageInfo        struct {
								HasNextPage bool
								EndCursor   github.Cursor
							}
							Nodes []github.Repository
						}
					}
				}{}

				resp.Data.Search.RepositoryCount = 5
				resp.Data.Search.Nodes = []github.Repository{
					{DatabaseID: 1}, {DatabaseID: 2}, {DatabaseID: 3}, {DatabaseID: 4},
				}

				if requestCounter >= tc.requestsBeforeFullSet {
					resp.Data.Search.Nodes = append(resp.Data.Search.Nodes, github.Repository{DatabaseID: 5})
				}

				encoder := json.NewEncoder(w)
				require.NoError(t, encoder.Encode(resp))
			}))
			defer srv.Close()

			apiURL, err := url.Parse(srv.URL)
			require.NoError(t, err)
			ghCli := github.NewV4Client("", apiURL, nil, nil)
			q := newRepositoryQuery("stars:>=5", ghCli, logtest.NoOp(t))
			q.Limit = 5

			// Fetch the repositories
			results := make(chan *githubResult)
			go func() {
				q.doRecursively(ctx, results)
				close(results)
			}()

			repos := []github.Repository{}
			for res := range results {
				repos = append(repos, *res.repo)
			}

			// Confirm that we received 5 repositories, confirming that we retried the request.
			assert.Len(t, repos, tc.expectedRepoCount)
		})
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
			githubSrc, err := NewGithubSource(ctx, logtest.Scoped(t), svc, cf)
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
	githubSrc, err := NewGithubSource(ctx, logtest.Scoped(t), svc, nil)
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
	githubSrc, err := NewGithubSource(ctx, logtest.Scoped(t), svc, nil)
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
		githubSrc, err := NewGithubSource(ctx, logger, svc, nil)
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
		githubSrc, err := NewGithubSource(ctx, logger, svc, cf)
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

func TestRepositoryQuery_DoWithRefinedWindow(t *testing.T) {
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
			limit: 20, // We simulate a lower limit than the 1000 limit on github.com
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
				q.DoWithRefinedWindow(context.Background(), results)
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

func TestRepositoryQuery_DoSingleRequest(t *testing.T) {
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
			limit: 20, // We simulate a lower limit than the 1000 limit on github.com
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
				q.DoSingleRequest(context.Background(), results)
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

func TestGithubSource_SearchRepositories(t *testing.T) {
	assertReposSearched := func(want []string) typestest.ReposAssertion {
		return func(t testing.TB, rs types.Repos) {
			t.Helper()

			have := rs.Names()
			sort.Strings(have)
			sort.Strings(want)

			if diff := cmp.Diff(want, have); diff != "" {
				t.Error(diff)
			}
		}
	}

	testCases := []struct {
		name         string
		query        string
		first        int
		excludeRepos []string
		assert       typestest.ReposAssertion
		mw           httpcli.Middleware
		conf         *schema.GitHubConnection
		err          string
	}{
		{
			name:         "query string found",
			query:        "sourcegraph sourcegraph",
			first:        5,
			excludeRepos: []string{},
			assert: assertReposSearched([]string{
				"github.com/sourcegraph/about",
				"github.com/sourcegraph/sourcegraph",
				"github.com/sourcegraph/src-cli",
				"github.com/sourcegraph/deploy-sourcegraph-docker",
				"github.com/sourcegraph/deploy-sourcegraph",
			}),
			conf: &schema.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{},
			},
			err: "<nil>",
		},
		{
			name:         "query string found reduced first",
			query:        "sourcegraph sourcegraph",
			first:        1,
			excludeRepos: []string{},
			assert: assertReposSearched([]string{
				"github.com/sourcegraph/sourcegraph",
			}),
			conf: &schema.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{},
			},
			err: "<nil>",
		},
		{
			name:         "query string empty results",
			query:        "horsegraph",
			first:        5,
			excludeRepos: []string{},
			assert:       assertReposSearched([]string{}),
			conf: &schema.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{},
			},
			err: "<nil>",
		},
		{
			name:         "query string exclude one positive match",
			query:        "sourcegraph sourcegraph",
			first:        5,
			excludeRepos: []string{"sourcegraph/about"},
			assert: assertReposSearched([]string{
				"github.com/sourcegraph/sourcegraph",
				"github.com/sourcegraph/src-cli",
				"github.com/sourcegraph/deploy-sourcegraph-docker",
				"github.com/sourcegraph/deploy-sourcegraph",
				"github.com/sourcegraph/handbook",
			}),
			conf: &schema.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{},
			},
			err: "<nil>",
		},
		{
			name:         "empty query string found",
			query:        "",
			first:        5,
			excludeRepos: []string{},
			assert: assertReposSearched([]string{
				"github.com/sourcegraph/vulnerable-js-test",
				"github.com/sourcegraph/scip-excel",
				"github.com/sourcegraph/controller-cdktf",
				"github.com/sourcegraph/deploy-sourcegraph-k8s",
				"github.com/sourcegraph/embedded-postgres",
			}),
			conf: &schema.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{},
			},
			err: "<nil>",
		},
		{
			name:         "empty query string found reduced first",
			query:        "",
			first:        1,
			excludeRepos: []string{},
			assert: assertReposSearched([]string{
				"github.com/sourcegraph/vulnerable-js-test",
			}),
			conf: &schema.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{},
			},
			err: "<nil>",
		},
		{
			name:  "empty query string exclude two positive match",
			query: "",
			first: 5,
			excludeRepos: []string{
				"sourcegraph/vulnerable-js-test",
				"sourcegraph/scip-excel",
			},
			assert: assertReposSearched([]string{
				"github.com/sourcegraph/controller-cdktf",
				"github.com/sourcegraph/deploy-sourcegraph-k8s",
				"github.com/sourcegraph/embedded-postgres",
				"github.com/sourcegraph/deploy-sourcegraph-docker-customer-replica-1",
				"github.com/sourcegraph/tf-dag",
			}),
			conf: &schema.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{},
			},
			err: "<nil>",
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "GITHUB-SEARCH-REPOS/" + tc.name
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
			githubSrc, err := NewGithubSource(ctx, logtest.Scoped(t), svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			repos, err := searchRepositories(context.Background(), githubSrc, tc.query, tc.first, tc.excludeRepos)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if tc.assert != nil {
				tc.assert(t, repos)
			}
		})
	}
}

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}
