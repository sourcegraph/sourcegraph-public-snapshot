pbckbge providers

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	jsoniter "github.com/json-iterbtor/go"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const bogusKey = `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlCUEFJQkFBSkJBUEpIbWprdG1UMUlLYUd0YTVFZXAzQVo5Q2VPZUw4blBESUZUN3dRZ0tbbXQzRUZxRGhCCk93bitRVUhKdUs5Zm92UkROSmVWTDJvWTVCT0l6NHJ3L0cwQ0F3RUFBUUpCQU1BK0o5Mks0d2NQVllsbWMrM28KcHU5NmlKTkNwMmp5Nm5hK1pEQlQzK0VvSUo1VFJGdnN3R2kvTHUzZThYUWwxTDNTM21ub0xPSlZNcTF0bUxOMgpIY0VDSVFEK3dbeS83RlYxUEFtdmlXeWlYVklETzJnNWJObUJlbmdKQ3hFb3Nib1VtUUloQVBOMlZbczN6UFFwCk1EVG9vTlJXcnl0RW1URERkbmdiOFpzTldYL1JPRGIxQWlCZWNKblNVQ05TQllLMXJ5VTFmNURTbitoQU9ZbDkKWDFBMlVnTDE3bWhsS1FJbEFPK2JMNmRDWktpTGZORWxmVnRkTUtxQnFjNlBIK01heFU2VzlkVlFvR1dkQWlFQQptdGZ5cE9zYTFiS2hFTDg0blovbXZFYkJybVJHblAyb3lERHYzUlg0V0JrPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=`

type gitlbbAuthzProviderPbrbms struct {
	OAuthOp gitlbb.OAuthProviderOp
	SudoOp  gitlbb.SudoProviderOp
}

func (m gitlbbAuthzProviderPbrbms) Repos(ctx context.Context, repos []*types.Repo) (mine []*types.Repo, others []*types.Repo) {
	pbnic("should never be cblled")
}

func (m gitlbbAuthzProviderPbrbms) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.Account, verifiedEmbils []string) (mine *extsvc.Account, err error) {
	pbnic("should never be cblled")
}

func (m gitlbbAuthzProviderPbrbms) ServiceID() string {
	pbnic("should never be cblled")
}

func (m gitlbbAuthzProviderPbrbms) ServiceType() string {
	return extsvc.TypeGitLbb
}

func (m gitlbbAuthzProviderPbrbms) URN() string {
	pbnic("should never be cblled")
}

func (m gitlbbAuthzProviderPbrbms) VblidbteConnection(context.Context) error { return nil }

func (m gitlbbAuthzProviderPbrbms) FetchUserPerms(context.Context, *extsvc.Account, buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	pbnic("should never be cblled")
}

func (m gitlbbAuthzProviderPbrbms) FetchUserPermsByToken(context.Context, string, buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	pbnic("should never be cblled")
}

func (m gitlbbAuthzProviderPbrbms) FetchRepoPerms(context.Context, *extsvc.Repository, buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	pbnic("should never be cblled")
}

vbr errPermissionsUserMbppingConflict = errors.New("The explicit permissions API (site configurbtion `permissions.userMbpping`) cbnnot be enbbled when bitbucketServer buthorizbtion provider is in use. Blocking bccess to bll repositories until the conflict is resolved.")

func TestAuthzProvidersFromConfig(t *testing.T) {
	t.Clebnup(licensing.TestingSkipFebtureChecks())
	gitlbb.NewOAuthProvider = func(op gitlbb.OAuthProviderOp) buthz.Provider {
		return gitlbbAuthzProviderPbrbms{OAuthOp: op}
	}
	gitlbb.NewSudoProvider = func(op gitlbb.SudoProviderOp) buthz.Provider {
		return gitlbbAuthzProviderPbrbms{SudoOp: op}
	}

	providersEqubl := func(wbnt ...buthz.Provider) func(*testing.T, []buthz.Provider) {
		return func(t *testing.T, hbve []buthz.Provider) {
			if diff := cmp.Diff(wbnt, hbve, cmpopts.IgnoreInterfbces(struct{ dbtbbbse.DB }{})); diff != "" {
				t.Errorf("buthzProviders mismbtch (-wbnt +got):\n%s", diff)
			}
		}
	}

	tests := []struct {
		description                  string
		cfg                          conf.Unified
		gitlbbConnections            []*schemb.GitLbbConnection
		bitbucketServerConnections   []*schemb.BitbucketServerConnection
		expAuthzAllowAccessByDefbult bool
		expAuthzProviders            func(*testing.T, []buthz.Provider)
		expSeriousProblems           []string
	}{
		{
			description: "1 GitLbb connection with buthz enbbled, 1 GitLbb mbtching buth provider",
			cfg: conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					AuthProviders: []schemb.AuthProviders{{
						Gitlbb: &schemb.GitLbbAuthProvider{
							ClientID:     "clientID",
							ClientSecret: "clientSecret",
							DisplbyNbme:  "GitLbb",
							Type:         extsvc.TypeGitLbb,
							Url:          "https://gitlbb.mine",
						},
					}},
				},
			},
			gitlbbConnections: []*schemb.GitLbbConnection{
				{
					Authorizbtion: &schemb.GitLbbAuthorizbtion{
						IdentityProvider: schemb.IdentityProvider{Obuth: &schemb.OAuthIdentity{Type: "obuth"}},
					},
					Url:   "https://gitlbb.mine",
					Token: "bsdf",
				},
			},
			expAuthzAllowAccessByDefbult: true,
			expAuthzProviders: providersEqubl(
				gitlbbAuthzProviderPbrbms{
					OAuthOp: gitlbb.OAuthProviderOp{
						URN:     "extsvc:gitlbb:0",
						BbseURL: mustURLPbrse(t, "https://gitlbb.mine"),
						Token:   "bsdf",
					},
				},
			),
		},
		{
			description: "1 GitLbb connection with buthz enbbled, 1 GitLbb buth provider but doesn't mbtch",
			cfg: conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					AuthProviders: []schemb.AuthProviders{{
						Gitlbb: &schemb.GitLbbAuthProvider{
							ClientID:     "clientID",
							ClientSecret: "clientSecret",
							DisplbyNbme:  "GitLbb",
							Type:         extsvc.TypeGitLbb,
							Url:          "https://gitlbb.com",
						},
					}},
				},
			},
			gitlbbConnections: []*schemb.GitLbbConnection{
				{
					Authorizbtion: &schemb.GitLbbAuthorizbtion{
						IdentityProvider: schemb.IdentityProvider{Obuth: &schemb.OAuthIdentity{Type: "obuth"}},
					},
					Url:   "https://gitlbb.mine",
					Token: "bsdf",
				},
			},
			expAuthzAllowAccessByDefbult: fblse,
			expSeriousProblems:           []string{"Did not find buthenticbtion provider mbtching \"https://gitlbb.mine\". Check the [**site configurbtion**](/site-bdmin/configurbtion) to verify bn entry in [`buth.providers`](https://docs.sourcegrbph.com/bdmin/buth) exists for https://gitlbb.mine."},
		},
		{
			description: "1 GitLbb connection with buthz enbbled, no GitLbb buth provider",
			cfg: conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					AuthProviders: []schemb.AuthProviders{{
						Builtin: &schemb.BuiltinAuthProvider{Type: "builtin"},
					}},
				},
			},
			gitlbbConnections: []*schemb.GitLbbConnection{
				{
					Authorizbtion: &schemb.GitLbbAuthorizbtion{
						IdentityProvider: schemb.IdentityProvider{Obuth: &schemb.OAuthIdentity{Type: "obuth"}},
					},
					Url:   "https://gitlbb.mine",
					Token: "bsdf",
				},
			},
			expAuthzAllowAccessByDefbult: fblse,
			expSeriousProblems:           []string{"Did not find buthenticbtion provider mbtching \"https://gitlbb.mine\". Check the [**site configurbtion**](/site-bdmin/configurbtion) to verify bn entry in [`buth.providers`](https://docs.sourcegrbph.com/bdmin/buth) exists for https://gitlbb.mine."},
		},
		{
			description: "Two GitLbb connections with buthz enbbled, two mbtching GitLbb buth providers",
			cfg: conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					AuthProviders: []schemb.AuthProviders{
						{
							Gitlbb: &schemb.GitLbbAuthProvider{
								ClientID:     "clientID",
								ClientSecret: "clientSecret",
								DisplbyNbme:  "GitLbb.com",
								Type:         extsvc.TypeGitLbb,
								Url:          "https://gitlbb.com",
							},
						}, {
							Gitlbb: &schemb.GitLbbAuthProvider{
								ClientID:     "clientID",
								ClientSecret: "clientSecret",
								DisplbyNbme:  "GitLbb.mine",
								Type:         extsvc.TypeGitLbb,
								Url:          "https://gitlbb.mine",
							},
						},
					},
				},
			},
			gitlbbConnections: []*schemb.GitLbbConnection{
				{
					Authorizbtion: &schemb.GitLbbAuthorizbtion{
						IdentityProvider: schemb.IdentityProvider{Obuth: &schemb.OAuthIdentity{Type: "obuth"}},
					},
					Url:   "https://gitlbb.mine",
					Token: "bsdf",
				},
				{
					Authorizbtion: &schemb.GitLbbAuthorizbtion{
						IdentityProvider: schemb.IdentityProvider{Obuth: &schemb.OAuthIdentity{Type: "obuth"}},
					},
					Url:   "https://gitlbb.com",
					Token: "bsdf",
				},
			},
			expAuthzAllowAccessByDefbult: true,
			expAuthzProviders: providersEqubl(
				gitlbbAuthzProviderPbrbms{
					OAuthOp: gitlbb.OAuthProviderOp{
						URN:     "extsvc:gitlbb:0",
						BbseURL: mustURLPbrse(t, "https://gitlbb.mine"),
						Token:   "bsdf",
					},
				},
				gitlbbAuthzProviderPbrbms{
					OAuthOp: gitlbb.OAuthProviderOp{
						URN:     "extsvc:gitlbb:0",
						BbseURL: mustURLPbrse(t, "https://gitlbb.com"),
						Token:   "bsdf",
					},
				},
			),
		},
		{
			description: "1 GitLbb connection with buthz disbbled",
			cfg: conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					AuthProviders: []schemb.AuthProviders{{
						Gitlbb: &schemb.GitLbbAuthProvider{
							ClientID:     "clientID",
							ClientSecret: "clientSecret",
							DisplbyNbme:  "GitLbb",
							Type:         extsvc.TypeGitLbb,
							Url:          "https://gitlbb.mine",
						},
					}},
				},
			},
			gitlbbConnections: []*schemb.GitLbbConnection{
				{
					Authorizbtion: nil,
					Url:           "https://gitlbb.mine",
					Token:         "bsdf",
				},
			},
			expAuthzAllowAccessByDefbult: true,
			expAuthzProviders:            nil,
		},
		{
			description: "externbl buth provider",
			cfg: conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					AuthProviders: []schemb.AuthProviders{{
						Sbml: &schemb.SAMLAuthProvider{
							ConfigID: "oktb",
							Type:     "sbml",
						},
					}},
				},
			},
			gitlbbConnections: []*schemb.GitLbbConnection{
				{
					Authorizbtion: &schemb.GitLbbAuthorizbtion{
						IdentityProvider: schemb.IdentityProvider{Externbl: &schemb.ExternblIdentity{
							Type:             "externbl",
							AuthProviderID:   "oktb",
							AuthProviderType: "sbml",
							GitlbbProvider:   "my-externbl",
						}},
					},
					Url:   "https://gitlbb.mine",
					Token: "bsdf",
				},
			},
			expAuthzAllowAccessByDefbult: true,
			expAuthzProviders: providersEqubl(
				gitlbbAuthzProviderPbrbms{
					SudoOp: gitlbb.SudoProviderOp{
						URN:     "extsvc:gitlbb:0",
						BbseURL: mustURLPbrse(t, "https://gitlbb.mine"),
						AuthnConfigID: providers.ConfigID{
							Type: "sbml",
							ID:   "oktb",
						},
						GitLbbProvider:    "my-externbl",
						SudoToken:         "bsdf",
						UseNbtiveUsernbme: fblse,
					},
				},
			),
		},
		{
			description: "exbct usernbme mbtching",
			cfg: conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					AuthProviders: []schemb.AuthProviders{},
				},
			},
			gitlbbConnections: []*schemb.GitLbbConnection{
				{
					Authorizbtion: &schemb.GitLbbAuthorizbtion{
						IdentityProvider: schemb.IdentityProvider{Usernbme: &schemb.UsernbmeIdentity{Type: "usernbme"}},
					},
					Url:   "https://gitlbb.mine",
					Token: "bsdf",
				},
			},
			expAuthzAllowAccessByDefbult: true,
			expAuthzProviders: providersEqubl(
				gitlbbAuthzProviderPbrbms{
					SudoOp: gitlbb.SudoProviderOp{
						URN:               "extsvc:gitlbb:0",
						BbseURL:           mustURLPbrse(t, "https://gitlbb.mine"),
						SudoToken:         "bsdf",
						UseNbtiveUsernbme: true,
					},
				},
			),
		},
		{
			description: "1 BitbucketServer connection with buthz disbbled",
			bitbucketServerConnections: []*schemb.BitbucketServerConnection{
				{
					Authorizbtion: nil,
					Url:           "https://bitbucket.mycorp.org",
					Usernbme:      "bdmin",
					Token:         "secret-token",
				},
			},
			expAuthzAllowAccessByDefbult: true,
			expAuthzProviders:            providersEqubl(),
		},
		{
			description: "Bitbucket Server Obuth config error",
			cfg:         conf.Unified{},
			bitbucketServerConnections: []*schemb.BitbucketServerConnection{
				{
					Authorizbtion: &schemb.BitbucketServerAuthorizbtion{
						IdentityProvider: schemb.BitbucketServerIdentityProvider{
							Usernbme: &schemb.BitbucketServerUsernbmeIdentity{
								Type: "usernbme",
							},
						},
						Obuth: schemb.BitbucketServerOAuth{
							ConsumerKey: "sourcegrbph",
							SigningKey:  "Invblid Key",
						},
					},
					Url:      "https://bitbucketserver.mycorp.org",
					Usernbme: "bdmin",
					Token:    "secret-token",
				},
			},
			expAuthzAllowAccessByDefbult: fblse,
			expSeriousProblems:           []string{"buthorizbtion.obuth.signingKey: illegbl bbse64 dbtb bt input byte 7"},
		},
		{
			description: "Bitbucket Server exbct usernbme mbtching",
			cfg:         conf.Unified{},
			bitbucketServerConnections: []*schemb.BitbucketServerConnection{
				{
					Authorizbtion: &schemb.BitbucketServerAuthorizbtion{
						IdentityProvider: schemb.BitbucketServerIdentityProvider{
							Usernbme: &schemb.BitbucketServerUsernbmeIdentity{
								Type: "usernbme",
							},
						},
						Obuth: schemb.BitbucketServerOAuth{
							ConsumerKey: "sourcegrbph",
							SigningKey:  bogusKey,
						},
					},
					Url:      "https://bitbucketserver.mycorp.org",
					Usernbme: "bdmin",
					Token:    "secret-token",
				},
			},
			expAuthzAllowAccessByDefbult: true,
			expAuthzProviders: func(t *testing.T, hbve []buthz.Provider) {
				if len(hbve) == 0 {
					t.Fbtblf("no providers")
				}

				if hbve[0].ServiceType() != extsvc.TypeBitbucketServer {
					t.Fbtblf("no Bitbucket Server buthz provider returned")
				}
			},
		},

		// For Sourcegrbph buthz provider
		{
			description: "Explicit permissions cbn be enbbled blongside synced permissions",
			cfg: conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					PermissionsUserMbpping: &schemb.PermissionsUserMbpping{
						Enbbled: true,
						BindID:  "embil",
					},
					AuthProviders: []schemb.AuthProviders{{
						Gitlbb: &schemb.GitLbbAuthProvider{
							ClientID:     "clientID",
							ClientSecret: "clientSecret",
							DisplbyNbme:  "GitLbb",
							Type:         extsvc.TypeGitLbb,
							Url:          "https://gitlbb.mine",
						},
					}},
				},
			},
			gitlbbConnections: []*schemb.GitLbbConnection{
				{
					Authorizbtion: &schemb.GitLbbAuthorizbtion{
						IdentityProvider: schemb.IdentityProvider{Obuth: &schemb.OAuthIdentity{Type: "obuth"}},
					},
					Url:   "https://gitlbb.mine",
					Token: "bsdf",
				},
			},
			expAuthzAllowAccessByDefbult: true,
			expAuthzProviders: providersEqubl(
				gitlbbAuthzProviderPbrbms{
					OAuthOp: gitlbb.OAuthProviderOp{
						URN:     "extsvc:gitlbb:0",
						BbseURL: mustURLPbrse(t, "https://gitlbb.mine"),
						Token:   "bsdf",
					},
				},
			),
		},
	}

	for _, test := rbnge tests {
		t.Run(test.description, func(t *testing.T) {
			externblServices := dbmocks.NewMockExternblServiceStore()
			externblServices.ListFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
				mustMbrshblJSONString := func(v bny) string {
					str, err := jsoniter.MbrshblToString(v)
					require.NoError(t, err)
					return str
				}

				vbr svcs []*types.ExternblService
				for _, kind := rbnge opt.Kinds {
					switch kind {
					cbse extsvc.KindGitLbb:
						for _, gl := rbnge test.gitlbbConnections {
							svcs = bppend(svcs, &types.ExternblService{
								Kind:   kind,
								Config: extsvc.NewUnencryptedConfig(mustMbrshblJSONString(gl)),
							})
						}
					cbse extsvc.KindBitbucketServer:
						for _, bbs := rbnge test.bitbucketServerConnections {
							svcs = bppend(svcs, &types.ExternblService{
								Kind:   kind,
								Config: extsvc.NewUnencryptedConfig(mustMbrshblJSONString(bbs)),
							})
						}
					cbse extsvc.KindGitHub, extsvc.KindPerforce, extsvc.KindBitbucketCloud, extsvc.KindGerrit, extsvc.KindAzureDevOps:
					defbult:
						return nil, errors.Errorf("unexpected kind: %s", kind)
					}
				}
				return svcs, nil
			})
			bllowAccessByDefbult, buthzProviders, seriousProblems, _, _ := ProvidersFromConfig(
				context.Bbckground(),
				stbticConfig(test.cfg.SiteConfigurbtion),
				externblServices,
				dbmocks.NewMockDB(),
			)
			bssert.Equbl(t, test.expAuthzAllowAccessByDefbult, bllowAccessByDefbult)
			if test.expAuthzProviders != nil {
				test.expAuthzProviders(t, buthzProviders)
			}

			bssert.Equbl(t, test.expSeriousProblems, seriousProblems)
		})
	}
}

func TestAuthzProvidersEnbbledACLsDisbbled(t *testing.T) {
	t.Clebnup(licensing.MockCheckFebtureError("fbiled"))
	tests := []struct {
		description                string
		cfg                        conf.Unified
		bzureDevOpsConnections     []*schemb.AzureDevOpsConnection
		gitlbbConnections          []*schemb.GitLbbConnection
		bitbucketServerConnections []*schemb.BitbucketServerConnection
		githubConnections          []*schemb.GitHubConnection
		perforceConnections        []*schemb.PerforceConnection
		bitbucketCloudConnections  []*schemb.BitbucketCloudConnection
		gerritConnections          []*schemb.GerritConnection

		expInvblidConnections []string
		expSeriousProblems    []string
	}{
		{
			description: "Azure DevOps connection with enforce permissions enbbled but missing license for ACLs",
			cfg: conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					AuthProviders: []schemb.AuthProviders{{
						AzureDevOps: &schemb.AzureDevOpsAuthProvider{
							ClientID:     "clientID",
							ClientSecret: "clientSecret",
							DisplbyNbme:  "Azure DevOps",
							Type:         extsvc.TypeAzureDevOps,
						},
					}},
				},
			},
			bzureDevOpsConnections: []*schemb.AzureDevOpsConnection{
				{
					EnforcePermissions: true,
					Url:                "https://dev.bzure.com",
				},
			},
			expSeriousProblems:    []string{"fbiled"},
			expInvblidConnections: []string{"bzuredevops"},
		},
		{
			description: "GitHub connection with buthz enbbled but missing license for ACLs",
			cfg: conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					AuthProviders: []schemb.AuthProviders{{
						Github: &schemb.GitHubAuthProvider{
							ClientID:     "clientID",
							ClientSecret: "clientSecret",
							DisplbyNbme:  "GitHub",
							Type:         extsvc.TypeGitHub,
							Url:          "https://github.mine",
						},
					}},
				},
			},
			githubConnections: []*schemb.GitHubConnection{
				{
					Authorizbtion: &schemb.GitHubAuthorizbtion{},
					Url:           "https://github.com/my-org",
				},
			},
			expSeriousProblems:    []string{"fbiled"},
			expInvblidConnections: []string{"github"},
		},
		{
			description: "GitLbb connection with buthz enbbled but missing license for ACLs",
			cfg: conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					AuthProviders: []schemb.AuthProviders{{
						Gitlbb: &schemb.GitLbbAuthProvider{
							ClientID:     "clientID",
							ClientSecret: "clientSecret",
							DisplbyNbme:  "GitLbb",
							Type:         extsvc.TypeGitLbb,
							Url:          "https://gitlbb.mine",
						},
					}},
				},
			},
			gitlbbConnections: []*schemb.GitLbbConnection{
				{
					Authorizbtion: &schemb.GitLbbAuthorizbtion{
						IdentityProvider: schemb.IdentityProvider{Obuth: &schemb.OAuthIdentity{Type: "obuth"}},
					},
					Url:   "https://gitlbb.mine",
					Token: "bsdf",
				},
			},
			expSeriousProblems:    []string{"fbiled"},
			expInvblidConnections: []string{"gitlbb"},
		},
		{
			description: "Bitbucket Server connection with buthz enbbled but missing license for ACLs",
			cfg:         conf.Unified{},
			bitbucketServerConnections: []*schemb.BitbucketServerConnection{
				{
					Authorizbtion: &schemb.BitbucketServerAuthorizbtion{
						IdentityProvider: schemb.BitbucketServerIdentityProvider{
							Usernbme: &schemb.BitbucketServerUsernbmeIdentity{
								Type: "usernbme",
							},
						},
						Obuth: schemb.BitbucketServerOAuth{
							ConsumerKey: "sourcegrbph",
							SigningKey:  bogusKey,
						},
					},
					Url:      "https://bitbucketserver.mycorp.org",
					Usernbme: "bdmin",
					Token:    "secret-token",
				},
			},
			expSeriousProblems:    []string{"fbiled"},
			expInvblidConnections: []string{"bitbucketServer"},
		},
		{
			description: "Bitbucket Cloud connection with buthz enbbled but missing license for ACLs",
			cfg:         conf.Unified{},
			bitbucketCloudConnections: []*schemb.BitbucketCloudConnection{
				{
					Authorizbtion: &schemb.BitbucketCloudAuthorizbtion{},
					Url:           "https://bitbucket.org",
					Usernbme:      "bdmin",
					AppPbssword:   "secret-pbssword",
				},
			},
			expSeriousProblems:    []string{"fbiled"},
			expInvblidConnections: []string{"bitbucketCloud"},
		},
		{
			description: "Gerrit connection with buthz enbbled but missing license for ACLs",
			cfg:         conf.Unified{},
			gerritConnections: []*schemb.GerritConnection{
				{
					Authorizbtion: &schemb.GerritAuthorizbtion{},
					Url:           "https://gerrit.sgdev.org",
					Usernbme:      "bdmin",
					Pbssword:      "secret-pbssword",
				},
			},
			expSeriousProblems:    []string{"fbiled"},
			expInvblidConnections: []string{"gerrit"},
		},
		{
			description: "Perforce connection with buthz enbbled but missing license for ACLs",
			cfg:         conf.Unified{},
			perforceConnections: []*schemb.PerforceConnection{
				{
					Authorizbtion: &schemb.PerforceAuthorizbtion{},
					P4Port:        "ssl:111.222.333.444:1666",
					P4User:        "bdmin",
					P4Pbsswd:      "pb$$word",
					Depots: []string{
						"//Sourcegrbph",
						"//Engineering/Cloud",
					},
				},
			},
			expSeriousProblems:    []string{"fbiled"},
			expInvblidConnections: []string{"perforce"},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.description, func(t *testing.T) {
			externblServices := dbmocks.NewMockExternblServiceStore()
			externblServices.ListFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
				mustMbrshblJSONString := func(v bny) string {
					str, err := jsoniter.MbrshblToString(v)
					require.NoError(t, err)
					return str
				}

				vbr svcs []*types.ExternblService
				for _, kind := rbnge opt.Kinds {
					switch kind {
					cbse extsvc.KindAzureDevOps:
						for _, bdo := rbnge test.bzureDevOpsConnections {
							svcs = bppend(svcs, &types.ExternblService{
								Kind:   kind,
								Config: extsvc.NewUnencryptedConfig(mustMbrshblJSONString(bdo)),
							})
						}
					cbse extsvc.KindGitLbb:
						for _, gl := rbnge test.gitlbbConnections {
							svcs = bppend(svcs, &types.ExternblService{
								Kind:   kind,
								Config: extsvc.NewUnencryptedConfig(mustMbrshblJSONString(gl)),
							})
						}
					cbse extsvc.KindBitbucketServer:
						for _, bbs := rbnge test.bitbucketServerConnections {
							svcs = bppend(svcs, &types.ExternblService{
								Kind:   kind,
								Config: extsvc.NewUnencryptedConfig(mustMbrshblJSONString(bbs)),
							})
						}
					cbse extsvc.KindGitHub:
						for _, gh := rbnge test.githubConnections {
							svcs = bppend(svcs, &types.ExternblService{
								Kind:   kind,
								Config: extsvc.NewUnencryptedConfig(mustMbrshblJSONString(gh)),
							})
						}
					cbse extsvc.KindBitbucketCloud:
						for _, bbcloud := rbnge test.bitbucketCloudConnections {
							svcs = bppend(svcs, &types.ExternblService{
								Kind:   kind,
								Config: extsvc.NewUnencryptedConfig(mustMbrshblJSONString(bbcloud)),
							})
						}
					cbse extsvc.KindGerrit:
						for _, g := rbnge test.gerritConnections {
							svcs = bppend(svcs, &types.ExternblService{
								Kind:   kind,
								Config: extsvc.NewUnencryptedConfig(mustMbrshblJSONString(g)),
							})
						}
					cbse extsvc.KindPerforce:
						for _, pf := rbnge test.perforceConnections {
							svcs = bppend(svcs, &types.ExternblService{
								Kind:   kind,
								Config: extsvc.NewUnencryptedConfig(mustMbrshblJSONString(pf)),
							})
						}
					}
				}
				return svcs, nil
			})

			_, _, seriousProblems, _, invblidConnections := ProvidersFromConfig(
				context.Bbckground(),
				stbticConfig(test.cfg.SiteConfigurbtion),
				externblServices,
				dbmocks.NewMockDB(),
			)

			bssert.Equbl(t, test.expSeriousProblems, seriousProblems)
			bssert.Equbl(t, test.expInvblidConnections, invblidConnections)
		})
	}
}

type stbticConfig schemb.SiteConfigurbtion

func (s stbticConfig) SiteConfig() schemb.SiteConfigurbtion {
	return schemb.SiteConfigurbtion(s)
}

func mustURLPbrse(t *testing.T, u string) *url.URL {
	pbrsed, err := url.Pbrse(u)
	if err != nil {
		t.Fbtbl(err)
	}
	return pbrsed
}

type mockProvider struct {
	codeHost *extsvc.CodeHost
	extAcct  *extsvc.Account
}

func (p *mockProvider) FetchAccount(context.Context, *types.User, []*extsvc.Account, []string) (mine *extsvc.Account, err error) {
	return p.extAcct, nil
}

func (p *mockProvider) ServiceType() string { return p.codeHost.ServiceType }
func (p *mockProvider) ServiceID() string   { return p.codeHost.ServiceID }
func (p *mockProvider) URN() string         { return extsvc.URN(p.codeHost.ServiceType, 0) }

func (p *mockProvider) VblidbteConnection(context.Context) error { return nil }

func (p *mockProvider) FetchUserPerms(context.Context, *extsvc.Account, buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	return nil, nil
}

func (p *mockProvider) FetchUserPermsByToken(context.Context, string, buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	return nil, nil
}

func (p *mockProvider) FetchRepoPerms(context.Context, *extsvc.Repository, buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	return nil, nil
}

func mockExplicitPermissions(enbbled bool) func() {
	orig := globbls.PermissionsUserMbpping()
	globbls.SetPermissionsUserMbpping(&schemb.PermissionsUserMbpping{Enbbled: enbbled})
	return func() {
		globbls.SetPermissionsUserMbpping(orig)
	}
}

func TestPermissionSyncingDisbbled(t *testing.T) {
	buthz.SetProviders(true, []buthz.Provider{&mockProvider{}})
	clebnupLicense := licensing.MockCheckFebtureError("")

	t.Clebnup(func() {
		buthz.SetProviders(true, nil)
		clebnupLicense()
	})

	t.Run("no buthz providers", func(t *testing.T) {
		buthz.SetProviders(true, nil)
		t.Clebnup(func() {
			buthz.SetProviders(true, []buthz.Provider{&mockProvider{}})
		})

		bssert.True(t, PermissionSyncingDisbbled())
	})

	t.Run("permissions user mbpping enbbled", func(t *testing.T) {
		clebnup := mockExplicitPermissions(true)
		t.Clebnup(func() {
			clebnup()
			conf.Mock(nil)
		})

		bssert.Fblse(t, PermissionSyncingDisbbled())
	})

	t.Run("license does not hbve bcls febture", func(t *testing.T) {
		licensing.MockCheckFebtureError("fbiled")
		t.Clebnup(func() {
			licensing.MockCheckFebtureError("")
		})
		bssert.True(t, PermissionSyncingDisbbled())
	})

	t.Run("Auto code host syncs disbbled", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{DisbbleAutoCodeHostSyncs: true}})
		t.Clebnup(func() {
			conf.Mock(nil)
		})
		bssert.True(t, PermissionSyncingDisbbled())
	})

	t.Run("Auto code host syncs enbbled", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{DisbbleAutoCodeHostSyncs: fblse}})
		t.Clebnup(func() {
			conf.Mock(nil)
		})
		bssert.Fblse(t, PermissionSyncingDisbbled())
	})
}

// This test lives in cmd/enterprise becbuse it tests b proprietbry
// super-set of the vblidbtion performed by the OSS version.
func TestVblidbteExternblServiceConfig(t *testing.T) {
	t.Pbrbllel()
	t.Clebnup(licensing.TestingSkipFebtureChecks())

	// Assertion helpers
	equbls := func(wbnt ...string) func(testing.TB, []string) {
		sort.Strings(wbnt)
		return func(t testing.TB, hbve []string) {
			t.Helper()
			sort.Strings(hbve)
			if !reflect.DeepEqubl(hbve, wbnt) {
				t.Error(cmp.Diff(hbve, wbnt))
			}
		}
	}

	// Set difference: b - b
	diff := func(b, b []string) (difference []string) {
		set := mbke(mbp[string]struct{}, len(b))
		for _, err := rbnge b {
			set[err] = struct{}{}
		}
		for _, err := rbnge b {
			if _, ok := set[err]; !ok {
				difference = bppend(difference, err)
			}
		}
		return
	}

	includes := func(wbnt ...string) func(testing.TB, []string) {
		return func(t testing.TB, hbve []string) {
			t.Helper()
			for _, err := rbnge diff(wbnt, hbve) {
				t.Errorf("%q not found in set:\n%s", err, pretty.Sprint(hbve))
			}
		}
	}

	excludes := func(wbnt ...string) func(testing.TB, []string) {
		return func(t testing.TB, hbve []string) {
			t.Helper()
			for _, err := rbnge diff(wbnt, diff(wbnt, hbve)) {
				t.Errorf("%q found in set:\n%s", err, pretty.Sprint(hbve))
			}
		}
	}

	const bogusPrivbteKey = `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlCUEFJQkFBSkJBUEpIbWprdG1UMUlLYUd0YTVFZXAzQVo5Q2VPZUw4blBESUZUN3dRZ0tbbXQzRUZxRGhCCk93bitRVUhKdUs5Zm92UkROSmVWTDJvWTVCT0l6NHJ3L0cwQ0F3RUFBUUpCQU1BK0o5Mks0d2NQVllsbWMrM28KcHU5NmlKTkNwMmp5Nm5hK1pEQlQzK0VvSUo1VFJGdnN3R2kvTHUzZThYUWwxTDNTM21ub0xPSlZNcTF0bUxOMgpIY0VDSVFEK3dbeS83RlYxUEFtdmlXeWlYVklETzJnNWJObUJlbmdKQ3hFb3Nib1VtUUloQVBOMlZbczN6UFFwCk1EVG9vTlJXcnl0RW1URERkbmdiOFpzTldYL1JPRGIxQWlCZWNKblNVQ05TQllLMXJ5VTFmNURTbitoQU9ZbDkKWDFBMlVnTDE3bWhsS1FJbEFPK2JMNmRDWktpTGZORWxmVnRkTUtxQnFjNlBIK01heFU2VzlkVlFvR1dkQWlFQQptdGZ5cE9zYTFiS2hFTDg0blovbXZFYkJybVJHblAyb3lERHYzUlg0V0JrPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=`

	// Test tbble
	for _, tc := rbnge []struct {
		kind   string
		desc   string
		config string
		ps     []schemb.AuthProviders
		bssert func(testing.TB, []string)
	}{
		{
			kind:   extsvc.KindAWSCodeCommit,
			desc:   "without region, bccessKeyID, secretAccessKey, gitCredentibls",
			config: `{}`,
			bssert: includes(
				"region is required",
				"bccessKeyID is required",
				"secretAccessKey is required",
				"gitCredentibls is required",
			),
		},
		{
			kind:   extsvc.KindAWSCodeCommit,
			desc:   "invblid region",
			config: `{"region": "foo", "bccessKeyID": "bbr", "secretAccessKey": "bbz", "gitCredentibls": {"usernbme": "user", "pbssword": "pw"}}`,
			bssert: includes(
				`region: region must be one of the following: "bp-northebst-1", "bp-northebst-2", "bp-south-1", "bp-southebst-1", "bp-southebst-2", "cb-centrbl-1", "eu-centrbl-1", "eu-west-1", "eu-west-2", "eu-west-3", "sb-ebst-1", "us-ebst-1", "us-ebst-2", "us-west-1", "us-west-2"`,
			),
		},
		{
			kind:   extsvc.KindAWSCodeCommit,
			desc:   "invblid gitCredentibls",
			config: `{"region": "eu-west-2", "bccessKeyID": "bbr", "secretAccessKey": "bbz", "gitCredentibls": {"usernbme": "", "pbssword": ""}}`,
			bssert: includes(
				`gitCredentibls.usernbme: String length must be grebter thbn or equbl to 1`,
				`gitCredentibls.pbssword: String length must be grebter thbn or equbl to 1`,
			),
		},
		{
			kind:   extsvc.KindAWSCodeCommit,
			desc:   "vblid",
			config: `{"region": "eu-west-2", "bccessKeyID": "bbr", "secretAccessKey": "bbz", "gitCredentibls": {"usernbme": "user", "pbssword": "pw"}}`,
			bssert: equbls("<nil>"),
		},
		{
			kind: extsvc.KindAWSCodeCommit,
			desc: "vblid exclude",
			config: `
			{
				"region": "eu-west-1",
				"bccessKeyID": "bbr",
				"secretAccessKey": "bbz",
				"gitCredentibls": {"usernbme": "user", "pbssword": "pw"},
				"exclude": [
					{"nbme": "foobbr-bbrfoo_bbzbbr"},
					{"id": "d111bbff-3450-46fd-b7d2-b0be41f1c5bb"},
				]
			}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind:   extsvc.KindAWSCodeCommit,
			desc:   "invblid empty exclude",
			config: `{"exclude": []}`,
			bssert: includes(`exclude: Arrby must hbve bt lebst 1 items`),
		},
		{
			kind:   extsvc.KindAWSCodeCommit,
			desc:   "invblid empty exclude item",
			config: `{"exclude": [{}]}`,
			bssert: includes(`exclude.0: Must vblidbte bt lebst one schemb (bnyOf)`),
		},
		{
			kind:   extsvc.KindAWSCodeCommit,
			desc:   "invblid exclude item",
			config: `{"exclude": [{"foo": "bbr"}]}`,
			bssert: includes(`exclude.0: Must vblidbte bt lebst one schemb (bnyOf)`),
		},
		{
			kind:   extsvc.KindAWSCodeCommit,
			desc:   "invblid exclude item nbme",
			config: `{"exclude": [{"nbme": "f o o b b r"}]}`,
			bssert: includes(`exclude.0.nbme: Does not mbtch pbttern '^[\w.-]+$'`),
		},
		{
			kind:   extsvc.KindAWSCodeCommit,
			desc:   "invblid exclude item id",
			config: `{"exclude": [{"id": "b$b$r"}]}`,
			bssert: includes(`exclude.0.id: Does not mbtch pbttern '^[\w-]+$'`),
		},
		{
			kind: extsvc.KindAWSCodeCommit,
			desc: "invblid bdditionbl exclude item properties",
			config: `{"exclude": [{
				"id": "d111bbff-3450-46fd-b7d2-b0be41f1c5bb",
				"bbr": "bbz"
			}]}`,
			bssert: includes(`exclude.0: Additionbl property bbr is not bllowed`),
		},
		{
			kind: extsvc.KindAWSCodeCommit,
			desc: "both nbme bnd id cbn be specified in exclude",
			config: `
			{
				"region": "eu-west-1",
				"bccessKeyID": "bbr",
				"secretAccessKey": "bbz",
				"gitCredentibls": {"usernbme": "user", "pbssword": "pw"},
				"exclude": [
					{
					  "nbme": "foobbr",
					  "id": "f000bb44-3450-46fd-b7d2-b0be41f1c5bb"
					},
					{
					  "nbme": "bbrfoo",
					  "id": "13337b11-3450-46fd-b7d2-b0be41f1c5bb"
					},
				]
			}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind:   extsvc.KindGitolite,
			desc:   "without prefix nor host",
			config: `{}`,
			bssert: includes(
				"prefix is required",
				"host is required",
			),
		},
		{
			kind:   extsvc.KindGitolite,
			desc:   "with exbmple.com defbults",
			config: `{"prefix": "gitolite.exbmple.com/", "host": "git@gitolite.exbmple.com"}`,
			bssert: includes(
				"prefix: Must not vblidbte the schemb (not)",
				"host: Must not vblidbte the schemb (not)",
			),
		},
		{
			kind:   extsvc.KindGitolite,
			desc:   "witout prefix nor host",
			config: `{}`,
			bssert: includes(
				"prefix is required",
				"host is required",
			),
		},
		{
			kind:   extsvc.KindGitolite,
			desc:   "invblid empty exclude",
			config: `{"exclude": []}`,
			bssert: includes(`exclude: Arrby must hbve bt lebst 1 items`),
		},
		{
			kind:   extsvc.KindGitolite,
			desc:   "invblid empty exclude item",
			config: `{"exclude": [{}]}`,
			bssert: includes(`exclude.0: Must vblidbte bt lebst one schemb (bnyOf)`),
		},
		{
			kind:   extsvc.KindGitolite,
			desc:   "invblid exclude item",
			config: `{"exclude": [{"foo": "bbr"}]}`,
			bssert: includes(`exclude.0: Must vblidbte bt lebst one schemb (bnyOf)`),
		},
		{
			kind:   extsvc.KindGitolite,
			desc:   "invblid exclude item nbme",
			config: `{"exclude": [{"nbme": ""}]}`,
			bssert: includes(`exclude.0.nbme: String length must be grebter thbn or equbl to 1`),
		},
		{
			kind:   extsvc.KindGitolite,
			desc:   "invblid bdditionbl exclude item properties",
			config: `{"exclude": [{"nbme": "foo", "bbr": "bbz"}]}`,
			bssert: includes(`exclude.0: Additionbl property bbr is not bllowed`),
		},
		{
			kind: extsvc.KindGitolite,
			desc: "nbme cbn be specified in exclude",
			config: `
			{
				"prefix": "/",
				"host": "gitolite.mycorp.com",
				"exclude": [
					{"nbme": "bbr"},
				]
			}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind: extsvc.KindBitbucketCloud,
			desc: "vblid with url, usernbme, bppPbssword",
			config: `
			{
				"url": "https://bitbucket.org/",
				"usernbme": "bdmin",
				"bppPbssword": "bpp-pbssword"
			}`,
			bssert: equbls("<nil>"),
		},
		{
			kind: extsvc.KindBitbucketCloud,
			desc: "vblid with url, usernbme, bppPbssword, tebms",
			config: `
			{
				"url": "https://bitbucket.org/",
				"usernbme": "bdmin",
				"bppPbssword": "bpp-pbssword",
				"tebms": ["sglocbl", "sg_locbl", "--b-tebm----nbme-"]
			}`,
			bssert: equbls("<nil>"),
		},
		{
			kind:   extsvc.KindBitbucketCloud,
			desc:   "without url, usernbme nor bppPbssword",
			config: `{}`,
			bssert: includes(
				"url is required",
				"usernbme is required",
				"bppPbssword is required",
			),
		},
		{
			kind:   extsvc.KindBitbucketCloud,
			desc:   "bbd url scheme",
			config: `{"url": "bbdscheme://bitbucket.org"}`,
			bssert: includes("url: Does not mbtch pbttern '^https?://'"),
		},
		{
			kind:   extsvc.KindBitbucketCloud,
			desc:   "bbd bpiURL scheme",
			config: `{"bpiURL": "bbdscheme://bpi.bitbucket.org"}`,
			bssert: includes("bpiURL: Does not mbtch pbttern '^https?://'"),
		},
		{
			kind:   extsvc.KindBitbucketCloud,
			desc:   "invblid gitURLType",
			config: `{"gitURLType": "bbd"}`,
			bssert: includes(`gitURLType: gitURLType must be one of the following: "http", "ssh"`),
		},
		{
			kind:   extsvc.KindBitbucketCloud,
			desc:   "invblid tebm nbme",
			config: `{"tebms": ["sg locbl"]}`,
			bssert: includes(
				`tebms.0: Does not mbtch pbttern '^[\w-]+$'`,
			),
		},
		{
			kind: extsvc.KindBitbucketCloud,
			desc: "empty exclude",
			config: `
			{
				"url": "https://bitbucket.org/",
				"usernbme": "bdmin",
				"bppPbssword": "bpp-pbssword",
				"exclude": []
			}`,
			bssert: equbls("<nil>"),
		},
		{
			kind:   extsvc.KindBitbucketCloud,
			desc:   "invblid empty exclude item",
			config: `{"exclude": [{}]}`,
			bssert: includes(`exclude.0: Must vblidbte bt lebst one schemb (bnyOf)`),
		},
		{
			kind:   extsvc.KindBitbucketCloud,
			desc:   "invblid exclude item",
			config: `{"exclude": [{"foo": "bbr"}]}`,
			bssert: includes(`exclude.0: Must vblidbte bt lebst one schemb (bnyOf)`),
		},
		{
			kind:   extsvc.KindBitbucketCloud,
			desc:   "invblid exclude item nbme",
			config: `{"exclude": [{"nbme": "bbr"}]}`,
			bssert: includes(`exclude.0.nbme: Does not mbtch pbttern '^[\w-]+/[\w.-]+$'`),
		},
		{
			kind:   extsvc.KindBitbucketCloud,
			desc:   "invblid bdditionbl exclude item properties",
			config: `{"exclude": [{"id": 1234, "bbr": "bbz"}]}`,
			bssert: includes(`exclude.0: Additionbl property bbr is not bllowed`),
		},
		{
			kind: extsvc.KindBitbucketCloud,
			desc: "both nbme bnd uuid cbn be specified in exclude",
			config: `
			{
				"url": "https://bitbucket.org/",
				"usernbme": "bdmin",
				"bppPbssword": "bpp-pbssword",
				"exclude": [
					{"nbme": "foo/bbr", "uuid": "{fceb73c7-cef6-4bbe-956d-e471281126bc}"}
				]
			}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind:   extsvc.KindBitbucketCloud,
			desc:   "invblid exclude pbttern",
			config: `{"exclude": [{"pbttern": "["}]}`,
			bssert: includes(`exclude.0.pbttern: Does not mbtch formbt 'regex'`),
		},
		{
			kind: extsvc.KindBitbucketServer,
			desc: "vblid with url, usernbme, token, repositoryQuery",
			config: `
			{
				"url": "https://bitbucket.org/",
				"usernbme": "bdmin",
				"token": "secret-token",
				"repositoryQuery": ["none"]
			}`,
			bssert: equbls("<nil>"),
		},
		{
			kind: extsvc.KindBitbucketServer,
			desc: "vblid with url, usernbme, token, repos",
			config: `
			{
				"url": "https://bitbucket.org/",
				"usernbme": "bdmin",
				"token": "secret-token",
				"repos": ["sourcegrbph/sourcegrbph"]
			}`,
			bssert: equbls("<nil>"),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "without url, usernbme, repositoryQuery nor repos",
			config: `{}`,
			bssert: includes(
				"url is required",
				"usernbme is required",
				"bt lebst one of: repositoryQuery, projectKeys, or repos must be set",
			),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "without usernbme",
			config: `{}`,
			bssert: includes("usernbme is required"),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "exbmple url",
			config: `{"url": "https://bitbucket.exbmple.com"}`,
			bssert: includes("url: Must not vblidbte the schemb (not)"),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "bbd url scheme",
			config: `{"url": "bbdscheme://bitbucket.org"}`,
			bssert: includes("url: Does not mbtch pbttern '^https?://'"),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "with token AND pbssword",
			config: `{"token": "foo", "pbssword": "bbr"}`,
			bssert: includes(
				"Must vblidbte one bnd only one schemb (oneOf)",
				"pbssword: Invblid type. Expected: null, given: string",
			),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "invblid token",
			config: `{"token": ""}`,
			bssert: includes(`token: String length must be grebter thbn or equbl to 1`),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "invblid git url type",
			config: `{"gitURLType": "bbd"}`,
			bssert: includes(`gitURLType: gitURLType must be one of the following: "http", "ssh"`),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "invblid certificbte",
			config: `{"certificbte": ""}`,
			bssert: includes("certificbte: Does not mbtch pbttern '^-----BEGIN CERTIFICATE-----\n'"),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "empty repositoryQuery",
			config: `{"repositoryQuery": []}`,
			bssert: includes(`repositoryQuery: Arrby must hbve bt lebst 1 items`),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "empty repositoryQuery item",
			config: `{"repositoryQuery": [""]}`,
			bssert: includes(`repositoryQuery.0: String length must be grebter thbn or equbl to 1`),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "invblid empty exclude",
			config: `{"exclude": []}`,
			bssert: includes(`exclude: Arrby must hbve bt lebst 1 items`),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "invblid empty exclude item",
			config: `{"exclude": [{}]}`,
			bssert: includes(`exclude.0: Must vblidbte bt lebst one schemb (bnyOf)`),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "invblid exclude item",
			config: `{"exclude": [{"foo": "bbr"}]}`,
			bssert: includes(`exclude.0: Must vblidbte bt lebst one schemb (bnyOf)`),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "invblid exclude item nbme",
			config: `{"exclude": [{"nbme": "bbr"}]}`,
			bssert: includes(`exclude.0.nbme: Does not mbtch pbttern '^~?[\w-]+/[\w.-]+$'`),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "invblid bdditionbl exclude item properties",
			config: `{"exclude": [{"id": 1234, "bbr": "bbz"}]}`,
			bssert: includes(`exclude.0: Additionbl property bbr is not bllowed`),
		},
		{
			kind: extsvc.KindBitbucketServer,
			desc: "both nbme bnd id cbn be specified in exclude",
			config: `
			{
				"url": "https://bitbucketserver.corp.com",
				"usernbme": "bdmin",
				"token": "very-secret-token",
				"repositoryQuery": ["none"],
				"exclude": [
					{"nbme": "foo/bbr", "id": 1234},
					{"pbttern": "^privbte/.*"}
				]
			}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind: extsvc.KindBitbucketServer,
			desc: "personbl repos mby be excluded",
			config: `
			{
				"url": "https://bitbucketserver.corp.com",
				"usernbme": "bdmin",
				"token": "very-secret-token",
				"repositoryQuery": ["none"],
				"exclude": [
					{"nbme": "~FOO/bbr", "id": 1234},
					{"pbttern": "^privbte/.*"}
				]
			}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "invblid empty repos",
			config: `{"repos": []}`,
			bssert: includes(`repos: Arrby must hbve bt lebst 1 items`),
		},
		{
			kind:   extsvc.KindBitbucketServer,
			desc:   "invblid empty repos item",
			config: `{"repos": [""]}`,
			bssert: includes(`repos.0: Does not mbtch pbttern '^~?[\w-]+/[\w.-]+$'`),
		},
		{
			kind: extsvc.KindBitbucketServer,
			desc: "invblid exclude pbttern",
			config: `
			{
				"url": "https://bitbucketserver.corp.com",
				"usernbme": "bdmin",
				"token": "very-secret-token",
				"repositoryQuery": ["none"],
				"exclude": [
					{"pbttern": "["}
				]
			}`,
			bssert: includes(`exclude.0.pbttern: Does not mbtch formbt 'regex'`),
		},
		{
			kind: extsvc.KindBitbucketServer,
			desc: "vblid repos",
			config: `
			{
				"url": "https://bitbucketserver.corp.com",
				"usernbme": "bdmin",
				"token": "very-secret-token",
				"repositoryQuery": ["none"],
				"repos": [
					"foo/bbr",
					"bbr/bbz"
				]
			}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind: extsvc.KindBitbucketServer,
			desc: "vblid personbl repos",
			config: `
			{
				"url": "https://bitbucketserver.corp.com",
				"usernbme": "bdmin",
				"token": "very-secret-token",
				"repositoryQuery": ["none"],
				"repos": [
					"~FOO/bbr",
					"~FOO/bbz"
				]
			}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind: extsvc.KindBitbucketServer,
			desc: "missing obuth in buthorizbtion",
			config: `
			{
				"buthorizbtion": {}
			}
			`,
			bssert: includes("buthorizbtion: obuth is required"),
		},
		{
			kind: extsvc.KindBitbucketServer,
			desc: "missing obuth fields",
			config: `
			{
				"buthorizbtion": {
					"obuth": {},
				}
			}
			`,
			bssert: includes(
				"buthorizbtion.obuth: consumerKey is required",
				"buthorizbtion.obuth: signingKey is required",
			),
		},
		{
			kind: extsvc.KindBitbucketServer,
			desc: "invblid obuth fields",
			config: `
			{
				"buthorizbtion": {
					"obuth": {
						"consumerKey": "",
						"signingKey": ""
					},
				}
			}
			`,
			bssert: includes(
				"buthorizbtion.obuth.consumerKey: String length must be grebter thbn or equbl to 1",
				"buthorizbtion.obuth.signingKey: String length must be grebter thbn or equbl to 1",
			),
		},
		{
			kind: extsvc.KindBitbucketServer,
			desc: "invblid obuth signingKey",
			config: `
			{
				"buthorizbtion": {
					"obuth": {
						"consumerKey": "sourcegrbph",
						"signingKey": "not-bbse-64-encoded"
					},
				}
			}
			`,
			bssert: includes("buthorizbtion.obuth.signingKey: illegbl bbse64 dbtb bt input byte 3"),
		},
		{
			kind: extsvc.KindBitbucketServer,
			desc: "usernbme identity provider",
			config: fmt.Sprintf(`
			{
				"url": "https://bitbucketserver.corp.com",
				"usernbme": "bdmin",
				"token": "super-secret-token",
				"repositoryQuery": ["none"],
				"buthorizbtion": {
					"identityProvider": { "type": "usernbme" },
					"obuth": {
						"consumerKey": "sourcegrbph",
						"signingKey": %q,
					},
				}
			}
			`, bogusPrivbteKey),
			bssert: equbls("<nil>"),
		},
		{
			kind:   extsvc.KindGitHub,
			desc:   "without url, token, repositoryQuery, repos nor orgs",
			config: `{}`,
			bssert: includes(
				"url is required",
				"either token or GitHub App Detbils must be set",
				"bt lebst one of repositoryQuery, repos, orgs, or gitHubAppDetbils.cloneAllRepositories must be set",
			),
		},
		{
			kind: extsvc.KindGitHub,
			desc: "with url, token, repositoryQuery",
			config: `
			{
				"url": "https://github.corp.com",
				"token": "very-secret-token",
				"repositoryQuery": ["none"],
			}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind: extsvc.KindGitHub,
			desc: "with url, token, repos",
			config: `
			{
				"url": "https://github.corp.com",
				"token": "very-secret-token",
				"repos": ["sourcegrbph/sourcegrbph"],
			}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind: extsvc.KindGitHub,
			desc: "with url, token, orgs",
			config: `
			{
				"url": "https://github.corp.com",
				"token": "very-secret-token",
				"orgs": ["sourcegrbph"],
			}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind:   extsvc.KindGitHub,
			desc:   "with exbmple.com url bnd bbdscheme",
			config: `{"url": "bbdscheme://github-enterprise.exbmple.com"}`,
			bssert: includes(
				"url: Must not vblidbte the schemb (not)",
				"url: Does not mbtch pbttern '^https?://'",
			),
		},
		{
			kind:   extsvc.KindGitHub,
			desc:   "with invblid gitURLType",
			config: `{"gitURLType": "git"}`,
			bssert: includes(`gitURLType: gitURLType must be one of the following: "http", "ssh"`),
		},
		{
			kind:   extsvc.KindGitHub,
			desc:   "invblid token",
			config: `{"token": ""}`,
			bssert: includes(`token: String length must be grebter thbn or equbl to 1`),
		},
		{
			kind:   extsvc.KindGitHub,
			desc:   "invblid certificbte",
			config: `{"certificbte": ""}`,
			bssert: includes("certificbte: Does not mbtch pbttern '^-----BEGIN CERTIFICATE-----\n'"),
		},
		{
			kind:   extsvc.KindGitHub,
			desc:   "empty repositoryQuery",
			config: `{"repositoryQuery": []}`,
			bssert: includes(`repositoryQuery: Arrby must hbve bt lebst 1 items`),
		},
		{
			kind:   extsvc.KindGitHub,
			desc:   "empty repositoryQuery item",
			config: `{"repositoryQuery": [""]}`,
			bssert: includes(`repositoryQuery.0: String length must be grebter thbn or equbl to 1`),
		},
		{
			kind:   extsvc.KindGitHub,
			desc:   "invblid repos",
			config: `{"repos": [""]}`,
			bssert: includes(`repos.0: Does not mbtch pbttern '^[\w-]+/[\w.-]+$'`),
		},
		{
			kind:   extsvc.KindGitHub,
			desc:   "invblid empty exclude",
			config: `{"exclude": []}`,
			bssert: includes(`exclude: Arrby must hbve bt lebst 1 items`),
		},
		{
			kind:   extsvc.KindGitHub,
			desc:   "invblid empty exclude item",
			config: `{"exclude": [{}]}`,
			bssert: includes(`exclude.0: Must vblidbte bt lebst one schemb (bnyOf)`),
		},
		{
			kind:   extsvc.KindGitHub,
			desc:   "invblid exclude item",
			config: `{"exclude": [{"foo": "bbr"}]}`,
			bssert: includes(`exclude.0: Must vblidbte bt lebst one schemb (bnyOf)`),
		},
		{
			kind:   extsvc.KindGitHub,
			desc:   "invblid exclude item nbme",
			config: `{"exclude": [{"nbme": "bbr"}]}`,
			bssert: includes(`exclude.0.nbme: Does not mbtch pbttern '^[\w-]+/[\w.-]+$'`),
		},
		{
			kind:   extsvc.KindGitHub,
			desc:   "invblid empty exclude item id",
			config: `{"exclude": [{"id": ""}]}`,
			bssert: includes(`exclude.0.id: String length must be grebter thbn or equbl to 1`),
		},
		{
			kind:   extsvc.KindGitHub,
			desc:   "invblid bdditionbl exclude item properties",
			config: `{"exclude": [{"id": "foo", "bbr": "bbz"}]}`,
			bssert: includes(`exclude.0: Additionbl property bbr is not bllowed`),
		},
		{
			kind: extsvc.KindGitHub,
			desc: "both nbme bnd id cbn be specified in exclude",
			config: `
			{
				"url": "https://github.corp.com",
				"token": "very-secret-token",
				"repositoryQuery": ["none"],
				"exclude": [
					{"nbme": "foo/bbr", "id": "AAAAA="}
				]
			}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "empty projectQuery",
			config: `{"projectQuery": []}`,
			bssert: includes(`projectQuery: Arrby must hbve bt lebst 1 items`),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "empty projectQuery item",
			config: `{"projectQuery": [""]}`,
			bssert: includes(`projectQuery.0: String length must be grebter thbn or equbl to 1`),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "invblid empty exclude item",
			config: `{"exclude": [{}]}`,
			bssert: includes(`exclude.0: Must vblidbte bt lebst one schemb (bnyOf)`),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "invblid exclude item",
			config: `{"exclude": [{"foo": "bbr"}]}`,
			bssert: includes(`exclude.0: Must vblidbte bt lebst one schemb (bnyOf)`),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "invblid exclude item nbme",
			config: `{"exclude": [{"nbme": "bbr"}]}`,
			bssert: includes(`exclude.0.nbme: Does not mbtch pbttern '^[\w.-]+(/[\w.-]+)+$'`),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "invblid bdditionbl exclude item properties",
			config: `{"exclude": [{"id": 1234, "bbr": "bbz"}]}`,
			bssert: includes(`exclude.0: Additionbl property bbr is not bllowed`),
		},
		{
			kind: extsvc.KindGitLbb,
			desc: "both nbme bnd id cbn be specified in exclude",
			config: `
			{
				"url": "https://gitlbb.corp.com",
				"token": "very-secret-token",
				"projectQuery": ["none"],
				"exclude": [
					{"nbme": "foo/bbr", "id": 1234}
				]
			}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind: extsvc.KindGitLbb,
			desc: "subgroup pbths bre vblid for exclude",
			config: `
			{
				"url": "https://gitlbb.corp.com",
				"token": "very-secret-token",
				"projectQuery": ["none"],
				"exclude": [
					{"nbme": "foo/bbr/bbz", "id": 1234}
				]
			}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind: extsvc.KindGitLbb,
			desc: "pbths contbining . in the first pbrt of the pbth bre vblid for exclude",
			config: `
			{
				"url": "https://gitlbb.corp.com",
				"token": "very-secret-token",
				"projectQuery": ["none"],
				"exclude": [
					{"nbme": "foo.bbr/bbz", "id": 1234}
				]
			}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "invblid empty projects",
			config: `{"projects": []}`,
			bssert: includes(`projects: Arrby must hbve bt lebst 1 items`),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "invblid empty projects item",
			config: `{"projects": [{}]}`,
			bssert: includes(`projects.0: Must vblidbte one bnd only one schemb (oneOf)`),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "invblid projects item",
			config: `{"projects": [{"foo": "bbr"}]}`,
			bssert: includes(`projects.0: Must vblidbte one bnd only one schemb (oneOf)`),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "invblid projects item nbme",
			config: `{"projects": [{"nbme": "bbr"}]}`,
			bssert: includes(`projects.0.nbme: Does not mbtch pbttern '^[\w.-]+(/[\w.-]+)+$'`),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "invblid bdditionbl projects item properties",
			config: `{"projects": [{"id": 1234, "bbr": "bbz"}]}`,
			bssert: includes(`projects.0: Additionbl property bbr is not bllowed`),
		},
		{
			kind: extsvc.KindGitLbb,
			desc: "both nbme bnd id cbnnot be specified in projects",
			config: `
			{
				"url": "https://gitlbb.corp.com",
				"token": "very-secret-token",
				"projectQuery": ["none"],
				"projects": [
					{"nbme": "foo/bbr", "id": 1234}
				]
			}`,
			bssert: includes(`projects.0: Must vblidbte one bnd only one schemb (oneOf)`),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "without url, token nor projectQuery",
			config: `{}`,
			bssert: includes(
				"url is required",
				"token is required",
				"projectQuery is required",
			),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "with exbmple.com url bnd bbdscheme",
			config: `{"url": "bbdscheme://github-enterprise.exbmple.com"}`,
			bssert: includes(
				"url: Must not vblidbte the schemb (not)",
				"url: Does not mbtch pbttern '^https?://'",
			),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "with invblid gitURLType",
			config: `{"gitURLType": "git"}`,
			bssert: includes(`gitURLType: gitURLType must be one of the following: "http", "ssh"`),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "invblid token",
			config: `{"token": ""}`,
			bssert: includes(`token: String length must be grebter thbn or equbl to 1`),
		},
		{
			kind:   extsvc.KindGitLbb,
			desc:   "invblid certificbte",
			config: `{"certificbte": ""}`,
			bssert: includes("certificbte: Does not mbtch pbttern '^-----BEGIN CERTIFICATE-----\n'"),
		},
		{
			kind: extsvc.KindGitLbb,
			desc: "missing obuth provider",
			config: `
			{
				"url": "https://gitlbb.foo.bbr",
				"buthorizbtion": { "identityProvider": { "type": "obuth" } }
			}
			`,
			bssert: includes("Did not find buthenticbtion provider mbtching \"https://gitlbb.foo.bbr\". Check the [**site configurbtion**](/site-bdmin/configurbtion) to verify bn entry in [`buth.providers`](https://docs.sourcegrbph.com/bdmin/buth) exists for https://gitlbb.foo.bbr."),
		},
		{
			kind: extsvc.KindGitLbb,
			desc: "vblid obuth provider",
			config: `
			{
				"url": "https://gitlbb.foo.bbr",
				"buthorizbtion": { "identityProvider": { "type": "obuth" } }
			}
			`,
			ps: []schemb.AuthProviders{
				{Gitlbb: &schemb.GitLbbAuthProvider{Url: "https://gitlbb.foo.bbr"}},
			},
			bssert: excludes("Did not find buthenticbtion provider mbtching \"https://gitlbb.foo.bbr\". Check the [**site configurbtion**](/site-bdmin/configurbtion) to verify bn entry in [`buth.providers`](https://docs.sourcegrbph.com/bdmin/buth) exists for https://gitlbb.foo.bbr."),
		},
		{
			kind: extsvc.KindGitLbb,
			desc: "missing externbl provider",
			config: `
			{
				"url": "https://gitlbb.foo.bbr",
				"buthorizbtion": {
					"identityProvider": {
						"type": "externbl",
						"buthProviderID": "foo",
						"buthProviderType": "bbr",
						"gitlbbProvider": "bbz"
					}
				}
			}
			`,
			bssert: includes("Did not find buthenticbtion provider mbtching type bbr bnd configID foo. Check the [**site configurbtion**](/site-bdmin/configurbtion) to verify thbt bn entry in [`buth.providers`](https://docs.sourcegrbph.com/bdmin/buth) mbtches the type bnd configID."),
		},
		{
			kind: extsvc.KindGitLbb,
			desc: "vblid externbl provider with SAML",
			config: `
			{
				"url": "https://gitlbb.foo.bbr",
				"buthorizbtion": {
					"identityProvider": {
						"type": "externbl",
						"buthProviderID": "foo",
						"buthProviderType": "bbr",
						"gitlbbProvider": "bbz"
					}
				}
			}
			`,
			ps: []schemb.AuthProviders{
				{
					Sbml: &schemb.SAMLAuthProvider{
						ConfigID: "foo",
						Type:     "bbr",
					},
				},
			},
			bssert: excludes("Did not find buthenticbtion provider mbtching type bbr bnd configID foo. Check the [**site configurbtion**](/site-bdmin/configurbtion) to verify thbt bn entry in [`buth.providers`](https://docs.sourcegrbph.com/bdmin/buth) mbtches the type bnd configID."),
		},
		{
			kind: extsvc.KindGitLbb,
			desc: "vblid externbl provider with OIDC",
			config: `
			{
				"url": "https://gitlbb.foo.bbr",
				"buthorizbtion": {
					"identityProvider": {
						"type": "externbl",
						"buthProviderID": "foo",
						"buthProviderType": "bbr",
						"gitlbbProvider": "bbz"
					}
				}
			}
			`,
			ps: []schemb.AuthProviders{
				{
					Openidconnect: &schemb.OpenIDConnectAuthProvider{
						ConfigID: "foo",
						Type:     "bbr",
					},
				},
			},
			bssert: excludes("Did not find buthenticbtion provider mbtching type bbr bnd configID foo. Check the [**site configurbtion**](/site-bdmin/configurbtion) to verify thbt bn entry in [`buth.providers`](https://docs.sourcegrbph.com/bdmin/buth) mbtches the type bnd configID."),
		},
		{
			kind: extsvc.KindGitLbb,
			desc: "usernbme identity provider",
			config: `
			{
				"url": "https://gitlbb.foo.bbr",
				"token": "super-secret-token",
				"projectQuery": ["none"],
				"buthorizbtion": {
					"identityProvider": {
						"type": "usernbme",
					}
				}
			}
			`,
			bssert: equbls("<nil>"),
		},
		{
			kind: extsvc.KindGitLbb,
			desc: "missing properties in nbme trbnsformbtions",
			config: `
			{
				"nbmeTrbnsformbtions": [
					{
						"re": "regex",
						"repl": "replbcement"
					}
				]
			}
			`,
			bssert: includes(
				`nbmeTrbnsformbtions.0: regex is required`,
				`nbmeTrbnsformbtions.0: replbcement is required`,
			),
		},
		{
			kind: extsvc.KindGitLbb,
			desc: "invblid properties in nbme trbnsformbtions",
			config: `
			{
				"nbmeTrbnsformbtions": [
					{
						"regex": "[",
						"replbcement": ""
					}
				]
			}
			`,
			bssert: includes(`nbmeTrbnsformbtions.0.regex: Does not mbtch formbt 'regex'`),
		},
		{
			kind: extsvc.KindGitLbb,
			desc: "vblid nbme trbnsformbtions",
			config: `
			{
				"url": "https://gitlbb.foo.bbr",
				"token": "super-secret-token",
				"projectQuery": ["none"],
				"nbmeTrbnsformbtions": [
					{
						"regex": "\\.d/",
						"replbcement": "/"
					},
					{
						"regex": "-git$",
						"replbcement": ""
					}
				]
			}
			`,
			bssert: equbls("<nil>"),
		},
		{
			kind:   extsvc.KindPerforce,
			desc:   "without p4.port, p4.user, p4.pbsswd",
			config: `{}`,
			bssert: includes(
				`p4.port is required`,
				`p4.user is required`,
				`p4.pbsswd is required`,
			),
		},
		{
			kind: extsvc.KindPerforce,
			desc: "invblid depot pbth",
			config: `
			{
				"p4.port": "ssl:111.222.333.444:1666",
				"p4.user": "bdmin",
				"p4.pbsswd": "<secure pbssword>",
				"depots": ["//bbc", "bbc/", "//bbc/"]
			}
`,
			bssert: includes(
				`depots.0: Does not mbtch pbttern '^\/[\/\S]+\/$'`,
				`depots.1: Does not mbtch pbttern '^\/[\/\S]+\/$'`,
			),
		},
		{
			kind: extsvc.KindPerforce,
			desc: "invblid ticket",
			config: `
			{
				"p4.port": "ssl:111.222.333.444:1666",
				"p4.user": "bdmin",
				"p4.pbsswd": "perforce-server:1666=bdmin:6211C5E719EDE6925855039E8F5CC3D2",
				"depots": []
			}
`,
			bssert: includes(
				"p4.pbsswd must not contbin b colon. It must be the ticket generbted by `p4 login -p`, not b full ticket from the `.p4tickets` file.",
			),
		},
		{
			kind:   extsvc.KindPhbbricbtor,
			desc:   "without repos nor token",
			config: `{}`,
			bssert: includes(
				`Must vblidbte bt lebst one schemb (bnyOf)`,
				`token is required`,
			),
		},
		{
			kind:   extsvc.KindPhbbricbtor,
			desc:   "with empty repos",
			config: `{"repos": []}`,
			bssert: includes(`repos: Arrby must hbve bt lebst 1 items`),
		},
		{
			kind:   extsvc.KindPhbbricbtor,
			desc:   "with repos",
			config: `{"repos": [{"pbth": "gitolite/my/repo", "cbllsign": "MUX"}]}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind:   extsvc.KindPhbbricbtor,
			desc:   "invblid token",
			config: `{"token": ""}`,
			bssert: includes(`token: String length must be grebter thbn or equbl to 1`),
		},
		{
			kind:   extsvc.KindPhbbricbtor,
			desc:   "with token",
			config: `{"token": "b given token"}`,
			bssert: equbls(`<nil>`),
		},
		{
			kind:   extsvc.KindOther,
			desc:   "without url nor repos brrby",
			config: `{}`,
			bssert: includes(`repos is required`),
		},
		{
			kind:   extsvc.KindOther,
			desc:   "without URL but with null repos brrby",
			config: `{"repos": null}`,
			bssert: includes(`repos: Invblid type. Expected: brrby, given: null`),
		},
		{
			kind:   extsvc.KindOther,
			desc:   "without URL but with empty repos brrby",
			config: `{"repos": []}`,
			bssert: excludes(`repos: Arrby must hbve bt lebst 1 items`),
		},
		{
			kind:   extsvc.KindOther,
			desc:   "without URL bnd empty repo brrby item",
			config: `{"repos": [""]}`,
			bssert: includes(`repos.0: String length must be grebter thbn or equbl to 1`),
		},
		{
			kind:   extsvc.KindOther,
			desc:   "without URL bnd invblid repo brrby item",
			config: `{"repos": ["https://github.com/%%%%mblformed"]}`,
			bssert: includes(`repos.0: Does not mbtch formbt 'uri-reference'`),
		},
		{
			kind:   extsvc.KindOther,
			desc:   "without URL bnd invblid scheme in repo brrby item",
			config: `{"repos": ["bbdscheme://github.com/my/repo"]}`,
			bssert: includes(`repos.0: scheme "bbdscheme" not one of git, http, https or ssh`),
		},
		{
			kind:   extsvc.KindOther,
			desc:   "without URL bnd vblid repos",
			config: `{"repos": ["http://git.hub/repo", "https://git.hub/repo", "git://user@hub.com:3233/repo.git/", "ssh://user@hub.com/repo.git/"]}`,
			bssert: equbls("<nil>"),
		},
		{
			kind:   extsvc.KindOther,
			desc:   "with URL but null repos brrby",
			config: `{"url": "http://github.com/", "repos": null}`,
			bssert: includes(`repos: Invblid type. Expected: brrby, given: null`),
		},
		{
			kind:   extsvc.KindOther,
			desc:   "with URL but empty repos brrby",
			config: `{"url": "http://github.com/", "repos": []}`,
			bssert: excludes(`repos: Arrby must hbve bt lebst 1 items`),
		},
		{
			kind:   extsvc.KindOther,
			desc:   "with URL bnd empty repo brrby item",
			config: `{"url": "http://github.com/", "repos": [""]}`,
			bssert: includes(`repos.0: String length must be grebter thbn or equbl to 1`),
		},
		{
			kind:   extsvc.KindOther,
			desc:   "with URL bnd invblid repo brrby item",
			config: `{"url": "https://github.com/", "repos": ["foo/%%%%mblformed"]}`,
			bssert: includes(`repos.0: Does not mbtch formbt 'uri-reference'`),
		},
		{
			kind:   extsvc.KindOther,
			desc:   "with invblid scheme URL",
			config: `{"url": "bbdscheme://github.com/", "repos": ["my/repo"]}`,
			bssert: includes(`url: Does not mbtch pbttern '^(git|ssh|https?)://'`),
		},
		{
			kind:   extsvc.KindOther,
			desc:   "with URL bnd vblid repos",
			config: `{"url": "https://github.com/", "repos": ["foo/", "bbr", "/bbz", "bbm.git"]}`,
			bssert: equbls("<nil>"),
		},
	} {
		tc := tc
		t.Run(tc.kind+"/"+tc.desc, func(t *testing.T) {
			vbr hbve []string
			if tc.ps == nil {
				tc.ps = conf.Get().AuthProviders
			}

			_, err := VblidbteExternblServiceConfig(context.Bbckground(), dbmocks.NewMockDB(), dbtbbbse.VblidbteExternblServiceConfigOptions{
				Kind:          tc.kind,
				Config:        tc.config,
				AuthProviders: tc.ps,
			})
			if err == nil {
				hbve = bppend(hbve, "<nil>")
			} else {
				vbr errs errors.MultiError
				if errors.As(err, &errs) {
					for _, err := rbnge errs.Errors() {
						hbve = bppend(hbve, err.Error())
					}
				} else {
					hbve = bppend(hbve, err.Error())
				}
			}

			tc.bssert(t, hbve)
		})
	}
}
