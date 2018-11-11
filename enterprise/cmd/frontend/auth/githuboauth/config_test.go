package githuboauth

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/auth"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/oauth2"
)

func Test_parseConfig(t *testing.T) {
	spew.Config.DisablePointerAddresses = true
	spew.Config.SortKeys = true
	spew.Config.SpewKeys = true

	type args struct {
		cfg *schema.SiteConfiguration
	}
	tests := []struct {
		name          string
		args          args
		wantProviders map[schema.GitHubAuthProvider]auth.Provider
		wantProblems  []string
	}{
		{
			name:          "No configs",
			args:          args{cfg: &schema.SiteConfiguration{}},
			wantProviders: map[schema.GitHubAuthProvider]auth.Provider{},
		},
		{
			name: "1 GitHub.com config",
			args: args{cfg: &schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{{
					Github: &schema.GitHubAuthProvider{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						DisplayName:  "GitHub",
						Type:         "github",
						Url:          "https://github.com",
					},
				}},
			}},
			wantProviders: map[schema.GitHubAuthProvider]auth.Provider{
				schema.GitHubAuthProvider{
					ClientID:     "my-client-id",
					ClientSecret: "my-client-secret",
					DisplayName:  "GitHub",
					Type:         "github",
					Url:          "https://github.com",
				}: &provider{
					config: oauth2.Config{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						Endpoint: oauth2.Endpoint{
							AuthURL:  "https://github.com/login/oauth/authorize",
							TokenURL: "https://github.com/login/oauth/access_token",
						},
						Scopes: []string{"repo"},
					},
					serviceID: "https://github.com/",
				},
			},
		},
		{
			name: "2 GitHub configs",
			args: args{cfg: &schema.SiteConfiguration{
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
			}},
			wantProviders: map[schema.GitHubAuthProvider]auth.Provider{
				schema.GitHubAuthProvider{
					ClientID:     "my-client-id",
					ClientSecret: "my-client-secret",
					DisplayName:  "GitHub",
					Type:         "github",
					Url:          "https://github.com",
				}: &provider{
					config: oauth2.Config{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						Endpoint: oauth2.Endpoint{
							AuthURL:  "https://github.com/login/oauth/authorize",
							TokenURL: "https://github.com/login/oauth/access_token",
						},
						Scopes: []string{"repo"},
					},
					serviceID: "https://github.com/",
				},
				schema.GitHubAuthProvider{
					ClientID:     "my-client-id-2",
					ClientSecret: "my-client-secret-2",
					DisplayName:  "GitHub Enterprise",
					Type:         "github",
					Url:          "https://mycompany.com",
				}: &provider{
					config: oauth2.Config{
						ClientID:     "my-client-id-2",
						ClientSecret: "my-client-secret-2",
						Endpoint: oauth2.Endpoint{
							AuthURL:  "https://mycompany.com/login/oauth/authorize",
							TokenURL: "https://mycompany.com/login/oauth/access_token",
						},
						Scopes: []string{"repo"},
					},
					serviceID: "https://mycompany.com/",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProviders, gotProblems := parseConfig(tt.args.cfg)
			for _, p := range gotProviders {
				if p, ok := p.(*provider); ok {
					p.login, p.callback = nil, nil
				}
			}
			for k, p := range tt.wantProviders {
				k := k
				if q, ok := p.(*provider); ok {
					q.sourceConfig = schema.AuthProviders{Github: &k}
				}
			}
			if !reflect.DeepEqual(gotProviders, tt.wantProviders) {
				dmp := diffmatchpatch.New()

				// t.Errorf("parseConfig() gotProviders = %s, want %s", , spew.Sdump(tt.wantProviders))
				t.Errorf("parseConfig() gotProviders != tt.wantProviders, diff:\n%s",
					dmp.DiffPrettyText(dmp.DiffMain(spew.Sdump(gotProviders), spew.Sdump(tt.wantProviders), false)),
				)
			}
			if !reflect.DeepEqual(gotProblems, tt.wantProblems) {
				t.Errorf("parseConfig() gotProblems = %v, want %v", gotProblems, tt.wantProblems)
			}
		})
	}
}
