package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCreateUser(t *testing.T) {
	resetMocks()
	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	database.Mocks.Users.Create = func(context.Context, database.NewUser) (*types.User, error) {
		return &types.User{ID: 1, Username: "alice"}, nil
	}

	calledGrantPendingPermissions := false
	database.Mocks.Authz.GrantPendingPermissions = func(context.Context, *database.GrantPendingPermissionsArgs) error {
		calledGrantPendingPermissions = true
		return nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
				mutation {
					createUser(username: "alice") {
						user {
							id
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"createUser": {
						"user": {
							"id": "VXNlcjox"
						}
					}
				}
			`,
		},
	})
	if !calledGrantPendingPermissions {
		t.Fatal("!calledGrantPendingPermissions")
	}
}
