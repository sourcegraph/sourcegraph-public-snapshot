package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestReposList(t *testing.T) {
	defaultRepos := []string{"github.com/popular/foo", "github.com/popular/bar"}
	allRepos := append(defaultRepos, "github.com/alice/foo", "github.com/alice/bar")

	cases := []struct {
		name string
		srv  *reposListServer
		body string
		want []string
	}{{
		name: "no indexers",
		srv: &reposListServer{
			Repos: &mockRepos{
				defaultRepos: defaultRepos,
				repos:        allRepos,
			},
			Indexers: suffixIndexers(false),
		},
		body: `{"Enabled": true, "Index": true}`,
		want: allRepos,
	}, {
		name: "dot-com no indexers",
		srv: &reposListServer{
			SourcegraphDotComMode: true,
			Repos: &mockRepos{
				defaultRepos: defaultRepos,
				repos:        allRepos,
			},
			Indexers: suffixIndexers(false),
		},
		body: `{"Enabled": true, "Index": true}`,
		want: defaultRepos,
	}, {
		name: "indexers",
		srv: &reposListServer{
			Repos: &mockRepos{
				defaultRepos: defaultRepos,
				repos:        allRepos,
			},
			Indexers: suffixIndexers(true),
		},
		body: `{"Hostname": "foo", "Enabled": true, "Index": true}`,
		want: []string{"github.com/popular/foo", "github.com/alice/foo"},
	}, {
		name: "dot-com indexers",
		srv: &reposListServer{
			SourcegraphDotComMode: true,
			Repos: &mockRepos{
				defaultRepos: defaultRepos,
				repos:        allRepos,
			},
			Indexers: suffixIndexers(true),
		},
		body: `{"Hostname": "foo", "Enabled": true, "Index": true}`,
		want: []string{"github.com/popular/foo"},
	}, {
		name: "none",
		srv: &reposListServer{
			Repos: &mockRepos{
				defaultRepos: defaultRepos,
				repos:        allRepos,
			},
			Indexers: suffixIndexers(true),
		},
		body: `{"Hostname": "baz", "Enabled": true, "Index": true}`,
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(tc.body)))
			w := httptest.NewRecorder()
			if err := tc.srv.serve(w, req); err != nil {
				t.Fatal(err)
			}

			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			if resp.StatusCode != http.StatusOK {
				t.Errorf("got status %v", resp.StatusCode)
			}

			// Parse the response as in zoekt-sourcegraph-indexserver/main.go.
			var repos []struct{ URI string }
			if err := json.Unmarshal(body, &repos); err != nil {
				t.Fatal(err)
			}

			var got []string
			for _, r := range repos {
				got = append(got, r.URI)
			}

			if !cmp.Equal(tc.want, got) {
				t.Fatalf("mismatch (-want +got):\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}

type mockRepos struct {
	defaultRepos []string
	repos        []string
}

func (r *mockRepos) ListDefault(context.Context) ([]*types.Repo, error) {
	var repos []*types.Repo
	for _, name := range r.defaultRepos {
		repos = append(repos, &types.Repo{
			Name: api.RepoName(name),
		})
	}
	return repos, nil
}

func (r *mockRepos) List(ctx context.Context, opt db.ReposListOptions) ([]*types.Repo, error) {
	if opt.Index == nil || !*opt.Index {
		return nil, errors.New("reposList test expects Index=true options")
	}

	var repos []*types.Repo
	for _, name := range r.repos {
		repos = append(repos, &types.Repo{
			Name: api.RepoName(name),
		})
	}
	return repos, nil
}

// suffixIndexers mocks Indexers. ReposSubset will return all repoNames with
// the suffix of hostname.
type suffixIndexers bool

func (b suffixIndexers) ReposSubset(ctx context.Context, hostname string, repoNames []string) ([]string, error) {
	if !b.Enabled() {
		return nil, errors.New("indexers disabled")
	}
	if hostname == "" {
		return nil, errors.New("empty hostname")
	}

	var filter []string
	for _, name := range repoNames {
		if strings.HasSuffix(name, hostname) {
			filter = append(filter, name)
		}
	}
	return filter, nil
}

func (b suffixIndexers) Enabled() bool {
	return bool(b)
}
