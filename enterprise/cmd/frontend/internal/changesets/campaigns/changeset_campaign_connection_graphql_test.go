package campaigns

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/projects"
)

func TestGraphQL_Project_ChangesetCampaignConnection(t *testing.T) {
	resetMocks()
	const (
		wantProjectID           = 3
		wantChangesetCampaignID = 2
	)
	projects.MockProjectByDBID = func(id int64) (graphqlbackend.Project, error) {
		return projects.TestNewProject(wantProjectID, "", 0, 0), nil
	}
	mocks.campaigns.List = func(dbChangesetCampaignsListOptions) ([]*dbChangesetCampaign, error) {
		return []*dbChangesetCampaign{{ID: wantChangesetCampaignID, Name: "n"}}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				{
					node(id: "UHJvamVjdDoy") {
						... on Project {
							changesetCampaigns {
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
						"changesetCampaigns": {
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
