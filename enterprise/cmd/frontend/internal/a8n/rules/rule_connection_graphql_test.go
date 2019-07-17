package rules

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/projects"
)

func TestGraphQL_Project_RuleConnection(t *testing.T) {
	resetMocks()
	const (
		wantProjectID = 3
		wantRuleID    = 2
	)
	projects.MockProjectByDBID = func(id int64) (graphqlbackend.Project, error) {
		return projects.TestNewProject(wantProjectID, "", 0, 0), nil
	}
	mocks.rules.List = func(dbRulesListOptions) ([]*dbRule, error) {
		return []*dbRule{{ID: wantRuleID, Name: "n"}}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				{
					node(id: "UHJvamVjdDoz") {
						... on Project {
							rules {
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
						"rules": {
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
