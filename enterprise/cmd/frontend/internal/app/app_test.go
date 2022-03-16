package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	gogithub "github.com/google/go-github/v41/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	a "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestNewGitHubAppCloudSetupHandler(t *testing.T) {
	orig := envvar.SourcegraphDotComMode()
	envvar.MockSourcegraphDotComMode(true)
	defer envvar.MockSourcegraphDotComMode(orig)

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)
	orgMembers := database.NewMockOrgMemberStore()
	orgs := database.NewMockOrgStore()
	orgs.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.Org, error) {
		return &types.Org{
			ID:   id,
			Name: "abc-org",
		}, nil
	})
	externalServices := database.NewMockExternalServiceStore()
	featureFlags := database.NewMockFeatureFlagStore()
	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.OrgMembersFunc.SetDefaultReturn(orgMembers)
	db.OrgsFunc.SetDefaultReturn(orgs)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)

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

	h := newGitHubAppCloudSetupHandler(db, apiURL, client)

	t.Run("user not logged in (no actor in context)", func(t *testing.T) {
		resp := httptest.NewRecorder()
		h.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusFound, resp.Code)
		assert.Equal(t, "/install-github-app-success", resp.Header().Get("Location"))
	})

	t.Run("invalid setup action", func(t *testing.T) {
		resp := httptest.NewRecorder()
		badReq, err := http.NewRequest(http.MethodGet, "/.setup/github-app-cloud?installation_id=21994992&setup_action=incorrect&state=T3JnOjE%3D", nil)
		require.Nil(t, err)

		h.ServeHTTP(resp, badReq)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "Invalid setup action 'incorrect'", resp.Body.String())
	})

	ctx := a.WithActor(req.Context(), &a.Actor{UID: 1})
	req = req.WithContext(ctx)

	t.Run("feature flag not enabled", func(t *testing.T) {
		resp := httptest.NewRecorder()
		h.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusForbidden, resp.Code)
		assert.Equal(t, "Sourcegraph Cloud GitHub App setup is not enabled for the organization", resp.Body.String())
	})
	featureFlags.GetOrgFeatureFlagFunc.SetDefaultReturn(true, nil)

	t.Run("not an organization member", func(t *testing.T) {
		orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultReturn(nil, nil)

		resp := httptest.NewRecorder()
		h.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "the authenticated user does not belong to the organization requested", resp.Body.String())
	})

	t.Run("create new", func(t *testing.T) {
		orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultReturn(&types.OrgMembership{}, nil)
		externalServices.UpsertFunc.SetDefaultHook(func(ctx context.Context, svcs ...*types.ExternalService) error {
			require.Len(t, svcs, 1)

			svc := svcs[0]
			assert.Equal(t, extsvc.KindGitHub, svc.Kind)
			assert.Equal(t, "GitHub (abc-org)", svc.DisplayName)

			wantConfig := `
{
  "url": "https://github.com",
  "repos": [],
  "githubAppInstallationID": "21994992",
  "pending": false
}
`
			assert.Equal(t, wantConfig, svc.Config)
			return nil
		})

		resp := httptest.NewRecorder()
		h.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusFound, resp.Code)
		assert.Equal(t, "/organizations/abc-org/settings/code-hosts", resp.Header().Get("Location"))

		mockrequire.Called(t, externalServices.UpsertFunc)
	})

	t.Run("update existing", func(t *testing.T) {
		externalServices.ListFunc.SetDefaultReturn(
			[]*types.ExternalService{
				{
					Kind:        extsvc.KindGitHub,
					DisplayName: "GitHub (old)",
					Config: `
{
  "url": "https://github.com",
  "repos": []
}
`,
				},
			},
			nil,
		)
		externalServices.UpsertFunc.SetDefaultHook(func(ctx context.Context, svcs ...*types.ExternalService) error {
			require.Len(t, svcs, 1)

			svc := svcs[0]
			assert.Equal(t, extsvc.KindGitHub, svc.Kind)
			assert.Equal(t, "GitHub (abc-org)", svc.DisplayName)

			wantConfig := `
{
  "url": "https://github.com",
  "repos": [],
  "githubAppInstallationID": "21994992",
  "pending": false
}
`
			assert.Equal(t, wantConfig, svc.Config)
			return nil
		})

		resp := httptest.NewRecorder()
		h.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusFound, resp.Code)
		assert.Equal(t, "/organizations/abc-org/settings/code-hosts", resp.Header().Get("Location"))

		mockrequire.Called(t, externalServices.ListFunc)
		mockrequire.Called(t, externalServices.UpsertFunc)
	})
}
