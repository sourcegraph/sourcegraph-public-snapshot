package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
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

func TestAddRepository(t *testing.T) {
	resetMocks()
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	backend.Mocks.Repos.Add = func(uri api.RepoURI) error { return nil }
	db.Mocks.Repos.MockGetByURI(t, "my/repo", 123)
	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				mutation {
					addRepository(name: "my/repo") {
					    id
					}
				}
			`,
			ExpectedResult: `
				{
					"addRepository": {
						"id": "UmVwb3NpdG9yeToxMjM="
					}
				}
			`,
		},
	})
}
