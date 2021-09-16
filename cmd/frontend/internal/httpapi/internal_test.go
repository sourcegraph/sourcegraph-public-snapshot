package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"
	"github.com/google/zoekt"
	"github.com/gorilla/mux"

	apirouter "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/router"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGitServiceHandlers(t *testing.T) {
	m := apirouter.NewInternal(mux.NewRouter())

	gitService := &gitServiceHandler{
		Gitserver: mockAddrForRepo{},
	}
	m.Get(apirouter.GitInfoRefs).Handler(http.HandlerFunc(gitService.serveInfoRefs))
	m.Get(apirouter.GitUploadPack).Handler(http.HandlerFunc(gitService.serveGitUploadPack))

	cases := map[string]string{
		"/git/foo/bar/info/refs?service=git-upload-pack": "http://foo.bar.gitserver/git/foo/bar/info/refs?service=git-upload-pack",
		"/git/foo/bar/git-upload-pack":                   "http://foo.bar.gitserver/git/foo/bar/git-upload-pack",
	}

	for target, want := range cases {
		req := httptest.NewRequest("GET", target, nil)
		w := httptest.NewRecorder()
		m.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusTemporaryRedirect {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("expected redirect for %q, got status %d. Body: %s", target, resp.StatusCode, body)
			continue
		}

		got := resp.Header.Get("Location")
		if got != want {
			t.Errorf("mismatched location for %q:\ngot:  %s\nwant: %s", target, got, want)
		}
	}
}

type mockAddrForRepo struct{}

func (mockAddrForRepo) AddrForRepo(name api.RepoName) string {
	return strings.ReplaceAll(string(name), "/", ".") + ".gitserver"
}

func TestReposIndex(t *testing.T) {
	allRepos := []types.RepoName{
		{ID: 1, Name: "github.com/popular/foo"},
		{ID: 2, Name: "github.com/popular/bar"},
		{ID: 3, Name: "github.com/alice/foo"},
		{ID: 4, Name: "github.com/alice/bar"},
	}

	indexableRepos := allRepos[:2]

	cases := []struct {
		name string
		srv  *reposListServer
		body string
		want []string
	}{{
		name: "indexers",
		srv: &reposListServer{
			ListIndexable:   fakeListIndexable(allRepos),
			StreamRepoNames: fakeStreamRepoNames(allRepos),
			Indexers:        suffixIndexers(true),
		},
		body: `{"Hostname": "foo"}`,
		want: []string{"github.com/popular/foo", "github.com/alice/foo"},
	}, {
		name: "indexers",
		srv: &reposListServer{
			ListIndexable:   fakeListIndexable(allRepos),
			StreamRepoNames: fakeStreamRepoNames(allRepos),
			Indexers:        suffixIndexers(true),
		},
		body: `{"Hostname": "foo", "Indexed": ["github.com/alice/bar"]}`,
		want: []string{"github.com/popular/foo", "github.com/alice/foo", "github.com/alice/bar"},
	}, {
		name: "dot-com indexers",
		srv: &reposListServer{
			ListIndexable:   fakeListIndexable(indexableRepos),
			StreamRepoNames: fakeStreamRepoNames(allRepos),
			Indexers:        suffixIndexers(true),
		},
		body: `{"Hostname": "foo"}`,
		want: []string{"github.com/popular/foo"},
	}, {
		name: "none",
		srv: &reposListServer{
			ListIndexable:   fakeListIndexable(allRepos),
			StreamRepoNames: fakeStreamRepoNames(allRepos),
			Indexers:        suffixIndexers(true),
		},
		want: []string{},
		body: `{"Hostname": "baz"}`,
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(tc.body)))
			w := httptest.NewRecorder()
			if err := tc.srv.serveIndex(w, req); err != nil {
				t.Fatal(err)
			}

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != http.StatusOK {
				t.Errorf("got status %v", resp.StatusCode)
			}

			var data struct {
				RepoNames []string
			}
			if err := json.Unmarshal(body, &data); err != nil {
				t.Fatal(err)
			}
			got := data.RepoNames

			if !cmp.Equal(tc.want, got) {
				t.Fatalf("mismatch (-want +got):\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}

func fakeListIndexable(indexable []types.RepoName) func(context.Context) ([]types.RepoName, error) {
	return func(context.Context) ([]types.RepoName, error) {
		return indexable, nil
	}
}

func fakeStreamRepoNames(repos []types.RepoName) func(context.Context, database.ReposListOptions, func(*types.RepoName)) error {
	return func(ctx context.Context, opt database.ReposListOptions, cb func(*types.RepoName)) error {
		names := make(map[string]bool, len(opt.Names))
		for _, name := range opt.Names {
			names[name] = true
		}

		ids := make(map[api.RepoID]bool, len(opt.IDs))
		for _, id := range opt.IDs {
			ids[id] = true
		}

		for i := range repos {
			r := &repos[i]
			if names[string(r.Name)] || ids[r.ID] {
				cb(&repos[i])
			}
		}

		return nil
	}
}

// suffixIndexers mocks Indexers. ReposSubset will return all repoNames with
// the suffix of hostname.
type suffixIndexers bool

func (b suffixIndexers) ReposSubset(ctx context.Context, hostname string, indexed map[uint32]*zoekt.MinimalRepoListEntry, indexable []types.RepoName) ([]types.RepoName, error) {
	if !b.Enabled() {
		return nil, errors.New("indexers disabled")
	}
	if hostname == "" {
		return nil, errors.New("empty hostname")
	}

	var filter []types.RepoName
	for _, r := range indexable {
		if strings.HasSuffix(string(r.Name), hostname) {
			filter = append(filter, r)
		} else if _, ok := indexed[uint32(r.ID)]; ok {
			filter = append(filter, r)
		}
	}
	return filter, nil
}

func (b suffixIndexers) Enabled() bool {
	return bool(b)
}

func TestRepoRankFromConfig(t *testing.T) {
	cases := []struct {
		name       string
		rankScores map[string]float64
		want       float64
	}{
		{"gh.test/sg/sg", nil, 0},
		{"gh.test/sg/sg", map[string]float64{"gh.test": 100}, 100},
		{"gh.test/sg/sg", map[string]float64{"gh.test": 100, "gh.test/sg": 50}, 150},
		{"gh.test/sg/sg", map[string]float64{"gh.test": 100, "gh.test/sg": 50, "gh.test/sg/sg": -20}, 130},
		{"gh.test/sg/ex", map[string]float64{"gh.test": 100, "gh.test/sg": 50, "gh.test/sg/sg": -20}, 150},
	}
	for _, tc := range cases {
		config := schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{
			Ranking: &schema.Ranking{
				RepoScores: tc.rankScores,
			},
		}}
		got := repoRankFromConfig(config, tc.name)
		if got != tc.want {
			t.Errorf("got score %v, want %v, repo %q config %v", got, tc.want, tc.name, tc.rankScores)
		}
	}
}
