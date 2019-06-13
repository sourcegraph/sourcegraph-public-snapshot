package rules_test

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/rules"
)

func TestGraphQL_RuleContainer_RuleConnection(t *testing.T) {
	rules.ResetMocks()
	const (
		wantContainerCampaignID = 1
		wantRuleID              = 2
	)
	campaigns.MockCampaignByID = func(graphql.ID) (graphqlbackend.Campaign, error) {
		return mockCampaign{id: wantContainerCampaignID}, nil
	}
	defer func() { campaigns.MockCampaignByID = nil }()
	rules.Mocks.Rules.List = func(rules.DBRulesListOptions) ([]*rules.DBRule, error) {
		return []*rules.DBRule{{ID: wantRuleID, Name: "n"}}, nil
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
