package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/schema"
)

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
	for _, tc := range []struct {
		name   string
		ctx    context.Context
		svc    ExternalService
		assert ReposAssertion
		err    string
	}{
		{
			name: "github/yielded repos are always enabled",
			svc: ExternalService{
				Kind: "GITHUB",
				Config: marshalJSON(t, &schema.GitHubConnection{
					Url: "https://github.com",
					RepositoryQuery: []string{
						"user:tsenart in:name patrol",
					},
					Repos: []string{"sourcegraph/sourcegraph"},
				}),
			},
			assert: func(t testing.TB, rs Repos) {
				for _, r := range rs {
					if !r.Enabled {
						t.Errorf("repo %q is not enabled", r.Name)
					}
				}
			},
			err: "<nil>",
		},
		{
			name: "gitlab/yielded repos are always enabled",
			svc: ExternalService{
				Kind: "GITLAB",
				Config: marshalJSON(t, &schema.GitLabConnection{
					Url: "https://gitlab.com",
					ProjectQuery: []string{
						"?search=vegeta",
					},
					// Repos: []string{"sourcegraph/sourcegraph"},
				}),
			},
			assert: func(t testing.TB, rs Repos) {
				if len(rs) == 0 {
					t.Fatal("expected Gitlab repositories")
				}

				for _, r := range rs {
					if !r.Enabled {
						t.Errorf("repo %q is not enabled", r.Name)
					}
				}
			},
			err: "<nil>",
		},
		{
			name: "github/excluded repos are never yielded",
			svc: ExternalService{
				Kind: "GITHUB",
				Config: marshalJSON(t, &schema.GitHubConnection{
					Url: "https://github.com",
					RepositoryQuery: []string{
						"user:tsenart in:name patrol",
					},
					Repos: []string{
						"sourcegraph/Sourcegraph",
						"tsenart/VEGETA",
					},
					Exclude: []*schema.Exclude{
						{Name: "tsenart/Vegeta"},
						{Id: "MDEwOlJlcG9zaXRvcnkxNTM2NTcyNDU="}, // tsenart/patrol ID
					},
				}),
			},
			assert: func(t testing.TB, rs Repos) {
				t.Helper()

				have := rs.Names()
				want := []string{"github.com/sourcegraph/sourcegraph"}

				sort.Strings(have)
				sort.Strings(want)

				if !reflect.DeepEqual(have, want) {
					t.Errorf("\nhave: %s\nwant: %s", have, want)
				}
			},
			err: "<nil>",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cf, save := newClientFactory(t, tc.name, &tc.svc)
			defer save(t)

			s, err := NewSource(&tc.svc, cf)
			if err != nil {
				t.Error(err)
				return // Let defers run
			}

			ctx := tc.ctx
			if ctx == nil {
				ctx = context.Background()
			}

			repos, err := s.ListRepos(ctx)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if tc.assert != nil {
				tc.assert(t, repos)
			}
		})
	}
}

func newClientFactory(t testing.TB, name string, svc *ExternalService) (httpcli.Factory, func(testing.TB)) {
	cassete := filepath.Join("testdata", strings.Replace(name, " ", "-", -1))
	rec := newRecorder(t, cassete, *update)

	var mw httpcli.Middleware
	switch strings.ToLower(svc.Kind) {
	case "github":
		mw = httpcli.NewMiddleware(
			redirect(map[string]string{"github-proxy": "api.github.com"}),
			auth(os.Getenv("GITHUB_ACCESS_TOKEN")),
		)
	case "gitlab":
		mw = httpcli.NewMiddleware(
			auth(os.Getenv("GITLAB_ACCESS_TOKEN")),
		)
	}

	return httpcli.NewFactory(mw, newRecorderOpt(t, rec)),
		func(t testing.TB) { save(t, rec) }
}

func redirect(rules map[string]string) httpcli.Middleware {
	return func(cli httpcli.Doer) httpcli.Doer {
		return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
			if host, ok := rules[req.URL.Hostname()]; ok {
				req.URL.Host = host
			}
			return cli.Do(req)
		})
	}
}

func auth(token string) httpcli.Middleware {
	return func(cli httpcli.Doer) httpcli.Doer {
		return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
			if token != "" {
				req.Header.Set("Authorization", "bearer "+token)
			}
			return cli.Do(req)
		})
	}
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
