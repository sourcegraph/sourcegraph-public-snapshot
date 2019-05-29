package repos

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
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

func TestGithubSource_ListRepos(t *testing.T) {
	assertAllReposListed := func(t testing.TB, rs Repos) {
		t.Helper()

		want := []string{
			"github.com/sourcegraph/about",
			"github.com/sourcegraph/sourcegraph",
		}

		have := rs.Names()
		sort.Strings(have)

		if !reflect.DeepEqual(have, want) {
			t.Error(cmp.Diff(have, want))
		}
	}

	testCases := []struct {
		name   string
		assert ReposAssertion
		mw     httpcli.Middleware
		err    string
	}{
		{
			name:   "found",
			assert: assertAllReposListed,
			err:    "<nil>",
		},
		{
			name:   "graphql fallback",
			mw:     githubGraphQLFailureMiddleware,
			assert: assertAllReposListed,
			err:    "<nil>",
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
				Kind: "GITHUB",
				Config: marshalJSON(t, &schema.GitHubConnection{
					Url:   "https://github.com",
					Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
					Repos: []string{
						"sourcegraph/about",
						"sourcegraph/sourcegraph",
					},
				}),
			}

			githubSrc, err := NewGithubSource(svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			repos, err := githubSrc.ListRepos(context.Background())
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
