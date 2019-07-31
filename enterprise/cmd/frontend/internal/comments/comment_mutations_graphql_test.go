package comments

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/testutil"
)

func TestGraphQL_CreateComment(t *testing.T) {
	resetMocks()
	const wantUserID = 1
	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: wantUserID}, nil
	}
	threadlike.MockThreadOrIssueOrChangesetByDBID = func(int64) (graphqlbackend.ThreadOrIssueOrChangeset, error) {
		return graphqlbackend.ThreadOrIssueOrChangeset{Thread: testutil.ThreadFixture}, nil
	}
	defer func() { threadlike.MockThreadOrIssueOrChangesetByDBID = nil }()
	wantComment := &dbComment{
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
					createComment(input: { node: "VGhyZWFkOjE=", body: "b" }) {
						id
						body
					}
				}
			`,
			ExpectedResult: `
				{
					"createComment": {
						"id": "Q2FtcGFpZ246Mg==",
						"body": "b"
					}
				}
			`,
		},
	})
}

func TestGraphQL_UpdateComment(t *testing.T) {
	resetMocks()
	const wantID = 2
	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1}, nil
	}
	mocks.comments.GetByID = func(id int64) (*dbComment, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbComment{ID: wantID}, nil
	}
	mocks.comments.Update = func(id int64, update dbCommentUpdate) (*dbComment, error) {
		if want := (dbCommentUpdate{Body: strptr("b1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &dbComment{
			ID:           2,
			AuthorUserID: 1,
			Body:         "b1",
		}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					updateComment(input: { id: "Q2FtcGFpZ246Mg==", body: "b1" }) {
						id
						body
					}
				}
			`,
			ExpectedResult: `
				{
					"updateComment": {
						"id": "Q2FtcGFpZ246Mg==",
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
	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1}, nil
	}
	mocks.comments.GetByID = func(id int64) (*dbComment, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
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
					deleteComment(comment: "Q2FtcGFpZ246Mg==") {
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
