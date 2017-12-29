package graphqlbackend

import (
	"context"
	"testing"

	"github.com/neelance/graphql-go/gqltesting"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

func TestUsers(t *testing.T) {
	resetMocks()
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*sourcegraph.User, error) {
		return &sourcegraph.User{SiteAdmin: true}, nil
	}
	backend.Mocks.Users.MockList(t, "user1", "user2")
	db.Mocks.Users.Count = func(context.Context) (int, error) { return 2, nil }
	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
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
