package gerrit

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
)

type mockClient struct {
	url                             *url.URL
	mockWithAuthenticator           func(a auth.Authenticator) gerrit.Client
	mockListAccountsByEmail         func(ctx context.Context, email string) (gerrit.ListAccountsResponse, error)
	mockListAccountsByUsername      func(ctx context.Context, username string) (gerrit.ListAccountsResponse, error)
	mockGetGroup                    func(ctx context.Context, groupName string) (gerrit.Group, error)
	mockListProjects                func(ctx context.Context, opts gerrit.ListProjectsArgs) (*gerrit.ListProjectsResponse, bool, error)
	mockGetAuthenticatedUserAccount func(ctx context.Context) (*gerrit.Account, error)
}

func (m *mockClient) WithAuthenticator(a auth.Authenticator) gerrit.Client {
	if m.mockWithAuthenticator != nil {
		return m.mockWithAuthenticator(a)
	}

	return nil
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

func (m *mockClient) GetGroup(ctx context.Context, groupName string) (gerrit.Group, error) {
	if m.mockGetGroup != nil {
		return m.mockGetGroup(ctx, groupName)
	}
	return gerrit.Group{}, nil
}

func (m *mockClient) ListProjects(ctx context.Context, opts gerrit.ListProjectsArgs) (*gerrit.ListProjectsResponse, bool, error) {
	if m.mockListProjects != nil {
		return m.mockListProjects(ctx, opts)
	}
	return nil, false, nil
}

func (m *mockClient) URL() *url.URL {
	return m.url
}

func (m *mockClient) GetAuthenticatedUserAccount(ctx context.Context) (*gerrit.Account, error) {
	if m.mockGetAuthenticatedUserAccount != nil {
		return m.mockGetAuthenticatedUserAccount(ctx)
	}

	return nil, nil
}
