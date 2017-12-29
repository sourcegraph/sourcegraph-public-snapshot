package graphqlbackend

import (
	"context"
	"testing"

	"github.com/neelance/graphql-go/gqltesting"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

func TestOrgs(t *testing.T) {
	resetMocks()
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*sourcegraph.User, error) {
		return &sourcegraph.User{SiteAdmin: true}, nil
	}
	backend.Mocks.Orgs.MockList(t, "org1", "org2")
	db.Mocks.Orgs.Count = func(context.Context) (int, error) { return 2, nil }
	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					orgs {
						nodes { name }
						totalCount
					}
				}
			`,
			ExpectedResult: `
				{
					"orgs": {
						"nodes": [
							{
								"name": "org1"
							},
							{
								"name": "org2"
							}
						],
						"totalCount": 2
					}
				}
			`,
		},
	})
}
