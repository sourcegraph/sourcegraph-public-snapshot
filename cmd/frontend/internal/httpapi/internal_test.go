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
	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	apirouter "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/router"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
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
			body, _ := ioutil.ReadAll(resp.Body)
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

func (mockAddrForRepo) AddrForRepo(_ context.Context, name api.RepoName) string {
	return strings.ReplaceAll(string(name), "/", ".") + ".gitserver"
}

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
			if err := tc.srv.serveList(w, req); err != nil {
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

func TestReposIndex(t *testing.T) {
	defaultRepos := []string{"github.com/popular/foo", "github.com/popular/bar"}
	allRepos := append(defaultRepos, "github.com/alice/foo", "github.com/alice/bar")

	cases := []struct {
		name string
		srv  *reposListServer
		body string
		want []string
	}{{
		name: "indexers",
		srv: &reposListServer{
			Repos: &mockRepos{
				defaultRepos: defaultRepos,
				repos:        allRepos,
			},
			Indexers: suffixIndexers(true),
		},
		body: `{"Hostname": "foo"}`,
		want: []string{"github.com/popular/foo", "github.com/alice/foo"},
	}, {
		name: "indexers",
		srv: &reposListServer{
			Repos: &mockRepos{
				defaultRepos: defaultRepos,
				repos:        allRepos,
			},
			Indexers: suffixIndexers(true),
		},
		body: `{"Hostname": "foo", "Indexed": ["github.com/alice/bar"]}`,
		want: []string{"github.com/popular/foo", "github.com/alice/foo", "github.com/alice/bar"},
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
		body: `{"Hostname": "foo"}`,
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
			body, _ := ioutil.ReadAll(resp.Body)

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

func (b suffixIndexers) ReposSubset(ctx context.Context, hostname string, indexed map[string]struct{}, repoNames []string) ([]string, error) {
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
		} else if _, ok := indexed[name]; ok {
			filter = append(filter, name)
		}
	}
	return filter, nil
}

func (b suffixIndexers) Enabled() bool {
	return bool(b)
}

func TestServeSearchConfiguration(t *testing.T) {
	var cfg conf.Unified
	cfg.SearchLargeFiles = []string{"**/*.jar", "*.bin"}
	symbolsEnabled := true
	cfg.SearchIndexSymbolsEnabled = &symbolsEnabled
	cfg.ExperimentalFeatures = &schema.ExperimentalFeatures{
		SearchMultipleBranchIndexing: []*schema.SearchMultipleBranchIndexing{
			{
				Name: "foo",
				Branches: []*schema.SearchMultipleBranchIndexingConfig{
					{Name: "branch1"},
					{Name: "branch2", Version: "abcde"},
				},
			},
		},
	}

	conf.Mock(&cfg)

	cases := []struct {
		name   string
		repo   string // optional repo query parameter
		status int    // optional, defaults to 200
		want   string
	}{
		{
			name: "default",
			want: `{"LargeFiles":["**/*.jar","*.bin"],"Symbols":true}`,
		},
		{
			name: "with valid repo",
			repo: "foo",
			want: `{"LargeFiles":["**/*.jar","*.bin"],"Symbols":true, "Branches": [{"Name": "branch1"}, {"Name": "branch2", "Version": "abcde"}]}`,
		},
		{
			name:   "with unknown repo",
			repo:   "bar",
			status: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/"
			if tc.repo != "" {
				url = "/?repo=" + tc.repo
			}
			req := httptest.NewRequest("POST", url, nil)
			w := httptest.NewRecorder()
			err := serveSearchConfiguration(w, req)
			if err != nil {
				t.Fatal(err)
			}

			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)
			expectedStatus := http.StatusOK
			if tc.status != 0 {
				expectedStatus = tc.status
			}
			if resp.StatusCode != expectedStatus {
				t.Fatalf("got status %v", resp.StatusCode)
			}
			if resp.StatusCode == http.StatusNotFound {
				return
			}

			var got, want struct {
				LargeFiles []string
				Symbols    bool
				Branches   []struct {
					Name    string
					Version string
				}
			}
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(body, &want); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(want, got) {
				t.Fatalf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
