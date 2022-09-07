package github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestNewAuthzProviders(t *testing.T) {
	t.Run("no authorization", func(t *testing.T) {
		providers, problems, warnings, invalidConnections := NewAuthzProviders(
			database.NewMockExternalServiceStore(),
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

		assert := assert.New(t)

		assert.Len(providers, 0, "unexpected a providers: %+v", providers)
		assert.Len(problems, 0, "unexpected problems: %+v", problems)
		assert.Len(warnings, 0, "unexpected warnings: %+v", warnings)
		assert.Len(invalidConnections, 0, "unexpected invalidConnections: %+v", invalidConnections)
	})

	t.Run("no matching auth provider", func(t *testing.T) {
		licensing.MockCheckFeatureError("")
		providers, problems, warnings, invalidConnections := NewAuthzProviders(
			database.NewMockExternalServiceStore(),
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

		require.Len(t, providers, 1, "expect exactly one provider")
		assert.NotNil(t, providers[0])

		assert.Empty(t, problems)
		assert.Empty(t, invalidConnections)

		require.Len(t, warnings, 1, "expect exactly one warning")
		assert.Contains(t, warnings[0], "no authentication provider")
	})

	t.Run("matching auth provider found", func(t *testing.T) {
		t.Run("default case", func(t *testing.T) {
			licensing.MockCheckFeatureError("")
			providers, problems, warnings, invalidConnections := NewAuthzProviders(
				database.NewMockExternalServiceStore(),
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

			require.Len(t, providers, 1, "expect exactly one provider")
			assert.NotNil(t, providers[0])

			assert.Empty(t, problems)
			assert.Empty(t, warnings)
			assert.Empty(t, invalidConnections)
		})

		t.Run("license does not have ACLs feature", func(t *testing.T) {
			licensing.MockCheckFeatureError("failed")
			providers, problems, warnings, invalidConnections := NewAuthzProviders(
				database.NewMockExternalServiceStore(),
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
			assert.Equal(t, expectedError, problems)
			assert.Equal(t, expInvalidConnectionErr, invalidConnections)
			assert.Empty(t, providers)
			assert.Empty(t, warnings)
		})

		t.Run("groups cache enabled, but not allowGroupsPermissionsSync", func(t *testing.T) {
			licensing.MockCheckFeatureError("")
			providers, problems, warnings, invalidConnections := NewAuthzProviders(
				database.NewMockExternalServiceStore(),
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

			require.Len(t, providers, 1, "expect exactly one provider")
			assert.NotNil(t, providers[0])
			assert.Nil(t, providers[0].(*Provider).groupsCache, "expect groups cache to be disabled")

			assert.Empty(t, problems)

			require.Len(t, warnings, 1, "expect exactly one warning")
			assert.Contains(t, warnings[0], "allowGroupsPermissionsSync")
			assert.Empty(t, invalidConnections)
		})

		t.Run("groups cache and allowGroupsPermissionsSync enabled", func(t *testing.T) {
			github.MockGetAuthenticatedOAuthScopes = func(context.Context) ([]string, error) {
				return []string{"read:org"}, nil
			}
			providers, problems, warnings, invalidConnections := NewAuthzProviders(
				database.NewMockExternalServiceStore(),
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

			require.Len(t, providers, 1, "expect exactly one provider")
			assert.NotNil(t, providers[0])
			assert.NotNil(t, providers[0].(*Provider).groupsCache, "expect groups cache to be enabled")

			assert.Empty(t, problems)
			assert.Empty(t, warnings)
			assert.Empty(t, invalidConnections)
		})

		t.Run("github app installation id available", func(t *testing.T) {
			const bogusKey = `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlCUEFJQkFBSkJBUEpIaWprdG1UMUlLYUd0YTVFZXAzQVo5Q2VPZUw4alBESUZUN3dRZ0tabXQzRUZxRGhCCk93bitRVUhKdUs5Zm92UkROSmVWTDJvWTVCT0l6NHJ3L0cwQ0F3RUFBUUpCQU1BK0o5Mks0d2NQVllsbWMrM28KcHU5NmlKTkNwMmp5Nm5hK1pEQlQzK0VvSUo1VFJGdnN3R2kvTHUzZThYUWwxTDNTM21ub0xPSlZNcTF0bUxOMgpIY0VDSVFEK3daeS83RlYxUEFtdmlXeWlYVklETzJnNWJOaUJlbmdKQ3hFa3Nia1VtUUloQVBOMlZaczN6UFFwCk1EVG9vTlJXcnl0RW1URERkamdiOFpzTldYL1JPRGIxQWlCZWNKblNVQ05TQllLMXJ5VTFmNURTbitoQU9ZaDkKWDFBMlVnTDE3bWhsS1FJaEFPK2JMNmRDWktpTGZORWxmVnRkTUtxQnFjNlBIK01heFU2VzlkVlFvR1dkQWlFQQptdGZ5cE9zYTFiS2hFTDg0blovaXZFYkJyaVJHalAya3lERHYzUlg0V0JrPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=`

			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					GitHubApp: &schema.GitHubApp{
						AppID:        "1234",
						ClientID:     "1234",
						ClientSecret: "1234",
						Slug:         "test-app",
						PrivateKey:   bogusKey,
					},
				},
			})
			defer conf.Mock(nil)

			providers, problems, warnings, invalidConnections := NewAuthzProviders(
				database.NewMockExternalServiceStore(),
				[]*ExternalConnection{
					{
						ExternalService: &types.ExternalService{
							ID:     1,
							Kind:   extsvc.KindGitHub,
							Config: extsvc.NewEmptyConfig(),
						},
						GitHubConnection: &types.GitHubConnection{
							URN: "extsvc:github:1",
							GitHubConnection: &schema.GitHubConnection{
								Url: schema.DefaultGitHubURL,
								Authorization: &schema.GitHubAuthorization{
									GroupsCacheTTL: 72,
								},
								GithubAppInstallationID: "1234",
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

			require.Len(t, providers, 1, "expect exactly one provider")
			assert.NotNil(t, providers[0])

			assert.Empty(t, problems)
			assert.Empty(t, warnings)
			assert.Empty(t, invalidConnections)
		})
	})
}
