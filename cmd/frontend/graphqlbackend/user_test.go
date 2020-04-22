package graphqlbackend

import (
	"context"
	"errors"
	"testing"

	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestUser(t *testing.T) {
	t.Run("by username", func(t *testing.T) {
		checkUserByUsername := func(t *testing.T) {
			t.Helper()
			gqltesting.RunTests(t, []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t),
					Query: `
				{
					user(username: "alice") {
						username
					}
				}
			`,
					ExpectedResult: `
				{
					"user": {
						"username": "alice"
					}
				}
			`,
				},
			})
		}

		resetMocks()
		db.Mocks.Users.GetByUsername = func(_ context.Context, username string) (*types.User, error) {
			if want := "alice"; username != want {
				t.Errorf("got %q, want %q", username, want)
			}
			return &types.User{ID: 1, Username: "alice"}, nil
		}

		t.Run("allowed on Sourcegraph.com", func(t *testing.T) {
			orig := envvar.SourcegraphDotComMode()
			envvar.MockSourcegraphDotComMode(true)
			defer envvar.MockSourcegraphDotComMode(orig) // reset

			checkUserByUsername(t)
		})

		t.Run("allowed on non-Sourcegraph.com", func(t *testing.T) {
			checkUserByUsername(t)
		})
	})

	t.Run("by email", func(t *testing.T) {
		resetMocks()
		db.Mocks.Users.GetByVerifiedEmail = func(_ context.Context, email string) (*types.User, error) {
			if want := "alice@example.com"; email != want {
				t.Errorf("got %q, want %q", email, want)
			}
			return &types.User{ID: 1, Username: "alice"}, nil
		}

		t.Run("disallowed on Sourcegraph.com", func(t *testing.T) {
			checkUserByEmailError := func(t *testing.T, wantErr string) {
				t.Helper()
				gqltesting.RunTests(t, []*gqltesting.Test{
					{
						Schema: mustParseGraphQLSchema(t),
						Query: `
				{
					user(email: "alice@example.com") {
						username
					}
				}
			`,
						ExpectedResult: `{"user": null}`,
						ExpectedErrors: []*gqlerrors.QueryError{{Message: wantErr, Path: []interface{}{"user"}, ResolverError: errors.New(wantErr)}},
					},
				})
			}

			orig := envvar.SourcegraphDotComMode()
			envvar.MockSourcegraphDotComMode(true)
			defer envvar.MockSourcegraphDotComMode(orig) // reset

			t.Run("for anonymous viewer", func(t *testing.T) {
				db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
					return nil, db.ErrNoCurrentUser
				}
				checkUserByEmailError(t, "not authenticated")
			})
			t.Run("for non-site-admin viewer", func(t *testing.T) {
				db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
					return &types.User{SiteAdmin: false}, nil
				}
				checkUserByEmailError(t, "must be site admin")
			})
		})

		t.Run("allowed on non-Sourcegraph.com", func(t *testing.T) {
			gqltesting.RunTests(t, []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t),
					Query: `
				{
					user(email: "alice@example.com") {
						username
					}
				}
			`,
					ExpectedResult: `
				{
					"user": {
						"username": "alice"
					}
				}
			`,
				},
			})
		})
	})
}

func TestNode_User(t *testing.T) {
	resetMocks()
	db.Mocks.Users.MockGetByID_Return(t, &types.User{ID: 1, Username: "alice"}, nil)

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
				{
					node(id: "VXNlcjox") {
						id
						... on User {
							username
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"node": {
						"id": "VXNlcjox",
						"username": "alice"
					}
				}
			`,
		},
	})
}
