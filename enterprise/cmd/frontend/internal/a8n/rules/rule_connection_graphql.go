package rules

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func (GraphQLResolver) RulesDefinedIn(ctx context.Context, projectID graphql.ID, arg *graphqlutil.ConnectionArgs) (graphqlbackend.RuleConnection, error) {
	// Check existence.
	project, err := graphqlbackend.ProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	list, err := dbRules{}.List(ctx, dbRulesListOptions{ProjectID: project.DBID()})
	if err != nil {
		return nil, err
	}
	rules := make([]*gqlRule, len(list))
	for i, a := range list {
		rules[i] = &gqlRule{db: a}
	}
	return &ruleConnection{arg: arg, rules: rules}, nil
}

type ruleConnection struct {
	arg   *graphqlutil.ConnectionArgs
	rules []*gqlRule
}

func (r *ruleConnection) Nodes(ctx context.Context) ([]graphqlbackend.Rule, error) {
	rules := r.rules
	if first := r.arg.First; first != nil && len(rules) > int(*first) {
		rules = rules[:int(*first)]
	}

	rules2 := make([]graphqlbackend.Rule, len(rules))
	for i, l := range rules {
		rules2[i] = l
	}
	return rules2, nil
}

func (r *ruleConnection) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.rules)), nil
}

func (r *ruleConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(r.arg.First != nil && int(*r.arg.First) < len(r.rules)), nil
}
