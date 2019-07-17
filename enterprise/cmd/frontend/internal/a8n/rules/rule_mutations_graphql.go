package rules

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (GraphQLResolver) CreateRule(ctx context.Context, arg *graphqlbackend.CreateRuleArgs) (graphqlbackend.Rule, error) {
	project, err := graphqlbackend.ProjectByID(ctx, arg.Input.Project)
	if err != nil {
		return nil, err
	}

	var settings string
	if arg.Input.Settings != nil {
		settings = *arg.Input.Settings
	} else {
		settings = "{}"
	}

	rule, err := dbRules{}.Create(ctx, &dbRule{
		ProjectID:   project.DBID(),
		Name:        arg.Input.Name,
		Description: arg.Input.Description,
		Settings:    settings,
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
	rule, err := dbRules{}.Update(ctx, l.db.ID, dbRuleUpdate{
		Name:        arg.Input.Name,
		Description: arg.Input.Description,
		Settings:    arg.Input.Settings,
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
	return nil, dbRules{}.DeleteByID(ctx, gqlRule.db.ID)
}
