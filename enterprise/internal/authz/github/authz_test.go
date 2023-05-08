package github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestNewAuthzProviders(t *testing.T) {
	db := database.NewMockDB()
	t.Run("no authorization", func(t *testing.T) {
		initResults := NewAuthzProviders(
			db,
			[]*ExternalConnection{
				{
					GitHubConnection: &types.GitHubConnection{
						URN: "",
						GitHubConnection: &schema.GitHubConnection{
							Url:           schema.DefaultGitHubURL,
							Authorization: nil,
						},
					},
				},
			},
			[]schema.AuthProviders{},
			false,
		)

		assertion := assert.New(t)

		assertion.Len(initResults.Providers, 0, "unexpected a providers: %+v", initResults.Providers)
		assertion.Len(initResults.Problems, 0, "unexpected problems: %+v", initResults.Problems)
		assertion.Len(initResults.Warnings, 0, "unexpected warnings: %+v", initResults.Warnings)
		assertion.Len(initResults.InvalidConnections, 0, "unexpected invalidConnections: %+v", initResults.InvalidConnections)
	})

	t.Run("no matching auth provider", func(t *testing.T) {
		t.Cleanup(licensing.TestingSkipFeatureChecks())
		initResults := NewAuthzProviders(
			db,
			[]*ExternalConnection{
				{
					GitHubConnection: &types.GitHubConnection{
						URN: "",
						GitHubConnection: &schema.GitHubConnection{
							Url:           "https://github.com/my-org", // incorrect
							Authorization: &schema.GitHubAuthorization{},
						},
					},
				},
			},
			[]schema.AuthProviders{{
				Github: &schema.GitHubAuthProvider{
					Url: schema.DefaultGitHubURL,
				},
			}},
			false,
		)

		require.Len(t, initResults.Providers, 1, "expect exactly one provider")
		assert.NotNil(t, initResults.Providers[0])

		assert.Empty(t, initResults.Problems)
		assert.Empty(t, initResults.InvalidConnections)

		require.Len(t, initResults.Warnings, 1, "expect exactly one warning")
		assert.Contains(t, initResults.Warnings[0], "no authentication provider")
	})

	t.Run("matching auth provider found", func(t *testing.T) {
		t.Run("default case", func(t *testing.T) {
			t.Cleanup(licensing.TestingSkipFeatureChecks())
			initResults := NewAuthzProviders(
				db,
				[]*ExternalConnection{
					{
						GitHubConnection: &types.GitHubConnection{
							URN: "",
							GitHubConnection: &schema.GitHubConnection{
								Url:           schema.DefaultGitHubURL,
								Authorization: &schema.GitHubAuthorization{},
							},
						},
					},
				},
				[]schema.AuthProviders{{
					// falls back to schema.DefaultGitHubURL
					Github: &schema.GitHubAuthProvider{},
				}},
				false,
			)

			require.Len(t, initResults.Providers, 1, "expect exactly one provider")
			assert.NotNil(t, initResults.Providers[0])

			assert.Empty(t, initResults.Problems)
			assert.Empty(t, initResults.Warnings)
			assert.Empty(t, initResults.InvalidConnections)
		})

		t.Run("license does not have ACLs feature", func(t *testing.T) {
			t.Cleanup(licensing.MockCheckFeatureError("failed"))
			initResults := NewAuthzProviders(
				db,
				[]*ExternalConnection{
					{
						GitHubConnection: &types.GitHubConnection{
							URN: "",
							GitHubConnection: &schema.GitHubConnection{
								Url:           schema.DefaultGitHubURL,
								Authorization: &schema.GitHubAuthorization{},
							},
						},
					},
				},
				[]schema.AuthProviders{{
					Github: &schema.GitHubAuthProvider{},
				}},
				false,
			)

			expectedError := []string{"failed"}
			expInvalidConnectionErr := []string{"github"}
			assert.Equal(t, expectedError, initResults.Problems)
			assert.Equal(t, expInvalidConnectionErr, initResults.InvalidConnections)
			assert.Empty(t, initResults.Providers)
			assert.Empty(t, initResults.Warnings)
		})

		t.Run("groups cache enabled, but not allowGroupsPermissionsSync", func(t *testing.T) {
			t.Cleanup(licensing.TestingSkipFeatureChecks())
			initResults := NewAuthzProviders(
				db,
				[]*ExternalConnection{
					{
						GitHubConnection: &types.GitHubConnection{
							URN: "",
							GitHubConnection: &schema.GitHubConnection{
								Url: schema.DefaultGitHubURL,
								Authorization: &schema.GitHubAuthorization{
									GroupsCacheTTL: 72,
								},
							},
						},
					},
				},
				[]schema.AuthProviders{{
					Github: &schema.GitHubAuthProvider{
						Url:                        schema.DefaultGitHubURL,
						AllowGroupsPermissionsSync: false,
					},
				}},
				false,
			)

			require.Len(t, initResults.Providers, 1, "expect exactly one provider")
			assert.NotNil(t, initResults.Providers[0])
			assert.Nil(t, initResults.Providers[0].(*Provider).groupsCache, "expect groups cache to be disabled")

			assert.Empty(t, initResults.Problems)

			require.Len(t, initResults.Warnings, 1, "expect exactly one warning")
			assert.Contains(t, initResults.Warnings[0], "allowGroupsPermissionsSync")
			assert.Empty(t, initResults.InvalidConnections)
		})

		t.Run("groups cache and allowGroupsPermissionsSync enabled", func(t *testing.T) {
			t.Cleanup(licensing.TestingSkipFeatureChecks())
			github.MockGetAuthenticatedOAuthScopes = func(context.Context) ([]string, error) {
				return []string{"read:org"}, nil
			}
			initResults := NewAuthzProviders(
				db,
				[]*ExternalConnection{
					{
						GitHubConnection: &types.GitHubConnection{
							URN: "",
							GitHubConnection: &schema.GitHubConnection{
								Url: schema.DefaultGitHubURL,
								Authorization: &schema.GitHubAuthorization{
									GroupsCacheTTL: 72,
								},
							},
						},
					},
				},
				[]schema.AuthProviders{{
					Github: &schema.GitHubAuthProvider{
						Url:                        "https://github.com",
						AllowGroupsPermissionsSync: true,
					},
				}},
				false,
			)

			require.Len(t, initResults.Providers, 1, "expect exactly one provider")
			assert.NotNil(t, initResults.Providers[0])
			assert.NotNil(t, initResults.Providers[0].(*Provider).groupsCache, "expect groups cache to be enabled")

			assert.Empty(t, initResults.Problems)
			assert.Empty(t, initResults.Warnings)
			assert.Empty(t, initResults.InvalidConnections)
		})
	})
}
