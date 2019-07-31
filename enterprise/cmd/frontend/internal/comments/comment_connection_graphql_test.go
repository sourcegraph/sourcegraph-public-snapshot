package comments

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestGraphQL_Namespace_CommentConnection(t *testing.T) {
	resetMocks()
	const (
		wantOrgID      = 3
		wantCommentID = 2
	)
	db.Mocks.Orgs.GetByID = func(context.Context, int32) (*types.Org, error) {
		return &types.Org{ID: wantOrgID}, nil
	}
	mocks.comments.List = func(dbCommentsListOptions) ([]*dbComment, error) {
		return []*dbComment{{ID: wantCommentID, Name: "n"}}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				{
					node(id: "T3JnOjM=") {
						... on Org {
							comments {
								nodes {
									name
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
						"comments": {
							"nodes": [
								{
									"name": "n"
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
