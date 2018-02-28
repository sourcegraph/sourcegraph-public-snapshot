package graphqlbackend

import (
	"context"
	"testing"

	"github.com/neelance/graphql-go/gqltesting"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func TestRepositoryResolver_Packages(t *testing.T) {
	resetMocks()

	db.Mocks.Pkgs.ListPackages = func(ctx context.Context, op *api.ListPackagesOp) ([]api.PackageInfo, error) {
		return []api.PackageInfo{{
			RepoID: 1,
			Lang:   "python",
			Pkg: map[string]interface{}{
				"name": "p",
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
						packages {
							lang
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
					"packages": [{
						"lang": "python",
						"name": "p",
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
