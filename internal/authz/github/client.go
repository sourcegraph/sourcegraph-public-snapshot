package github

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

// client defines the set of GitHub API client methods used by the authz provider.
//
// NOTE: All methods are sorted in alphabetical order.
type client interface {
	ListAffiliatedRepositories(ctx context.Context, visibility github.Visibility, page int, affiliations ...github.Affiliation) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)
	ListRepositoryCollaborators(ctx context.Context, owner, repo string, page int) (users []*github.Collaborator, hasNextPage bool, _ error)

	GetAuthenticatedUserOrgsDetails(ctx context.Context, page int) (orgs []*github.OrgDetails, hasNextPage bool, rateLimitCost int, err error)
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
	MockListAffiliatedRepositories      func(ctx context.Context, visibility github.Visibility, page int, affiliations ...github.Affiliation) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error)
	MockListRepositoryCollaborators     func(ctx context.Context, owner, repo string, page int) (users []*github.Collaborator, hasNextPage bool, _ error)
	MockGetAuthenticatedUserOrgsDetails func(ctx context.Context, page int) (orgs []*github.OrgDetails, hasNextPage bool, rateLimitCost int, err error)
	MockGetAuthenticatedUserTeams       func(ctx context.Context, page int) (teams []*github.Team, hasNextPage bool, rateLimitCost int, err error)
	MockWithToken                       func(token string) client
}

func (m *mockClient) ListAffiliatedRepositories(ctx context.Context, visibility github.Visibility, page int, affiliations ...github.Affiliation) ([]*github.Repository, bool, int, error) {
	return m.MockListAffiliatedRepositories(ctx, visibility, page)
}

func (m *mockClient) ListRepositoryCollaborators(ctx context.Context, owner, repo string, page int) ([]*github.Collaborator, bool, error) {
	return m.MockListRepositoryCollaborators(ctx, owner, repo, page)
}

func (m *mockClient) GetAuthenticatedUserOrgsDetails(ctx context.Context, page int) (orgs []*github.OrgDetails, hasNextPage bool, rateLimitCost int, err error) {
	return m.MockGetAuthenticatedUserOrgsDetails(ctx, page)
}

func (m *mockClient) GetAuthenticatedUserTeams(ctx context.Context, page int) (teams []*github.Team, hasNextPage bool, rateLimitCost int, err error) {
	return m.MockGetAuthenticatedUserTeams(ctx, page)
}

func (m *mockClient) WithToken(token string) client {
	return m.MockWithToken(token)
}
