package comments

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/actor"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/internal"
	comments_types "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/types"
)

func TestGraphQL_AddCommentReply(t *testing.T) {
	internal.ResetMocks()
	const (
		wantUserID          = 1
		wantThreadID        = 1
		wantThreadCommentID = 3
	)
	wantThreadGQLID := graphqlbackend.MarshalThreadID(wantThreadID)
	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: wantUserID}, nil
	}
	mockCommentByGQLID = func(id graphql.ID) (*internal.DBComment, error) {
		if id != wantThreadGQLID {
			t.Errorf("got thread ID %q, want %q", id, wantThreadGQLID)
		}
		return &internal.DBComment{ID: wantThreadCommentID}, nil
	}
	defer func() { mockCommentByGQLID = nil }()
	mockNewGQLToComment = func(v *internal.DBComment) (graphqlbackend.Comment, error) { return &mockComment{body: v.Body}, nil }
	defer func() { mockNewGQLToComment = nil }()
	wantComment := &internal.DBComment{
		Object: comments_types.CommentObject{ParentCommentID: wantThreadCommentID},
		Author: actor.DBColumns{UserID: wantUserID},
		Body:   "b",
	}
	internal.Mocks.Comments.Create = func(comment *internal.DBComment) (*internal.DBComment, error) {
		if !reflect.DeepEqual(comment, wantComment) {
			t.Errorf("got comment %+v, want %+v", comment, wantComment)
		}
		tmp := *comment
		tmp.ID = 2
		return &tmp, nil
	}
	events.MockCreateEvent = func(event events.CreationData) error {
		return nil
	}
	defer func() { events.MockCreateEvent = nil }()

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					addCommentReply(input: { parentComment: "` + string(wantThreadGQLID) + `", body: "b" }) {
						body
					}
				}
			`,
			ExpectedResult: `
				{
					"addCommentReply": {
						"body": "b"
					}
				}
			`,
		},
	})
}

func TestGraphQL_EditComment(t *testing.T) {
	internal.ResetMocks()
	const (
		wantID       = 2
		wantThreadID = 1
	)
	wantThreadGQLID := graphqlbackend.MarshalThreadID(wantThreadID)
	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1}, nil
	}
	mockCommentByGQLID = func(id graphql.ID) (*internal.DBComment, error) {
		if id != wantThreadGQLID {
			t.Errorf("got thread ID %q, want %q", id, wantThreadGQLID)
		}
		return &internal.DBComment{ID: wantID}, nil
	}
	defer func() { mockCommentByGQLID = nil }()

	mockNewGQLToComment = func(v *internal.DBComment) (graphqlbackend.Comment, error) { return &mockComment{body: v.Body}, nil }
	defer func() { mockNewGQLToComment = nil }()
	internal.Mocks.Comments.Update = func(id int64, update internal.DBCommentUpdate) (*internal.DBComment, error) {
		if want := (internal.DBCommentUpdate{Body: strptr("b1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &internal.DBComment{
			ID:     2,
			Object: comments_types.CommentObject{ThreadID: 1},
			Author: actor.DBColumns{UserID: 1},
			Body:   "b1",
		}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					editComment(input: { id: "` + string(wantThreadGQLID) + `", body: "b1" }) {
						body
					}
				}
			`,
			ExpectedResult: `
				{
					"editComment": {
						"body": "b1"
					}
				}
			`,
		},
	})
}

func TestGraphQL_DeleteComment(t *testing.T) {
	internal.ResetMocks()
	const wantID = 2
	wantThreadGQLID := graphqlbackend.MarshalThreadID(1)
	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1}, nil
	}
	mockCommentByGQLID = func(id graphql.ID) (*internal.DBComment, error) {
		if id != wantThreadGQLID {
			t.Errorf("got thread ID %q, want %q", id, wantThreadGQLID)
		}
		return &internal.DBComment{ID: wantID}, nil
	}
	defer func() { mockCommentByGQLID = nil }()
	internal.Mocks.Comments.DeleteByID = func(id int64) error {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					deleteComment(comment: "` + string(wantThreadGQLID) + `") {
						alwaysNil
					}
				}
			`,
			ExpectedResult: `
				{
					"deleteComment": null
				}
			`,
		},
	})
}

func strptr(s string) *string { return &s }
