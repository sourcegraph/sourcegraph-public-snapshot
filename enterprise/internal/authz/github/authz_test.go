package github

import (
	"strings"
	"testing"

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
					Url:           "https://github.com/my-org",
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
		if len(warnings) != 1 || !strings.Contains(warnings[0], "not find authentication provider") {
			t.Fatalf("unexpected warnings: %+v", warnings)
		}
	})

	t.Run("auth enabled, matching auth provider", func(t *testing.T) {
		providers, problems, warnings := NewAuthzProviders(
			[]*types.GitHubConnection{{
				GitHubConnection: &schema.GitHubConnection{
					Url:           "https://github.com/",
					Authorization: &schema.GitHubAuthorization{},
				},
			}},
			[]schema.AuthProviders{{
				Github: &schema.GitHubAuthProvider{
					Url: "https://github.com",
				},
			}})
		t.Log(providers[0].ServiceID())
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
}
