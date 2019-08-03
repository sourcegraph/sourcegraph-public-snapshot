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
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestGraphQL_CreateThread(t *testing.T) {
	internal.ResetMocks()
	const wantRepositoryID = 1
	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1}, nil
	}
	db.Mocks.Repos.Get = func(context.Context, api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: wantRepositoryID}, nil
	}
	wantThread := &internal.DBThread{
		Type:         internal.DBThreadTypeThread,
		RepositoryID: wantRepositoryID,
		Title:        "t",
		ExternalURL:  strptr("u"),
		Status:       string(graphqlbackend.ThreadStatusOpen),
	}
	internal.Mocks.Threads.Create = func(thread *internal.DBThread) (*internal.DBThread, error) {
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
					createThread(input: { repository: "T3JnOjE=", title: "t", externalURL: "u" }) {
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
	internal.ResetMocks()
	const wantID = 2
	internal.Mocks.Threads.GetByID = func(id int64) (*internal.DBThread, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &internal.DBThread{ID: wantID}, nil
	}
	internal.Mocks.Threads.Update = func(id int64, update internal.DBThreadUpdate) (*internal.DBThread, error) {
		if want := (internal.DBThreadUpdate{Title: strptr("t1"), ExternalURL: strptr("u1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &internal.DBThread{
			Type:         internal.DBThreadTypeThread,
			ID:           2,
			RepositoryID: 1,
			Title:        "t1",
			ExternalURL:  strptr("u1"),
		}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					updateThread(input: { id: "VGhyZWFkOjI=", title: "t1", externalURL: "u1" }) {
						id
						title
						externalURL
					}
				}
			`,
			ExpectedResult: `
				{
					"updateThread": {
						"id": "VGhyZWFkOjI=",
						"title": "t1",
						"externalURL": "u1"
					}
				}
			`,
		},
	})
}

func TestGraphQL_DeleteThread(t *testing.T) {
	internal.ResetMocks()
	const wantID = 2
	internal.Mocks.Threads.GetByID = func(id int64) (*internal.DBThread, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &internal.DBThread{ID: wantID}, nil
	}
	internal.Mocks.Threads.DeleteByID = func(id int64) error {
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
