package main

import (
	"context"
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/schema"
)

type gitlabAuthzProviderParams struct {
	OAuthOp gitlab.OAuthAuthzProviderOp
	SudoOp  gitlab.SudoProviderOp
}

func (m gitlabAuthzProviderParams) RepoPerms(ctx context.Context, account *extsvc.ExternalAccount, repos []*types.Repo) ([]authz.RepoPerms, error) {
	panic("should never be called")
}

func (m gitlabAuthzProviderParams) Repos(ctx context.Context, repos []*types.Repo) (mine []*types.Repo, others []*types.Repo) {
	panic("should never be called")
}

func (m gitlabAuthzProviderParams) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.ExternalAccount) (mine *extsvc.ExternalAccount, err error) {
	panic("should never be called")
}

func (m gitlabAuthzProviderParams) ServiceID() string {
	panic("should never be called")
}

func (m gitlabAuthzProviderParams) ServiceType() string {
	return "gitlab"
}
func (m gitlabAuthzProviderParams) Validate() []string { return nil }

func Test_authzProvidersFromConfig(t *testing.T) {
	gitlab.NewOAuthProvider = func(op gitlab.OAuthAuthzProviderOp) authz.Provider {
		op.MockCache = nil // ignore cache value
		return gitlabAuthzProviderParams{OAuthOp: op}
	}
	gitlab.NewSudoProvider = func(op gitlab.SudoProviderOp) authz.Provider {
		op.MockCache = nil // ignore cache value
		return gitlabAuthzProviderParams{SudoOp: op}
	}

	providersEqual := func(want ...authz.Provider) func(*testing.T, []authz.Provider) {
		return func(t *testing.T, have []authz.Provider) {
			if !reflect.DeepEqual(have, want) {
				t.Errorf("authzProviders: (actual) %+v != (expected) %+v", asJSON(t, have), asJSON(t, want))
			}
		}
	}

	const bogusKey = `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlCUEFJQkFBSkJBUEpIaWprdG1UMUlLYUd0YTVFZXAzQVo5Q2VPZUw4alBESUZUN3dRZ0tabXQzRUZxRGhCCk93bitRVUhKdUs5Zm92UkROSmVWTDJvWTVCT0l6NHJ3L0cwQ0F3RUFBUUpCQU1BK0o5Mks0d2NQVllsbWMrM28KcHU5NmlKTkNwMmp5Nm5hK1pEQlQzK0VvSUo1VFJGdnN3R2kvTHUzZThYUWwxTDNTM21ub0xPSlZNcTF0bUxOMgpIY0VDSVFEK3daeS83RlYxUEFtdmlXeWlYVklETzJnNWJOaUJlbmdKQ3hFa3Nia1VtUUloQVBOMlZaczN6UFFwCk1EVG9vTlJXcnl0RW1URERkamdiOFpzTldYL1JPRGIxQWlCZWNKblNVQ05TQllLMXJ5VTFmNURTbitoQU9ZaDkKWDFBMlVnTDE3bWhsS1FJaEFPK2JMNmRDWktpTGZORWxmVnRkTUtxQnFjNlBIK01heFU2VzlkVlFvR1dkQWlFQQptdGZ5cE9zYTFiS2hFTDg0blovaXZFYkJyaVJHalAya3lERHYzUlg0V0JrPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=`

	tests := []struct {
		description                  string
		cfg                          conf.Unified
		gitlabConnections            []*schema.GitLabConnection
		bitbucketServerConnections   []*schema.BitbucketServerConnection
		expAuthzAllowAccessByDefault bool
		expAuthzProviders            func(*testing.T, []authz.Provider)
		expSeriousProblems           []string
	}{
		{
			description: "1 GitLab connection with authz enabled, 1 GitLab matching auth provider",
			cfg: conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthProviders: []schema.AuthProviders{{
						Gitlab: &schema.GitLabAuthProvider{
							ClientID:     "clientID",
							ClientSecret: "clientSecret",
							DisplayName:  "GitLab",
							Type:         "gitlab",
							Url:          "https://gitlab.mine",
						},
					}},
				},
			},
			gitlabConnections: []*schema.GitLabConnection{
				{
					Authorization: &schema.GitLabAuthorization{
						IdentityProvider: schema.IdentityProvider{Oauth: &schema.OAuthIdentity{Type: "oauth"}},
						Ttl:              "48h",
					},
					Url:   "https://gitlab.mine",
					Token: "asdf",
				},
			},
			expAuthzAllowAccessByDefault: true,
			expAuthzProviders: providersEqual(
				gitlabAuthzProviderParams{
					OAuthOp: gitlab.OAuthAuthzProviderOp{
						BaseURL:  mustURLParse(t, "https://gitlab.mine"),
						CacheTTL: 48 * time.Hour,
					},
				},
			),
		},
		{
			description: "1 GitLab connection with authz enabled, 1 GitLab auth provider but doesn't match",
			cfg: conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthProviders: []schema.AuthProviders{{
						Gitlab: &schema.GitLabAuthProvider{
							ClientID:     "clientID",
							ClientSecret: "clientSecret",
							DisplayName:  "GitLab",
							Type:         "gitlab",
							Url:          "https://gitlab.com",
						},
					}},
				},
			},
			gitlabConnections: []*schema.GitLabConnection{
				{
					Authorization: &schema.GitLabAuthorization{
						IdentityProvider: schema.IdentityProvider{Oauth: &schema.OAuthIdentity{Type: "oauth"}},
						Ttl:              "48h",
					},
					Url:   "https://gitlab.mine",
					Token: "asdf",
				},
			},
			expAuthzAllowAccessByDefault: false,
			expSeriousProblems:           []string{"Did not find authentication provider matching \"https://gitlab.mine\". Check the [**site configuration**](/site-admin/configuration) to verify an entry in [`auth.providers`](https://docs.sourcegraph.com/admin/auth) exists for https://gitlab.mine."},
		},
		{
			description: "1 GitLab connection with authz enabled, no GitLab auth provider",
			cfg: conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthProviders: []schema.AuthProviders{{
						Builtin: &schema.BuiltinAuthProvider{Type: "builtin"},
					}},
				},
			},
			gitlabConnections: []*schema.GitLabConnection{
				{
					Authorization: &schema.GitLabAuthorization{
						IdentityProvider: schema.IdentityProvider{Oauth: &schema.OAuthIdentity{Type: "oauth"}},
						Ttl:              "48h",
					},
					Url:   "https://gitlab.mine",
					Token: "asdf",
				},
			},
			expAuthzAllowAccessByDefault: false,
			expSeriousProblems:           []string{"Did not find authentication provider matching \"https://gitlab.mine\". Check the [**site configuration**](/site-admin/configuration) to verify an entry in [`auth.providers`](https://docs.sourcegraph.com/admin/auth) exists for https://gitlab.mine."},
		},
		{
			description: "Two GitLab connections with authz enabled, two matching GitLab auth providers",
			cfg: conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthProviders: []schema.AuthProviders{
						{
							Gitlab: &schema.GitLabAuthProvider{
								ClientID:     "clientID",
								ClientSecret: "clientSecret",
								DisplayName:  "GitLab.com",
								Type:         "gitlab",
								Url:          "https://gitlab.com",
							},
						}, {
							Gitlab: &schema.GitLabAuthProvider{
								ClientID:     "clientID",
								ClientSecret: "clientSecret",
								DisplayName:  "GitLab.mine",
								Type:         "gitlab",
								Url:          "https://gitlab.mine",
							},
						},
					},
				},
			},
			gitlabConnections: []*schema.GitLabConnection{
				{
					Authorization: &schema.GitLabAuthorization{
						IdentityProvider: schema.IdentityProvider{Oauth: &schema.OAuthIdentity{Type: "oauth"}},
					},
					Url:   "https://gitlab.mine",
					Token: "asdf",
				},
				{
					Authorization: &schema.GitLabAuthorization{
						IdentityProvider: schema.IdentityProvider{Oauth: &schema.OAuthIdentity{Type: "oauth"}},
					},
					Url:   "https://gitlab.com",
					Token: "asdf",
				},
			},
			expAuthzAllowAccessByDefault: true,
			expAuthzProviders: providersEqual(
				gitlabAuthzProviderParams{
					OAuthOp: gitlab.OAuthAuthzProviderOp{
						BaseURL:  mustURLParse(t, "https://gitlab.mine"),
						CacheTTL: 3 * time.Hour,
					},
				},
				gitlabAuthzProviderParams{
					OAuthOp: gitlab.OAuthAuthzProviderOp{
						BaseURL:  mustURLParse(t, "https://gitlab.com"),
						CacheTTL: 3 * time.Hour,
					},
				},
			),
		},
		{
			description: "1 GitLab connection with authz disabled",
			cfg: conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthProviders: []schema.AuthProviders{{
						Gitlab: &schema.GitLabAuthProvider{
							ClientID:     "clientID",
							ClientSecret: "clientSecret",
							DisplayName:  "GitLab",
							Type:         "gitlab",
							Url:          "https://gitlab.mine",
						},
					}},
				},
			},
			gitlabConnections: []*schema.GitLabConnection{
				{
					Authorization: nil,
					Url:           "https://gitlab.mine",
					Token:         "asdf",
				},
			},
			expAuthzAllowAccessByDefault: true,
			expAuthzProviders:            nil,
		},
		{
			description: "TTL error",
			cfg: conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthProviders: []schema.AuthProviders{{
						Gitlab: &schema.GitLabAuthProvider{
							ClientID:     "clientID",
							ClientSecret: "clientSecret",
							DisplayName:  "GitLab",
							Type:         "gitlab",
							Url:          "https://gitlab.mine",
						},
					}},
				},
			},
			gitlabConnections: []*schema.GitLabConnection{
				{
					Authorization: &schema.GitLabAuthorization{
						IdentityProvider: schema.IdentityProvider{Oauth: &schema.OAuthIdentity{Type: "oauth"}},
						Ttl:              "invalid",
					},
					Url:   "https://gitlab.mine",
					Token: "asdf",
				},
			},
			expAuthzAllowAccessByDefault: false,
			expSeriousProblems:           []string{"authorization.ttl: time: invalid duration invalid"},
		},
		{
			description: "external auth provider",
			cfg: conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthProviders: []schema.AuthProviders{{
						Saml: &schema.SAMLAuthProvider{
							ConfigID: "okta",
							Type:     "saml",
						},
					}},
				},
			},
			gitlabConnections: []*schema.GitLabConnection{
				{
					Authorization: &schema.GitLabAuthorization{
						IdentityProvider: schema.IdentityProvider{External: &schema.ExternalIdentity{
							Type:             "external",
							AuthProviderID:   "okta",
							AuthProviderType: "saml",
							GitlabProvider:   "my-external",
						}},
					},
					Url:   "https://gitlab.mine",
					Token: "asdf",
				},
			},
			expAuthzAllowAccessByDefault: true,
			expAuthzProviders: providersEqual(
				gitlabAuthzProviderParams{
					SudoOp: gitlab.SudoProviderOp{
						BaseURL: mustURLParse(t, "https://gitlab.mine"),
						AuthnConfigID: providers.ConfigID{
							Type: "saml",
							ID:   "okta",
						},
						GitLabProvider:    "my-external",
						SudoToken:         "asdf",
						CacheTTL:          3 * time.Hour,
						UseNativeUsername: false,
					},
				},
			),
		},
		{
			description: "exact username matching",
			cfg: conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthProviders: []schema.AuthProviders{},
				},
			},
			gitlabConnections: []*schema.GitLabConnection{
				{
					Authorization: &schema.GitLabAuthorization{
						IdentityProvider: schema.IdentityProvider{Username: &schema.UsernameIdentity{Type: "username"}},
					},
					Url:   "https://gitlab.mine",
					Token: "asdf",
				},
			},
			expAuthzAllowAccessByDefault: true,
			expAuthzProviders: providersEqual(
				gitlabAuthzProviderParams{
					SudoOp: gitlab.SudoProviderOp{
						BaseURL:           mustURLParse(t, "https://gitlab.mine"),
						SudoToken:         "asdf",
						CacheTTL:          3 * time.Hour,
						UseNativeUsername: true,
					},
				},
			),
		},
		{
			description: "1 BitbucketServer connection with authz disabled",
			bitbucketServerConnections: []*schema.BitbucketServerConnection{
				{
					Authorization: nil,
					Url:           "https://bitbucket.mycorp.org",
					Username:      "admin",
					Token:         "secret-token",
				},
			},
			expAuthzAllowAccessByDefault: true,
			expAuthzProviders:            providersEqual(),
		},
		{
			description: "Bitbucket Server TTL error",
			cfg:         conf.Unified{},
			bitbucketServerConnections: []*schema.BitbucketServerConnection{
				{
					Authorization: &schema.BitbucketServerAuthorization{
						IdentityProvider: schema.BitbucketServerIdentityProvider{
							Username: &schema.BitbucketServerUsernameIdentity{
								Type: "username",
							},
						},
						Oauth: schema.BitbucketServerOAuth{
							ConsumerKey: "sourcegraph",
							SigningKey:  bogusKey,
						},
						Ttl: "invalid",
					},
					Url:      "https://bitbucketserver.mycorp.org",
					Username: "admin",
					Token:    "secret-token",
				},
			},
			expAuthzAllowAccessByDefault: false,
			expSeriousProblems:           []string{"1 error occurred:\n\t* authorization.ttl: time: invalid duration invalid\n\n"},
		},
		{
			description: "Bitbucket Server Oauth config error",
			cfg:         conf.Unified{},
			bitbucketServerConnections: []*schema.BitbucketServerConnection{
				{
					Authorization: &schema.BitbucketServerAuthorization{
						IdentityProvider: schema.BitbucketServerIdentityProvider{
							Username: &schema.BitbucketServerUsernameIdentity{
								Type: "username",
							},
						},
						Oauth: schema.BitbucketServerOAuth{
							ConsumerKey: "sourcegraph",
							SigningKey:  "Invalid Key",
						},
						Ttl: "15m",
					},
					Url:      "https://bitbucketserver.mycorp.org",
					Username: "admin",
					Token:    "secret-token",
				},
			},
			expAuthzAllowAccessByDefault: false,
			expSeriousProblems:           []string{"1 error occurred:\n\t* authorization.oauth.signingKey: illegal base64 data at input byte 7\n\n"},
		},
		{
			description: "Bitbucket Server exact username matching",
			cfg:         conf.Unified{},
			bitbucketServerConnections: []*schema.BitbucketServerConnection{
				{
					Authorization: &schema.BitbucketServerAuthorization{
						IdentityProvider: schema.BitbucketServerIdentityProvider{
							Username: &schema.BitbucketServerUsernameIdentity{
								Type: "username",
							},
						},
						Oauth: schema.BitbucketServerOAuth{
							ConsumerKey: "sourcegraph",
							SigningKey:  bogusKey,
						},
						Ttl: "15m",
					},
					Url:      "https://bitbucketserver.mycorp.org",
					Username: "admin",
					Token:    "secret-token",
				},
			},
			expAuthzAllowAccessByDefault: true,
			expAuthzProviders: func(t *testing.T, have []authz.Provider) {
				if len(have) == 0 {
					t.Fatalf("no providers")
				}

				if have[0].ServiceType() != bitbucketserver.ServiceType {
					t.Fatalf("no Bitbucket Server authz provider returned")
				}
			},
		},

		// For Sourcegraph authz provider
		{
			description: "Conflicted configuration between Sourcegraph and GitLab authz provider",
			cfg: conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					PermissionsUserMapping: &schema.PermissionsUserMapping{
						Enabled: true,
						BindID:  "email",
					},
					AuthProviders: []schema.AuthProviders{{
						Gitlab: &schema.GitLabAuthProvider{
							ClientID:     "clientID",
							ClientSecret: "clientSecret",
							DisplayName:  "GitLab",
							Type:         "gitlab",
							Url:          "https://gitlab.mine",
						},
					}},
				},
			},
			gitlabConnections: []*schema.GitLabConnection{
				{
					Authorization: &schema.GitLabAuthorization{
						IdentityProvider: schema.IdentityProvider{Oauth: &schema.OAuthIdentity{Type: "oauth"}},
						Ttl:              "48h",
					},
					Url:   "https://gitlab.mine",
					Token: "asdf",
				},
			},
			expAuthzAllowAccessByDefault: false,
			expSeriousProblems:           []string{"The Sourcegraph permissions (`permissions.userMapping`) cannot be enabled when \"gitlab\" authorization providers are in use. Blocking access to all repositories until the conflict is resolved."},
		},
		{
			description: "Conflicted configuration between Sourcegraph and Bitbucket Server authz provider",
			cfg: conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					PermissionsUserMapping: &schema.PermissionsUserMapping{
						Enabled: true,
						BindID:  "email",
					},
				},
			},
			bitbucketServerConnections: []*schema.BitbucketServerConnection{
				{
					Authorization: &schema.BitbucketServerAuthorization{
						IdentityProvider: schema.BitbucketServerIdentityProvider{
							Username: &schema.BitbucketServerUsernameIdentity{
								Type: "username",
							},
						},
						Oauth: schema.BitbucketServerOAuth{
							ConsumerKey: "sourcegraph",
							SigningKey:  bogusKey,
						},
						Ttl: "15m",
					},
					Url:      "https://bitbucketserver.mycorp.org",
					Username: "admin",
					Token:    "secret-token",
				},
			},
			expAuthzAllowAccessByDefault: false,
			expSeriousProblems:           []string{"The Sourcegraph permissions (`permissions.userMapping`) cannot be enabled when \"bitbucketServer\" authorization providers are in use. Blocking access to all repositories until the conflict is resolved."},
		},
	}

	for _, test := range tests {
		t.Logf("Test %q", test.description)

		store := fakeStore{
			gitlabs:          test.gitlabConnections,
			bitbucketServers: test.bitbucketServerConnections,
		}

		allowAccessByDefault, authzProviders, seriousProblems, _ :=
			authzProvidersFromConfig(context.Background(), &test.cfg, &store, nil)
		if allowAccessByDefault != test.expAuthzAllowAccessByDefault {
			t.Errorf("allowAccessByDefault: (actual) %v != (expected) %v", asJSON(t, allowAccessByDefault), asJSON(t, test.expAuthzAllowAccessByDefault))
		}
		if test.expAuthzProviders != nil {
			test.expAuthzProviders(t, authzProviders)
		}
		if !reflect.DeepEqual(seriousProblems, test.expSeriousProblems) {
			t.Errorf("seriousProblems: (actual) %+v != (expected) %+v", asJSON(t, seriousProblems), asJSON(t, test.expSeriousProblems))
		}
	}
}

func mustURLParse(t *testing.T, u string) *url.URL {
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func asJSON(t *testing.T, v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

type fakeStore struct {
	gitlabs          []*schema.GitLabConnection
	githubs          []*schema.GitHubConnection
	bitbucketServers []*schema.BitbucketServerConnection
}

func (s fakeStore) ListGitHubConnections(context.Context) ([]*schema.GitHubConnection, error) {
	return s.githubs, nil
}

func (s fakeStore) ListGitLabConnections(context.Context) ([]*schema.GitLabConnection, error) {
	return s.gitlabs, nil
}

func (s fakeStore) ListBitbucketServerConnections(context.Context) ([]*schema.BitbucketServerConnection, error) {
	return s.bitbucketServers, nil
}
