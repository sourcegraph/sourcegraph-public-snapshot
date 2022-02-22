package github

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestNewAuthzProviders(t *testing.T) {
	t.Run("no authorization", func(t *testing.T) {
		providers, problems, warnings := NewAuthzProviders(
			[]*types.GitHubConnection{{
				GitHubConnection: &schema.GitHubConnection{
					Url:           "https://github.com",
					Authorization: nil,
				},
			}},
			[]schema.AuthProviders{},
			false,
		)
		if len(providers) != 0 {
			t.Fatalf("unexpected providers: %+v", providers)
		}
		if len(problems) != 0 {
			t.Fatalf("unexpected problems: %+v", problems)
		}
		if len(warnings) != 0 {
			t.Fatalf("unexpected warnings: %+v", warnings)
		}
	})

	t.Run("no matching auth provider", func(t *testing.T) {
		providers, problems, warnings := NewAuthzProviders(
			[]*types.GitHubConnection{{
				GitHubConnection: &schema.GitHubConnection{
					Url:           "https://github.com/my-org", // incorrect
					Authorization: &schema.GitHubAuthorization{},
				},
			}},
			[]schema.AuthProviders{{
				Github: &schema.GitHubAuthProvider{
					Url: "https://github.com",
				},
			}},
			false,
		)
		if len(providers) != 1 || providers[0] == nil {
			t.Fatal("expected a provider")
		}
		if len(problems) != 0 {
			t.Fatalf("unexpected problems: %+v", problems)
		}
		if len(warnings) != 1 || !strings.Contains(warnings[0], "no authentication provider") {
			t.Fatalf("unexpected warnings: %+v", warnings)
		}
	})

	t.Run("matching auth provider found", func(t *testing.T) {
		t.Run("default case", func(t *testing.T) {
			providers, problems, warnings := NewAuthzProviders(
				[]*types.GitHubConnection{{
					GitHubConnection: &schema.GitHubConnection{
						Url:           schema.DefaultGitHubURL,
						Authorization: &schema.GitHubAuthorization{},
					},
				}},
				[]schema.AuthProviders{{
					// falls back to schema.DefaultGitHubURL
					Github: &schema.GitHubAuthProvider{},
				}},
				false,
			)
			if len(providers) != 1 || providers[0] == nil {
				t.Fatal("expected a provider")
			}
			if len(problems) != 0 {
				t.Fatalf("unexpected problems: %+v", problems)
			}
			if len(warnings) != 0 {
				t.Fatalf("unexpected warnings: %+v", warnings)
			}
		})

		t.Run("groups cache enabled, but not allowGroupsPermissionsSync", func(t *testing.T) {
			providers, problems, warnings := NewAuthzProviders(
				[]*types.GitHubConnection{{
					GitHubConnection: &schema.GitHubConnection{
						Url: "https://github.com/",
						Authorization: &schema.GitHubAuthorization{
							GroupsCacheTTL: 72,
						},
					},
				}},
				[]schema.AuthProviders{{
					Github: &schema.GitHubAuthProvider{
						Url:                        "https://github.com",
						AllowGroupsPermissionsSync: false,
					},
				}},
				false,
			)
			if len(providers) != 1 || providers[0] == nil {
				t.Fatal("expected a provider")
			}
			if len(problems) != 0 {
				t.Fatalf("unexpected problems: %+v", problems)
			}
			if len(warnings) != 1 || !strings.Contains(warnings[0], "`allowGroupsPermissionsSync`") {
				t.Fatalf("unexpected warnings: %+v", warnings)
			}
			// assert groups cache is forcibly disabled
			if (providers[0]).(*Provider).groupsCache != nil {
				t.Fatal("expected groups cache to be disabled")
			}
		})

		t.Run("groups cache and allowGroupsPermissionsSync enabled", func(t *testing.T) {
			github.MockGetAuthenticatedOAuthScopes = func(context.Context) ([]string, error) {
				return []string{"read:org"}, nil
			}
			providers, problems, warnings := NewAuthzProviders(
				[]*types.GitHubConnection{{
					GitHubConnection: &schema.GitHubConnection{
						Url: "https://github.com/",
						Authorization: &schema.GitHubAuthorization{
							GroupsCacheTTL: 72,
						},
					},
				}},
				[]schema.AuthProviders{{
					Github: &schema.GitHubAuthProvider{
						Url:                        "https://github.com",
						AllowGroupsPermissionsSync: true,
					},
				}},
				false,
			)
			if len(providers) != 1 || providers[0] == nil {
				t.Fatal("expected a provider")
			}
			if len(problems) != 0 {
				t.Fatalf("unexpected problems: %+v", problems)
			}
			if len(warnings) != 0 {
				t.Fatalf("unexpected warnings: %+v", warnings)
			}
			// assert groups cache is available
			if (providers[0]).(*Provider).groupsCache == nil {
				t.Fatal("expected groups cache to be enabled")
			}
		})

		t.Run("github app installation id available", func(t *testing.T) {
			const bogusKey = `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlCUEFJQkFBSkJBUEpIaWprdG1UMUlLYUd0YTVFZXAzQVo5Q2VPZUw4alBESUZUN3dRZ0tabXQzRUZxRGhCCk93bitRVUhKdUs5Zm92UkROSmVWTDJvWTVCT0l6NHJ3L0cwQ0F3RUFBUUpCQU1BK0o5Mks0d2NQVllsbWMrM28KcHU5NmlKTkNwMmp5Nm5hK1pEQlQzK0VvSUo1VFJGdnN3R2kvTHUzZThYUWwxTDNTM21ub0xPSlZNcTF0bUxOMgpIY0VDSVFEK3daeS83RlYxUEFtdmlXeWlYVklETzJnNWJOaUJlbmdKQ3hFa3Nia1VtUUloQVBOMlZaczN6UFFwCk1EVG9vTlJXcnl0RW1URERkamdiOFpzTldYL1JPRGIxQWlCZWNKblNVQ05TQllLMXJ5VTFmNURTbitoQU9ZaDkKWDFBMlVnTDE3bWhsS1FJaEFPK2JMNmRDWktpTGZORWxmVnRkTUtxQnFjNlBIK01heFU2VzlkVlFvR1dkQWlFQQptdGZ5cE9zYTFiS2hFTDg0blovaXZFYkJyaVJHalAya3lERHYzUlg0V0JrPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=`

			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					Dotcom: &schema.Dotcom{
						GithubAppCloud: &schema.GithubAppCloud{
							AppID:        "1234",
							ClientID:     "1234",
							ClientSecret: "1234",
							Slug:         "test-app",
							PrivateKey:   bogusKey,
						},
					},
				},
			})
			defer conf.Mock(nil)

			providers, problems, warnings := NewAuthzProviders(
				[]*types.GitHubConnection{{
					GitHubConnection: &schema.GitHubConnection{
						Url: "https://github.com/",
						Authorization: &schema.GitHubAuthorization{
							GroupsCacheTTL: 72,
						},
						GithubAppInstallationID: "1234",
					},
				}},
				[]schema.AuthProviders{{
					// falls back to schema.DefaultGitHubURL
					Github: &schema.GitHubAuthProvider{},
				}},
				false,
			)
			if len(providers) != 1 || providers[0] == nil {
				t.Fatal("expected a provider")
			}
			if len(problems) != 0 {
				t.Fatalf("unexpected problems: %+v", problems)
			}
			if len(warnings) != 0 {
				t.Fatalf("unexpected warnings: %+v", warnings)
			}
		})
	})
}
