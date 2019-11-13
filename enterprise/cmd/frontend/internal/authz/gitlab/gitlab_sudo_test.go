package gitlab

import (
	"context"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

func Test_GitLab_FetchAccount(t *testing.T) {
	// Test structures
	type call struct {
		description string

		user    *types.User
		current []*extsvc.ExternalAccount

		expMine *extsvc.ExternalAccount
	}
	type test struct {
		description string

		// authnProviders is the list of auth providers that are mocked
		authnProviders []providers.Provider

		// op configures the SudoProvider instance
		op SudoProviderOp

		calls []call
	}

	// Mocks
	gitlabMock := newMockGitLab(mockGitLabOp{
		t: t,
		users: []*gitlab.User{
			{
				ID:       101,
				Username: "b.l",
				Identities: []gitlab.Identity{
					{Provider: "okta.mine", ExternUID: "bl"},
					{Provider: "onelogin.mine", ExternUID: "bl"},
				},
			},
			{
				ID:         102,
				Username:   "k.l",
				Identities: []gitlab.Identity{{Provider: "okta.mine", ExternUID: "kl"}},
			},
			{
				ID:         199,
				Username:   "user-without-extern-id",
				Identities: nil,
			},
		},
	})
	gitlab.MockListUsers = gitlabMock.ListUsers

	// Test cases
	tests := []test{
		{
			description: "1 authn provider, basic authz provider",
			authnProviders: []providers.Provider{
				mockAuthnProvider{
					configID:  providers.ConfigID{ID: "okta.mine", Type: "saml"},
					serviceID: "https://okta.mine/",
				},
			},
			op: SudoProviderOp{
				BaseURL:           mustURL(t, "https://gitlab.mine"),
				AuthnConfigID:     providers.ConfigID{ID: "okta.mine", Type: "saml"},
				GitLabProvider:    "okta.mine",
				UseNativeUsername: false,
			},
			calls: []call{
				{
					description: "1 account, matches",
					user:        &types.User{ID: 123},
					current:     []*extsvc.ExternalAccount{acct(t, 1, "saml", "https://okta.mine/", "bl", "")},
					expMine:     acct(t, 123, gitlab.ServiceType, "https://gitlab.mine/", "101", ""),
				},
				{
					description: "many accounts, none match",
					user:        &types.User{ID: 123},
					current: []*extsvc.ExternalAccount{
						acct(t, 1, "saml", "https://okta.mine/", "nomatch", ""),
						acct(t, 1, "saml", "nomatch", "bl", ""),
						acct(t, 1, "nomatch", "https://okta.mine/", "bl", ""),
					},
					expMine: nil,
				},
				{
					description: "many accounts, 1 match",
					user:        &types.User{ID: 123},
					current: []*extsvc.ExternalAccount{
						acct(t, 1, "saml", "nomatch", "bl", ""),
						acct(t, 1, "nomatch", "https://okta.mine/", "bl", ""),
						acct(t, 1, "saml", "https://okta.mine/", "bl", ""),
					},
					expMine: acct(t, 123, gitlab.ServiceType, "https://gitlab.mine/", "101", ""),
				},
				{
					description: "no user",
					user:        nil,
					current:     nil,
					expMine:     nil,
				},
			},
		},
		{
			description:    "0 authn providers, native username",
			authnProviders: nil,
			op: SudoProviderOp{
				BaseURL:           mustURL(t, "https://gitlab.mine"),
				UseNativeUsername: true,
			},
			calls: []call{
				{
					description: "username match",
					user:        &types.User{ID: 123, Username: "b.l"},
					expMine:     acct(t, 123, gitlab.ServiceType, "https://gitlab.mine/", "101", ""),
				},
				{
					description: "no username match",
					user:        &types.User{ID: 123, Username: "nomatch"},
					expMine:     nil,
				},
			},
		},
		{
			description:    "0 authn providers, basic authz provider",
			authnProviders: nil,
			op: SudoProviderOp{
				BaseURL:           mustURL(t, "https://gitlab.mine"),
				AuthnConfigID:     providers.ConfigID{ID: "okta.mine", Type: "saml"},
				GitLabProvider:    "okta.mine",
				UseNativeUsername: false,
			},
			calls: []call{
				{
					description: "no matches",
					user:        &types.User{ID: 123, Username: "b.l"},
					expMine:     nil,
				},
			},
		},
		{
			description: "2 authn providers, basic authz provider",
			authnProviders: []providers.Provider{
				mockAuthnProvider{
					configID:  providers.ConfigID{ID: "okta.mine", Type: "saml"},
					serviceID: "https://okta.mine/",
				},
				mockAuthnProvider{
					configID:  providers.ConfigID{ID: "onelogin.mine", Type: "openidconnect"},
					serviceID: "https://onelogin.mine/",
				},
			},
			op: SudoProviderOp{
				BaseURL:           mustURL(t, "https://gitlab.mine"),
				AuthnConfigID:     providers.ConfigID{ID: "onelogin.mine", Type: "openidconnect"},
				GitLabProvider:    "onelogin.mine",
				UseNativeUsername: false,
			},
			calls: []call{
				{
					description: "1 authn provider matches",
					user:        &types.User{ID: 123},
					current:     []*extsvc.ExternalAccount{acct(t, 1, "openidconnect", "https://onelogin.mine/", "bl", "")},
					expMine:     acct(t, 123, gitlab.ServiceType, "https://gitlab.mine/", "101", ""),
				},
				{
					description: "0 authn providers match",
					user:        &types.User{ID: 123},
					current:     []*extsvc.ExternalAccount{acct(t, 1, "openidconnect", "https://onelogin.mine/", "nomatch", "")},
					expMine:     nil,
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			providers.MockProviders = test.authnProviders
			defer func() { providers.MockProviders = nil }()

			ctx := context.Background()
			authzProvider := newSudoProvider(test.op)
			for _, c := range test.calls {
				t.Run(c.description, func(t *testing.T) {
					acct, err := authzProvider.FetchAccount(ctx, c.user, c.current)
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					// ignore AccountData field in comparison
					if acct != nil {
						acct.AccountData, c.expMine.AccountData = nil, nil
					}

					if !reflect.DeepEqual(acct, c.expMine) {
						dmp := diffmatchpatch.New()
						t.Errorf("wantUser != user\n%s",
							dmp.DiffPrettyText(dmp.DiffMain(spew.Sdump(c.expMine), spew.Sdump(acct), false)))
					}
				})
			}
		})
	}
}

func Test_SudoProvider_RepoPerms(t *testing.T) {
	type call struct {
		description string
		account     *extsvc.ExternalAccount
		repos       []*types.Repo
		expPerms    []authz.RepoPerms
	}
	type test struct {
		description string
		op          SudoProviderOp
		calls       []call
	}

	// Mock the following scenario:
	// - public projects begin with 99
	// - internal projects begin with 98
	// - private projects begin with the digit of the user that owns them (other users may have access)
	// - user 1 owns its own repositories and nothing else
	// - user 2 owns its own repos and has guest access to user 1's
	// - user 3 owns its own repos and has full access to user 1's and guest access to user 2's
	gitlabMock := newMockGitLab(mockGitLabOp{
		t: t,
		publicProjs: []int{ // public projects
			991,
		},
		internalProjs: []int{ // internal projects
			981,
		},
		privateProjs: map[int][2][]int32{ // private projects
			10: {
				{ // guests
					2,
				},
				{ // content ("full access")
					1,
					3,
				},
			},
			20: {
				{
					3,
				},
				{
					2,
				},
			},
			30: {
				{},
				{3},
			},
		},
		sudoTok: "sudo-token",
	})
	gitlab.MockGetProject = gitlabMock.GetProject
	gitlab.MockListTree = gitlabMock.ListTree

	repos := map[string]*types.Repo{
		"u1/repo1":       repo("u1/repo1", gitlab.ServiceType, "https://gitlab.mine/", "10"),
		"u2/repo1":       repo("u2/repo1", gitlab.ServiceType, "https://gitlab.mine/", "20"),
		"u3/repo1":       repo("u3/repo1", gitlab.ServiceType, "https://gitlab.mine/", "30"),
		"internal/repo1": repo("internal/repo1", gitlab.ServiceType, "https://gitlab.mine/", "981"),
		"public/repo1":   repo("public/repo1", gitlab.ServiceType, "https://gitlab.mine/", "991"),
	}

	tests := []test{
		{
			description: "standard config",
			op: SudoProviderOp{
				BaseURL:   mustURL(t, "https://gitlab.mine"),
				SudoToken: "sudo-token",
			},
			calls: []call{
				{
					description: "u1 user has expected perms",
					account:     acct(t, 1, "gitlab", "https://gitlab.mine/", "1", "oauth-u1"),
					repos: []*types.Repo{
						repos["u1/repo1"],
						repos["u2/repo1"],
						repos["u3/repo1"],
						repos["internal/repo1"],
						repos["public/repo1"],
					},
					expPerms: []authz.RepoPerms{
						{Repo: repos["u1/repo1"], Perms: authz.Read},
						{Repo: repos["internal/repo1"], Perms: authz.Read},
						{Repo: repos["public/repo1"], Perms: authz.Read},
					},
				},
				{
					description: "u2 user has expected perms",
					account:     acct(t, 2, "gitlab", "https://gitlab.mine/", "2", "oauth-u2"),
					repos: []*types.Repo{
						repos["u1/repo1"],
						repos["u2/repo1"],
						repos["u3/repo1"],
						repos["internal/repo1"],
						repos["public/repo1"],
					},
					expPerms: []authz.RepoPerms{
						{Repo: repos["u2/repo1"], Perms: authz.Read},
						{Repo: repos["internal/repo1"], Perms: authz.Read},
						{Repo: repos["public/repo1"], Perms: authz.Read},
					},
				},
				{
					description: "other user has expected perms (internal and public)",
					account:     acct(t, 4, "gitlab", "https://gitlab.mine/", "555", "oauth-other"),
					repos: []*types.Repo{
						repos["u1/repo1"],
						repos["u2/repo1"],
						repos["u3/repo1"],
						repos["internal/repo1"],
						repos["public/repo1"],
					},
					expPerms: []authz.RepoPerms{
						{Repo: repos["internal/repo1"], Perms: authz.Read},
						{Repo: repos["public/repo1"], Perms: authz.Read},
					},
				},
				{
					description: "no token means only public and internal repos",
					account:     acct(t, 4, "gitlab", "https://gitlab.mine/", "555", ""),
					repos: []*types.Repo{
						repos["u1/repo1"],
						repos["u2/repo1"],
						repos["u3/repo1"],
						repos["internal/repo1"],
						repos["public/repo1"],
					},
					expPerms: []authz.RepoPerms{
						{Repo: repos["internal/repo1"], Perms: authz.Read},
						{Repo: repos["public/repo1"], Perms: authz.Read},
					},
				},
				{
					description: "unauthenticated means only public repos",
					account:     nil,
					repos: []*types.Repo{
						repos["u1/repo1"],
						repos["u2/repo1"],
						repos["u3/repo1"],
						repos["internal/repo1"],
						repos["public/repo1"],
					},
					expPerms: []authz.RepoPerms{
						{Repo: repos["public/repo1"], Perms: authz.Read},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			for _, c := range test.calls {
				t.Logf("Call %q", c.description)

				// Recreate the authz provider cache every time, before running twice (once uncached, once cached)
				ctx := context.Background()
				op := test.op
				op.MockCache = make(mockCache)
				authzProvider := newSudoProvider(op)

				for i := 0; i < 2; i++ {
					t.Logf("iter %d", i)
					perms, err := authzProvider.RepoPerms(ctx, c.account, c.repos)
					if err != nil {
						t.Errorf("unexpected error: %v", err)
						continue
					}
					if !reflect.DeepEqual(perms, c.expPerms) {
						t.Errorf("expected %s, but got %s", asJSON(t, c.expPerms), asJSON(t, perms))
					}
				}
			}
		})
	}
}
