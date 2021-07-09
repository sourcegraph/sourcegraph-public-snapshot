package graphqlbackend

import (
	"context"
	"fmt"
	"testing"

	"github.com/graph-gophers/graphql-go"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestNamespace(t *testing.T) {
	t.Run("user", func(t *testing.T) {
		resetMocks()
		const wantUserID = 3
		database.Mocks.Users.GetByID = func(_ context.Context, id int32) (*types.User, error) {
			if id != wantUserID {
				t.Errorf("got %d, want %d", id, wantUserID)
			}
			return &types.User{ID: wantUserID, Username: "alice"}, nil
		}
		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t),
				Query: `
				{
					namespace(id: "VXNlcjoz") {
						__typename
						... on User { username }
					}
				}
			`,
				ExpectedResult: `
				{
					"namespace": {
						"__typename": "User",
						"username": "alice"
					}
				}
			`,
			},
		})
	})

	t.Run("organization", func(t *testing.T) {
		resetMocks()
		const wantOrgID = 3
		database.Mocks.Orgs.GetByID = func(_ context.Context, id int32) (*types.Org, error) {
			if id != wantOrgID {
				t.Errorf("got %d, want %d", id, wantOrgID)
			}
			return &types.Org{ID: wantOrgID, Name: "acme"}, nil
		}
		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t),
				Query: `
				{
					namespace(id: "T3JnOjM=") {
						__typename
						... on Org { name }
					}
				}
			`,
				ExpectedResult: `
				{
					"namespace": {
						"__typename": "Org",
						"name": "acme"
					}
				}
			`,
			},
		})
	})

	t.Run("invalid", func(t *testing.T) {
		resetMocks()

		invalidID := "aW52YWxpZDoz"
		wantErr := InvalidNamespaceIDErr{id: graphql.ID(invalidID)}

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t),
				Query: fmt.Sprintf(`
				{
					namespace(id: %q) {
						__typename
					}
				}
			`, invalidID),
				ExpectedResult: `
				{
					"namespace": null
				}
			`,
				ExpectedErrors: []*gqlerrors.QueryError{
					{
						Path:          []interface{}{"namespace"},
						Message:       wantErr.Error(),
						ResolverError: wantErr,
					},
				},
			},
		})
	})
}

func TestNamespaceByName(t *testing.T) {
	t.Run("user", func(t *testing.T) {
		resetMocks()
		const (
			wantName   = "alice"
			wantUserID = 123
		)
		database.Mocks.Namespaces.GetByName = func(ctx context.Context, name string) (*database.Namespace, error) {
			if name != wantName {
				t.Errorf("got %q, want %q", name, wantName)
			}
			return &database.Namespace{Name: "alice", User: wantUserID}, nil
		}
		database.Mocks.Users.GetByID = func(_ context.Context, id int32) (*types.User, error) {
			if id != wantUserID {
				t.Errorf("got %d, want %d", id, wantUserID)
			}
			return &types.User{ID: wantUserID, Username: wantName}, nil
		}
		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t),
				Query: `
				{
					namespaceByName(name: "alice") {
						__typename
						... on User { username }
					}
				}
			`,
				ExpectedResult: `
				{
					"namespaceByName": {
						"__typename": "User",
						"username": "alice"
					}
				}
			`,
			},
		})
	})

	t.Run("organization", func(t *testing.T) {
		resetMocks()
		const (
			wantName  = "acme"
			wantOrgID = 3
		)
		database.Mocks.Namespaces.GetByName = func(ctx context.Context, name string) (*database.Namespace, error) {
			if name != wantName {
				t.Errorf("got %q, want %q", name, wantName)
			}
			return &database.Namespace{Name: "alice", Organization: wantOrgID}, nil
		}
		database.Mocks.Orgs.GetByID = func(_ context.Context, id int32) (*types.Org, error) {
			if id != wantOrgID {
				t.Errorf("got %d, want %d", id, wantOrgID)
			}
			return &types.Org{ID: wantOrgID, Name: "acme"}, nil
		}
		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t),
				Query: `
				{
					namespaceByName(name: "acme") {
						__typename
						... on Org { name }
					}
				}
			`,
				ExpectedResult: `
				{
					"namespaceByName": {
						"__typename": "Org",
						"name": "acme"
					}
				}
			`,
			},
		})
	})

	t.Run("invalid", func(t *testing.T) {
		resetMocks()
		database.Mocks.Namespaces.GetByName = func(ctx context.Context, name string) (*database.Namespace, error) {
			return nil, database.ErrNamespaceNotFound
		}
		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t),
				Query: `
				{
					namespaceByName(name: "doesntexist") {
						__typename
					}
				}
			`,
				ExpectedResult: `
				{
					"namespaceByName": null
				}
			`,
			},
		})
	})
}
