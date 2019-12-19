package githuboauth

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/oauth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	githubcodehost "github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/oauth2"
)

func Test_parseConfig(t *testing.T) {
	spew.Config.DisablePointerAddresses = true
	spew.Config.SortKeys = true
	spew.Config.SpewKeys = true

	type args struct {
		cfg *conf.Unified
	}
	tests := []struct {
		name          string
		args          args
		wantProviders map[schema.GitHubAuthProvider]providers.Provider
		wantProblems  []string
	}{
		{
			name:          "No configs",
			args:          args{cfg: &conf.Unified{}},
			wantProviders: map[schema.GitHubAuthProvider]providers.Provider{},
		},
		{
			name: "1 GitHub.com config",
			args: args{cfg: &conf.Unified{SiteConfiguration: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{{
					Github: &schema.GitHubAuthProvider{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						DisplayName:  "GitHub",
						Type:         "github",
						Url:          "https://github.com",
					},
				}},
			}}},
			wantProviders: map[schema.GitHubAuthProvider]providers.Provider{
				{
					ClientID:     "my-client-id",
					ClientSecret: "my-client-secret",
					DisplayName:  "GitHub",
					Type:         "github",
					Url:          "https://github.com",
				}: provider("https://github.com/", oauth2.Config{
					ClientID:     "my-client-id",
					ClientSecret: "my-client-secret",
					Endpoint: oauth2.Endpoint{
						AuthURL:  "https://github.com/login/oauth/authorize",
						TokenURL: "https://github.com/login/oauth/access_token",
					},
					Scopes: []string{"repo", "user:email"},
				}),
			},
		},
		{
			name: "2 GitHub configs",
			args: args{cfg: &conf.Unified{SiteConfiguration: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{{
					Github: &schema.GitHubAuthProvider{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						DisplayName:  "GitHub",
						Type:         "github",
						Url:          "https://github.com",
					},
				}, {
					Github: &schema.GitHubAuthProvider{
						ClientID:     "my-client-id-2",
						ClientSecret: "my-client-secret-2",
						DisplayName:  "GitHub Enterprise",
						Type:         "github",
						Url:          "https://mycompany.com",
					},
				}},
			}}},
			wantProviders: map[schema.GitHubAuthProvider]providers.Provider{
				{
					ClientID:     "my-client-id",
					ClientSecret: "my-client-secret",
					DisplayName:  "GitHub",
					Type:         "github",
					Url:          "https://github.com",
				}: provider("https://github.com/", oauth2.Config{
					ClientID:     "my-client-id",
					ClientSecret: "my-client-secret",
					Endpoint: oauth2.Endpoint{
						AuthURL:  "https://github.com/login/oauth/authorize",
						TokenURL: "https://github.com/login/oauth/access_token",
					},
					Scopes: []string{"repo", "user:email"},
				}),
				{
					ClientID:     "my-client-id-2",
					ClientSecret: "my-client-secret-2",
					DisplayName:  "GitHub Enterprise",
					Type:         "github",
					Url:          "https://mycompany.com",
				}: provider("https://mycompany.com/", oauth2.Config{
					ClientID:     "my-client-id-2",
					ClientSecret: "my-client-secret-2",
					Endpoint: oauth2.Endpoint{
						AuthURL:  "https://mycompany.com/login/oauth/authorize",
						TokenURL: "https://mycompany.com/login/oauth/access_token",
					},
					Scopes: []string{"repo", "user:email"},
				}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProviders, gotProblems := parseConfig(tt.args.cfg)
			for _, p := range gotProviders {
				if p, ok := p.(*oauth.Provider); ok {
					p.Login, p.Callback = nil, nil
					p.ProviderOp.Login, p.ProviderOp.Callback = nil, nil
				}
			}
			for k, p := range tt.wantProviders {
				k := k
				if q, ok := p.(*oauth.Provider); ok {
					q.SourceConfig = schema.AuthProviders{Github: &k}
				}
			}
			if !reflect.DeepEqual(gotProviders, tt.wantProviders) {
				dmp := diffmatchpatch.New()

				t.Errorf("parseConfig() gotProviders != tt.wantProviders, diff:\n%s",
					dmp.DiffPrettyText(dmp.DiffMain(spew.Sdump(tt.wantProviders), spew.Sdump(gotProviders), false)),
				)
			}
			if !reflect.DeepEqual(gotProblems.Messages(), tt.wantProblems) {
				t.Errorf("parseConfig() gotProblems = %v, want %v", gotProblems, tt.wantProblems)
			}
		})
	}
}

func provider(serviceID string, oauth2Config oauth2.Config) *oauth.Provider {
	op := oauth.ProviderOp{
		AuthPrefix:   authPrefix,
		OAuth2Config: oauth2Config,
		StateConfig:  getStateConfig(),
		ServiceID:    serviceID,
		ServiceType:  githubcodehost.ServiceType,
	}
	return &oauth.Provider{ProviderOp: op}
}
