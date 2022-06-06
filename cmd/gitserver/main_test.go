// gitserver is the gitserver server.
package main

import (
	"bytes"
	"context"
	"flag"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		logtest.InitWithLevel(m, log.LevelNone)
	}
	os.Exit(m.Run())
}

func TestParsePercent(t *testing.T) {
	tests := []struct {
		i       int
		want    int
		wantErr bool
	}{
		{i: -1, wantErr: true},
		{i: -4, wantErr: true},
		{i: 300, wantErr: true},
		{i: 0, want: 0},
		{i: 50, want: 50},
		{i: 100, want: 100},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got, err := getPercent(tt.i)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePercent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parsePercent() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func TestGetRemoteURLFunc_GitHubAppCloud(t *testing.T) {
	externalServiceStore := database.NewMockExternalServiceStore()
	externalServiceStore.GetByIDFunc.SetDefaultReturn(
		&types.ExternalService{
			ID:   1,
			Kind: extsvc.KindGitHub,
			Config: `
{
  "url": "https://github.com",
  "githubAppInstallationID": "21994992",
  "repos": []
}`,
		},
		nil,
	)

	repoStore := database.NewMockRepoStore()
	repoStore.GetByNameFunc.SetDefaultReturn(
		&types.Repo{
			ID:   1,
			Name: "test-repo-1",
			Sources: map[string]*types.SourceInfo{
				"extsvc:github:1": {
					ID:       "extsvc:github:1",
					CloneURL: "https://github.com/sgtest/test-repo-1",
				},
			},
			Metadata: &github.Repository{
				URL: "https://github.com/sgtest/test-repo-1",
			},
		},
		nil,
	)

	doer := &mockDoer{
		do: func(r *http.Request) (*http.Response, error) {
			want := "http://github-proxy/app/installations/21994992/access_tokens"
			if r.URL.String() != want {
				return nil, errors.Errorf("URL: want %q but got %q", want, r.URL)
			}

			body := `{"token": "mock-installtion-access-token"}`
			return &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(body))),
			}, nil
		},
	}

	orig := envvar.SourcegraphDotComMode()
	envvar.MockSourcegraphDotComMode(true)
	defer envvar.MockSourcegraphDotComMode(orig)

	const bogusKey = `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlCUEFJQkFBSkJBUEpIaWprdG1UMUlLYUd0YTVFZXAzQVo5Q2VPZUw4alBESUZUN3dRZ0tabXQzRUZxRGhCCk93bitRVUhKdUs5Zm92UkROSmVWTDJvWTVCT0l6NHJ3L0cwQ0F3RUFBUUpCQU1BK0o5Mks0d2NQVllsbWMrM28KcHU5NmlKTkNwMmp5Nm5hK1pEQlQzK0VvSUo1VFJGdnN3R2kvTHUzZThYUWwxTDNTM21ub0xPSlZNcTF0bUxOMgpIY0VDSVFEK3daeS83RlYxUEFtdmlXeWlYVklETzJnNWJOaUJlbmdKQ3hFa3Nia1VtUUloQVBOMlZaczN6UFFwCk1EVG9vTlJXcnl0RW1URERkamdiOFpzTldYL1JPRGIxQWlCZWNKblNVQ05TQllLMXJ5VTFmNURTbitoQU9ZaDkKWDFBMlVnTDE3bWhsS1FJaEFPK2JMNmRDWktpTGZORWxmVnRkTUtxQnFjNlBIK01heFU2VzlkVlFvR1dkQWlFQQptdGZ5cE9zYTFiS2hFTDg0blovaXZFYkJyaVJHalAya3lERHYzUlg0V0JrPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=`
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			Dotcom: &schema.Dotcom{
				GithubAppCloud: &schema.GithubAppCloud{
					AppID:      "404",
					PrivateKey: bogusKey,
					Slug:       "test-app",
				},
			},
		},
	})
	defer conf.Mock(nil)

	got, err := getRemoteURLFunc(context.Background(), externalServiceStore, repoStore, doer, "test-repo-1")
	require.NoError(t, err)

	want := "https://x-access-token:mock-installtion-access-token@github.com/sgtest/test-repo-1"
	assert.Equal(t, want, got)
}

func TestGetVCSSyncer(t *testing.T) {
	repo := api.RepoName("foo/bar")
	extsvcStore := database.NewMockExternalServiceStore()
	repoStore := database.NewMockRepoStore()
	depsSvc := new(dependencies.Service)

	repoStore.GetByNameFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
		return &types.Repo{
			ExternalRepo: api.ExternalRepoSpec{
				ServiceType: extsvc.TypePerforce,
			},
			Sources: map[string]*types.SourceInfo{
				"a": {
					ID:       "abc",
					CloneURL: "example.com",
				},
			},
		}, nil
	})

	extsvcStore.GetByIDFunc.SetDefaultHook(func(ctx context.Context, i int64) (*types.ExternalService, error) {
		return &types.ExternalService{
			ID:          1,
			Kind:        extsvc.KindPerforce,
			DisplayName: "test",
			Config:      `{}`,
		}, nil
	})

	s, err := getVCSSyncer(context.Background(), extsvcStore, repoStore, depsSvc, repo)
	if err != nil {
		t.Fatal(err)
	}

	_, ok := s.(*server.PerforceDepotSyncer)
	if !ok {
		t.Fatalf("Want *server.PerforceDepotSyncer, got %T", s)
	}
}
