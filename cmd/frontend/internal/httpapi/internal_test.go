package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestReposList(t *testing.T) {
	defaultRepos := []string{"github.com/vim/vim", "github.com/torvalds/linux"}
	allRepos := append(defaultRepos, "github.com/alice/rabbitmq", "github.com/bob/jabberd")

	cases := []struct {
		name string
		srv  *reposListServer
		body string
		want []string
	}{{
		name: "enabled",
		srv: &reposListServer{
			Repos: &mockRepos{
				defaultRepos: defaultRepos,
				repos:        allRepos,
			},
		},
		body: `{"Enabled": true, "Index": true}`,
		want: allRepos,
	}, {
		name: "sourcegraph.com",
		srv: &reposListServer{
			SourcegraphDotComMode: true,
			Repos: &mockRepos{
				defaultRepos: defaultRepos,
				repos:        allRepos,
			},
		},
		body: `{"Enabled": true, "Index": true}`,
		want: defaultRepos,
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

func (r *mockRepos) List(context.Context, db.ReposListOptions) ([]*types.Repo, error) {
	var repos []*types.Repo
	for _, name := range r.repos {
		repos = append(repos, &types.Repo{
			Name: api.RepoName(name),
		})
	}
	return repos, nil
}
