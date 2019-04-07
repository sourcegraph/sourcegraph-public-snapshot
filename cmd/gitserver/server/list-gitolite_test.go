package server

import (
	"context"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/pkg/tst"
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
				"git@gitolite.example.com": []*gitolite.Repo{
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
				config: stubConfig{
					Gitolite_: func(ctx context.Context) ([]*schema.GitoliteConnection, error) {
						return test.configs, nil
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
			tst.CheckStrEqual(t, test.expResponseBody, string(respBody), "response body did not match")
			tst.CheckDeepEqual(t, test.expResponseCode, resp.StatusCode, "response status codes did not match")
		})
	}
}

type stubConfig struct {
	Gitolite_ func(ctx context.Context) ([]*schema.GitoliteConnection, error)
}

func (c stubConfig) Gitolite(ctx context.Context) ([]*schema.GitoliteConnection, error) {
	return c.Gitolite_(ctx)
}

type stubGitoliteClient struct {
	ListRepos_ func(ctx context.Context, host string) ([]*gitolite.Repo, error)
}

func (c stubGitoliteClient) ListRepos(ctx context.Context, host string) ([]*gitolite.Repo, error) {
	return c.ListRepos_(ctx, host)
}
