package graphqlbackend

import (
	"testing"

	"github.com/neelance/graphql-go/gqltesting"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func TestRepositories(t *testing.T) {
	resetMocks()
	localstore.Mocks.Repos.MockList(t, "repo1", "repo2")
	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					root {
						repositories {
							nodes { uri }
							totalCount
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"root": {
						"repositories": {
							"nodes": [
								{
									"uri": "repo1"
								},
								{
									"uri": "repo2"
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
