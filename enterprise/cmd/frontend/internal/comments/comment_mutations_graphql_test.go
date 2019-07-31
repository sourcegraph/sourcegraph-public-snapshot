package comments

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestGraphQL_CreateComment(t *testing.T) {
	resetMocks()
	const wantOrgID = 1
	db.Mocks.Orgs.GetByID = func(context.Context, int32) (*types.Org, error) {
		return &types.Org{ID: wantOrgID}, nil
	}
	wantComment := &dbComment{
		NamespaceOrgID: wantOrgID,
		Name:           "n",
		Description:    strptr("d"),
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
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					createComment(input: { namespace: "T3JnOjE=", name: "n", description: "d" }) {
						id
						name
					}
				}
			`,
			ExpectedResult: `
				{
					"createComment": {
						"id": "Q2FtcGFpZ246Mg==",
						"name": "n"
					}
				}
			`,
		},
	})
}

func TestGraphQL_UpdateComment(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.comments.GetByID = func(id int64) (*dbComment, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbComment{ID: wantID}, nil
	}
	mocks.comments.Update = func(id int64, update dbCommentUpdate) (*dbComment, error) {
		if want := (dbCommentUpdate{Name: strptr("n1"), Description: strptr("d1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &dbComment{
			ID:             2,
			NamespaceOrgID: 1,
			Name:           "n1",
			Description:    strptr("d1"),
		}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					updateComment(input: { id: "Q2FtcGFpZ246Mg==", name: "n1", description: "d1" }) {
						id
						name
						description
					}
				}
			`,
			ExpectedResult: `
				{
					"updateComment": {
						"id": "Q2FtcGFpZ246Mg==",
						"name": "n1",
						"description": "d1"
					}
				}
			`,
		},
	})
}

func TestGraphQL_DeleteComment(t *testing.T) {
	resetMocks()
	const wantID = 2
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
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
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

func TestGraphQL_AddRemoveThreadsToFromComment(t *testing.T) {
	resetMocks()
	const wantCommentID = 2
	mocks.comments.GetByID = func(id int64) (*dbComment, error) {
		if id != wantCommentID {
			t.Errorf("got ID %d, want %d", id, wantCommentID)
		}
		return &dbComment{ID: wantCommentID}, nil
	}
	const wantThreadID = 3
	mockGetThreadDBIDs = func([]graphql.ID) ([]int64, error) { return []int64{wantThreadID}, nil }
	defer func() { mockGetThreadDBIDs = nil }()
	addRemoveThreadsToFromComment := func(comment int64, threads []int64) error {
		if comment != wantCommentID {
			t.Errorf("got %d, want %d", comment, wantCommentID)
		}
		if want := []int64{wantThreadID}; !reflect.DeepEqual(threads, want) {
			t.Errorf("got %v, want %v", threads, want)
		}
		return nil
	}

	tests := map[string]*func(thread int64, threads []int64) error{
		"addThreadsToComment":      &mocks.commentsThreads.AddThreadsToComment,
		"removeThreadsFromComment": &mocks.commentsThreads.RemoveThreadsFromComment,
	}
	for name, mockFn := range tests {
		t.Run(name, func(t *testing.T) {
			*mockFn = addRemoveThreadsToFromComment
			defer func() { *mockFn = nil }()

			gqltesting.RunTests(t, []*gqltesting.Test{
				{
					Context: backend.WithAuthzBypass(context.Background()),
					Schema:  graphqlbackend.GraphQLSchema,
					Query: `
				mutation {
					` + name + `(comment: "Q2FtcGFpZ246Mg==", threads: ["RGlzY3Vzc2lvblRocmVhZDoiMyI="]) {
						__typename
					}
				}
			`,
					ExpectedResult: `
				{
					"` + name + `": {
						"__typename": "EmptyResponse"
					}
				}
			`,
				},
			})
		})
	}
}
