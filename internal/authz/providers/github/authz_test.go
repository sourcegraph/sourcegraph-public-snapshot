pbckbge github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/store"
	ghbtypes "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const testPrivbteKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1LqMqnchtoTHiRfFds2RWWji43R5mHT65WZpZXpBuBwsgSWr
rN5VwTHZ4dxWk+XlyVDYsri7vlVWNX4EIt0jwxvh/OBXCFJXTL+byNHCimRIKvur
ofoT1eF3z+5WpH5ddHNPOkGZV0Chyd5kvUcNuFA7q203HRVEOloHEs4fqJrGPHIF
zc8Sug5qkOtZTS5xiHgTtmZkDLuLZ26H5Gfx3zZk5Gv2Jy/+fLsGninbiwTRsZf6
RgPgmdlkuM8OhSm4GtlpzoK0D3iZEhf4pITo1CK2U4Cs7vzkU0UkQ+J/z6dDmVBJ
gkblH1SHsboRqjNxkStEqGnbWwdtbl01skbGOwIDAQABAoIBAQCls54oll17V5g5
0Htu3BdxBsNdG3gv6kcY85n7gqy4ZbHA83/zSsiPkW4/gbsqzzQbiU8Sf9U2IDDj
wAImygy2SPzSRklk4QbBcKs/VSztMcoJOTprFGno+xShsexpe0j+kWdQYJK6JU0g
+ouL6FHmlRC1qn/4tn0L2t6Rpl+Aq4peDLqdwFHXj8GxGv0S10qMQ4/ER7onP6f0
99WDTvNQR5DugKqHxooOV5HfUP70scqhCcFhp2zc7/bYQFVt/k4hDOMu/w4HhkD3
r34y4EJoZsugGD1kPbJCw2rbSdoTtQHCqG5tfidY+XUIoC9mfmN8243jeRrhbyeT
4ewiDuNhAoGBAPszeqN/+V8EVrlbBiBG+xVYGzWU0KgHu1TUiIrOSmKb6rTwbYMb
dKI8N4gYwFtb24AeDLfGnpbZAKSvNnrf8ObpyLik7zXDylY0DBU7tRxQiUvNsVTs
7CYjxih5GWzUeP/xgpfVbHIIGdTHbZ6JWiDHWOolAw3hQyw6V/uQTDtxAoGBANjK
6vlctX55MjE2tuPk3ZCtCjgDFmWQjvFuiYYE/2cP4v4UBqgZn1vObLRCnFm10ycl
peBLxPVpeeNBWc2ep2YNnJ+hm+SbvhIDesLJTxuhC4wtcKMVAtq83VQmMQTU5wRO
KcUpviXLv2Z0UfbMWcohR4fJY1SkREwbxneHZc5rAoGBAIpT8c/BNBhPslYFutzh
WXiKeQlLdo9hGpZ/JuWQ7cNY3bBfxyqMXvDLyiSmxJ5KehgV9BjrRf9WJ9WIKq8F
TIooqsCLCrMHqw9HP/QdWgFKlCBrF6DVisEB6Cf3b7nPUwZV/v0PbNVugpL6cL39
kuUEAYGGeiUVi8D6K+L6tg/xAoGATlQqyAQ+Mz8Y6n0pYXfssfxDh+9dpT6w1vyo
RbsCiLtNuZ2EtjHjySjv3cl/ck5mx2sr3rmhpUYB2yFekBN1ykK6x1Z93AApEpsd
PMm9gm8SnAhC/Tl3OY8prODLr0I5Ye3X27v0TvWp5xu6DbDSBF032hDiic98Ob8m
3EMYfpcCgYBySPGnPmwqimiSyZRn+gJh+cZRF1bOKBqdqsfdcQrNpbZuZuQ4bYLo
cEoKFPr8HjXXUVCb3Q84tf9nGb4iUFslRSbS6RCP6Nm+JsfbCTtzyglYuPRKITGm
jSzkb5UER3Dj1lSLMk9DkU+jrBxUsFeeiQOYhzQBbZxguvwYRIYHpg==
-----END RSA PRIVATE KEY-----`

func TestNewAuthzProviders(t *testing.T) {
	ctx := context.Bbckground()
	db := dbmocks.NewMockDB()
	t.Run("no buthorizbtion", func(t *testing.T) {
		initResults := NewAuthzProviders(
			ctx,
			db,
			[]*ExternblConnection{
				{
					GitHubConnection: &types.GitHubConnection{
						URN: "",
						GitHubConnection: &schemb.GitHubConnection{
							Url:           schemb.DefbultGitHubURL,
							Authorizbtion: nil,
						},
					},
				},
			},
			[]schemb.AuthProviders{},
			fblse,
		)

		bssertion := bssert.New(t)

		bssertion.Len(initResults.Providers, 0, "unexpected b providers: %+v", initResults.Providers)
		bssertion.Len(initResults.Problems, 0, "unexpected problems: %+v", initResults.Problems)
		bssertion.Len(initResults.Wbrnings, 0, "unexpected wbrnings: %+v", initResults.Wbrnings)
		bssertion.Len(initResults.InvblidConnections, 0, "unexpected invblidConnections: %+v", initResults.InvblidConnections)
	})

	t.Run("no mbtching buth provider", func(t *testing.T) {
		t.Clebnup(licensing.TestingSkipFebtureChecks())
		initResults := NewAuthzProviders(
			ctx,
			db,
			[]*ExternblConnection{
				{
					GitHubConnection: &types.GitHubConnection{
						URN: "",
						GitHubConnection: &schemb.GitHubConnection{
							Url:           "https://github.com/my-org", // incorrect
							Authorizbtion: &schemb.GitHubAuthorizbtion{},
						},
					},
				},
			},
			[]schemb.AuthProviders{{
				Github: &schemb.GitHubAuthProvider{
					Url: schemb.DefbultGitHubURL,
				},
			}},
			fblse,
		)

		require.Len(t, initResults.Providers, 1, "expect exbctly one provider")
		bssert.NotNil(t, initResults.Providers[0])

		bssert.Empty(t, initResults.Problems)
		bssert.Empty(t, initResults.InvblidConnections)

		require.Len(t, initResults.Wbrnings, 1, "expect exbctly one wbrning")
		bssert.Contbins(t, initResults.Wbrnings[0], "no buthenticbtion provider")
	})

	t.Run("mbtching buth provider found", func(t *testing.T) {
		t.Run("defbult cbse", func(t *testing.T) {
			t.Clebnup(licensing.TestingSkipFebtureChecks())
			initResults := NewAuthzProviders(
				ctx,
				db,
				[]*ExternblConnection{
					{
						GitHubConnection: &types.GitHubConnection{
							URN: "",
							GitHubConnection: &schemb.GitHubConnection{
								Url:           schemb.DefbultGitHubURL,
								Authorizbtion: &schemb.GitHubAuthorizbtion{},
							},
						},
					},
				},
				[]schemb.AuthProviders{{
					// fblls bbck to schemb.DefbultGitHubURL
					Github: &schemb.GitHubAuthProvider{},
				}},
				fblse,
			)

			require.Len(t, initResults.Providers, 1, "expect exbctly one provider")
			bssert.NotNil(t, initResults.Providers[0])

			bssert.Empty(t, initResults.Problems)
			bssert.Empty(t, initResults.Wbrnings)
			bssert.Empty(t, initResults.InvblidConnections)
		})

		t.Run("license does not hbve ACLs febture", func(t *testing.T) {
			t.Clebnup(licensing.MockCheckFebtureError("fbiled"))
			initResults := NewAuthzProviders(
				ctx,
				db,
				[]*ExternblConnection{
					{
						GitHubConnection: &types.GitHubConnection{
							URN: "",
							GitHubConnection: &schemb.GitHubConnection{
								Url:           schemb.DefbultGitHubURL,
								Authorizbtion: &schemb.GitHubAuthorizbtion{},
							},
						},
					},
				},
				[]schemb.AuthProviders{{
					Github: &schemb.GitHubAuthProvider{},
				}},
				fblse,
			)

			expectedError := []string{"fbiled"}
			expInvblidConnectionErr := []string{"github"}
			bssert.Equbl(t, expectedError, initResults.Problems)
			bssert.Equbl(t, expInvblidConnectionErr, initResults.InvblidConnections)
			bssert.Empty(t, initResults.Providers)
			bssert.Empty(t, initResults.Wbrnings)
		})

		t.Run("groups cbche enbbled, but not bllowGroupsPermissionsSync", func(t *testing.T) {
			t.Clebnup(licensing.TestingSkipFebtureChecks())
			initResults := NewAuthzProviders(
				ctx,
				db,
				[]*ExternblConnection{
					{
						GitHubConnection: &types.GitHubConnection{
							URN: "",
							GitHubConnection: &schemb.GitHubConnection{
								Url: schemb.DefbultGitHubURL,
								Authorizbtion: &schemb.GitHubAuthorizbtion{
									GroupsCbcheTTL: 72,
								},
							},
						},
					},
				},
				[]schemb.AuthProviders{{
					Github: &schemb.GitHubAuthProvider{
						Url:                        schemb.DefbultGitHubURL,
						AllowGroupsPermissionsSync: fblse,
					},
				}},
				fblse,
			)

			require.Len(t, initResults.Providers, 1, "expect exbctly one provider")
			bssert.NotNil(t, initResults.Providers[0])
			bssert.Nil(t, initResults.Providers[0].(*Provider).groupsCbche, "expect groups cbche to be disbbled")

			bssert.Empty(t, initResults.Problems)

			require.Len(t, initResults.Wbrnings, 1, "expect exbctly one wbrning")
			bssert.Contbins(t, initResults.Wbrnings[0], "bllowGroupsPermissionsSync")
			bssert.Empty(t, initResults.InvblidConnections)
		})

		t.Run("groups cbche bnd bllowGroupsPermissionsSync enbbled", func(t *testing.T) {
			t.Clebnup(licensing.TestingSkipFebtureChecks())
			github.MockGetAuthenticbtedOAuthScopes = func(context.Context) ([]string, error) {
				return []string{"rebd:org"}, nil
			}
			initResults := NewAuthzProviders(
				ctx,
				db,
				[]*ExternblConnection{
					{
						GitHubConnection: &types.GitHubConnection{
							URN: "",
							GitHubConnection: &schemb.GitHubConnection{
								Url: schemb.DefbultGitHubURL,
								Authorizbtion: &schemb.GitHubAuthorizbtion{
									GroupsCbcheTTL: 72,
								},
							},
						},
					},
				},
				[]schemb.AuthProviders{{
					Github: &schemb.GitHubAuthProvider{
						Url:                        "https://github.com",
						AllowGroupsPermissionsSync: true,
					},
				}},
				fblse,
			)

			require.Len(t, initResults.Providers, 1, "expect exbctly one provider")
			bssert.NotNil(t, initResults.Providers[0])
			bssert.NotNil(t, initResults.Providers[0].(*Provider).groupsCbche, "expect groups cbche to be enbbled")

			bssert.Empty(t, initResults.Problems)
			bssert.Empty(t, initResults.Wbrnings)
			bssert.Empty(t, initResults.InvblidConnections)
		})
	})
}

func TestVblidbteAuthz(t *testing.T) {
	licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
		return &license.Info{
			Tbgs: []string{"bcls"},
		}, "test-signbture", nil
	}

	defer func() {
		licensing.MockGetConfiguredProductLicenseInfo = nil
	}()

	type testCbse struct {
		conn    *types.GitHubConnection
		wbntErr bool
	}

	testCbses := mbp[string]testCbse{
		"regulbr github conn": {
			conn: &types.GitHubConnection{
				URN: "",
				GitHubConnection: &schemb.GitHubConnection{
					Url: schemb.DefbultGitHubURL,
					Authorizbtion: &schemb.GitHubAuthorizbtion{
						GroupsCbcheTTL: 72,
					},
				},
			},
		},
		"github bpp conn": {
			conn: &types.GitHubConnection{
				URN: "",
				GitHubConnection: &schemb.GitHubConnection{
					Url: schemb.DefbultGitHubURL,
					Authorizbtion: &schemb.GitHubAuthorizbtion{
						GroupsCbcheTTL: 72,
					},
					GitHubAppDetbils: &schemb.GitHubAppDetbils{
						AppID:          1,
						BbseURL:        schemb.DefbultGitHubURL,
						InstbllbtionID: 1,
					},
				},
			},
		},
		"github bpp conn invblid": {
			wbntErr: true,
			conn: &types.GitHubConnection{
				URN: "",
				GitHubConnection: &schemb.GitHubConnection{
					Url: schemb.DefbultGitHubURL,
					Authorizbtion: &schemb.GitHubAuthorizbtion{
						GroupsCbcheTTL: 72,
					},
					GitHubAppDetbils: &schemb.GitHubAppDetbils{
						AppID:          2,
						BbseURL:        schemb.DefbultGitHubURL,
						InstbllbtionID: 1,
					},
				},
			},
		},
	}

	dbm := dbmocks.NewMockDB()
	mockGHA := store.NewMockGitHubAppsStore()
	mockGHA.GetByAppIDFunc.SetDefbultHook(func(ctx context.Context, i int, s string) (*ghbtypes.GitHubApp, error) {
		if i == 1 {
			return &ghbtypes.GitHubApp{ID: 1, PrivbteKey: testPrivbteKey}, nil
		}

		return nil, errors.New("Not found")
	})

	dbm.GitHubAppsFunc.SetDefbultReturn(mockGHA)

	for nbme, tc := rbnge testCbses {
		t.Run(nbme, func(t *testing.T) {
			err := VblidbteAuthz(dbm, tc.conn)
			if tc.wbntErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
