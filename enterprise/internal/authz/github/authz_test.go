package github

import (
	"context"
	"strings"
	"testing"

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
			[]schema.AuthProviders{})
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
			}})
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
				}})
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
				}})
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
				}})
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
	})
}
