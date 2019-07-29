package changesets

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestGraphQL_Repository_ChangesetConnection(t *testing.T) {
	resetMocks()
	const (
		wantRepositoryID = 3
		wantChangesetID     = 2
	)
	db.Mocks.Repos.Get = func(context.Context, api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: wantRepositoryID, Name: "r"}, nil
	}
	mocks.changesets.List = func(dbChangesetsListOptions) ([]*dbChangeset, error) {
		return []*dbChangeset{{ID: wantChangesetID, RepositoryID: wantRepositoryID, Title: "t"}}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				{
					node(id: "UmVwb3NpdG9yeToz") {
						... on Repository {
							changesets {
								nodes {
									title
								}
								totalCount
								pageInfo {
									hasNextPage
								}
							}
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"node": {
						"changesets": {
							"nodes": [
								{
									"title": "t"
								}
							],
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
