package gerrit

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestProvider_ValidateConnection(t *testing.T) {
	testCases := []struct {
		name    string
		client  mockClient
		wantErr string
	}{
		{
			name: "GetGroup fails",
			client: mockClient{
				mockGetGroup: func(ctx context.Context, email string) (gerrit.Group, error) {
					return gerrit.Group{}, errors.New("fake error")
				},
			},
			wantErr: fmt.Sprintf("Unable to get %s group: %v", adminGroupName, errors.New("fake error")),
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
			wantErr: fmt.Sprintf("Gerrit credentials not sufficent enough to query %s group", adminGroupName),
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
			wantErr: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestProvider(&tc.client)
			err := p.ValidateConnection(context.Background())
			errMessage := ""
			if err != nil {
				errMessage = err.Error()
			}
			if diff := cmp.Diff(errMessage, tc.wantErr); diff != "" {
				t.Fatalf("warnings did not match: %s", diff)
			}

		})
	}
}

func TestProvider_FetchUserPerms(t *testing.T) {
	accountData := extsvc.AccountData{}
	err := gerrit.SetExternalAccountData(&accountData, &gerrit.Account{}, &gerrit.AccountCredentials{
		Username: "test-user",
		Password: "test-password",
	})
	if err != nil {
		t.Fatal(err)
	}

	mClient := mockClient{}
	mClient.mockListProjects = func(ctx context.Context, opts gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error) {
		resp := gerrit.ListProjectsResponse{
			"test-project": &gerrit.Project{
				ID: "test-project",
			},
		}

		return resp, false, nil
	}

	testCases := map[string]struct {
		client    mockClient
		account   *extsvc.Account
		wantErr   bool
		wantPerms *authz.ExternalUserPermissions
	}{
		"nil account gives error": {
			client:  mClient,
			account: nil,
			wantErr: true,
		},
		"account of wrong service type gives error": {
			client: mClient,
			account: &extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: "github",
					ServiceID:   "https://gerrit.sgdev.org/",
				},
			},
			wantErr: true,
		},
		"account of wrong service id gives error": {
			client: mClient,
			account: &extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: "gerrit",
					ServiceID:   "https://github.sgdev.org/",
				},
			},
			wantErr: true,
		},
		"account with no data gives error": {
			client: mClient,
			account: &extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: "gerrit",
					ServiceID:   "https://gerrit.sgdev.org/",
				},
				AccountData: extsvc.AccountData{},
			},
			wantErr: true,
		},
		"correct account gives correct permissions": {
			client: mClient,
			account: &extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: "gerrit",
					ServiceID:   "https://gerrit.sgdev.org/",
				},
				AccountData: accountData,
			},
			wantPerms: &authz.ExternalUserPermissions{
				Exacts: []extsvc.RepoID{"test-project"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			p := NewTestProvider(&tc.client)
			perms, err := p.FetchUserPerms(context.Background(), tc.account, authz.FetchPermsOptions{})
			if err != nil && !tc.wantErr {
				t.Fatalf("unexpected error: %s", err)
			}
			if err == nil && tc.wantErr {
				t.Fatalf("expected error but got none")
			}
			if diff := cmp.Diff(perms, tc.wantPerms); diff != "" {
				t.Fatalf("permissions did not match: %s", diff)
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
