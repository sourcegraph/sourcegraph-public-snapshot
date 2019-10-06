package rules

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (GraphQLResolver) CreateRule(ctx context.Context, arg *graphqlbackend.CreateRuleArgs) (graphqlbackend.Rule, error) {
	data := &DBRule{
		Name:        arg.Input.Rule.Name,
		Description: arg.Input.Rule.Description,
		Definition:  string(arg.Input.Rule.Definition),
	}

	var err error
	data.Container, err = dbRuleContainerByID(arg.Input.Container)
	if err != nil {
		return nil, err
	}

	if arg.Input.Rule.Template != nil {
		data.TemplateID = &arg.Input.Rule.Template.Template
		var err error
		data.TemplateContext, err = arg.Input.Rule.Template.ContextJSONCString()
		if err != nil {
			return nil, err
		}
	}

	rule, err := DBRules{}.Create(ctx, data)
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

	data := dbRuleUpdate{
		Name:          arg.Input.Name,
		Description:   arg.Input.Description,
		ClearTemplate: arg.Input.ClearTemplate != nil && *arg.Input.ClearTemplate,
		Definition:    (*string)(arg.Input.Definition),
	}
	if arg.Input.Template != nil {
		data.TemplateID = &arg.Input.Template.Template
		var err error
		data.TemplateContext, err = arg.Input.Template.ContextJSONCString()
		if err != nil {
			return nil, err
		}
	}

	rule, err := DBRules{}.Update(ctx, l.db.ID, data)
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
