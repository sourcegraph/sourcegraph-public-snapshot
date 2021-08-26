package github

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

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
//
// NOTE: All methods are sorted in alphabetical order.
type client interface {
	ListAffiliatedRepositories(ctx context.Context, visibility github.Visibility, page int, affiliations ...github.Affiliation) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)
	ListOrgRepositories(ctx context.Context, org string, page int, repoType string) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)
	ListTeamRepositories(ctx context.Context, org, team string, page int) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)

	ListRepositoryCollaborators(ctx context.Context, owner, repo string, page int) (users []*github.Collaborator, hasNextPage bool, _ error)

	GetAuthenticatedUserOrgsDetailsAndMembership(ctx context.Context, page int) (orgs []github.OrgDetailsAndMembership, hasNextPage bool, rateLimitCost int, err error)
	GetAuthenticatedUserTeams(ctx context.Context, page int) (teams []*github.Team, hasNextPage bool, rateLimitCost int, err error)

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
	MockListAffiliatedRepositories                   func(ctx context.Context, visibility github.Visibility, page int, affiliations ...github.Affiliation) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)
	MockListOrgRepositories                          func(ctx context.Context, org string, page int, repoType string) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)
	MockListTeamRepositories                         func(ctx context.Context, org, team string, page int) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)
	MockListRepositoryCollaborators                  func(ctx context.Context, owner, repo string, page int) (users []*github.Collaborator, hasNextPage bool, _ error)
	MockGetAuthenticatedUserOrgsDetailsAndMembership func(ctx context.Context, page int) (orgs []github.OrgDetailsAndMembership, hasNextPage bool, rateLimitCost int, err error)
	MockGetAuthenticatedUserTeams                    func(ctx context.Context, page int) (teams []*github.Team, hasNextPage bool, rateLimitCost int, err error)
	MockWithToken                                    func(token string) client
}

func (m *mockClient) ListAffiliatedRepositories(ctx context.Context, visibility github.Visibility, page int, affiliations ...github.Affiliation) ([]*github.Repository, bool, int, error) {
	return m.MockListAffiliatedRepositories(ctx, visibility, page)
}

func (m *mockClient) ListOrgRepositories(ctx context.Context, org string, page int, repoType string) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error) {
	return m.MockListOrgRepositories(ctx, org, page, repoType)
}

func (m *mockClient) ListTeamRepositories(ctx context.Context, org, team string, page int) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error) {
	return m.MockListTeamRepositories(ctx, org, team, page)
}

func (m *mockClient) ListRepositoryCollaborators(ctx context.Context, owner, repo string, page int) ([]*github.Collaborator, bool, error) {
	return m.MockListRepositoryCollaborators(ctx, owner, repo, page)
}

func (m *mockClient) GetAuthenticatedUserOrgsDetailsAndMembership(ctx context.Context, page int) (orgs []github.OrgDetailsAndMembership, hasNextPage bool, rateLimitCost int, err error) {
	return m.MockGetAuthenticatedUserOrgsDetailsAndMembership(ctx, page)
}

func (m *mockClient) GetAuthenticatedUserTeams(ctx context.Context, page int) (teams []*github.Team, hasNextPage bool, rateLimitCost int, err error) {
	return m.MockGetAuthenticatedUserTeams(ctx, page)
}

func (m *mockClient) WithToken(token string) client {
	if m.MockWithToken == nil {
		return m
	}
	return m.MockWithToken(token)
}
