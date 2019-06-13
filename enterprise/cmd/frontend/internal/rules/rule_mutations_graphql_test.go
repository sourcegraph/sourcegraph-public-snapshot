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
	ResetMocks()
	const wantContainerCampaignID = 1
	wantRule := &DBRule{
		Container:   RuleContainer{Campaign: wantContainerCampaignID},
		Name:        "n",
		Description: strptr("d"),
		Definition:  "[1]",
	}
	Mocks.Rules.Create = func(rule *DBRule) (*DBRule, error) {
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
					createRule(input: { container: $container, rule: { name: "n", description: "d", definition: "[1]" } }) {
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
	ResetMocks()
	const wantID = 2
	Mocks.Rules.GetByID = func(id int64) (*DBRule, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &DBRule{ID: wantID}, nil
	}
	Mocks.Rules.Update = func(id int64, update dbRuleUpdate) (*DBRule, error) {
		if want := (dbRuleUpdate{Name: strptr("n1"), Description: strptr("d1"), Definition: strptr("h1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &DBRule{
			ID:          2,
			Name:        "n1",
			Description: strptr("d1"),
			Definition:  "true",
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
							"raw": "true"
						}
					}
				}
			`,
		},
	})
}

func TestGraphQL_DeleteRule(t *testing.T) {
	ResetMocks()
	const wantID = 2
	Mocks.Rules.GetByID = func(id int64) (*DBRule, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &DBRule{ID: wantID}, nil
	}
	Mocks.Rules.DeleteByID = func(id int64) error {
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

func strptr(s string) *string { return &s }
