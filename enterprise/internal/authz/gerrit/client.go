package gerrit

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

type client interface {
	ListProjects(ctx context.Context, opts gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error)
	GetGroup(ctx context.Context, groupName string) (gerrit.Group, error)
	WithAuthenticator(a auth.Authenticator) client
}

var _ client = (*ClientAdapter)(nil)

// ClientAdapter is an adapter for Gerrit API client.
type ClientAdapter struct {
	*gerrit.Client
}

// NewClient creates a new Gerrit client and wraps it in a ClientAdapter.
func NewClient(urn string, baseURL *url.URL, creds *gerrit.AccountCredentials, httpClient httpcli.Doer) (client, error) {
	c, err := gerrit.NewClient(urn, baseURL, creds, httpClient)
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
	mockListProjects func(ctx context.Context, opts gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error)
	mockGetGroup     func(ctx context.Context, groupName string) (gerrit.Group, error)
}

func (m *mockClient) ListProjects(ctx context.Context, opts gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error) {
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
