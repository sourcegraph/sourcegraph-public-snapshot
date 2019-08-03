package changesets

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

func TestGraphQL_CreateChangeset(t *testing.T) {
	internal.ResetMocks()
	const wantRepositoryID = 1
	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1}, nil
	}
	db.Mocks.Repos.Get = func(context.Context, api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: wantRepositoryID}, nil
	}
	wantChangeset := &internal.DBThread{
		Type:         graphqlbackend.ThreadlikeTypeChangeset,
		RepositoryID: wantRepositoryID,
		Title:        "t",
		ExternalURL:  strptr("u"),
		Status:       string(graphqlbackend.ChangesetStatusOpen),
		BaseRef:      "b",
		HeadRef:      "h",
	}
	internal.Mocks.Threads.Create = func(changeset *internal.DBThread) (*internal.DBThread, error) {
		if !reflect.DeepEqual(changeset, wantChangeset) {
			t.Errorf("got changeset %+v, want %+v", changeset, wantChangeset)
		}
		tmp := *changeset
		tmp.ID = 2
		return &tmp, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					createChangeset(input: { repository: "T3JnOjE=", title: "t", externalURL: "u", baseRef: "b", headRef: "h" }) {
						id
						title
					}
				}
			`,
			ExpectedResult: `
				{
					"createChangeset": {
						"id": "Q2hhbmdlc2V0OjI=",
						"title": "t"
					}
				}
			`,
		},
	})
}

func TestGraphQL_UpdateChangeset(t *testing.T) {
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
			Type:         graphqlbackend.ThreadlikeTypeChangeset,
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
					updateChangeset(input: { id: "Q2hhbmdlc2V0OjI=", title: "t1", externalURL: "u1" }) {
						id
						title
						externalURL
					}
				}
			`,
			ExpectedResult: `
				{
					"updateChangeset": {
						"id": "Q2hhbmdlc2V0OjI=",
						"title": "t1",
						"externalURL": "u1"
					}
				}
			`,
		},
	})
}

func TestGraphQL_DeleteChangeset(t *testing.T) {
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
					deleteChangeset(changeset: "Q2hhbmdlc2V0OjI=") {
						alwaysNil
					}
				}
			`,
			ExpectedResult: `
				{
					"deleteChangeset": null
				}
			`,
		},
	})
}

func strptr(s string) *string { return &s }
