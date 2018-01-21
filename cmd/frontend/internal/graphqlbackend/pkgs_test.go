package graphqlbackend

import (
	"context"
	"testing"

	"github.com/neelance/graphql-go/gqltesting"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func TestPkgs(t *testing.T) {
	resetMocks()

	db.Mocks.Pkgs.ListPackages = func(ctx context.Context, op *api.ListPackagesOp) ([]api.PackageInfo, error) {
		if op.Lang == "python" && op.PkgQuery["name"] == "fflask" {
			return []api.PackageInfo{{
				RepoID: 1,
				Lang:   "python",
				Pkg: map[string]interface{}{
					"name": "fflask",
				},
			}}, nil
		}
		return nil, nil
	}
	db.Mocks.Repos.MockGet_Return(t, &types.Repo{ID: 1, URI: "github.com/pallets/fflask"})

	gqltesting.RunTests(t, []*gqltesting.Test{{
		Schema: GraphQLSchema,
		Query: `
				{
					packages(lang: "python", name: "fflask") {
						lang
						name
					}
				}
		`,
		ExpectedResult: `
			{
				"packages": [{
					"lang": "python",
					"name": "fflask"
				}]
			}
		`,
	}, {
		Schema: GraphQLSchema,
		Query: `
				{
					packages(lang: "python", name: "fflask") {
						lang
						name
						repo {
							uri
						}
					}
				}
		`,
		ExpectedResult: `
			{
				"packages": [{
					"lang": "python",
					"name": "fflask",
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
					packages(lang: "go", name: "fflask") {
						lang
						name
						repo {
							uri
						}
					}
				}
		`,
		ExpectedResult: `
			{
				"packages": []
			}
		`,
	}})
}
