package issues

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
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestGraphQL_CreateIssue(t *testing.T) {
	internal.ResetMocks()
	const wantRepositoryID = 1
	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1}, nil
	}
	db.Mocks.Repos.Get = func(context.Context, api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: wantRepositoryID}, nil
	}
	wantIssue := &internal.DBThread{
		Type:         internal.DBThreadTypeIssue,
		RepositoryID: wantRepositoryID,
		Title:        "t",
		State:        string(graphqlbackend.ThreadStateOpen),
	}
	internal.Mocks.Threads.Create = func(issue *internal.DBThread) (*internal.DBThread, error) {
		if !reflect.DeepEqual(issue, wantIssue) {
			t.Errorf("got issue %+v, want %+v", issue, wantIssue)
		}
		tmp := *issue
		tmp.ID = 2
		return &tmp, nil
	}
	extsvc.MockImportGitHubThreadEvents = func() error { return nil }
	defer func() { extsvc.MockImportGitHubThreadEvents = nil }()

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					createIssue(input: { repository: "T3JnOjE=", title: "t" }) {
						id
						title
					}
				}
			`,
			ExpectedResult: `
				{
					"createIssue": {
						"id": "SXNzdWU6Mg==",
						"title": "t"
					}
				}
			`,
		},
	})
}

func TestGraphQL_UpdateIssue(t *testing.T) {
	internal.ResetMocks()
	const wantID = 2
	internal.Mocks.Threads.GetByID = func(id int64) (*internal.DBThread, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &internal.DBThread{ID: wantID}, nil
	}
	internal.Mocks.Threads.Update = func(id int64, update internal.DBThreadUpdate) (*internal.DBThread, error) {
		if want := (internal.DBThreadUpdate{Title: strptr("t1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &internal.DBThread{
			Type:         internal.DBThreadTypeIssue,
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
					updateIssue(input: { id: "SXNzdWU6Mg==", title: "t1" }) {
						id
						title
					}
				}
			`,
			ExpectedResult: `
				{
					"updateIssue": {
						"id": "SXNzdWU6Mg==",
						"title": "t1"
					}
				}
			`,
		},
	})
}

func TestGraphQL_DeleteIssue(t *testing.T) {
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
					deleteIssue(issue: "Q2hhbmdlc2V0OjI=") {
						alwaysNil
					}
				}
			`,
			ExpectedResult: `
				{
					"deleteIssue": null
				}
			`,
		},
	})
}

func strptr(s string) *string { return &s }
