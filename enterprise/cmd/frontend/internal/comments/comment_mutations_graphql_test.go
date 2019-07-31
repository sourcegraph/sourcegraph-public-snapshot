package comments

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike"
)

func TestGraphQL_CreateComment(t *testing.T) {
	resetMocks()
	const wantUserID = 1
	wantThreadGQLID := threadlike.MarshalID(threadlike.GQLTypeThread, 1)
	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: wantUserID}, nil
	}
	mocks.newGQLToComment = func(v *dbComment) (graphqlbackend.Comment, error) { return &mockComment{body: v.Body}, nil }
	wantComment := &dbComment{
		Object:       dbCommentObject{ThreadID: 1},
		AuthorUserID: wantUserID,
		Body:         "b",
	}
	mocks.comments.Create = func(comment *dbComment) (*dbComment, error) {
		if !reflect.DeepEqual(comment, wantComment) {
			t.Errorf("got comment %+v, want %+v", comment, wantComment)
		}
		tmp := *comment
		tmp.ID = 2
		return &tmp, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					createComment(input: { node: "` + string(wantThreadGQLID) + `", body: "b" }) {
						body
					}
				}
			`,
			ExpectedResult: `
				{
					"createComment": {
						"body": "b"
					}
				}
			`,
		},
	})
}

func TestGraphQL_EditComment(t *testing.T) {
	resetMocks()
	const (
		wantID       = 2
		wantThreadID = 1
	)
	wantThreadGQLID := threadlike.MarshalID(threadlike.GQLTypeThread, wantThreadID)
	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1}, nil
	}
	mocks.commentByGQLID = func(id graphql.ID) (*dbComment, error) {
		if id != wantThreadGQLID {
			t.Errorf("got thread ID %q, want %q", id, wantThreadGQLID)
		}
		return &dbComment{ID: wantID}, nil
	}
	mocks.newGQLToComment = func(v *dbComment) (graphqlbackend.Comment, error) { return &mockComment{body: v.Body}, nil }
	mocks.comments.Update = func(id int64, update dbCommentUpdate) (*dbComment, error) {
		if want := (dbCommentUpdate{Body: strptr("b1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &dbComment{
			ID:           2,
			Object:       dbCommentObject{ThreadID: 1},
			AuthorUserID: 1,
			Body:         "b1",
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
	resetMocks()
	const wantID = 2
	wantThreadGQLID := threadlike.MarshalID(threadlike.GQLTypeThread, 1)
	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1}, nil
	}
	mocks.commentByGQLID = func(id graphql.ID) (*dbComment, error) {
		if id != wantThreadGQLID {
			t.Errorf("got thread ID %q, want %q", id, wantThreadGQLID)
		}
		return &dbComment{ID: wantID}, nil
	}
	mocks.comments.DeleteByID = func(id int64) error {
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
