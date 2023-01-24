package gerrit

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/schema"
)

type client interface {
	ListAccountsByEmail(ctx context.Context, email string) (gerrit.ListAccountsResponse, error)
	ListAccountsByUsername(ctx context.Context, username string) (gerrit.ListAccountsResponse, error)
	ListProjects(ctx context.Context, opts gerrit.ListProjectsArgs) (*gerrit.ListProjectsResponse, bool, error)
	GetGroup(ctx context.Context, groupName string) (gerrit.Group, error)
	WithAuthenticator(a auth.Authenticator) client
}

var _ client = (*ClientAdapter)(nil)

// ClientAdapter is an adapter for Gerrit API client.
type ClientAdapter struct {
	*gerrit.Client
}

func NewClient(urn string, config *schema.GerritConnection, httpClient httpcli.Doer) (client, error) {
	c, err := gerrit.NewClient(urn, config, httpClient)
	if err != nil {
		return nil, err
	}
	return &ClientAdapter{c}, nil
}

// WithAuthenticator returns a new ClientAdapter with the given authenticator.
func (m *ClientAdapter) WithAuthenticator(a auth.Authenticator) client {
	return &ClientAdapter{m.Client.WithAuthenticator(a)}
}

type mockClient struct {
	mockListAccountsByEmail    func(ctx context.Context, email string) (gerrit.ListAccountsResponse, error)
	mockListAccountsByUsername func(ctx context.Context, username string) (gerrit.ListAccountsResponse, error)
	mockListProjects           func(ctx context.Context, opts gerrit.ListProjectsArgs) (*gerrit.ListProjectsResponse, bool, error)
	mockGetGroup               func(ctx context.Context, groupName string) (gerrit.Group, error)
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

func (m *mockClient) ListProjects(ctx context.Context, opts gerrit.ListProjectsArgs) (*gerrit.ListProjectsResponse, bool, error) {
	if m.mockListProjects != nil {
		return m.mockListProjects(ctx, opts)
	}

	return nil, false, nil
}

func (m *mockClient) GetGroup(ctx context.Context, groupName string) (gerrit.Group, error) {
	if m.mockGetGroup != nil {
		return m.mockGetGroup(ctx, groupName)
	}
	return gerrit.Group{}, nil
}

func (m *mockClient) WithAuthenticator(a auth.Authenticator) client {
	return m
}
