package rules

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (GraphQLResolver) CreateRule(ctx context.Context, arg *graphqlbackend.CreateRuleArgs) (graphqlbackend.Rule, error) {
	dbContainer, err := dbRuleContainerByID(arg.Input.Container)
	if err != nil {
		return nil, err
	}

	rule, err := DBRules{}.Create(ctx, &DBRule{
		Container:   dbContainer,
		Name:        arg.Input.Rule.Name,
		Description: arg.Input.Rule.Description,
		Definition:  string(arg.Input.Rule.Definition),
	})
	if err != nil {
		return nil, err
	}
	return &gqlRule{db: rule}, nil
}

func (GraphQLResolver) UpdateRule(ctx context.Context, arg *graphqlbackend.UpdateRuleArgs) (graphqlbackend.Rule, error) {
	l, err := ruleByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	rule, err := DBRules{}.Update(ctx, l.db.ID, dbRuleUpdate{
		Name:        arg.Input.Name,
		Description: arg.Input.Description,
		Definition:  (*string)(arg.Input.Definition),
	})
	if err != nil {
		return nil, err
	}
	return &gqlRule{db: rule}, nil
}

func (GraphQLResolver) DeleteRule(ctx context.Context, arg *graphqlbackend.DeleteRuleArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlRule, err := ruleByID(ctx, arg.Rule)
	if err != nil {
		return nil, err
	}
	return nil, DBRules{}.DeleteByID(ctx, gqlRule.db.ID)
}
