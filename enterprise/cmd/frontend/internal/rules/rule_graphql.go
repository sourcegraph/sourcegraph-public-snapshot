package rules

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// gqlRule implements the GraphQL type Rule.
type gqlRule struct{ db *dbRule }

// ruleByID looks up and returns the Rule with the given GraphQL ID. If no such Rule exists, it
// returns a non-nil error.
func ruleByID(ctx context.Context, id graphql.ID) (*gqlRule, error) {
	dbID, err := graphqlbackend.UnmarshalRuleID(id)
	if err != nil {
		return nil, err
	}
	return ruleByDBID(ctx, dbID)
}

func (GraphQLResolver) RuleByID(ctx context.Context, id graphql.ID) (graphqlbackend.Rule, error) {
	return ruleByID(ctx, id)
}

// ruleByDBID looks up and returns the Rule with the given database ID. If no such Rule exists,
// it returns a non-nil error.
func ruleByDBID(ctx context.Context, dbID int64) (*gqlRule, error) {
	v, err := dbRules{}.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return &gqlRule{db: v}, nil
}

func (v *gqlRule) ID() graphql.ID {
	return graphqlbackend.MarshalRuleID(v.db.ID)
}

func (v *gqlRule) Container(ctx context.Context) (*graphqlbackend.ToRuleContainer, error) {
	return GraphQLResolver{}.RuleContainerByID(ctx, v.db.Container.graphqlID())
}

func (v *gqlRule) Name() string { return v.db.Name }

func (v *gqlRule) Description() *string { return v.db.Description }

func (v *gqlRule) Definition() string { return v.db.Definition }
