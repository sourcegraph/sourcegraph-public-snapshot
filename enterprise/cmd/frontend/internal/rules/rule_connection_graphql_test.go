package rules

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/campaigns"
)

func TestGraphQL_RuleContainer_RuleConnection(t *testing.T) {
	resetMocks()
	const (
		wantContainerCampaignID = 1
		wantRuleID              = 2
	)
	campaigns.MockCampaignByID = func(graphql.ID) (graphqlbackend.Campaign, error) {
		return mockCampaign{id: wantContainerCampaignID}, nil
	}
	defer func() { campaigns.MockCampaignByID = nil }()
	mocks.rules.List = func(dbRulesListOptions) ([]*dbRule, error) {
		return []*dbRule{{ID: wantRuleID, Name: "n"}}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				query($container: ID!) {
					node(id: $container) {
						... on Campaign {
							rules {
								nodes {
									name
								}
							}
						}
					}
				}
			`,
			Variables: map[string]interface{}{
				"container": string(graphqlbackend.MarshalCampaignID(wantContainerCampaignID)),
			},
			ExpectedResult: `
				{
					"node": {
						"rules": {
							"nodes": [
								{
									"name": "n"
								}
							]
						}
					}
				}
			`,
		},
	})
}
