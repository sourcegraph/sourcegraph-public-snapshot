package authz

import (
	"context"
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/gitlab"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

type newGitLabAuthzProviderParams struct {
	Op gitlab.GitLabAuthzProviderOp
}

func (m newGitLabAuthzProviderParams) RepoPerms(ctx context.Context, account *extsvc.ExternalAccount, repos map[authz.Repo]struct{}) (map[api.RepoName]map[authz.Perm]bool, error) {
	panic("should never be called")
}
func (m newGitLabAuthzProviderParams) Repos(ctx context.Context, repos map[authz.Repo]struct{}) (mine map[authz.Repo]struct{}, others map[authz.Repo]struct{}) {
	panic("should never be called")
}
func (m newGitLabAuthzProviderParams) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.ExternalAccount) (mine *extsvc.ExternalAccount, err error) {
	panic("should never be called")
}
func (m newGitLabAuthzProviderParams) ServiceID() string {
	panic("should never be called")
}
func (m newGitLabAuthzProviderParams) ServiceType() string {
	panic("should never be called")
}
func (m newGitLabAuthzProviderParams) Validate() []string { return nil }

func Test_providersFromConfig(t *testing.T) {
	NewGitLabProvider = func(op gitlab.GitLabAuthzProviderOp) authz.Provider {
		op.MockCache = nil // ignore cache value
		return newGitLabAuthzProviderParams{op}
	}

	tests := []struct {
		description                  string
		cfg                          conf.Unified
		expAuthzAllowAccessByDefault bool
		expAuthzProviders            []authz.Provider
		expSeriousProblems           []string
		expWarnings                  []string
	}{
		{
			description: "1 auth provider (okta), 1 GitLab referencing okta",
			cfg: conf.Unified{
				Critical: schema.CriticalConfiguration{
					AuthProviders: []schema.AuthProviders{
						schema.AuthProviders{
							Saml: &schema.SAMLAuthProvider{
								ConfigID: "okta-config-id",
								Type:     "saml",
							},
						},
					},
				},
				SiteConfiguration: schema.SiteConfiguration{
					Gitlab: []*schema.GitLabConnection{
						{
							Authorization: &schema.GitLabAuthorization{
								AuthnProvider: schema.AuthnProvider{
									ConfigID:       "okta-config-id",
									Type:           "saml",
									GitlabProvider: "okta",
								},
								Ttl: "48h",
							},
							Url:   "https://gitlab.mine",
							Token: "asdf",
						},
					},
				},
			},
			expAuthzAllowAccessByDefault: true,
			expAuthzProviders: []authz.Provider{
				newGitLabAuthzProviderParams{
					Op: gitlab.GitLabAuthzProviderOp{
						BaseURL:        mustURLParse(t, "https://gitlab.mine"),
						AuthnConfigID:  auth.ProviderConfigID{Type: "saml", ID: "okta-config-id"},
						SudoToken:      "asdf",
						GitLabProvider: "okta",
						CacheTTL:       48 * time.Hour,
					},
				},
			},
		},
		{
			description: "2 auth providers (okta, onelogin), 2 GitLabs referencing okta and onelogin, respectively",
			cfg: conf.Unified{
				Critical: schema.CriticalConfiguration{
					AuthProviders: []schema.AuthProviders{
						schema.AuthProviders{
							Saml: &schema.SAMLAuthProvider{
								ConfigID: "okta-config-id",
								Type:     "saml",
							},
						},
						schema.AuthProviders{
							Openidconnect: &schema.OpenIDConnectAuthProvider{
								ConfigID: "onelogin-config-id",
								Type:     "openidconnect",
							},
						},
					},
				},
				SiteConfiguration: schema.SiteConfiguration{
					Gitlab: []*schema.GitLabConnection{
						{
							Authorization: &schema.GitLabAuthorization{
								AuthnProvider: schema.AuthnProvider{
									ConfigID:       "onelogin-config-id",
									GitlabProvider: "onelogin",
									Type:           "openidconnect",
								},
							},
							Url:   "https://gitlab-0.mine",
							Token: "asdf",
						},
						{
							Authorization: &schema.GitLabAuthorization{
								AuthnProvider: schema.AuthnProvider{
									ConfigID:       "okta-config-id",
									GitlabProvider: "okta",
									Type:           "saml",
								},
							},
							Url:   "https://gitlab-1.mine",
							Token: "asdf",
						},
					},
				},
			},
			expAuthzAllowAccessByDefault: true,
			expAuthzProviders: []authz.Provider{
				newGitLabAuthzProviderParams{
					Op: gitlab.GitLabAuthzProviderOp{
						BaseURL:        mustURLParse(t, "https://gitlab-0.mine"),
						AuthnConfigID:  auth.ProviderConfigID{Type: "openidconnect", ID: "onelogin-config-id"},
						SudoToken:      "asdf",
						GitLabProvider: "onelogin",
						CacheTTL:       3 * time.Hour,
					},
				},
				newGitLabAuthzProviderParams{
					Op: gitlab.GitLabAuthzProviderOp{
						BaseURL:        mustURLParse(t, "https://gitlab-1.mine"),
						AuthnConfigID:  auth.ProviderConfigID{Type: "saml", ID: "okta-config-id"},
						SudoToken:      "asdf",
						GitLabProvider: "okta",
						CacheTTL:       3 * time.Hour,
					},
				},
			},
		},
		{
			description: "0 auth providers, 1 GitLab referencing non-existent auth provider",
			cfg: conf.Unified{SiteConfiguration: schema.SiteConfiguration{
				Gitlab: []*schema.GitLabConnection{
					{
						Authorization: &schema.GitLabAuthorization{
							AuthnProvider: schema.AuthnProvider{
								ConfigID:       "onelogin-config-id",
								GitlabProvider: "onelogin",
								Type:           "openidconnect",
							},
						},
						Url:   "https://gitlab-0.mine",
						Token: "asdf",
					},
				},
			}},
			expAuthzAllowAccessByDefault: false,
			expAuthzProviders: []authz.Provider{
				newGitLabAuthzProviderParams{
					Op: gitlab.GitLabAuthzProviderOp{
						BaseURL:        mustURLParse(t, "https://gitlab-0.mine"),
						AuthnConfigID:  auth.ProviderConfigID{Type: "openidconnect", ID: "onelogin-config-id"},
						SudoToken:      "asdf",
						GitLabProvider: "onelogin",
						CacheTTL:       3 * time.Hour,
					},
				},
			},
			expSeriousProblems: []string{"Could not find item in `auth.providers` with config ID \"onelogin-config-id\" and type \"openidconnect\""},
		},
		{
			description: "1 GitLab referencing no auth provider",
			cfg: conf.Unified{SiteConfiguration: schema.SiteConfiguration{
				Gitlab: []*schema.GitLabConnection{
					{
						Authorization: &schema.GitLabAuthorization{},
						Url:           "https://gitlab-0.mine",
						Token:         "asdf",
					},
				},
			}},
			expAuthzAllowAccessByDefault: false,
			expAuthzProviders: []authz.Provider{
				newGitLabAuthzProviderParams{
					Op: gitlab.GitLabAuthzProviderOp{
						BaseURL:           mustURLParse(t, "https://gitlab-0.mine"),
						AuthnConfigID:     auth.ProviderConfigID{},
						SudoToken:         "asdf",
						CacheTTL:          3 * time.Hour,
						UseNativeUsername: false,
					},
				},
			},
			expSeriousProblems: []string{"`authz.authnProvider.configID` was empty. No users will be granted access to these repositories."},
		},
		{
			description: "1 GitLab with permissions disabled",
			cfg: conf.Unified{SiteConfiguration: schema.SiteConfiguration{
				Gitlab: []*schema.GitLabConnection{
					{
						Url:   "https://gitlab-0.mine",
						Token: "asdf",
					},
				},
			}},
			expAuthzAllowAccessByDefault: true,
			expAuthzProviders:            nil,
			expSeriousProblems:           nil,
		},
		{
			description: "1 GitLab with incomplete auth provider descriptor, ttl error",
			cfg: conf.Unified{
				Critical: schema.CriticalConfiguration{
					AuthProviders: []schema.AuthProviders{
						schema.AuthProviders{
							Saml: &schema.SAMLAuthProvider{
								ConfigID: "okta-config-id",
								Type:     "saml",
							},
						},
					},
				},
				SiteConfiguration: schema.SiteConfiguration{
					Gitlab: []*schema.GitLabConnection{
						{
							Authorization: &schema.GitLabAuthorization{
								AuthnProvider: schema.AuthnProvider{
									ConfigID:       "okta-config-id",
									GitlabProvider: "okta",
								},
								Ttl: "invalid",
							},
							Url:   "https://gitlab-0.mine",
							Token: "asdf",
						},
					},
				},
			},
			expAuthzAllowAccessByDefault: false,
			expAuthzProviders: []authz.Provider{
				newGitLabAuthzProviderParams{
					Op: gitlab.GitLabAuthzProviderOp{
						BaseURL:        mustURLParse(t, "https://gitlab-0.mine"),
						AuthnConfigID:  auth.ProviderConfigID{ID: "okta-config-id"},
						SudoToken:      "asdf",
						GitLabProvider: "okta",
						CacheTTL:       3 * time.Hour,
					},
				},
			},
			expSeriousProblems: []string{"`authz.authnProvider.type` was not specified, which means GitLab users cannot be resolved."},
			expWarnings:        []string{"Could not parse time duration \"invalid\", falling back to 3h0m0s."},
		},
	}

	for _, test := range tests {
		t.Logf("Test %q", test.description)
		allowAccessByDefault, authzProviders, seriousProblems, warnings := providersFromConfig(&test.cfg)
		if allowAccessByDefault != test.expAuthzAllowAccessByDefault {
			t.Errorf("allowAccessByDefault: (actual) %v != (expected) %v", asJSON(t, allowAccessByDefault), asJSON(t, test.expAuthzAllowAccessByDefault))
		}
		if !reflect.DeepEqual(authzProviders, test.expAuthzProviders) {
			t.Errorf("authzProviders: (actual) %+v != (expected) %+v", asJSON(t, authzProviders), asJSON(t, test.expAuthzProviders))
		}
		if !reflect.DeepEqual(seriousProblems, test.expSeriousProblems) {
			t.Errorf("seriousProblems: (actual) %+v != (expected) %+v", asJSON(t, seriousProblems), asJSON(t, test.expSeriousProblems))
		}
		if !reflect.DeepEqual(warnings, test.expWarnings) {
			t.Errorf("warnings: (actual) %+v != (expected) %+v", asJSON(t, warnings), asJSON(t, test.expWarnings))
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
