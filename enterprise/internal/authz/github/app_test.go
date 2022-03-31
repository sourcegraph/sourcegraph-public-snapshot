package github

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func TestNewAppProvider(t *testing.T) {
	srvHit := false
	doer := &mockDoer{
		do: func(r *http.Request) (*http.Response, error) {
			if r.URL.Path == "/app/installations/1234/access_tokens" {
				tokenString := "app-token"
				tokenExpiry := time.Now()
				token := github.InstallationToken{
					Token:     &tokenString,
					ExpiresAt: &tokenExpiry,
				}

				respJSON, err := json.Marshal(token)
				require.NoError(t, err)

				return &http.Response{
					Status:     http.StatusText(http.StatusCreated),
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(bytes.NewReader(respJSON)),
				}, nil
			}

			srvHit = true
			assert.Equal(t, "Bearer app-token", r.Header.Get("Authorization"))
			return &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`[]`))),
			}, nil
		},
	}

	config, err := json.Marshal(
		&schema.GitHubConnection{
			Url:                     schema.DefaultGitHubURL,
			Authorization:           &schema.GitHubAuthorization{},
			GithubAppInstallationID: "1234",
		},
	)
	require.NoError(t, err)

	const bogusKey = `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlCUEFJQkFBSkJBUEpIaWprdG1UMUlLYUd0YTVFZXAzQVo5Q2VPZUw4alBESUZUN3dRZ0tabXQzRUZxRGhCCk93bitRVUhKdUs5Zm92UkROSmVWTDJvWTVCT0l6NHJ3L0cwQ0F3RUFBUUpCQU1BK0o5Mks0d2NQVllsbWMrM28KcHU5NmlKTkNwMmp5Nm5hK1pEQlQzK0VvSUo1VFJGdnN3R2kvTHUzZThYUWwxTDNTM21ub0xPSlZNcTF0bUxOMgpIY0VDSVFEK3daeS83RlYxUEFtdmlXeWlYVklETzJnNWJOaUJlbmdKQ3hFa3Nia1VtUUloQVBOMlZaczN6UFFwCk1EVG9vTlJXcnl0RW1URERkamdiOFpzTldYL1JPRGIxQWlCZWNKblNVQ05TQllLMXJ5VTFmNURTbitoQU9ZaDkKWDFBMlVnTDE3bWhsS1FJaEFPK2JMNmRDWktpTGZORWxmVnRkTUtxQnFjNlBIK01heFU2VzlkVlFvR1dkQWlFQQptdGZ5cE9zYTFiS2hFTDg0blovaXZFYkJyaVJHalAya3lERHYzUlg0V0JrPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=`

	baseURL, err := url.Parse(schema.DefaultGitHubURL)
	require.NoError(t, err)

	t.Run("with non-nil service", func(t *testing.T) {
		svc := &types.ExternalService{
			ID:     1,
			Kind:   extsvc.KindGitHub,
			Config: string(config),
		}

		provider, err := newAppProvider(database.NewMockExternalServiceStore(), svc, "", baseURL, "1234", bogusKey, 1234, doer)
		require.NoError(t, err)

		cli, err := provider.client()
		require.NoError(t, err)

		// call any endpoint so that test server can check Authorization header
		_, _, err = cli.ListTeamMembers(context.Background(), "anyOwner", "anyTeam", 0)
		require.NoError(t, err)

		assert.True(t, srvHit, "hit server endpoint")
	})

	t.Run("with nil service for validation", func(t *testing.T) {
		var svc *types.ExternalService

		// just validate that a new provider can be created for validation if svc is nil
		_, err = newAppProvider(database.NewMockExternalServiceStore(), svc, "", baseURL, "1234", bogusKey, 1234, doer)
		require.NoError(t, err)
	})

}
