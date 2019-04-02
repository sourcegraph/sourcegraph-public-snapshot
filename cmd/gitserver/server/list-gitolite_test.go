package server

import (
	"context"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server/mock_server"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/pkg/tst"
	"github.com/sourcegraph/sourcegraph/schema"
)

type gitoliteMockParams struct {
	repos           []*gitolite.Repo
	configs         []*schema.GitoliteConnection
	expGitoliteHost string
}

func (p gitoliteMockParams) newGitolite(ctrl *gomock.Controller) Gitolite {
	return Gitolite{
		client: func() IGitoliteClient {
			m := mock_server.NewMockIGitoliteClient(ctrl)
			m.
				EXPECT().
				ListRepos(gomock.Any(), gomock.Eq("git@gitolite.example.com")).
				Return([]*gitolite.Repo{
					{Name: "myrepo", URL: "git@gitolite.example.com:myrepo"},
				}, nil).
				AnyTimes()
			return m
		}(),
		config: func() IConfig {
			m := mock_server.NewMockIConfig(ctrl)
			m.
				EXPECT().
				Gitolite(gomock.Any()).
				Return([]*schema.GitoliteConnection{
					{
						Host:   "git@gitolite.example.com",
						Prefix: "gitolite.example.com/",
					},
				}, nil).
				AnyTimes()
			return m
		}(),
	}
}

func Test_Gitolite_listGitolite(t *testing.T) {
	tests := []struct {
		gitolite func(ctrl *gomock.Controller) Gitolite

		gitoliteHost    string
		expResponseCode int
		expResponseBody string
	}{
		{
			gitolite: gitoliteMockParams{
				repos: []*gitolite.Repo{
					{Name: "myrepo", URL: "git@gitolite.example.com:myrepo"},
				},
				configs: []*schema.GitoliteConnection{
					{
						Host:   "git@gitolite.example.com",
						Prefix: "gitolite.example.com/",
					},
				},
				expGitoliteHost: "git@gitolite.example.com",
			}.newGitolite,

			gitoliteHost:    "git@gitolite.example.com",
			expResponseCode: 200,
			expResponseBody: `[{"Name":"myrepo","URL":"git@gitolite.example.com:myrepo"}]` + "\n",
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			g := test.gitolite(ctrl)
			w := httptest.NewRecorder()
			g.listGitolite(context.Background(), test.gitoliteHost, w)

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
