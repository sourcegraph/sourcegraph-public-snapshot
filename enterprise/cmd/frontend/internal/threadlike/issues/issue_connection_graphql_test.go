package issues

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestGraphQL_Repository_IssueConnection(t *testing.T) {
	internal.ResetMocks()
	const (
		wantRepositoryID = 3
		wantIssueID  = 2
	)
	db.Mocks.Repos.Get = func(context.Context, api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: wantRepositoryID, Name: "r"}, nil
	}
	internal.Mocks.Threads.List = func(internal.DBThreadsListOptions) ([]*internal.DBThread, error) {
		return []*internal.DBThread{{ID: wantIssueID, RepositoryID: wantRepositoryID, Title: "t"}}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				{
					node(id: "UmVwb3NpdG9yeToz") {
						... on Repository {
							issues {
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
						"issues": {
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
