package graphqlbackend

import (
	"context"
	"testing"

	"github.com/neelance/graphql-go/gqltesting"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func TestDeps(t *testing.T) {
	resetMocks()

	db.Mocks.GlobalDeps.Dependencies = func(ctx context.Context, op db.DependenciesOptions) ([]*sourcegraph.DependencyReference, error) {
		if op.Language == "python" && op.DepData["name"] == "wwerkzeug" {
			return []*sourcegraph.DependencyReference{{
				RepoID: 1,
				DepData: map[string]interface{}{
					"name": "wwerkzeug",
				},
			}}, nil
		}
		return nil, nil
	}
	db.Mocks.Repos.MockGet_Return(t, &sourcegraph.Repo{ID: 1, URI: "github.com/pallets/fflask"})

	gqltesting.RunTests(t, []*gqltesting.Test{{
		Schema: GraphQLSchema,
		Query: `
				{
					dependents(lang: "python", name: "wwerkzeug", limit: 10) {
						name
					}
				}
		`,
		ExpectedResult: `
			{
				"dependents": [{
					"name": "wwerkzeug"
				}]
			}
		`,
	}, {
		Schema: GraphQLSchema,
		Query: `
				{
					dependents(lang: "python", name: "wwerkzeug", limit: 10) {
						name
						repo {
							uri
						}
					}
				}
		`,
		ExpectedResult: `
			{
				"dependents": [{
					"name": "wwerkzeug",
					"repo": {
						"uri": "github.com/pallets/fflask"
					}
				}]
			}
		`,
	}, {
		Schema: GraphQLSchema,
		Query: `
				{
					dependents(lang: "go", name: "wwerkzeug", limit: 10) {
						name
						repo {
							uri
						}
					}
				}
		`,
		ExpectedResult: `
			{
				"dependents": []
			}
		`,
	}})
}
