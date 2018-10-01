package graphqlbackend

import (
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func TestRepositoryResolver_Dependencies(t *testing.T) {
	resetMocks()

	backend.Mocks.Dependencies.List = func(*types.Repo, api.CommitID, bool) ([]*api.DependencyReference, error) {
		return []*api.DependencyReference{{
			Language: "go",
			RepoID:   1,
			DepData: map[string]interface{}{
				"name": "d",
			},
		}}, nil
	}
	backend.Mocks.Repos.MockResolveRev_NoCheck(t, "cccccccccccccccccccccccccccccccccccccccc")
	backend.Mocks.Repos.MockGetCommit_Return_NoCheck(t, &git.Commit{})
	db.Mocks.Repos.MockGetByURI(t, "r", 1)

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					repository(name: "r") {
						dependencies {
							nodes {
								language
								data {
									key
									value
								}
								dependingCommit {
									repository {
										name
									}
								}
							}
							totalCount
							pageInfo {
								hasNextPage
							}
						}
					}
				}
		`,
			ExpectedResult: `
			{
				"repository": {
					"dependencies": {
						"nodes": [{
							"language": "go",
							"data": [
								{
									"key": "name",
									"value": "d"
								}
							],
							"dependingCommit": {
								"repository": {
									"name": "r"
								}
							}
						}],
						"totalCount": 1,
						"pageInfo": {
							"hasNextPage": false
						}
					}
				}
			}
		`,
		},
	})
}
