package comments

import (
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func TestGraphQL_CommentConnection(t *testing.T) {
	resetMocks()
	const (
		wantCommentID = 2
		wantThreadID  = 3
	)
	mocks.newGQLToComment = func(v *DBComment) (graphqlbackend.Comment, error) { return &mockComment{body: v.Body}, nil }
	mocks.comments.List = func(dbCommentsListOptions) ([]*DBComment, error) {
		return []*DBComment{{ID: wantCommentID, ThreadID: wantThreadID, Body: "b"}}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: graphqlbackend.GraphQLSchema,
			Query: `
				{
					comments {
						nodes {
							body
						}
						totalCount
						pageInfo {
							hasNextPage
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"comments": {
						"nodes": [
							{
								"body": "b"
							}
						],
						"totalCount": 1,
						"pageInfo": {
							"hasNextPage": false
						}
					}
				}
			`,
		},
	})
}
