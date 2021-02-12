package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestOrganization(t *testing.T) {
	resetMocks()
	database.Mocks.Orgs.GetByName = func(context.Context, string) (*types.Org, error) {
		return &types.Org{ID: 1, Name: "acme"}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
				{
					organization(name: "acme") {
						name
					}
				}
			`,
			ExpectedResult: `
				{
					"organization": {
						"name": "acme"
					}
				}
			`,
		},
	})
}

func TestNode_Org(t *testing.T) {
	resetMocks()
	database.Mocks.Orgs.MockGetByID_Return(t, &types.Org{ID: 1, Name: "acme"}, nil)

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
				{
					node(id: "T3JnOjE=") {
						id
						... on Org {
							name
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"node": {
						"id": "T3JnOjE=",
						"name": "acme"
					}
				}
			`,
		},
	})
}
