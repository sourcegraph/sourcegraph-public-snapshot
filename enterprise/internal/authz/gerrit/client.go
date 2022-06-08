package gerrit

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
)

type client interface {
	ListAccountsByEmail(ctx context.Context, email string) (gerrit.ListAccountsResponse, error)
	ListAccountsByUsername(ctx context.Context, username string) (gerrit.ListAccountsResponse, error)
}

var _ client = (*ClientAdapter)(nil)

// ClientAdapter is an adapter for Gerrit API client.
type ClientAdapter struct {
	*gerrit.Client
}

type mockClient struct {
	mockListAccountsByEmail    func(ctx context.Context, email string) (gerrit.ListAccountsResponse, error)
	mockListAccountsByUsername func(ctx context.Context, username string) (gerrit.ListAccountsResponse, error)
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
