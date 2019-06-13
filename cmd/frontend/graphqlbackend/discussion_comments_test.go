package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestDiscussionsMutations_UpdateComment(t *testing.T) {
	resetMocks()
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) { return &types.User{}, nil }
	mockViewerCanUseDiscussions = func() error { return nil }
	defer func() { mockViewerCanUseDiscussions = nil }()
	const (
		wantCommentID = 123
		wantThreadID  = 456
		wantContents  = "b"
	)
	db.Mocks.DiscussionThreads.Get = func(threadID int64) (*types.DiscussionThread, error) {
		if threadID != wantThreadID {
			t.Errorf("got threadID %v, want %v", threadID, wantThreadID)
		}
		return &types.DiscussionThread{}, nil
	}
	db.Mocks.DiscussionComments.Get = func(commentID int64) (*types.DiscussionComment, error) {
		if commentID != wantCommentID {
			t.Errorf("got commentID %v, want %v", commentID, wantCommentID)
		}
		return &types.DiscussionComment{ThreadID: wantThreadID}, nil
	}
	db.Mocks.DiscussionComments.Update = func(_ context.Context, commentID int64, opts *db.DiscussionCommentsUpdateOptions) (*types.DiscussionComment, error) {
		if commentID != wantCommentID {
			t.Errorf("got commentID %v, want %v", commentID, wantCommentID)
		}
		if opts == nil || opts.Contents == nil || *opts.Contents != wantContents {
			var contents string
			if opts != nil && opts.Contents != nil {
				contents = *opts.Contents
			}
			t.Errorf("got contents %v, want %v", contents, wantContents)
		}
		return &types.DiscussionComment{ThreadID: wantThreadID, Contents: wantContents}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  GraphQLSchema,
			Query: `
                                mutation($contents: String!) {
                                        discussions {
                                                updateComment(input: {commentID: "123", contents: $contents}) {
                                                        __typename
                                                }
                                        }
                                }
                        `,
			Variables: map[string]interface{}{"contents": wantContents},
			ExpectedResult: `
                                {
                                        "discussions": {
                                                "updateComment": {
                                                        "__typename": "DiscussionThread"
                                                }
                                        }
                                }
                        `,
		},
	})
}
