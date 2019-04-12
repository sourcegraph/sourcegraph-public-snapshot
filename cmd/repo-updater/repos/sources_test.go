package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestOtherRepoName(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		out  string
	}{
		{"user and password elided", "https://user:pass@foo.bar/baz", "foo.bar/baz"},
		{"scheme elided", "https://user@foo.bar/baz", "foo.bar/baz"},
		{"raw query elided", "https://foo.bar/baz?secret_token=12345", "foo.bar/baz"},
		{"fragment elided", "https://foo.bar/baz#fragment", "foo.bar/baz"},
		{": replaced with -", "git://foo.bar/baz:bam", "foo.bar/baz-bam"},
		{"@ replaced with -", "ssh://foo.bar/baz@bam", "foo.bar/baz-bam"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cloneURL, err := url.Parse(tc.in)
			if err != nil {
				t.Fatal(err)
			}

			if have, want := otherRepoName(cloneURL), tc.out; have != want {
				t.Errorf("otherRepoName(%q):\nhave: %q\nwant: %q", tc.in, have, want)
			}
		})
	}
}

func TestNewSourcer(t *testing.T) {
	now := time.Now()

	github := ExternalService{
		Kind:        "GITHUB",
		DisplayName: "Github - Test",
		Config:      `{"url": "https://github.com"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	githubDotCom := ExternalService{Kind: "GITHUB"}

	gitlab := ExternalService{
		Kind:        "GITHUB",
		DisplayName: "Github - Test",
		Config:      `{"url": "https://github.com"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   now,
	}

	sources := func(es ...*ExternalService) (srcs []Source) {
		t.Helper()

		for _, e := range es {
			src, err := NewSource(e, nil)
			if err != nil {
				t.Fatal(err)
			}
			srcs = append(srcs, src)
		}

		return srcs
	}

	for _, tc := range []struct {
		name string
		svcs ExternalServices
		srcs Sources
		err  string
	}{
		{
			name: "deleted external services are excluded",
			svcs: ExternalServices{&github, &gitlab},
			srcs: sources(&github),
			err:  "<nil>",
		},
		{
			name: "github.com is added when not existent",
			svcs: ExternalServices{},
			srcs: sources(&githubDotCom),
			err:  "<nil>",
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			srcs, err := NewSourcer(nil)(tc.svcs...)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			have := srcs.ExternalServices()
			want := tc.srcs.ExternalServices()

			if !reflect.DeepEqual(have, want) {
				t.Errorf("sources:\n%s", cmp.Diff(have, want))
			}
		})
	}
}

func TestSources_ListRepos(t *testing.T) {

	// NOTE(tsenart): Updating this test's testdata with the `-update` flag requires
	// setting up a local instance of Bitbucket Server and populating it.
	// $ docker run \
	//    -v ~/.bitbucket/:/var/atlassian/application-data/bitbucket \
	//    --name="bitbucket"\
	//    -d -p 7990:7990 -p 7999:7999 \
	//    atlassian/bitbucket-server

	type testCase struct {
		name   string
		ctx    context.Context
		svcs   ExternalServices
		assert ReposAssertion
		err    string
	}

	var testCases []testCase
	{
		svcs := ExternalServices{
			{
				Kind: "GITHUB",
				Config: marshalJSON(t, &schema.GitHubConnection{
					Url:   "https://github.com",
					Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
					RepositoryQuery: []string{
						"user:tsenart in:name patrol",
					},
					Repos: []string{"sourcegraph/sourcegraph"},
				}),
			},
			{
				Kind: "GITLAB",
				Config: marshalJSON(t, &schema.GitLabConnection{
					Url:   "https://gitlab.com",
					Token: os.Getenv("GITLAB_ACCESS_TOKEN"),
					ProjectQuery: []string{
						"?search=gokulkarthick",
					},
				}),
			},
			{
				Kind: "BITBUCKETSERVER",
				Config: marshalJSON(t, &schema.BitbucketServerConnection{
					Url:   "http://127.0.0.1:7990",
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
					RepositoryQuery: []string{
						"?visibility=private",
						"?visibility=public",
					},
				}),
			},
			{
				Kind: "OTHER",
				Config: marshalJSON(t, &schema.OtherExternalServiceConnection{
					Url: "https://github.com",
					Repos: []string{
						"google/go-cmp",
					},
				}),
			},
		}

		kinds := make(map[string]struct{})
		for _, s := range svcs {
			kinds[strings.ToLower(s.Kind)] = struct{}{}
		}

		testCases = append(testCases, testCase{
			name: "yielded repos are always enabled",
			svcs: svcs,
			assert: func(t testing.TB, rs Repos) {
				t.Helper()

				set := make(map[string]bool)

				for _, r := range rs {
					set[strings.ToLower(r.ExternalRepo.ServiceType)] = true
					if !r.Enabled {
						t.Errorf("repo %q is not enabled", r.Name)
					}
				}

				for kind := range kinds {
					if !set[kind] {
						t.Errorf("external service of kind %q didn't yield any repos", kind)
					}
				}
			},
			err: "<nil>",
		})
	}

	{
		svcs := ExternalServices{
			{
				Kind: "GITHUB",
				Config: marshalJSON(t, &schema.GitHubConnection{
					Url:   "https://github.com",
					Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
					RepositoryQuery: []string{
						"user:tsenart in:name patrol",
					},
					Repos: []string{
						"sourcegraph/Sourcegraph",
						"tsenart/VEGETA",
					},
					Exclude: []*schema.ExcludedGitHubRepo{
						{Name: "tsenart/Vegeta"},
						{Id: "MDEwOlJlcG9zaXRvcnkxNTM2NTcyNDU="}, // tsenart/patrol ID
					},
				}),
			},
			{
				Kind: "GITLAB",
				Config: marshalJSON(t, &schema.GitLabConnection{
					Url:   "https://gitlab.com",
					Token: os.Getenv("GITLAB_ACCESS_TOKEN"),
					ProjectQuery: []string{
						"?search=gokulkarthick",
						"?search=dotfiles-vegetableman",
					},
					Exclude: []*schema.ExcludedGitLabProject{
						{Name: "gokulkarthick/gokulkarthick"},
						{Id: 7789240},
					},
				}),
			},
			{
				Kind: "BITBUCKETSERVER",
				Config: marshalJSON(t, &schema.BitbucketServerConnection{
					Url:   "http://127.0.0.1:7990",
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
					Repos: []string{
						"ORG/foo",
						"org/BAZ",
					},
					RepositoryQuery: []string{
						"?visibility=private",
					},
					Exclude: []*schema.ExcludedBitbucketServerRepo{
						{Name: "ORG/Foo"}, // test case insensitivity
						{Id: 3},           // baz id
					},
				}),
			},
		}

		testCases = append(testCases, testCase{
			name: "excluded repos are never yielded",
			svcs: svcs,
			assert: func(t testing.TB, rs Repos) {
				t.Helper()

				set := make(map[string]bool)
				for _, s := range svcs {
					c, err := s.Configuration()
					if err != nil {
						t.Fatal(err)
					}

					type excluded struct {
						name, id string
					}

					var ex []excluded
					switch cfg := c.(type) {
					case *schema.GitHubConnection:
						for _, e := range cfg.Exclude {
							ex = append(ex, excluded{name: e.Name, id: e.Id})
						}
					case *schema.GitLabConnection:
						for _, e := range cfg.Exclude {
							ex = append(ex, excluded{name: e.Name, id: strconv.Itoa(e.Id)})
						}
					case *schema.BitbucketServerConnection:
						for _, e := range cfg.Exclude {
							ex = append(ex, excluded{name: e.Name, id: strconv.Itoa(e.Id)})
						}
					}

					if len(ex) == 0 {
						t.Fatal("exclude list must not be empty")
					}

					for _, e := range ex {
						name := e.name
						switch s.Kind {
						case "GITHUB", "BITBUCKETSERVER":
							name = strings.ToLower(name)
						}
						set[name], set[e.id] = true, true
					}
				}

				for _, r := range rs {
					if set[r.Name] || set[r.ExternalRepo.ID] {
						t.Errorf("excluded repo{name=%s, id=%s} was yielded", r.Name, r.ExternalRepo.ID)
					}
				}
			},
			err: "<nil>",
		})
	}

	{
		svcs := ExternalServices{
			{
				Kind: "GITHUB",
				Config: marshalJSON(t, &schema.GitHubConnection{
					Url:   "https://github.com",
					Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
					Repos: []string{
						"sourcegraph/Sourcegraph",
						"tsenart/Vegeta",
						"tsenart/vegeta-missing",
					},
				}),
			},
			{
				Kind: "GITLAB",
				Config: marshalJSON(t, &schema.GitLabConnection{
					Url:          "https://gitlab.com",
					Token:        os.Getenv("GITLAB_ACCESS_TOKEN"),
					ProjectQuery: []string{"none"},
					Projects: []*schema.GitLabProject{
						{Name: "gnachman/iterm2"},
						{Name: "gnachman/iterm2-missing"},
						{Id: 13083}, // https://gitlab.com/gitlab-org/gitlab-ce
					},
				}),
			},
			{
				Kind: "BITBUCKETSERVER",
				Config: marshalJSON(t, &schema.BitbucketServerConnection{
					Url:             "http://127.0.0.1:7990",
					Token:           os.Getenv("BITBUCKET_SERVER_TOKEN"),
					RepositoryQuery: []string{"none"},
					Repos: []string{
						"Org/foO",
						"org/baz",
						"ORG/bar",
					},
				}),
			},
			{
				Kind: "OTHER",
				Config: marshalJSON(t, &schema.OtherExternalServiceConnection{
					Url: "https://github.com",
					Repos: []string{
						"google/go-cmp",
					},
				}),
			},
		}

		testCases = append(testCases, testCase{
			name: "included repos that exist are yielded",
			svcs: svcs,
			assert: func(t testing.TB, rs Repos) {
				t.Helper()

				have := rs.Names()
				sort.Strings(have)

				want := []string{
					"127.0.0.1/ORG/bar",
					"127.0.0.1/ORG/baz",
					"127.0.0.1/ORG/foo",
					"github.com/google/go-cmp",
					"github.com/sourcegraph/sourcegraph",
					"github.com/tsenart/vegeta",
					"gitlab.com/gitlab-org/gitlab-ce",
					"gitlab.com/gnachman/iterm2",
				}

				if !reflect.DeepEqual(have, want) {
					t.Error(cmp.Diff(have, want))
				}
			},
			err: "<nil>",
		})
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			srcs, err := NewSourcer(cf)(tc.svcs...)
			if err != nil {
				t.Fatal(err)
			}

			ctx := tc.ctx
			if ctx == nil {
				ctx = context.Background()
			}

			repos, err := srcs.ListRepos(ctx)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if tc.assert != nil {
				tc.assert(t, repos)
			}
		})
	}
}

func newClientFactory(t testing.TB, name string) (httpcli.Factory, func(testing.TB)) {
	cassete := filepath.Join("testdata", "sources", strings.Replace(name, " ", "-", -1))
	rec := newRecorder(t, cassete, *update)
	mw := httpcli.NewMiddleware(
		githubProxyRedirectMiddleware,
	)
	return httpcli.NewFactory(mw, newRecorderOpt(t, rec)),
		func(t testing.TB) { save(t, rec) }
}

func githubProxyRedirectMiddleware(cli httpcli.Doer) httpcli.Doer {
	return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Hostname() == "github-proxy" {
			req.URL.Host = "api.github.com"
			req.URL.Scheme = "https"
		}
		return cli.Do(req)
	})
}

func newRecorder(t testing.TB, file string, record bool) *recorder.Recorder {
	mode := recorder.ModeReplaying
	if record {
		mode = recorder.ModeRecording
	}

	rec, err := recorder.NewAsMode(file, mode, nil)
	if err != nil {
		t.Fatal(err)
	}

	rec.AddFilter(func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		// The ratelimit.Monitor type resets its internal timestamp if it's
		// updated with a timestamp in the past. This makes tests ran with
		// recorded interations just wait for a very long time. Removing
		// these headers from the casseste effectively disables rate-limiting
		// in tests which replay HTTP interactions, which is desired behaviour.
		for _, name := range [...]string{
			"RateLimit-Limit",
			"RateLimit-Observed",
			"RateLimit-Remaining",
			"RateLimit-Reset",
			"RateLimit-Resettime",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		} {
			i.Response.Headers.Del(name)
		}
		return nil
	})

	rec.AddFilter(func(i *cassette.Interaction) error {
		return nil
	})

	return rec
}

func save(t testing.TB, rec *recorder.Recorder) {
	if err := rec.Stop(); err != nil {
		t.Errorf("failed to update test data: %s", err)
	}
}

func newRecorderOpt(t testing.TB, rec *recorder.Recorder) httpcli.Opt {
	return func(c *http.Client) error {
		tr := c.Transport
		if tr == nil {
			tr = http.DefaultTransport
		}

		if testing.Verbose() {
			rec.SetTransport(logged(t, "transport")(tr))
			c.Transport = logged(t, "recorder")(rec)
		} else {
			rec.SetTransport(tr)
			c.Transport = rec
		}

		return nil
	}
}

func logged(t testing.TB, prefix string) func(http.RoundTripper) http.RoundTripper {
	return func(rt http.RoundTripper) http.RoundTripper {
		return roundTripFunc(func(req *http.Request) (*http.Response, error) {
			bs, err := httputil.DumpRequestOut(req, true)
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("[%s] request\n%s", prefix, bs)

			res, err := rt.RoundTrip(req)
			if err != nil {
				return res, err
			}

			bs, err = httputil.DumpResponse(res, true)
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("[%s] response\n%s", prefix, bs)

			return res, nil
		})
	}
}

type roundTripFunc httpcli.DoerFunc

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func marshalJSON(t testing.TB, v interface{}) string {
	t.Helper()

	bs, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return string(bs)
}
