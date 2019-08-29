package threads

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestGraphQL_CreateThread(t *testing.T) {
	resetMocks()
	const wantRepositoryID = 1
	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1}, nil
	}
	db.Mocks.Repos.Get = func(context.Context, api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: wantRepositoryID}, nil
	}
	wantThread := &DBThread{
		RepositoryID: wantRepositoryID,
		Title:        "t",
		State:        string(graphqlbackend.ThreadStateOpen),
	}
	mocks.threads.Create = func(thread *DBThread) (*DBThread, error) {
		if !reflect.DeepEqual(thread, wantThread) {
			t.Errorf("got thread %+v, want %+v", thread, wantThread)
		}
		tmp := *thread
		tmp.ID = 2
		return &tmp, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					createThread(input: { repository: "T3JnOjE=", title: "t" }) {
						id
						title
					}
				}
			`,
			ExpectedResult: `
				{
					"createThread": {
						"id": "VGhyZWFkOjI=",
						"title": "t"
					}
				}
			`,
		},
	})
}

func TestGraphQL_UpdateThread(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.threads.GetByID = func(id int64) (*DBThread, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &DBThread{ID: wantID}, nil
	}
	mocks.threads.Update = func(id int64, update dbThreadUpdate) (*DBThread, error) {
		if want := (dbThreadUpdate{Title: strptr("t1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &DBThread{
			ID:           2,
			RepositoryID: 1,
			Title:        "t1",
		}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					updateThread(input: { id: "VGhyZWFkOjI=", title: "t1" }) {
						id
						title
					}
				}
			`,
			ExpectedResult: `
				{
					"updateThread": {
						"id": "VGhyZWFkOjI=",
						"title": "t1"
					}
				}
			`,
		},
	})
}

func TestGraphQL_DeleteThread(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.threads.GetByID = func(id int64) (*DBThread, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &DBThread{ID: wantID}, nil
	}
	mocks.threads.DeleteByID = func(id int64) error {
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
					deleteThread(thread: "VGhyZWFkOjI=") {
						alwaysNil
					}
				}
			`,
			ExpectedResult: `
				{
					"deleteThread": null
				}
			`,
		},
	})
}

func strptr(s string) *string { return &s }
