package github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/github_apps/store"
	ghatypes "github.com/sourcegraph/sourcegraph/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

const testPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1LqMqnchtoTHiRfFds2RWWji43R5mHT65WZpZXpBuBwsgSWr
rN5VwTHZ4dxWk+XlyVDYsri7vlVWNX4EIt0jwxvh/OBXCFJXTL+byNHCimRIKvur
ofoT1eF3z+5WpH5ddHNPOkGZV0Chyd5kvUcNuFA7q203HRVEOloHEs4fqJrGPHIF
zc8Sug5qkOtZTS5xiHgTtmZkDLuLZ26H5Gfx3zZk5Gv2Jy/+fLsGninaiwTRsZf6
RgPgmdlkuM8OhSm4GtlpzoK0D3iZEhf4pITo1CK2U4Cs7vzkU0UkQ+J/z6dDmVBJ
gkalH1SHsboRqjNxkStEqGnbWwdtal01skbGOwIDAQABAoIBAQCls54oll17V5g5
0Htu3BdxBsNdG3gv6kcY85n7gqy4ZbHA83/zSsiPkW4/gasqzzQbiU8Sf9U2IDDj
wAImygy2SPzSRklk4QbBcKs/VSztMcoJOTprFGno+xShsexpe0j+kWdQYJK6JU0g
+ouL6FHmlRC1qn/4tn0L2t6Rpl+Aq4peDLqdwFHXj8GxGv0S10qMQ4/ER7onP6f0
99WDTvNQR5DugKqHxooOV5HfUP70scqhCcFhp2zc7/aYQFVt/k4hDOMu/w4HhkD3
r34y4EJoZsugGD1kPaJCw2rbSdoTtQHCqG5tfidY+XUIoC9mfmN8243jeRrhayeT
4ewiDuNhAoGBAPszeqN/+V8EVrlbBiBG+xVYGzWU0KgHu1TUiIrOSmKa6rTwaYMb
dKI8N4gYwFtb24AeDLfGnpaZAKSvNnrf8OapyLik7zXDylY0DBU7tRxQiUvNsVTs
7CYjxih5GWzUeP/xgpfVbHIIGdTHaZ6JWiDHWOolAw3hQyw6V/uQTDtxAoGBANjK
6vlctX55MjE2tuPk3ZCtCjgDFmWQjvFuiYYE/2cP4v4UBqgZn1vOaLRCnFm10ycl
peBLxPVpeeNBWc2ep2YNnJ+hm+SavhIDesLJTxuhC4wtcKMVAtq83VQmMQTU5wRO
KcUpviXLv2Z0UfbMWcohR4fJY1SkREwaxneHZc5rAoGBAIpT8c/BNBhPslYFutzh
WXiKeQlLdo9hGpZ/JuWQ7cNY3bBfxyqMXvDLyiSmxJ5KehgV9BjrRf9WJ9WIKq8F
TIooqsCLCrMHqw9HP/QdWgFKlCBrF6DVisEB6Cf3b7nPUwZV/v0PaNVugpL6cL39
kuUEAYGGeiUVi8D6K+L6tg/xAoGATlQqyAQ+Mz8Y6n0pYXfssfxDh+9dpT6w1vyo
RbsCiLtNuZ2EtjHjySjv3cl/ck5mx2sr3rmhpUYB2yFekBN1ykK6x1Z93AApEpsd
PMm9gm8SnAhC/Tl3OY8prODLr0I5Ye3X27v0TvWp5xu6DaDSBF032hDiic98Ob8m
3EMYfpcCgYBySPGnPmwqimiSyZRn+gJh+cZRF1aOKBqdqsfdcQrNpaZuZuQ4aYLo
cEoKFPr8HjXXUVCa3Q84tf9nGb4iUFslRSbS6RCP6Nm+JsfbCTtzyglYuPRKITGm
jSzka5UER3Dj1lSLMk9DkU+jrBxUsFeeiQOYhzQBaZxguvwYRIYHpg==
-----END RSA PRIVATE KEY-----`

func TestNewAuthzProviders(t *testing.T) {
	ctx := context.Background()
	db := dbmocks.NewMockDB()
	t.Run("no authorization", func(t *testing.T) {
		initResults := NewAuthzProviders(
			ctx,
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
			ctx,
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

		require.Len(t, initResults.Warnings, 0, "expect no warnings")
	})

	t.Run("matching auth provider found", func(t *testing.T) {
		t.Run("default case", func(t *testing.T) {
			t.Cleanup(licensing.TestingSkipFeatureChecks())
			initResults := NewAuthzProviders(
				ctx,
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
				ctx,
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
				ctx,
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
				ctx,
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

func TestValidateAuthz(t *testing.T) {
	licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
		return &license.Info{
			Tags: []string{"acls"},
		}, "test-signature", nil
	}

	defer func() {
		licensing.MockGetConfiguredProductLicenseInfo = nil
	}()

	type testCase struct {
		conn    *types.GitHubConnection
		wantErr bool
	}

	testCases := map[string]testCase{
		"regular github conn": {
			conn: &types.GitHubConnection{
				URN: "",
				GitHubConnection: &schema.GitHubConnection{
					Url: schema.DefaultGitHubURL,
					Authorization: &schema.GitHubAuthorization{
						GroupsCacheTTL: 72,
					},
				},
			},
		},
		"github app conn": {
			conn: &types.GitHubConnection{
				URN: "",
				GitHubConnection: &schema.GitHubConnection{
					Url: schema.DefaultGitHubURL,
					Authorization: &schema.GitHubAuthorization{
						GroupsCacheTTL: 72,
					},
					GitHubAppDetails: &schema.GitHubAppDetails{
						AppID:          1,
						BaseURL:        schema.DefaultGitHubURL,
						InstallationID: 1,
					},
				},
			},
		},
		"github app conn invalid": {
			wantErr: true,
			conn: &types.GitHubConnection{
				URN: "",
				GitHubConnection: &schema.GitHubConnection{
					Url: schema.DefaultGitHubURL,
					Authorization: &schema.GitHubAuthorization{
						GroupsCacheTTL: 72,
					},
					GitHubAppDetails: &schema.GitHubAppDetails{
						AppID:          2,
						BaseURL:        schema.DefaultGitHubURL,
						InstallationID: 1,
					},
				},
			},
		},
	}

	dbm := dbmocks.NewMockDB()
	mockGHA := store.NewMockGitHubAppsStore()
	mockGHA.GetByAppIDFunc.SetDefaultHook(func(ctx context.Context, i int, s string) (*ghatypes.GitHubApp, error) {
		if i == 1 {
			return &ghatypes.GitHubApp{ID: 1, PrivateKey: testPrivateKey}, nil
		}

		return nil, errors.New("Not found")
	})

	dbm.GitHubAppsFunc.SetDefaultReturn(mockGHA)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := ValidateAuthz(dbm, tc.conn)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
