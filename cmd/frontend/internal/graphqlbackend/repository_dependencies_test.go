package graphqlbackend

import (
	"context"
	"testing"

	"github.com/neelance/graphql-go/gqltesting"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func TestRepositoryResolver_Dependencies(t *testing.T) {
	resetMocks()

	db.Mocks.GlobalDeps.Dependencies = func(ctx context.Context, op db.DependenciesOptions) ([]*api.DependencyReference, error) {
		return []*api.DependencyReference{{
			RepoID: 1,
			DepData: map[string]interface{}{
				"name": "d",
			},
		}}, nil
	}
	db.Mocks.Repos.MockGetByURI(t, "r", 1)
	db.Mocks.Repos.MockGet_Return(t, &types.Repo{ID: 1, URI: "r"})

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					repository(uri: "r") {
						dependencies {
							name
							repository {
								uri
							}
						}
					}
				}
		`,
			ExpectedResult: `
			{
				"repository": {
					"dependencies": [{
						"name": "d",
						"repository": {
							"uri": "r"
						}
					}]
				}
			}
		`,
		},
	})
}
