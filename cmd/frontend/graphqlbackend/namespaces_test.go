package graphqlbackend

import (
	"context"
	"errors"
	"testing"

	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db"
)

func TestNamespace(t *testing.T) {
	t.Run("user", func(t *testing.T) {
		resetMocks()
		const wantUserID = 3
		db.Mocks.Users.GetByID = func(_ context.Context, id int32) (*types.User, error) {
			if id != wantUserID {
				t.Errorf("got %d, want %d", id, wantUserID)
			}
			return &types.User{ID: wantUserID, Username: "alice"}, nil
		}
		gqltesting.RunTests(t, []*gqltesting.Test{
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
		db.Mocks.Orgs.GetByID = func(_ context.Context, id int32) (*types.Org, error) {
			if id != wantOrgID {
				t.Errorf("got %d, want %d", id, wantOrgID)
			}
			return &types.Org{ID: wantOrgID, Name: "acme"}, nil
		}
		gqltesting.RunTests(t, []*gqltesting.Test{
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
		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Schema: mustParseGraphQLSchema(t),
				Query: `
				{
					namespace(id: "aW52YWxpZDoz") {
						__typename
					}
				}
			`,
				ExpectedResult: `
				{
					"namespace": null
				}
			`,
				ExpectedErrors: []*gqlerrors.QueryError{
					{
						Message:       "invalid ID for namespace",
						Path:          []interface{}{"namespace"},
						ResolverError: errors.New("invalid ID for namespace"),
					},
				},
			},
		})
	})
}
