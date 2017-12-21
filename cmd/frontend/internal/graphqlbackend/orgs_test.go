package graphqlbackend

import (
	"testing"

	"github.com/neelance/graphql-go/gqltesting"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
)

func TestOrgs(t *testing.T) {
	resetMocks()
	backend.Mocks.Orgs.MockList(t, "org1", "org2")
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
