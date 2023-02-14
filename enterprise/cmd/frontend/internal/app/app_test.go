package app

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	gogithub "github.com/google/go-github/v41/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestNewGitHubAppSetupHandler(t *testing.T) {
	const bogusKey = `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWHdJQkFBS0JnUUMvemZJMXRqSlUzbHIxQlFIUHMxYzFvbUNrMFJ0RVVQYXpKTTRYaXEvTmo5ZW13cXhnCmdseVNraEgrU0tKa1hJeXdzTjlBc2hpWm9EOFF1UEtKdy9pQkwrQXNNemU2VmlEa0hoMFMza0hqdGxNVWRlQTMKanMwVFluNnh2TXh1Z3lwTVdKV3BBaS9pdm5Ta3pYNmdtRStjVU4rbDl4aUlWNkx0bGl4M0hla3Nyd0lEQVFBQgpBb0dCQUt3bFp6SVY2RzZMY3c5ZUF4WXJYQ1pqS21KQzJ6b2hnSW1naXVoT0xTTk42cnRkRmVFNG4yVmRmSkRCCkdCOERnYkpEek52Ly9GeEZtdFNqYWV1RDI5QnBBVThvUnQzczBsOXo2K1hkaG5XRzhoNHdDOW83MUJiVTcyUVcKVkIyL0hCTkJMSzBSY1BqV2lvWnp5a3lhQ0dKYnhSemRNV3hMME8xcjJ0MmRtZWRCQWtFQTQ5RmoxVWlWWER5dApKcDVBdkJudk1WUHdjdlI3UnpRNko0RmdydlcwQWRlMzRjSVVPcCtuZm1vaTlZN0dNdGpzS2ZPSWJtZjdnZ3pxCllSWDl1bkQwNXdKQkFOZUlFaDlGSzV3L05lbUpRaXY5bzB6YW9RUXV6WGE3QzdaU3F6RExsaCttWUhVNXBBRFUKalZHS056TnJEaUp6c1NrOWNwb1d0Nk5FdmVHVFNtWkdTUGtDUVFDWFhkQ1BMYUxQbmlFTnY2Z1RVc2Z5Wm1zawpkZnhTMndpb3B2V3VTZUpJTnlRZUErMmM1ZWRMdndsclRtbXg3eDg2NEd5TnJ0a1ZGNi9Dd2ZITHByR1JBa0VBCmxvYnUrUzNxL2szYlRrWlJrNzJwN2tRSERvL05hYTNLeVVSRlVXZnVhaDVkNGFFbkhIbFdWV3R0a0JpbG40UWoKYUFVRlkvNlh0SXlPL050TXE4OU1xUUpCQUpzZ0U4UmlCZXh1aEtLcjZCVjVsSzBMdjU2QlFDaGpkUS84TFFqZAppQWYwYlJ4RE1IS0lzVHFHSW15UzMwVTNvdVkrekxqSVQxb3Fibm0rTFY5VEdtcz0KLS0tLS1FTkQgUlNBIFBSSVZBVEUgS0VZLS0tLS0=`
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{GitHubApp: &schema.GitHubApp{PrivateKey: bogusKey}}})
	defer conf.Mock(nil)

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)

	externalServices := database.NewMockExternalServiceStore()

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	apiURL, err := url.Parse("https://github.com")
	require.NoError(t, err)

	client := NewMockGithubClient()
	client.GetAppInstallationFunc.SetDefaultReturn(
		&gogithub.Installation{
			Account: &gogithub.User{
				Login: gogithub.String("abc-org"),
			},
		},
		nil,
	)

	req, err := http.NewRequest(http.MethodGet, "/.setup/github-app-cloud?installation_id=21994992&setup_action=install&state=T3JnOjE%3D", nil)

	require.Nil(t, err)

	h := newGitHubAppSetupHandler(db, apiURL, client)

	t.Run("user not logged in (no actor in context)", func(t *testing.T) {
		resp := httptest.NewRecorder()
		h.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusFound, resp.Code)

		uri, err := url.ParseRequestURI(resp.Header().Get("Location"))
		require.Nil(t, err)
		queryVals := uri.Query()

		installationID := queryVals.Get("installation_id")

		decodedKey, err := base64.StdEncoding.DecodeString(bogusKey)
		require.Nil(t, err)

		installationIDBytes, err := base64.StdEncoding.DecodeString(installationID)
		require.Nil(t, err)

		decryptedID, err := DecryptWithPrivateKey(string(installationIDBytes), decodedKey)
		require.Nil(t, err)

		assert.Equal(t, decryptedID, "21994992")
	})

	t.Run("invalid setup action", func(t *testing.T) {
		resp := httptest.NewRecorder()
		badReq, err := http.NewRequest(http.MethodGet, "/.setup/github-app-cloud?installation_id=21994992&setup_action=incorrect&state=T3JnOjE%3D", nil)
		require.Nil(t, err)

		h.ServeHTTP(resp, badReq)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, `Invalid setup action "incorrect"`, resp.Body.String())
	})

	ctx := sgactor.WithActor(req.Context(), &sgactor.Actor{UID: 1})
	req = req.WithContext(ctx)

	t.Run("create new", func(t *testing.T) {
		externalServices.UpsertFunc.SetDefaultHook(func(ctx context.Context, svcs ...*types.ExternalService) error {
			require.Len(t, svcs, 1)

			svc := svcs[0]
			assert.Equal(t, extsvc.KindGitHub, svc.Kind)
			assert.Equal(t, "GitHub (abc-org)", svc.DisplayName)

			wantConfig := extsvc.NewUnencryptedConfig(`{
  "url": "https://github.com",
  "repos": [],
  "githubAppInstallationID": "21994992",
  "pending": false
} `)
			assert.Equal(t, wantConfig, svc.Config)
			return nil
		})

		resp := httptest.NewRecorder()
		h.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusFound, resp.Code)
		assert.True(t, strings.Contains(resp.Header().Get("Location"), "/install-github-app-success?installation_id="))

		mockrequire.Called(t, externalServices.UpsertFunc)
	})
}
