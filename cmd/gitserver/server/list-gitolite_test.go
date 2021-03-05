package server

import (
	"context"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Test_Gitolite_listRepos(t *testing.T) {
	tests := []struct {
		listRepos       map[string][]*gitolite.Repo
		configs         []*schema.GitoliteConnection
		gitoliteHost    string
		expResponseCode int
		expResponseBody string
	}{
		{
			listRepos: map[string][]*gitolite.Repo{
				"git@gitolite.example.com": {
					{Name: "myrepo", URL: "git@gitolite.example.com:myrepo"},
				},
			},
			configs: []*schema.GitoliteConnection{
				{
					Host:   "git@gitolite.example.com",
					Prefix: "gitolite.example.com/",
				},
			},
			gitoliteHost:    "git@gitolite.example.com",
			expResponseCode: 200,
			expResponseBody: `[{"Name":"myrepo","URL":"git@gitolite.example.com:myrepo"}]` + "\n",
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			g := gitoliteFetcher{
				client: stubGitoliteClient{
					ListRepos_: func(ctx context.Context, host string) ([]*gitolite.Repo, error) {
						return test.listRepos[host], nil
					},
				},
			}
			w := httptest.NewRecorder()
			g.listRepos(context.Background(), test.gitoliteHost, w)
			resp := w.Result()
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test.expResponseBody, string(respBody)); diff != "" {
				t.Errorf("unexpected response body diff:\n%s", diff)
			}
			if diff := cmp.Diff(test.expResponseCode, resp.StatusCode); diff != "" {
				t.Errorf("unexpected response code diff:\n%s", diff)
			}
		})
	}
}

type stubGitoliteClient struct {
	ListRepos_ func(ctx context.Context, host string) ([]*gitolite.Repo, error)
}

func (c stubGitoliteClient) ListRepos(ctx context.Context, host string) ([]*gitolite.Repo, error) {
	return c.ListRepos_(ctx, host)
}
