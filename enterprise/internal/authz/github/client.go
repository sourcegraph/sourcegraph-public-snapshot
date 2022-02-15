package github

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

// 🚨 SECURITY: Call sites should take care to provide this valid values and use the return
// value appropriately to ensure org repo access are only provided to valid users.
func canViewOrgRepos(org *github.OrgDetailsAndMembership) bool {
	if org == nil {
		return false
	}
	// If user is active org admin, they can see all org repos
	if org.OrgMembership != nil && org.OrgMembership.State == "active" && org.OrgMembership.Role == "admin" {
		return true
	}
	// https://github.com/organizations/$ORG/settings/member_privileges -> "Base permissions"
	return org.OrgDetails != nil && (org.DefaultRepositoryPermission == "read" ||
		org.DefaultRepositoryPermission == "write" ||
		org.DefaultRepositoryPermission == "admin")
}

// client defines the set of GitHub API client methods used by the authz provider.
type client interface {
	ListAffiliatedRepositories(ctx context.Context, visibility github.Visibility, page int, affiliations ...github.RepositoryAffiliation) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)
	ListOrgRepositories(ctx context.Context, org string, page int, repoType string) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)
	ListTeamRepositories(ctx context.Context, org, team string, page int) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)

	ListRepositoryCollaborators(ctx context.Context, owner, repo string, page int, affiliations github.CollaboratorAffiliation) (users []*github.Collaborator, hasNextPage bool, _ error)
	ListRepositoryTeams(ctx context.Context, owner, repo string, page int) (teams []*github.Team, hasNextPage bool, _ error)

	ListOrganizations(ctx context.Context) (orgs []*github.Org, _ error)
	ListOrganizationMembers(ctx context.Context, owner string, page int, adminsOnly bool) (users []*github.Collaborator, hasNextPage bool, _ error)
	ListTeamMembers(ctx context.Context, owner, team string, page int) (users []*github.Collaborator, hasNextPage bool, _ error)

	GetAuthenticatedUserOrgsDetailsAndMembership(ctx context.Context, page int) (orgs []github.OrgDetailsAndMembership, hasNextPage bool, rateLimitCost int, err error)
	GetAuthenticatedUserTeams(ctx context.Context, page int) (teams []*github.Team, hasNextPage bool, rateLimitCost int, err error)
	GetOrganization(ctx context.Context, login string) (org *github.OrgDetails, err error)
	GetRepository(ctx context.Context, owner, name string) (*github.Repository, error)

	GetAuthenticatedOAuthScopes(ctx context.Context) ([]string, error)
	WithToken(token string) client
}

var _ client = (*ClientAdapter)(nil)

// ClientAdapter is an adapter for GitHub API client.
type ClientAdapter struct {
	*github.V3Client
}

func (c *ClientAdapter) WithToken(token string) client {
	return &ClientAdapter{
		V3Client: c.V3Client.WithAuthenticator(&auth.OAuthBearerToken{Token: token}),
	}
}

var _ client = (*mockClient)(nil)

type mockClient struct {
	MockListAffiliatedRepositories                   func(ctx context.Context, visibility github.Visibility, page int, affiliations ...github.RepositoryAffiliation) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)
	MockListOrgRepositories                          func(ctx context.Context, org string, page int, repoType string) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)
	MockListTeamRepositories                         func(ctx context.Context, org, team string, page int) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)
	MockListRepositoryCollaborators                  func(ctx context.Context, owner, repo string, page int, affiliation github.CollaboratorAffiliation) (users []*github.Collaborator, hasNextPage bool, _ error)
	MockListRepositoryTeams                          func(ctx context.Context, owner, repo string, page int) (teams []*github.Team, hasNextPage bool, _ error)
	MockListOrganizationMembers                      func(ctx context.Context, owner string, page int, adminOnly bool) (users []*github.Collaborator, hasNextPage bool, _ error)
	MockListTeamMembers                              func(ctx context.Context, owner, team string, page int) (users []*github.Collaborator, hasNextPage bool, _ error)
	MockGetAuthenticatedUserOrgsDetailsAndMembership func(ctx context.Context, page int) (orgs []github.OrgDetailsAndMembership, hasNextPage bool, rateLimitCost int, err error)
	MockGetAuthenticatedUserTeams                    func(ctx context.Context, page int) (teams []*github.Team, hasNextPage bool, rateLimitCost int, err error)
	MockGetOrganization                              func(ctx context.Context, login string) (org *github.OrgDetails, err error)
	MockGetRepository                                func(ctx context.Context, owner, repo string) (*github.Repository, error)

	MockGetAuthenticatedOAuthScopes func(ctx context.Context) ([]string, error)
	MockWithToken                   func(token string) client
}

func (m *mockClient) ListAffiliatedRepositories(ctx context.Context, visibility github.Visibility, page int, affiliations ...github.RepositoryAffiliation) ([]*github.Repository, bool, int, error) {
	return m.MockListAffiliatedRepositories(ctx, visibility, page, affiliations...)
}

func (m *mockClient) ListOrgRepositories(ctx context.Context, org string, page int, repoType string) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error) {
	return m.MockListOrgRepositories(ctx, org, page, repoType)
}

func (m *mockClient) ListTeamRepositories(ctx context.Context, org, team string, page int) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error) {
	return m.MockListTeamRepositories(ctx, org, team, page)
}

func (m *mockClient) ListRepositoryCollaborators(ctx context.Context, owner, repo string, page int, affiliation github.CollaboratorAffiliation) ([]*github.Collaborator, bool, error) {
	return m.MockListRepositoryCollaborators(ctx, owner, repo, page, affiliation)
}

func (m *mockClient) ListRepositoryTeams(ctx context.Context, owner, repo string, page int) ([]*github.Team, bool, error) {
	return m.MockListRepositoryTeams(ctx, owner, repo, page)
}

func (m *mockClient) ListOrganizationMembers(ctx context.Context, owner string, page int, adminOnly bool) (users []*github.Collaborator, hasNextPage bool, _ error) {
	return m.MockListOrganizationMembers(ctx, owner, page, adminOnly)
}

func (m *mockClient) ListTeamMembers(ctx context.Context, owner, team string, page int) (users []*github.Collaborator, hasNextPage bool, _ error) {
	return m.MockListTeamMembers(ctx, owner, team, page)
}

func (m *mockClient) GetAuthenticatedUserOrgsDetailsAndMembership(ctx context.Context, page int) (orgs []github.OrgDetailsAndMembership, hasNextPage bool, rateLimitCost int, err error) {
	return m.MockGetAuthenticatedUserOrgsDetailsAndMembership(ctx, page)
}

func (m *mockClient) GetAuthenticatedUserTeams(ctx context.Context, page int) (teams []*github.Team, hasNextPage bool, rateLimitCost int, err error) {
	return m.MockGetAuthenticatedUserTeams(ctx, page)
}

func (m *mockClient) GetOrganization(ctx context.Context, login string) (org *github.OrgDetails, err error) {
	return m.MockGetOrganization(ctx, login)
}

func (m *mockClient) GetRepository(ctx context.Context, owner, name string) (*github.Repository, error) {
	return m.MockGetRepository(ctx, owner, name)
}

func (m *mockClient) GetAuthenticatedOAuthScopes(ctx context.Context) ([]string, error) {
	return m.MockGetAuthenticatedOAuthScopes(ctx)
}

func (m *mockClient) WithToken(token string) client {
	if m.MockWithToken == nil {
		return m
	}
	return m.MockWithToken(token)
}
