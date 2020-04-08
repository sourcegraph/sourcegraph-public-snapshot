package github

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
)

var _ Client = (*MockClient)(nil)

type MockClient struct {
	MockWithToken                      func(token string) Client
	MockRateLimit                      func() *ratelimit.Monitor
	MockCreatePullRequest              func(ctx context.Context, in *CreatePullRequestInput) (*PullRequest, error)
	MockClosePullRequest               func(ctx context.Context, pr *PullRequest) error
	MockLoadPullRequests               func(ctx context.Context, prs ...*PullRequest) error
	MockUpdatePullRequest              func(ctx context.Context, in *UpdatePullRequestInput) (*PullRequest, error)
	MockGetOpenPullRequestByRefs       func(ctx context.Context, owner, name, baseRef, headRef string) (*PullRequest, error)
	MockGetRepository                  func(ctx context.Context, owner, name string) (*Repository, error)
	MockGetRepositoryByNodeID          func(ctx context.Context, token, id string) (*Repository, error)
	MockGetRepositoriesByNodeIDFromAPI func(ctx context.Context, token string, nodeIDs []string) (map[string]*Repository, error)
	MockGetReposByNameWithOwner        func(ctx context.Context, namesWithOwners ...string) ([]*Repository, error)
	MockGetAuthenticatedUserEmails     func(ctx context.Context) ([]*UserEmail, error)
	MockGetAuthenticatedUserOrgs       func(ctx context.Context) ([]*Org, error)
	MockListPublicRepositories         func(ctx context.Context, sinceRepoID int64) ([]*Repository, error)
	MockListRepositoriesForSearch      func(ctx context.Context, searchString string, page int) (RepositoryListPage, error)
	MockListOrgRepositories            func(ctx context.Context, org string, page int) (repos []*Repository, hasNextPage bool, rateLimitCost int, err error)
	MockListUserRepositories           func(ctx context.Context, user string, page int) (repos []*Repository, hasNextPage bool, rateLimitCost int, err error)
	MockListAffiliatedRepositories     func(ctx context.Context, page int) (repos []*Repository, hasNextPage bool, rateLimitCost int, err error)
	MockListInstallationRepositories   func(ctx context.Context) ([]*Repository, error)
	MockListRepositoryCollaborators    func(ctx context.Context, owner, repo string, page int) (users []*Collaborator, hasNextPage bool, err error)
}

func (m *MockClient) WithToken(token string) Client {
	return m.MockWithToken(token)
}

func (m *MockClient) RateLimit() *ratelimit.Monitor {
	return m.MockRateLimit()
}

func (m *MockClient) CreatePullRequest(ctx context.Context, in *CreatePullRequestInput) (*PullRequest, error) {
	return m.MockCreatePullRequest(ctx, in)
}

func (m *MockClient) ClosePullRequest(ctx context.Context, pr *PullRequest) error {
	return m.MockClosePullRequest(ctx, pr)
}

func (m *MockClient) LoadPullRequests(ctx context.Context, prs ...*PullRequest) error {
	return m.MockLoadPullRequests(ctx, prs...)
}

func (m *MockClient) UpdatePullRequest(ctx context.Context, in *UpdatePullRequestInput) (*PullRequest, error) {
	return m.MockUpdatePullRequest(ctx, in)
}

func (m *MockClient) GetOpenPullRequestByRefs(ctx context.Context, owner, name, baseRef, headRef string) (*PullRequest, error) {
	return m.MockGetOpenPullRequestByRefs(ctx, owner, name, baseRef, headRef)
}

func (m *MockClient) GetRepository(ctx context.Context, owner, name string) (*Repository, error) {
	return m.MockGetRepository(ctx, owner, name)
}

func (m *MockClient) GetRepositoryByNodeID(ctx context.Context, token, id string) (*Repository, error) {
	return m.MockGetRepositoryByNodeID(ctx, token, id)
}

func (m *MockClient) GetRepositoriesByNodeIDFromAPI(ctx context.Context, token string, nodeIDs []string) (map[string]*Repository, error) {
	return m.MockGetRepositoriesByNodeIDFromAPI(ctx, token, nodeIDs)
}

func (m *MockClient) GetReposByNameWithOwner(ctx context.Context, namesWithOwners ...string) ([]*Repository, error) {
	return m.MockGetReposByNameWithOwner(ctx, namesWithOwners...)
}

func (m *MockClient) GetAuthenticatedUserEmails(ctx context.Context) ([]*UserEmail, error) {
	return m.MockGetAuthenticatedUserEmails(ctx)
}

func (m *MockClient) GetAuthenticatedUserOrgs(ctx context.Context) ([]*Org, error) {
	return m.MockGetAuthenticatedUserOrgs(ctx)
}

func (m *MockClient) ListPublicRepositories(ctx context.Context, sinceRepoID int64) ([]*Repository, error) {
	return m.MockListPublicRepositories(ctx, sinceRepoID)
}

func (m *MockClient) ListRepositoriesForSearch(ctx context.Context, searchString string, page int) (RepositoryListPage, error) {
	return m.MockListRepositoriesForSearch(ctx, searchString, page)
}

func (m *MockClient) ListOrgRepositories(ctx context.Context, org string, page int) (repos []*Repository, hasNextPage bool, rateLimitCost int, err error) {
	return m.MockListOrgRepositories(ctx, org, page)
}

func (m *MockClient) ListUserRepositories(ctx context.Context, user string, page int) (repos []*Repository, hasNextPage bool, rateLimitCost int, err error) {
	return m.MockListUserRepositories(ctx, user, page)
}

func (m *MockClient) ListAffiliatedRepositories(ctx context.Context, page int) (repos []*Repository, hasNextPage bool, rateLimitCost int, err error) {
	return m.MockListAffiliatedRepositories(ctx, page)
}

func (m *MockClient) ListInstallationRepositories(ctx context.Context) ([]*Repository, error) {
	return m.MockListInstallationRepositories(ctx)
}

func (m *MockClient) ListRepositoryCollaborators(ctx context.Context, owner, repo string, page int) (users []*Collaborator, hasNextPage bool, err error) {
	return m.MockListRepositoryCollaborators(ctx, owner, repo, page)
}
