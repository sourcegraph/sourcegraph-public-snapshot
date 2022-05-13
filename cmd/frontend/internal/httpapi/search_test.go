package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/zoekt"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestServeConfiguration(t *testing.T) {
	repos := []types.MinimalRepo{{
		ID:    5,
		Name:  "5",
		Stars: 5,
	}, {
		ID:    6,
		Name:  "6",
		Stars: 6,
	}}
	srv := &searchIndexerServer{
		RepoStore: &fakeRepoStore{Repos: repos},
		SearchContextsRepoRevs: func(ctx context.Context, repoIDs []api.RepoID) (map[api.RepoID][]string, error) {
			return map[api.RepoID][]string{6: {"a", "b"}}, nil
		},
	}

	gitserver.Mocks.ResolveRevision = func(spec string, _ gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		return api.CommitID("!" + spec), nil
	}
	t.Cleanup(func() { gitserver.Mocks.ResolveRevision = nil })

	data := url.Values{
		"repoID": []string{"1", "5", "6"},
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	if err := srv.serveConfiguration(w, req); err != nil {
		t.Fatal(err)
	}

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	// This is a very fragile test since it will depend on changes to
	// searchbackend.GetIndexOptions. If this becomes a problem we can make it
	// more robust by shifting around responsibilities.
	want := `{"Name":"","RepoID":0,"Public":false,"Fork":false,"Archived":false,"LargeFiles":null,"Symbols":false,"Error":"repo not found: id=1"}
{"Name":"5","RepoID":5,"Public":true,"Fork":false,"Archived":false,"LargeFiles":null,"Symbols":true,"Branches":[{"Name":"HEAD","Version":"!HEAD"}],"Priority":5}
{"Name":"6","RepoID":6,"Public":true,"Fork":false,"Archived":false,"LargeFiles":null,"Symbols":true,"Branches":[{"Name":"HEAD","Version":"!HEAD"},{"Name":"a","Version":"!a"},{"Name":"b","Version":"!b"}],"Priority":6}`

	if d := cmp.Diff(want, string(body)); d != "" {
		t.Fatalf("mismatch (-want, +got):\n%s", d)
	}

	// when fingerprint is set we only return a subset. We simulate this by setting RepoStore to only list repo number 5
	srv.RepoStore = &fakeRepoStore{Repos: repos[:1]}
	req = httptest.NewRequest("POST", "/", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Sourcegraph-Config-Fingerprint", resp.Header.Get("X-Sourcegraph-Config-Fingerprint"))

	w = httptest.NewRecorder()
	if err := srv.serveConfiguration(w, req); err != nil {
		t.Fatal(err)
	}

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)

	// We want the same as before, except we only want to get back 5.
	//
	// This is a very fragile test since it will depend on changes to
	// searchbackend.GetIndexOptions. If this becomes a problem we can make it
	// more robust by shifting around responsibilities.
	want = `{"Name":"5","RepoID":5,"Public":true,"Fork":false,"Archived":false,"LargeFiles":null,"Symbols":true,"Branches":[{"Name":"HEAD","Version":"!HEAD"}],"Priority":5}`

	if d := cmp.Diff(want, string(body)); d != "" {
		t.Fatalf("mismatch (-want, +got):\n%s", d)
	}
}

func TestReposIndex(t *testing.T) {
	allRepos := []types.MinimalRepo{
		{ID: 1, Name: "github.com/popular/foo"},
		{ID: 2, Name: "github.com/popular/bar"},
		{ID: 3, Name: "github.com/alice/foo"},
		{ID: 4, Name: "github.com/alice/bar"},
	}

	indexableRepos := allRepos[:2]

	cases := []struct {
		name      string
		indexable []types.MinimalRepo
		body      string
		want      []string
	}{{
		name:      "indexers",
		indexable: allRepos,
		body:      `{"Hostname": "foo"}`,
		want:      []string{"github.com/popular/foo", "github.com/alice/foo"},
	}, {
		name:      "indexedids",
		indexable: allRepos,
		body:      `{"Hostname": "foo", "IndexedIDs": [4]}`,
		want:      []string{"github.com/popular/foo", "github.com/alice/foo", "github.com/alice/bar"},
	}, {
		name:      "dot-com indexers",
		indexable: indexableRepos,
		body:      `{"Hostname": "foo"}`,
		want:      []string{"github.com/popular/foo"},
	}, {
		name:      "dot-com indexedids",
		indexable: indexableRepos,
		body:      `{"Hostname": "foo", "IndexedIDs": [2]}`,
		want:      []string{"github.com/popular/foo", "github.com/popular/bar"},
	}, {
		name:      "none",
		indexable: allRepos,
		body:      `{"Hostname": "baz"}`,
		want:      []string{},
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := &searchIndexerServer{
				ListIndexable: fakeListIndexable(tc.indexable),
				RepoStore: &fakeRepoStore{
					Repos: allRepos,
				},
				Indexers: suffixIndexers(true),
			}

			req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(tc.body)))
			w := httptest.NewRecorder()
			if err := srv.serveList(w, req); err != nil {
				t.Fatal(err)
			}

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != http.StatusOK {
				t.Errorf("got status %v", resp.StatusCode)
			}

			var data struct {
				RepoIDs []api.RepoID
			}
			if err := json.Unmarshal(body, &data); err != nil {
				t.Fatal(err)
			}

			wantIDs := make([]api.RepoID, len(tc.want))
			for i, name := range tc.want {
				for _, repo := range allRepos {
					if string(repo.Name) == name {
						wantIDs[i] = repo.ID
					}
				}
			}
			if d := cmp.Diff(wantIDs, data.RepoIDs); d != "" {
				t.Fatalf("ids mismatch (-want +got):\n%s", d)
			}
		})
	}
}

func fakeListIndexable(indexable []types.MinimalRepo) func(context.Context) ([]types.MinimalRepo, error) {
	return func(context.Context) ([]types.MinimalRepo, error) {
		return indexable, nil
	}
}

type fakeRepoStore struct {
	Repos []types.MinimalRepo
}

func (f *fakeRepoStore) List(_ context.Context, opts database.ReposListOptions) ([]*types.Repo, error) {
	var repos []*types.Repo
	for _, r := range f.Repos {
		for _, id := range opts.IDs {
			if id == r.ID {
				repos = append(repos, r.ToRepo())
			}
		}
	}

	return repos, nil
}

func (f *fakeRepoStore) StreamMinimalRepos(ctx context.Context, opt database.ReposListOptions, cb func(*types.MinimalRepo)) error {
	names := make(map[string]bool, len(opt.Names))
	for _, name := range opt.Names {
		names[name] = true
	}

	ids := make(map[api.RepoID]bool, len(opt.IDs))
	for _, id := range opt.IDs {
		ids[id] = true
	}

	for i := range f.Repos {
		r := &f.Repos[i]
		if names[string(r.Name)] || ids[r.ID] {
			cb(&f.Repos[i])
		}
	}

	return nil
}

// suffixIndexers mocks Indexers. ReposSubset will return all repoNames with
// the suffix of hostname.
type suffixIndexers bool

func (b suffixIndexers) ReposSubset(ctx context.Context, hostname string, indexed map[uint32]*zoekt.MinimalRepoListEntry, indexable []types.MinimalRepo) ([]types.MinimalRepo, error) {
	if !b.Enabled() {
		return nil, errors.New("indexers disabled")
	}
	if hostname == "" {
		return nil, errors.New("empty hostname")
	}

	var filter []types.MinimalRepo
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
