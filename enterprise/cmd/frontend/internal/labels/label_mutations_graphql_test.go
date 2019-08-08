package labels

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
)

func TestGraphQL_CreateLabel(t *testing.T) {
	resetMocks()
	const wantRepositoryID = 1
	graphqlbackend.MockRepositoryByID = func(id graphql.ID) (*graphqlbackend.RepositoryResolver, error) {
		return graphqlbackend.NewRepositoryResolver(&types.Repo{ID: wantRepositoryID}), nil
	}
	defer func() { graphqlbackend.MockRepositoryByID = nil }()
	wantLabel := &dbLabel{
		RepositoryID: wantRepositoryID,
		Name:         "n",
		Description:  strptr("d"),
		Color:        "h",
	}
	mocks.labels.Create = func(label *dbLabel) (*dbLabel, error) {
		if !reflect.DeepEqual(label, wantLabel) {
			t.Errorf("got label %+v, want %+v", label, wantLabel)
		}
		tmp := *label
		tmp.ID = 2
		return &tmp, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation($repository: ID!) {
					createLabel(input: { repository: $repository, name: "n", description: "d", color: "h" }) {
						id
						name
					}
				}
			`,
			Variables: map[string]interface{}{"repository": string(graphqlbackend.MarshalRepositoryID(wantRepositoryID))},
			ExpectedResult: `
				{
					"createLabel": {
						"id": "TGFiZWw6Mg==",
						"name": "n"
					}
				}
			`,
		},
	})
}

func TestGraphQL_UpdateLabel(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.labels.GetByID = func(id int64) (*dbLabel, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbLabel{ID: wantID}, nil
	}
	mocks.labels.Update = func(id int64, update dbLabelUpdate) (*dbLabel, error) {
		if want := (dbLabelUpdate{Name: strptr("n1"), Description: strptr("d1"), Color: strptr("h1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &dbLabel{
			ID:           2,
			RepositoryID: 1,
			Name:         "n1",
			Description:  strptr("d1"),
			Color:        "h1",
		}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					updateLabel(input: { id: "TGFiZWw6Mg==", name: "n1", description: "d1", color: "h1" }) {
						id
						name
						description
						color
					}
				}
			`,
			ExpectedResult: `
				{
					"updateLabel": {
						"id": "TGFiZWw6Mg==",
						"name": "n1",
						"description": "d1",
						"color": "h1"
					}
				}
			`,
		},
	})
}

func TestGraphQL_DeleteLabel(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.labels.GetByID = func(id int64) (*dbLabel, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbLabel{ID: wantID}, nil
	}
	mocks.labels.DeleteByID = func(id int64) error {
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
					deleteLabel(label: "TGFiZWw6Mg==") {
						alwaysNil
					}
				}
			`,
			ExpectedResult: `
				{
					"deleteLabel": null
				}
			`,
		},
	})
}

func TestGraphQL_AddRemoveLabelsToFromLabelable(t *testing.T) {
	resetMocks()
	const wantThreadID = 3
	threads.MockThreadByID = func(id graphql.ID) (graphqlbackend.Thread, error) {
		return mockThread{id: wantThreadID}, nil
	}
	defer func() { threads.MockThreadByID = nil }()
	const wantID = 2
	mocks.labels.GetByID = func(id int64) (*dbLabel, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbLabel{ID: wantID}, nil
	}
	addRemoveLabelsToFromLabelable := func(thread int64, labels []int64) error {
		if thread != wantThreadID {
			t.Errorf("got %d, want %d", thread, wantThreadID)
		}
		if want := []int64{wantID}; !reflect.DeepEqual(labels, want) {
			t.Errorf("got %v, want %v", labels, want)
		}
		return nil
	}

	tests := map[string]*func(thread int64, labels []int64) error{
		"addLabelsToLabelable":      &mocks.labelsObjects.AddLabelsToThread,
		"removeLabelsFromLabelable": &mocks.labelsObjects.RemoveLabelsFromThread,
	}
	for name, mockFn := range tests {
		t.Run(name, func(t *testing.T) {
			*mockFn = addRemoveLabelsToFromLabelable
			defer func() { *mockFn = nil }()

			gqltesting.RunTests(t, []*gqltesting.Test{
				{
					Context: backend.WithAuthzBypass(context.Background()),
					Schema:  graphqlbackend.GraphQLSchema,
					Query: `
				mutation {
					` + name + `(labelable: "VGhyZWFkOjM=", labels: ["TGFiZWw6Mg=="]) {
						__typename
						... on Thread {
							id
						}
					}
				}
			`,
					ExpectedResult: `
				{
					"` + name + `": {
						"__typename": "Thread",
						"id": "VGhyZWFkOjM="
					}
				}
			`,
				},
			})
		})
	}
}
