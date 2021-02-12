package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestUsers(t *testing.T) {
	resetMocks()
	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	database.Mocks.Users.List = func(ctx context.Context, opt *database.UsersListOptions) ([]*types.User, error) {
		return []*types.User{{Username: "user1"}, {Username: "user2"}}, nil
	}
	database.Mocks.Users.Count = func(context.Context, *database.UsersListOptions) (int, error) { return 2, nil }
	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
				{
					users {
						nodes { username }
						totalCount
					}
				}
			`,
			ExpectedResult: `
				{
					"users": {
						"nodes": [
							{
								"username": "user1"
							},
							{
								"username": "user2"
							}
						],
						"totalCount": 2
					}
				}
			`,
		},
	})
}
