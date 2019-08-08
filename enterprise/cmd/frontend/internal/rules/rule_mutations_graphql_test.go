package rules

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func TestGraphQL_CreateRule(t *testing.T) {
	resetMocks()
	const wantContainerCampaignID = 1
	wantRule := &dbRule{
		Container:   ruleContainer{Campaign: wantContainerCampaignID},
		Name:        "n",
		Description: strptr("d"),
		Definition:  "h",
	}
	mocks.rules.Create = func(rule *dbRule) (*dbRule, error) {
		if !reflect.DeepEqual(rule, wantRule) {
			t.Errorf("got rule %+v, want %+v", rule, wantRule)
		}
		tmp := *rule
		tmp.ID = 2
		return &tmp, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation($container: ID!) {
					createRule(input: { container: $container, name: "n", description: "d", definition: "h" }) {
						id
						name
					}
				}
			`,
			Variables: map[string]interface{}{
				"container": string(graphqlbackend.MarshalCampaignID(wantContainerCampaignID)),
			},
			ExpectedResult: `
				{
					"createRule": {
						"id": "UnVsZToy",
						"name": "n"
					}
				}
			`,
		},
	})
}

func TestGraphQL_UpdateRule(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.rules.GetByID = func(id int64) (*dbRule, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbRule{ID: wantID}, nil
	}
	mocks.rules.Update = func(id int64, update dbRuleUpdate) (*dbRule, error) {
		if want := (dbRuleUpdate{Name: strptr("n1"), Description: strptr("d1"), Definition: strptr("h1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &dbRule{
			ID:          2,
			Name:        "n1",
			Description: strptr("d1"),
			Definition:  "h1",
		}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					updateRule(input: { id: "UnVsZToy", name: "n1", description: "d1", definition: "h1" }) {
						id
						name
						description
						definition {
							raw
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"updateRule": {
						"id": "UnVsZToy",
						"name": "n1",
						"description": "d1",
						"definition": {
							"raw": "h1"
						}
					}
				}
			`,
		},
	})
}

func TestGraphQL_DeleteRule(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.rules.GetByID = func(id int64) (*dbRule, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbRule{ID: wantID}, nil
	}
	mocks.rules.DeleteByID = func(id int64) error {
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
					deleteRule(rule: "UnVsZToy") {
						alwaysNil
					}
				}
			`,
			ExpectedResult: `
				{
					"deleteRule": null
				}
			`,
		},
	})
}
