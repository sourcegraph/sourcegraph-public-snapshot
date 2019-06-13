package projects

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestGraphQL_CreateProject(t *testing.T) {
	resetMocks()
	const wantOrgID = 1
	db.Mocks.Orgs.GetByID = func(context.Context, int32) (*types.Org, error) {
		return &types.Org{ID: wantOrgID}, nil
	}
	wantProject := &dbProject{
		NamespaceOrgID: wantOrgID,
		Name:           "n",
	}
	mocks.projects.Create = func(project *dbProject) (*dbProject, error) {
		if !reflect.DeepEqual(project, wantProject) {
			t.Errorf("got project %+v, want %+v", project, wantProject)
		}
		tmp := *project
		tmp.ID = 2
		return &tmp, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					projects {
						createProject(input: { namespace: "T3JnOjE=", name: "n" }) {
							id
							name
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"projects": {
						"createProject": {
							"id": "UHJvamVjdDoy",
							"name": "n"
						}
					}
				}
			`,
		},
	})
}

func TestGraphQL_UpdateProject(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.projects.GetByID = func(id int64) (*dbProject, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbProject{ID: wantID}, nil
	}
	mocks.projects.Update = func(id int64, update dbProjectUpdate) (*dbProject, error) {
		if want := (dbProjectUpdate{Name: strptr("n1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &dbProject{
			ID:             2,
			NamespaceOrgID: 1,
			Name:           "n1",
		}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					projects {
						updateProject(input: { id: "TGFiZWw6Mg==", name: "n1" }) {
							id
							name
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"projects": {
						"updateProject": {
							"id": "UHJvamVjdDoy",
							"name": "n1"
						}
					}
				}
			`,
		},
	})
}

func TestGraphQL_DeleteProject(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.projects.GetByID = func(id int64) (*dbProject, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbProject{ID: wantID}, nil
	}
	mocks.projects.DeleteByID = func(id int64) error {
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
					projects {
						deleteProject(project: "UHJvamVjdDoy") {
							alwaysNil
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"projects": {
						"deleteProject": null
					}
				}
			`,
		},
	})
}
