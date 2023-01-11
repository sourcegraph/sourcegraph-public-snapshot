package gerrit

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func TestProvider_ValidateConnection(t *testing.T) {
	testCases := []struct {
		name     string
		client   mockClient
		warnings []string
	}{
		{
			name: "GetGroup fails",
			client: mockClient{
				mockGetGroup: func(ctx context.Context, email string) (gerrit.Group, error) {
					return gerrit.Group{}, errors.New("fake error")
				},
			},
			warnings: []string{fmt.Sprintf("Unable to get %s group: %v", adminGroupName, errors.New("fake error"))},
		},
		{
			name: "no access to admin group",
			client: mockClient{
				mockGetGroup: func(ctx context.Context, email string) (gerrit.Group, error) {
					return gerrit.Group{
						ID: "",
					}, nil
				},
			},
			warnings: []string{fmt.Sprintf("Gerrit credentials not sufficent enough to query %s group", adminGroupName)},
		},
		{
			name: "admin group is valid",
			client: mockClient{
				mockGetGroup: func(ctx context.Context, email string) (gerrit.Group, error) {
					return gerrit.Group{
						ID:        "71242ef4aa1025f600bcefbe41d4902e231fc92a",
						CreatedOn: "2020-11-27 13:49:45.000000000",
						Name:      adminGroupName,
					}, nil
				},
			},
			warnings: []string{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestProvider(&tc.client)
			warnings := p.ValidateConnection(context.Background())
			if diff := cmp.Diff(warnings, tc.warnings); diff != "" {
				t.Fatalf("warnings did not match: %s", diff)
			}

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
