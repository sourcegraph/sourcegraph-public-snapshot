package comments

import (
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/internal"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/types"
)

func TestGraphQL_CommentConnection(t *testing.T) {
	internal.ResetMocks()
	const (
		wantCommentID = 2
		wantThreadID  = 3
	)
	mockNewGQLToComment = func(v *internal.DBComment) (graphqlbackend.Comment, error) { return &mockComment{body: v.Body}, nil }
	defer func() { mockNewGQLToComment = nil }()
	internal.Mocks.Comments.List = func(internal.DBCommentsListOptions) ([]*internal.DBComment, error) {
		return []*internal.DBComment{{ID: wantCommentID, Object: types.CommentObject{ThreadID: wantThreadID}, Body: "b"}}, nil
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
