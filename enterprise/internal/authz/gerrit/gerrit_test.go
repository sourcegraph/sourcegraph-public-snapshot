package gerrit

import (
	"context"
	"net/url"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestProvider_FetchAccount(t *testing.T) {
	userEmail := "test-email@example.com"
	userName := "test-user"
	testCases := []struct {
		name   string
		client mockClient
	}{
		{
			name: "no matching username but email match",
			client: mockClient{
				mockListAccountsByEmail: func(ctx context.Context, email string) (gerrit.ListAccountsResponse, error) {
					return []gerrit.Account{
						{
							Email: userEmail,
						},
					}, nil
				},
				mockListAccountsByUsername: nil,
			},
		},
		{
			name: "username matches and email valid",
			client: mockClient{
				mockListAccountsByEmail: nil,
				mockListAccountsByUsername: func(ctx context.Context, username string) (gerrit.ListAccountsResponse, error) {
					return []gerrit.Account{
						{
							Email:    userEmail,
							Username: username,
						},
					}, nil
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestProvider(&tc.client)
			user := types.User{
				Username: userName,
			}
			verifiedEmails := []string{
				userEmail,
			}
			acct, err := p.FetchAccount(context.Background(), &user, nil, verifiedEmails)
			if err != nil {
				t.Fatalf("error fetching account: %s", err)
			}
			if acct == nil {
				t.Fatalf("account was nil")
			}
			// TODO: validate account
		})
	}
}

func NewTestProvider(client client) *Provider {
	baseURL, _ := url.Parse("https://gerrit.sgdev.org")
	return &Provider{
		urn:      "Gerrit",
		client:   client,
		codeHost: extsvc.NewCodeHost(baseURL, extsvc.TypeGerrit),
	}
}
