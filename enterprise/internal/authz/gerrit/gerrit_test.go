package gerrit

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestProvider_FetchUserPerms(t *testing.T) {
	userEmail := "test-email@example.com"
	userName := "test-user"
	gerritUserID := int32(123456789)
	acctData, err := marshalAccountData(userName, userEmail, gerritUserID)
	if err != nil {
		t.Fatalf("error marshaling account data: %v", err)
	}
	account := extsvc.Account{
		ID:     1,
		UserID: 1,
		AccountData: extsvc.AccountData{
			Data: acctData,
		},
	}
	testCases := []struct {
		name          string
		client        mockClient
		expectedPerms []extsvc.RepoID
	}{
		{
			name: "one of user's groups has access to one project but not the other",
			expectedPerms: []extsvc.RepoID{
				"project-with-access",
			},
			client: mockClient{
				mockGetProjectAccess: func(ctx context.Context, projects ...string) (gerrit.GetProjectAccessResponse, error) {
					return gerrit.GetProjectAccessResponse{
						"project-with-access": {
							Groups: map[string]gerrit.GroupInfo{
								"group-with-access": {
									ID: "group-with-access",
								},
							},
						},
						"project-without-access": {
							Groups: map[string]gerrit.GroupInfo{
								"group-without-access": {
									ID: "group-without-access",
								},
							},
						},
					}, nil
				},
			},
		},
		{
			name:          "user has access to repo which inherits access",
			expectedPerms: []extsvc.RepoID{"project-with-access"},
			client: mockClient{
				mockGetProjectAccess: func(ctx context.Context, projects ...string) (gerrit.GetProjectAccessResponse, error) {
					if len(projects) == 1 {
						// return the inherited project access
						if projects[0] == "All-Access" {
							return gerrit.GetProjectAccessResponse{
								"All-Access": {
									Groups: map[string]gerrit.GroupInfo{"group-with-access": {ID: "group-with-access"}},
								},
							}, nil
						}
						if projects[0] == "No-Access" {
							return gerrit.GetProjectAccessResponse{
								"No-Access": {
									Groups: map[string]gerrit.GroupInfo{"group-without-access": {ID: "group-without-access"}},
								},
							}, nil
						}
						t.Errorf("unexpected project access request for project %s", projects[0])
					}
					return gerrit.GetProjectAccessResponse{
						"project-with-access": {
							InheritsFrom: gerrit.Project{
								ID: "All-Access",
							},
						},
						"project-without-access": {
							InheritsFrom: gerrit.Project{
								ID: "No-Access",
							},
						},
					}, nil
				},
			},
		},
		{
			name:          "user has access to repo which has multiple layers of inherited access",
			expectedPerms: []extsvc.RepoID{"project-with-access"},
			client: mockClient{
				mockGetProjectAccess: func(ctx context.Context, projects ...string) (gerrit.GetProjectAccessResponse, error) {
					if len(projects) == 1 {
						// return the inherited project access
						if projects[0] == "Maybe-Access" {
							return gerrit.GetProjectAccessResponse{
								"Maybe-Access": {
									InheritsFrom: gerrit.Project{
										ID: "All-Access",
									},
								},
							}, nil
						}
						if projects[0] == "All-Access" {
							return gerrit.GetProjectAccessResponse{
								"All-Access": {
									Groups: map[string]gerrit.GroupInfo{"group-with-access": {ID: "group-with-access"}},
								},
							}, nil
						}
						if projects[0] == "No-Access" {
							return gerrit.GetProjectAccessResponse{
								"No-Access": {
									Groups: map[string]gerrit.GroupInfo{"group-without-access": {ID: "group-without-access"}},
								},
							}, nil
						}
						t.Errorf("unexpected project access request for project %s", projects[0])
					}
					return gerrit.GetProjectAccessResponse{
						"project-with-access": {
							InheritsFrom: gerrit.Project{
								ID: "Maybe-Access",
							},
						},
						"project-without-access": {
							InheritsFrom: gerrit.Project{
								ID: "No-Access",
							},
						},
					}, nil
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.client.mockGetAccountGroups = func(ctx context.Context, acctID int32) (gerrit.GetAccountGroupsResponse, error) {
				return gerrit.GetAccountGroupsResponse{
					{
						ID: "group-with-access",
					},
				}, nil
			}
			tc.client.mockListProjects = func(ctx context.Context, opts gerrit.ListProjectsArgs) (projects *gerrit.ListProjectsResponse, nextPage bool, err error) {
				return &gerrit.ListProjectsResponse{
					"project-with-access": {
						ID: "project-with-access",
					},
					"project-without-access": {
						ID: "project-without-access",
					},
				}, false, nil
			}
			p := NewTestProvider(&tc.client)
			perms, err := p.FetchUserPerms(context.Background(), &account, authz.FetchPermsOptions{})
			if err != nil {
				t.Errorf("unexpected error fetching user perms: %v", err)
			}
			assert.Equal(t, tc.expectedPerms, perms.Exacts)

		})
	}
}

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
		urn:              "Gerrit",
		client:           client,
		codeHost:         extsvc.NewCodeHost(baseURL, extsvc.TypeGerrit),
		projectAccessMap: map[string]gerrit.ProjectAccessInfo{},
	}
}
