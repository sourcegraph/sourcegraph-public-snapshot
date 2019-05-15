package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestDiscussionThreadTargetConnection(t *testing.T) {
	resetMocks()
	mockViewerCanUseDiscussions = func() error { return nil }
	defer func() { mockViewerCanUseDiscussions = nil }()
	const wantThreadID = 123
	db.Mocks.DiscussionThreads.List = func(_ context.Context, opt *db.DiscussionThreadsListOptions) ([]*types.DiscussionThread, error) {
		return []*types.DiscussionThread{{ID: wantThreadID}}, nil
	}
	db.Mocks.DiscussionThreads.ListTargets = func(threadID int64) ([]*types.DiscussionThreadTargetRepo, error) {
		if threadID != wantThreadID {
			t.Fatalf("got threadID %v, want %v", threadID, wantThreadID)
		}
		return []*types.DiscussionThreadTargetRepo{{Path: strptr("foo/bar")}, {Path: strptr("qux")}, {Path: strptr("zap")}}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					discussionThreads {
						nodes {
							targets(first: 2) {
								nodes {
									__typename
									... on DiscussionThreadTargetRepo {
										path
									}
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
					"discussionThreads": {
						"nodes": [
							{
								"targets": {
									"nodes": [
										{
											"__typename": "DiscussionThreadTargetRepo",
											"path": "foo/bar"
										},
										{
											"__typename": "DiscussionThreadTargetRepo",
											"path": "qux"
										}
									],
									"totalCount": 3,
									"pageInfo": {
										"hasNextPage": true
									}
								}
							}
						]
					}
				}
			`,
		},
	})
}
