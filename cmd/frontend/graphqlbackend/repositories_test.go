package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

func TestRepositories(t *testing.T) {
	resetMocks()
	db.Mocks.Repos.MockList(t, "repo1", "repo2")
	db.Mocks.Repos.Count = func(context.Context, db.ReposListOptions) (int, error) { return 2, nil }
	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					repositories {
						nodes { name }
						totalCount
					}
				}
			`,
			ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{
								"name": "repo1"
							},
							{
								"name": "repo2"
							}
						],
						"totalCount": null
					}
				}
			`,
		},
	})
}
