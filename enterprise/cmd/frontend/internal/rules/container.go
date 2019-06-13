package rules

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type RuleContainer struct {
	Campaign int64
	Thread   int64
}

func (c RuleContainer) graphqlID() graphql.ID {
	switch {
	case c.Campaign != 0:
		return graphqlbackend.MarshalCampaignID(c.Campaign)
	case c.Thread != 0:
		return graphqlbackend.MarshalThreadID(c.Thread)
	default:
		panic("invalid RuleContainer")
	}
}

func dbRuleContainerByID(ruleContainer graphql.ID) (c RuleContainer, err error) {
	// TODO!(sqs) THIS MUST actually try to fetch the object so that we check and enforce perms SECURITY
	switch relay.UnmarshalKind(ruleContainer) {
	case graphqlbackend.GQLTypeCampaign:
		c.Campaign, err = graphqlbackend.UnmarshalCampaignID(ruleContainer)
	case graphqlbackend.GQLTypeThread:
		c.Thread, err = graphqlbackend.UnmarshalThreadID(ruleContainer)
	default:
		err = fmt.Errorf("node %q is not a RuleContainer", ruleContainer)
	}
	return
}

func (GraphQLResolver) RuleContainerByID(ctx context.Context, id graphql.ID) (*graphqlbackend.ToRuleContainer, error) {
	node, err := graphqlbackend.NodeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var c graphqlbackend.ToRuleContainer
	switch relay.UnmarshalKind(id) {
	case graphqlbackend.GQLTypeCampaign:
		c.Campaign = node.(graphqlbackend.Campaign)
	case graphqlbackend.GQLTypeThread:
		c.Thread = node.(graphqlbackend.Thread)
	default:
		return nil, fmt.Errorf("node %q is not a RuleContainer", id)
	}
	return &c, nil
}
