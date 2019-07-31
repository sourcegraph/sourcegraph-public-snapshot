package comments

import (
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/testutil"
)

func TestGraphQL_CommentConnection(t *testing.T) {
	resetMocks()
	const (
		wantCommentID = 2
		wantThreadID  = 3
	)
	threadlike.MockThreadOrIssueOrChangesetByDBID = func(int64) (graphqlbackend.ThreadOrIssueOrChangeset, error) {
		return graphqlbackend.ThreadOrIssueOrChangeset{Thread: testutil.ThreadFixture}, nil
	}
	defer func() { threadlike.MockThreadOrIssueOrChangesetByDBID = nil }()
	mocks.comments.List = func(dbCommentsListOptions) ([]*dbComment, error) {
		return []*dbComment{{ID: wantCommentID, ThreadID: wantThreadID, Body: "b"}}, nil
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
