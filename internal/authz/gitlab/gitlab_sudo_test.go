package gitlab

import (
	"context"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

func Test_GitLab_FetchAccount(t *testing.T) {
	// Test structures
	type call struct {
		description string

		user    *types.User
		current []*extsvc.Account

		expMine *extsvc.Account
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
					current:     []*extsvc.Account{acct(t, 1, "saml", "https://okta.mine/", "bl", "")},
					expMine:     acct(t, 123, extsvc.TypeGitLab, "https://gitlab.mine/", "101", ""),
				},
				{
					description: "many accounts, none match",
					user:        &types.User{ID: 123},
					current: []*extsvc.Account{
						acct(t, 1, "saml", "https://okta.mine/", "nomatch", ""),
						acct(t, 1, "saml", "nomatch", "bl", ""),
						acct(t, 1, "nomatch", "https://okta.mine/", "bl", ""),
					},
					expMine: nil,
				},
				{
					description: "many accounts, 1 match",
					user:        &types.User{ID: 123},
					current: []*extsvc.Account{
						acct(t, 1, "saml", "nomatch", "bl", ""),
						acct(t, 1, "nomatch", "https://okta.mine/", "bl", ""),
						acct(t, 1, "saml", "https://okta.mine/", "bl", ""),
					},
					expMine: acct(t, 123, extsvc.TypeGitLab, "https://gitlab.mine/", "101", ""),
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
					expMine:     acct(t, 123, extsvc.TypeGitLab, "https://gitlab.mine/", "101", ""),
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
					current:     []*extsvc.Account{acct(t, 1, "openidconnect", "https://onelogin.mine/", "bl", "")},
					expMine:     acct(t, 123, extsvc.TypeGitLab, "https://gitlab.mine/", "101", ""),
				},
				{
					description: "0 authn providers match",
					user:        &types.User{ID: 123},
					current:     []*extsvc.Account{acct(t, 1, "openidconnect", "https://onelogin.mine/", "nomatch", "")},
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
			authzProvider := newSudoProvider(test.op, nil)
			for _, c := range test.calls {
				t.Run(c.description, func(t *testing.T) {
					acct, err := authzProvider.FetchAccount(ctx, c.user, c.current)
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					// ignore Data field in comparison
					if acct != nil {
						acct.Data, c.expMine.Data = nil, nil
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
