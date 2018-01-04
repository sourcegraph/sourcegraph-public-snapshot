package graphqlbackend

import (
	"context"
	"testing"

	"github.com/neelance/graphql-go/gqltesting"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

func TestUsers(t *testing.T) {
	resetMocks()
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*sourcegraph.User, error) {
		return &sourcegraph.User{SiteAdmin: true}, nil
	}
	db.Mocks.Users.List = func(ctx context.Context, opt *db.UsersListOptions) ([]*sourcegraph.User, error) {
		return []*sourcegraph.User{{Username: "user1"}, {Username: "user2"}}, nil
	}
	db.Mocks.Users.Count = func(context.Context) (int, error) { return 2, nil }
	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					site {
						users {
							nodes { username }
							totalCount
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"site": {
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
				}
			`,
		},
	})
}
