package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestNewAppProvider(t *testing.T) {
	t.Run("test new app provider client", func(t *testing.T) {
		srvHit := false
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v3/app/installations/1234/access_tokens" {
				tokenString := "1234"
				tokenExpiry := time.Now()
				token := github.InstallationToken{
					Token:        &tokenString,
					ExpiresAt:    &tokenExpiry,
					Permissions:  &github.InstallationPermissions{},
					Repositories: []*github.Repository{},
				}

				respJson, err := json.Marshal(token)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				w.Write(respJson)
			} else {
				srvHit = true
				gotHeader := r.Header.Get("Authorization")
				wantHeader := "Bearer 1234"

				assert.Equal(t, wantHeader, gotHeader)
			}
		}))

		ghConnection := &types.GitHubConnection{
			URN: "",
			GitHubConnection: &schema.GitHubConnection{
				Url: srv.URL,
				Authorization: &schema.GitHubAuthorization{
					GroupsCacheTTL: 72,
				},
				GithubAppInstallationID: "1234",
			},
		}

		baseURL, err := url.Parse(ghConnection.Url)
		require.NoError(t, err)

		const bogusKey = `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlCUEFJQkFBSkJBUEpIaWprdG1UMUlLYUd0YTVFZXAzQVo5Q2VPZUw4alBESUZUN3dRZ0tabXQzRUZxRGhCCk93bitRVUhKdUs5Zm92UkROSmVWTDJvWTVCT0l6NHJ3L0cwQ0F3RUFBUUpCQU1BK0o5Mks0d2NQVllsbWMrM28KcHU5NmlKTkNwMmp5Nm5hK1pEQlQzK0VvSUo1VFJGdnN3R2kvTHUzZThYUWwxTDNTM21ub0xPSlZNcTF0bUxOMgpIY0VDSVFEK3daeS83RlYxUEFtdmlXeWlYVklETzJnNWJOaUJlbmdKQ3hFa3Nia1VtUUloQVBOMlZaczN6UFFwCk1EVG9vTlJXcnl0RW1URERkamdiOFpzTldYL1JPRGIxQWlCZWNKblNVQ05TQllLMXJ5VTFmNURTbitoQU9ZaDkKWDFBMlVnTDE3bWhsS1FJaEFPK2JMNmRDWktpTGZORWxmVnRkTUtxQnFjNlBIK01heFU2VzlkVlFvR1dkQWlFQQptdGZ5cE9zYTFiS2hFTDg0blovaXZFYkJyaVJHalAya3lERHYzUlg0V0JrPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=`

		provider, err := newAppProvider(ghConnection.URN, baseURL, "1234", bogusKey, 1234)
		require.NoError(t, err)

		cli, err := provider.client()
		require.NoError(t, err)

		// call any endpoint so that test server can check Authorization header
		cli.ListTeamMembers(context.Background(), "anyOwner", "anyTeam", 0)

		assert.True(t, srvHit, "hit server endpoint")
	})
}
