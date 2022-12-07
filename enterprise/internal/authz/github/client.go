package github

import (
	"context"

	gh "github.com/google/go-github/v48/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

// 🚨 SECURITY: Call sites should take care to provide this valid values and use the return
// value appropriately to ensure org repo access are only provided to valid users.
func canViewOrgRepos(org *gh.Organization, orgMembership *gh.Membership) bool {
	if org == nil && orgMembership == nil {
		return false
	}
	// If user is active org admin, they can see all org repos
	if orgMembership != nil && orgMembership.GetState() == "active" && orgMembership.GetRole() == "admin" {
		return true
	}
	// https://github.com/organizations/$ORG/settings/member_privileges -> "Base permissions"
	return (org.GetDefaultRepoPermission() == "read" ||
		org.GetDefaultRepoPermission() == "write" ||
		org.GetDefaultRepoPermission() == "admin")
}

// client defines the set of GitHub API client methods used by the authz provider.
type client interface {
	ListAffiliatedRepositories(ctx context.Context, visibility github.Visibility, page int, affiliations ...github.RepositoryAffiliation) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)
	ListOrgRepositories(ctx context.Context, org string, page int, repoType string) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)
	ListTeamRepositories(ctx context.Context, org, team string, page int) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)

	ListRepositoryCollaborators(ctx context.Context, owner, repo string, page int, affiliations github.CollaboratorAffiliation) (users []*github.Collaborator, hasNextPage bool, _ error)
	ListRepositoryTeams(ctx context.Context, owner, repo string, page int) (teams []*github.Team, hasNextPage bool, _ error)

	ListOrganizationMembers(ctx context.Context, owner string, page int, adminsOnly bool) (users []*github.Collaborator, hasNextPage bool, _ error)
	ListTeamMembers(ctx context.Context, owner, team string, page int) (users []*github.Collaborator, hasNextPage bool, _ error)

	GetAuthenticatedUserOrgsDetailsAndMembership(ctx context.Context, page int) (orgs []github.OrgDetailsAndMembership, hasNextPage bool, rateLimitCost int, err error)
	GetAuthenticatedUserTeams(ctx context.Context, page int) (teams []*github.Team, hasNextPage bool, rateLimitCost int, err error)
	GetOrganization(ctx context.Context, login string) (org *github.OrgDetails, err error)
	GetRepository(ctx context.Context, owner, name string) (*github.Repository, error)

	GetAuthenticatedOAuthScopes(ctx context.Context) ([]string, error)
	WithAuthenticator(auther auth.Authenticator) client
}

var _ client = (*ClientAdapter)(nil)

// ClientAdapter is an adapter for GitHub API client.
type ClientAdapter struct {
	*github.V3Client
}

func (c *ClientAdapter) WithAuthenticator(auther auth.Authenticator) client {
	return &ClientAdapter{
		V3Client: c.V3Client.WithAuthenticator(auther),
	}
}
