package authz

import (
	"context"
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/gitlab"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

type newGitLabAuthzProviderParams struct {
	Op gitlab.GitLabOAuthAuthzProviderOp
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
	NewGitLabProvider = func(op gitlab.GitLabOAuthAuthzProviderOp) authz.Provider {
		op.MockCache = nil // ignore cache value
		return newGitLabAuthzProviderParams{op}
	}

	db.Mocks = db.MockStores{}
	defer func() { db.Mocks = db.MockStores{} }()

	tests := []struct {
		description                  string
		cfg                          conf.Unified
		gitlabConnections            []*schema.GitLabConnection
		expAuthzAllowAccessByDefault bool
		expAuthzProviders            []authz.Provider
		expSeriousProblems           []string
	}{
		{
			description: "1 GitLab connection with authz enabled, 1 GitLab matching auth provider",
			cfg: conf.Unified{
				Critical: schema.CriticalConfiguration{
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
						Ttl: "48h",
					},
					Url:   "https://gitlab.mine",
					Token: "asdf",
				},
			},
			expAuthzAllowAccessByDefault: true,
			expAuthzProviders: []authz.Provider{
				newGitLabAuthzProviderParams{
					Op: gitlab.GitLabOAuthAuthzProviderOp{
						BaseURL:  mustURLParse(t, "https://gitlab.mine"),
						CacheTTL: 48 * time.Hour,
					},
				},
			},
		},
		{
			description: "1 GitLab connection with authz enabled, 1 GitLab auth provider but doesn't match",
			cfg: conf.Unified{
				Critical: schema.CriticalConfiguration{
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
						Ttl: "48h",
					},
					Url:   "https://gitlab.mine",
					Token: "asdf",
				},
			},
			expAuthzAllowAccessByDefault: false,
			expSeriousProblems:           []string{"Did not find authentication provider matching \"https://gitlab.mine\""},
		},
		{
			description: "1 GitLab connection with authz enabled, no GitLab auth provider",
			cfg: conf.Unified{
				Critical: schema.CriticalConfiguration{
					AuthProviders: []schema.AuthProviders{{
						Builtin: &schema.BuiltinAuthProvider{Type: "builtin"},
					}},
				},
			},
			gitlabConnections: []*schema.GitLabConnection{
				{
					Authorization: &schema.GitLabAuthorization{
						Ttl: "48h",
					},
					Url:   "https://gitlab.mine",
					Token: "asdf",
				},
			},
			expAuthzAllowAccessByDefault: false,
			expSeriousProblems:           []string{"Did not find authentication provider matching \"https://gitlab.mine\""},
		},
		{
			description: "Two GitLab connections with authz enabled, two matching GitLab auth providers",
			cfg: conf.Unified{
				Critical: schema.CriticalConfiguration{
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
					Authorization: &schema.GitLabAuthorization{},
					Url:           "https://gitlab.mine",
					Token:         "asdf",
				},
				{
					Authorization: &schema.GitLabAuthorization{},
					Url:           "https://gitlab.com",
					Token:         "asdf",
				},
			},
			expAuthzAllowAccessByDefault: true,
			expAuthzProviders: []authz.Provider{
				newGitLabAuthzProviderParams{
					Op: gitlab.GitLabOAuthAuthzProviderOp{
						BaseURL:  mustURLParse(t, "https://gitlab.mine"),
						CacheTTL: 3 * time.Hour,
					},
				},
				newGitLabAuthzProviderParams{
					Op: gitlab.GitLabOAuthAuthzProviderOp{
						BaseURL:  mustURLParse(t, "https://gitlab.com"),
						CacheTTL: 3 * time.Hour,
					},
				},
			},
		},
		{
			description: "1 GitLab connection with authz disabled",
			cfg: conf.Unified{
				Critical: schema.CriticalConfiguration{
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
				Critical: schema.CriticalConfiguration{
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
					Authorization: &schema.GitLabAuthorization{Ttl: "invalid"},
					Url:           "https://gitlab.mine",
					Token:         "asdf",
				},
			},
			expAuthzAllowAccessByDefault: false,
			expSeriousProblems:           []string{"Could not parse time duration \"invalid\"."},
		},
	}

	for _, test := range tests {
		t.Logf("Test %q", test.description)

		gitlabs := test.gitlabConnections
		db.Mocks.ExternalServices.List = func(opt db.ExternalServicesListOptions) ([]*types.ExternalService, error) {

			if reflect.DeepEqual(opt.Kinds, []string{"GITLAB"}) {
				externalServices := make([]*types.ExternalService, 0, len(gitlabs))
				for _, gl := range gitlabs {
					config, err := json.Marshal(gl)
					if err != nil {
						return nil, err
					}
					externalServices = append(externalServices, &types.ExternalService{
						ID:          2,
						Kind:        "GITLAB",
						DisplayName: "Test GitLab",
						Config:      string(config),
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					})
				}
				return externalServices, nil
			}
			return nil, nil
		}

		allowAccessByDefault, authzProviders, seriousProblems, _ := providersFromConfig(context.Background(), &test.cfg)
		if allowAccessByDefault != test.expAuthzAllowAccessByDefault {
			t.Errorf("allowAccessByDefault: (actual) %v != (expected) %v", asJSON(t, allowAccessByDefault), asJSON(t, test.expAuthzAllowAccessByDefault))
		}
		if !reflect.DeepEqual(authzProviders, test.expAuthzProviders) {
			t.Errorf("authzProviders: (actual) %+v != (expected) %+v", asJSON(t, authzProviders), asJSON(t, test.expAuthzProviders))
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
