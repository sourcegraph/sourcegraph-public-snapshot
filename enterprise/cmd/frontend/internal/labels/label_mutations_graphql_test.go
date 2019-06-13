package labels

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/projects"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestGraphQL_CreateLabel(t *testing.T) {
	resetMocks()
	const wantProjectID = 1
	projects.MockProjectByDBID = func(id int64) (graphqlbackend.Project, error) {
		return projects.TestNewProject(wantProjectID, "", 0, 0), nil
	}
	wantLabel := &dbLabel{
		ProjectID:   wantProjectID,
		Name:        "n",
		Description: strptr("d"),
		Color:       "h",
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
				mutation {
					labels {
						createLabel(input: { project: "T3JnOjE=", name: "n", description: "d", color: "h" }) {
							id
							name
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"labels": {
						"createLabel": {
							"id": "TGFiZWw6Mg==",
							"name": "n"
						}
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
			ID:          2,
			ProjectID:   1,
			Name:        "n1",
			Description: strptr("d1"),
			Color:       "h1",
		}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					labels {
						updateLabel(input: { id: "TGFiZWw6Mg==", name: "n1", description: "d1", color: "h1" }) {
							id
							name
							description
							color
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"labels": {
						"updateLabel": {
							"id": "TGFiZWw6Mg==",
							"name": "n1",
							"description": "d1",
							"color": "h1"
						}
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
					labels {
						deleteLabel(label: "TGFiZWw6Mg==") {
							alwaysNil
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"labels": {
						"deleteLabel": null
					}
				}
			`,
		},
	})
}

func TestGraphQL_AddRemoveLabelsToFromLabelable(t *testing.T) {
	resetMocks()
	const wantThreadID = 3
	db.Mocks.DiscussionThreads.Get = func(int64) (*types.DiscussionThread, error) {
		return &types.DiscussionThread{ID: wantThreadID}, nil
	}
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
					labels {
						` + name + `(labelable: "RGlzY3Vzc2lvblRocmVhZDoiMyI=", labels: ["TGFiZWw6Mg=="]) {
							__typename
							... on DiscussionThread {
								id
							}
						}
					}
				}
			`,
					ExpectedResult: `
				{
					"labels": {
						"` + name + `": {
							"__typename": "DiscussionThread",
							"id": "RGlzY3Vzc2lvblRocmVhZDoiMyI="
						}
					}
				}
			`,
				},
			})
		})
	}
}
