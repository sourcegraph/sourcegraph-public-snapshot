package gerrit

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
)

type client interface {
	ListAccountsByEmail(ctx context.Context, email string) (gerrit.ListAccountsResponse, error)
	ListAccountsByUsername(ctx context.Context, username string) (gerrit.ListAccountsResponse, error)
	GetAccountGroups(ctx context.Context, acctID int32) (gerrit.GetAccountGroupsResponse, error)
	ListProjects(ctx context.Context, opts gerrit.ListProjectsArgs) (projects *gerrit.ListProjectsResponse, nextPage bool, err error)
	GetProjectAccess(ctx context.Context, projects ...string) (gerrit.GetProjectAccessResponse, error)
}

var _ client = (*ClientAdapter)(nil)

// ClientAdapter is an adapter for Gerrit API client.
type ClientAdapter struct {
	*gerrit.Client
}

type mockClient struct {
	mockListAccountsByEmail    func(ctx context.Context, email string) (gerrit.ListAccountsResponse, error)
	mockListAccountsByUsername func(ctx context.Context, username string) (gerrit.ListAccountsResponse, error)
	mockGetAccountGroups       func(ctx context.Context, acctID int32) (gerrit.GetAccountGroupsResponse, error)
	mockListProjects           func(ctx context.Context, opts gerrit.ListProjectsArgs) (projects *gerrit.ListProjectsResponse, nextPage bool, err error)
	mockGetProjectAccess       func(ctx context.Context, projects ...string) (gerrit.GetProjectAccessResponse, error)
}

func (m *mockClient) ListAccountsByEmail(ctx context.Context, email string) (gerrit.ListAccountsResponse, error) {
	if m.mockListAccountsByEmail != nil {
		return m.mockListAccountsByEmail(ctx, email)
	}
	return nil, nil
}

func (m *mockClient) ListAccountsByUsername(ctx context.Context, username string) (gerrit.ListAccountsResponse, error) {
	if m.mockListAccountsByUsername != nil {
		return m.mockListAccountsByUsername(ctx, username)
	}
	return nil, nil
}

func (m *mockClient) GetAccountGroups(ctx context.Context, acctID int32) (gerrit.GetAccountGroupsResponse, error) {
	if m.mockGetAccountGroups != nil {
		return m.mockGetAccountGroups(ctx, acctID)
	}
	return nil, nil
}

func (m *mockClient) ListProjects(ctx context.Context, opts gerrit.ListProjectsArgs) (projects *gerrit.ListProjectsResponse, nextPage bool, err error) {
	if m.mockListProjects != nil {
		return m.mockListProjects(ctx, opts)
	}
	return nil, false, nil
}

func (m *mockClient) GetProjectAccess(ctx context.Context, projects ...string) (gerrit.GetProjectAccessResponse, error) {
	if m.mockGetProjectAccess != nil {
		return m.mockGetProjectAccess(ctx, projects...)
	}
	return nil, nil
}
