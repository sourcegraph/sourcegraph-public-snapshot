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
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestGraphQL_CreateChangeset(t *testing.T) {
	resetMocks()
	const wantRepositoryID = 1
	db.Mocks.Repos.Get = func(context.Context, api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: wantRepositoryID}, nil
	}
	wantChangeset := &dbChangeset{
		RepositoryID: wantRepositoryID,
		Title:        "t",
		ExternalURL:  strptr("u"),
		Status:       graphqlbackend.ChangesetStatusOpen,
		Type:         graphqlbackend.ChangesetTypeChangeset,
	}
	mocks.changesets.Create = func(changeset *dbChangeset) (*dbChangeset, error) {
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
					changesets {
						createChangeset(input: { repository: "T3JnOjE=", title: "t", externalURL: "u", type: THREAD }) {
							id
							title
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"changesets": {
						"createChangeset": {
							"id": "VGhyZWFkOjI=",
							"title": "t"
						}
					}
				}
			`,
		},
	})
}

func TestGraphQL_UpdateChangeset(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.changesets.GetByID = func(id int64) (*dbChangeset, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbChangeset{ID: wantID}, nil
	}
	mocks.changesets.Update = func(id int64, update dbChangesetUpdate) (*dbChangeset, error) {
		if want := (dbChangesetUpdate{Title: strptr("t1"), ExternalURL: strptr("u1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &dbChangeset{
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
					changesets {
						updateChangeset(input: { id: "VGhyZWFkOjI=", title: "t1", externalURL: "u1" }) {
							id
							title
							externalURL
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"changesets": {
						"updateChangeset": {
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

func TestGraphQL_DeleteChangeset(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.changesets.GetByID = func(id int64) (*dbChangeset, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbChangeset{ID: wantID}, nil
	}
	mocks.changesets.DeleteByID = func(id int64) error {
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
					changesets {
						deleteChangeset(changeset: "VGhyZWFkOjI=") {
							alwaysNil
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"changesets": {
						"deleteChangeset": null
					}
				}
			`,
		},
	})
}
