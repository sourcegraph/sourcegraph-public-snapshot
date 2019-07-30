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
	db.Mocks.Repos.Get = func(context.Context, api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: wantRepositoryID}, nil
	}
	wantThread := &dbThread{
		DBThreadCommon: DBThreadCommon{
			RepositoryID: wantRepositoryID,
			Title:        "t",
			ExternalURL:  strptr("u"),
		},
		Status: graphqlbackend.ThreadStatusOpen,
	}
	mocks.threads.Create = func(thread *dbThread) (*dbThread, error) {
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
					threads {
						createThread(input: { repository: "T3JnOjE=", title: "t", externalURL: "u" }) {
							id
							title
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"threads": {
						"createThread": {
							"id": "VGhyZWFkOjI=",
							"title": "t"
						}
					}
				}
			`,
		},
	})
}

func TestGraphQL_UpdateThread(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.threads.GetByID = func(id int64) (*dbThread, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbThread{DBThreadCommon: DBThreadCommon{ID: wantID}}, nil
	}
	mocks.threads.Update = func(id int64, update dbThreadUpdate) (*dbThread, error) {
		if want := (dbThreadUpdate{Title: strptr("t1"), ExternalURL: strptr("u1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &dbThread{
			DBThreadCommon: DBThreadCommon{
				ID:           2,
				RepositoryID: 1,
				Title:        "t1",
				ExternalURL:  strptr("u1"),
			},
		}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					threads {
						updateThread(input: { id: "VGhyZWFkOjI=", title: "t1", externalURL: "u1" }) {
							id
							title
							externalURL
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"threads": {
						"updateThread": {
							"id": "VGhyZWFkOjI=",
							"title": "t1",
							"externalURL": "u1"
						}
					}
				}
			`,
		},
	})
}

func TestGraphQL_DeleteThread(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.threads.GetByID = func(id int64) (*dbThread, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbThread{DBThreadCommon: DBThreadCommon{ID: wantID}}, nil
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
					threads {
						deleteThread(thread: "VGhyZWFkOjI=") {
							alwaysNil
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"threads": {
						"deleteThread": null
					}
				}
			`,
		},
	})
}
