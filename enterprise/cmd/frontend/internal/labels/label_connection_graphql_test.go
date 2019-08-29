package labels

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
)

func TestGraphQL_Labelable_LabelConnection(t *testing.T) {
	resetMocks()
	const (
		wantThreadID = 3
		wantLabelID  = 2
	)
	graphqlbackend.Threads = threads.GraphQLResolver{}
	defer func() { graphqlbackend.Threads = nil }()
	threads.MockThreadByID = func(id graphql.ID) (graphqlbackend.Thread, error) {
		return mockThread{id: wantThreadID}, nil
	}
	defer func() { threads.MockThreadByID = nil }()
	mocks.labelsObjects.List = func(dbLabelsObjectsListOptions) ([]*dbObjectLabel, error) {
		return []*dbObjectLabel{{Thread: wantThreadID, Label: wantLabelID}}, nil
	}
	mocks.labels.GetByID = func(id int64) (*dbLabel, error) {
		if id != wantLabelID {
			t.Errorf("got %d, want %d", id, wantLabelID)
		}
		return &dbLabel{Name: "n"}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				{
					node(id: "VGhyZWFkOjM=") {
						... on Thread {
							labels {
								nodes {
									name
								}
								totalCount
								pageInfo {
									hasNextPage
								}
							}
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"node": {
						"labels": {
							"nodes": [
								{
									"name": "n"
								}
							],
							"totalCount": 1,
							"pageInfo": {
								"hasNextPage": false
							}
						}
					}
				}
			`,
		},
	})
}

func TestGraphQL_Repository_LabelConnection(t *testing.T) {
	resetMocks()
	const (
		wantRepositoryID = 3
		wantLabelID      = 2
	)
	graphqlbackend.MockRepositoryByID = func(id graphql.ID) (*graphqlbackend.RepositoryResolver, error) {
		return graphqlbackend.NewRepositoryResolver(&types.Repo{ID: wantRepositoryID}), nil
	}
	defer func() { graphqlbackend.MockRepositoryByID = nil }()
	mocks.labels.List = func(dbLabelsListOptions) ([]*dbLabel, error) {
		return []*dbLabel{{ID: wantLabelID, Name: "n"}}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				{
					node(id: "UmVwb3NpdG9yeToz") {
						... on Repository {
							labels {
								nodes {
									name
								}
								totalCount
								pageInfo {
									hasNextPage
								}
							}
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"node": {
						"labels": {
							"nodes": [
								{
									"name": "n"
								}
							],
							"totalCount": 1,
							"pageInfo": {
								"hasNextPage": false
							}
						}
					}
				}
			`,
		},
	})
}
