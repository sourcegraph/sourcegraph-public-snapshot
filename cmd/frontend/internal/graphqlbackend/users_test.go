package graphqlbackend

import (
	"testing"

	"github.com/neelance/graphql-go/gqltesting"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
)

func TestUsers(t *testing.T) {
	resetMocks()
	backend.Mocks.Users.MockList(t, "user1", "user2")
	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					root {
						users {
							nodes { username }
							totalCount
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"root": {
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
