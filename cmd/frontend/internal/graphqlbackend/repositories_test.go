package graphqlbackend

import (
	"context"
	"testing"

	"github.com/neelance/graphql-go/gqltesting"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

func TestRepositories(t *testing.T) {
	resetMocks()
	db.Mocks.Repos.MockList(t, "repo1", "repo2")
	db.Mocks.Repos.Count = func(context.Context) (int, error) { return 2, nil }
	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					repositories {
						nodes { uri }
						totalCount
					}
				}
			`,
			ExpectedResult: `
				{
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
			`,
		},
	})
}
