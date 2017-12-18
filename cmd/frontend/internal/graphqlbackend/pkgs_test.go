package graphqlbackend

import (
	"context"
	"testing"

	"github.com/neelance/graphql-go/gqltesting"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func TestPkgs(t *testing.T) {
	resetMocks()

	localstore.Mocks.Pkgs.ListPackages = func(ctx context.Context, op *sourcegraph.ListPackagesOp) ([]sourcegraph.PackageInfo, error) {
		if op.Lang == "python" && op.PkgQuery["name"] == "fflask" {
			return []sourcegraph.PackageInfo{{
				RepoID: 1,
				Lang:   "python",
				Pkg: map[string]interface{}{
					"name": "fflask",
				},
			}}, nil
		}
		return nil, nil
	}
	localstore.Mocks.Repos.MockGet_Return(t, &sourcegraph.Repo{ID: 1, URI: "github.com/pallets/fflask"})

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
