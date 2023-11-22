package azureoauth

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/log/logtest"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/oauth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

func newOauthProvider(oauth2Config oauth2.Config) *oauth.Provider {
	return &oauth.Provider{
		ProviderOp: oauth.ProviderOp{
			AuthPrefix:   "/.auth/azuredevops",
			OAuth2Config: func() oauth2.Config { return oauth2Config },
			StateConfig:  oauth.GetStateConfig(stateCookie),
			ServiceID:    "https://dev.azure.com/",
			ServiceType:  extsvc.TypeAzureDevOps,
		},
	}
}

func newUnifiedConfig(s schema.SiteConfiguration) conf.Unified {
	return conf.Unified{SiteConfiguration: s}
}

func TestParseConfig(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	testCases := []struct {
		name          string
		config        conf.Unified
		wantProviders []Provider
		wantProblems  []string
	}{
		{
			name:          "empty config",
			config:        conf.Unified{},
			wantProviders: []Provider(nil),
		},
		{
			name: "Azure Dev Ops config with default scopes",
			config: newUnifiedConfig(schema.SiteConfiguration{
				ExternalURL: "https://example.com",
				AuthProviders: []schema.AuthProviders{{
					AzureDevOps: &schema.AzureDevOpsAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplayName:  "Azure Dev Ops",
						Type:         extsvc.TypeAzureDevOps,
					},
				}},
			}),
			wantProviders: []Provider{
				{
					AzureDevOpsAuthProvider: &schema.AzureDevOpsAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplayName:  "Azure Dev Ops",
						Type:         extsvc.TypeAzureDevOps,
						ApiScope:     "vso.code,vso.identity,vso.project",
					},
					Provider: newOauthProvider(oauth2.Config{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						Endpoint: oauth2.Endpoint{
							AuthURL:   "https://app.vssps.visualstudio.com/oauth2/authorize",
							TokenURL:  "https://app.vssps.visualstudio.com/oauth2/token",
							AuthStyle: oauth2.AuthStyleInParams,
						},
						Scopes:      []string{"vso.code", "vso.identity", "vso.project"},
						RedirectURL: "https://example.com/.auth/azuredevops/callback",
					}),
				},
			},
		},
		{
			name: "Azure Dev Ops config with custom scopes",
			config: newUnifiedConfig(schema.SiteConfiguration{
				ExternalURL: "https://example.com",
				AuthProviders: []schema.AuthProviders{{
					AzureDevOps: &schema.AzureDevOpsAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplayName:  "Azure Dev Ops",
						Type:         extsvc.TypeAzureDevOps,
						ApiScope:     "vso.code",
					},
				}},
			}),
			wantProviders: []Provider{
				{
					AzureDevOpsAuthProvider: &schema.AzureDevOpsAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplayName:  "Azure Dev Ops",
						Type:         extsvc.TypeAzureDevOps,
						ApiScope:     "vso.code",
					},
					Provider: newOauthProvider(oauth2.Config{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						Endpoint: oauth2.Endpoint{
							AuthURL:   "https://app.vssps.visualstudio.com/oauth2/authorize",
							TokenURL:  "https://app.vssps.visualstudio.com/oauth2/token",
							AuthStyle: oauth2.AuthStyleInParams,
						},
						Scopes:      []string{"vso.code"},
						RedirectURL: "https://example.com/.auth/azuredevops/callback",
					}),
				},
			},
		},
		{
			name: "Azure Dev Ops config with duplicate client ID config",
			config: newUnifiedConfig(schema.SiteConfiguration{
				ExternalURL: "https://example.com",
				AuthProviders: []schema.AuthProviders{
					{
						AzureDevOps: &schema.AzureDevOpsAuthProvider{
							ClientID:     "myclientid",
							ClientSecret: "myclientsecret",
							DisplayName:  "Azure Dev Ops",
							Type:         extsvc.TypeAzureDevOps,
						},
					},
					{
						AzureDevOps: &schema.AzureDevOpsAuthProvider{
							ClientID:     "myclientid",
							ClientSecret: "myclientsecret",
							DisplayName:  "Azure Dev Ops The Second",
							Type:         extsvc.TypeAzureDevOps,
						},
					},
				},
			}),
			wantProviders: []Provider{
				{
					AzureDevOpsAuthProvider: &schema.AzureDevOpsAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplayName:  "Azure Dev Ops",
						Type:         extsvc.TypeAzureDevOps,
						ApiScope:     "vso.code,vso.identity,vso.project",
					},
					Provider: newOauthProvider(oauth2.Config{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						Endpoint: oauth2.Endpoint{
							AuthURL:   "https://app.vssps.visualstudio.com/oauth2/authorize",
							TokenURL:  "https://app.vssps.visualstudio.com/oauth2/token",
							AuthStyle: oauth2.AuthStyleInParams,
						},
						Scopes:      []string{"vso.code", "vso.identity", "vso.project"},
						RedirectURL: "https://example.com/.auth/azuredevops/callback",
					}),
				},
			},
			wantProblems: []string{
				"Cannot have more than one auth provider for Azure Dev Ops with Client ID \"myclientid\", only the first one will be used",
			},
		},
		{
			name: "Azure Dev Ops config with separate client ID config",
			config: newUnifiedConfig(schema.SiteConfiguration{
				ExternalURL: "https://example.com",
				AuthProviders: []schema.AuthProviders{
					{
						AzureDevOps: &schema.AzureDevOpsAuthProvider{
							ClientID:     "myclientid",
							ClientSecret: "myclientsecret",
							DisplayName:  "Azure Dev Ops",
							Type:         extsvc.TypeAzureDevOps,
						},
					},
					{
						AzureDevOps: &schema.AzureDevOpsAuthProvider{
							ClientID:     "myclientid-second",
							ClientSecret: "myclientsecret",
							DisplayName:  "Azure Dev Ops The Second",
							Type:         extsvc.TypeAzureDevOps,
						},
					},
				},
			}),
			wantProviders: []Provider{
				{
					AzureDevOpsAuthProvider: &schema.AzureDevOpsAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplayName:  "Azure Dev Ops",
						Type:         extsvc.TypeAzureDevOps,
						ApiScope:     "vso.code,vso.identity,vso.project",
					},
					Provider: newOauthProvider(oauth2.Config{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						Endpoint: oauth2.Endpoint{
							AuthURL:   "https://app.vssps.visualstudio.com/oauth2/authorize",
							TokenURL:  "https://app.vssps.visualstudio.com/oauth2/token",
							AuthStyle: oauth2.AuthStyleInParams,
						},
						Scopes:      []string{"vso.code", "vso.identity", "vso.project"},
						RedirectURL: "https://example.com/.auth/azuredevops/callback",
					}),
				},
				{
					AzureDevOpsAuthProvider: &schema.AzureDevOpsAuthProvider{
						ClientID:     "myclientid-second",
						ClientSecret: "myclientsecret",
						DisplayName:  "Azure Dev Ops The Second",
						Type:         extsvc.TypeAzureDevOps,
						ApiScope:     "vso.code,vso.identity,vso.project",
					},
					Provider: newOauthProvider(oauth2.Config{
						ClientID:     "myclientid-second",
						ClientSecret: "myclientsecret",
						Endpoint: oauth2.Endpoint{
							AuthURL:   "https://app.vssps.visualstudio.com/oauth2/authorize",
							TokenURL:  "https://app.vssps.visualstudio.com/oauth2/token",
							AuthStyle: oauth2.AuthStyleInParams,
						},
						Scopes:      []string{"vso.code", "vso.identity", "vso.project"},
						RedirectURL: "https://example.com/.auth/azuredevops/callback",
					}),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotProviders, gotProblems := parseConfig(logtest.Scoped(t), tc.config, db)
			gotConfigs := make([]oauth2.Config, len(gotProviders))

			for i, p := range gotProviders {
				if pr, ok := p.Provider.(*oauth.Provider); ok {
					pr.Login, pr.Callback = nil, nil
					gotConfigs[i] = pr.OAuth2Config()
					pr.OAuth2Config = nil
					pr.ProviderOp.Login, pr.ProviderOp.Callback = nil, nil
				}
			}

			wantConfigs := make([]oauth2.Config, len(tc.wantProviders))

			for i, p := range tc.wantProviders {
				if pr, ok := p.Provider.(*oauth.Provider); ok {
					pr.SourceConfig = schema.AuthProviders{AzureDevOps: p.AzureDevOpsAuthProvider}
					wantConfigs[i] = pr.OAuth2Config()
					pr.OAuth2Config = nil
				}
			}

			if diff := cmp.Diff(tc.wantProviders, gotProviders); diff != "" {
				t.Errorf("mismatched providers: (-want,+got)\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantProblems, gotProblems.Messages()); diff != "" {
				t.Errorf("mismatched problems (-want,+got):\n%s", diff)
			}
			if diff := cmp.Diff(wantConfigs, gotConfigs, cmpopts.IgnoreUnexported()); diff != "" {
				t.Errorf("mismatched configs (-want,+got):\n%s", diff)
			}
		})
	}
}
