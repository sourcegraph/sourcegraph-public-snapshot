package graphqlbackend

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Rules is the implementation of the GraphQL type RulesMutation. If it is not set at runtime, a
// "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var Rules RulesResolver

const GQLTypeRule = "Rule"

func MarshalRuleID(id int64) graphql.ID {
	return relay.MarshalID(GQLTypeRule, id)
}

func UnmarshalRuleID(id graphql.ID) (dbID int64, err error) {
	if typ := relay.UnmarshalKind(id); typ != GQLTypeRule {
		return 0, fmt.Errorf("rule ID has unexpected type type %q", typ)
	}
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

var errRulesNotImplemented = errors.New("rules is not implemented")

// RuleByID is called to look up a Rule given its GraphQL ID.
func RuleByID(ctx context.Context, id graphql.ID) (Rule, error) {
	if Rules == nil {
		return nil, errRulesNotImplemented
	}
	return Rules.RuleByID(ctx, id)
}

// RulesInRuleContainer returns an instance of the GraphQL RuleConnection type with the list of
// rules defined in a rule container.
func RulesInRuleContainer(ctx context.Context, ruleContainer graphql.ID, arg *graphqlutil.ConnectionArgs) (RuleConnection, error) {
	if Rules == nil {
		return nil, errRulesNotImplemented
	}
	return Rules.RulesInRuleContainer(ctx, ruleContainer, arg)
}

func (schemaResolver) RuleContainer(ctx context.Context, arg *struct{ ID graphql.ID }) (*ToRuleContainer, error) {
	if Rules == nil {
		return nil, errRulesNotImplemented
	}
	return Rules.RuleContainerByID(ctx, arg.ID)
}

func (schemaResolver) CreateRule(ctx context.Context, arg *CreateRuleArgs) (Rule, error) {
	if Rules == nil {
		return nil, errRulesNotImplemented
	}
	return Rules.CreateRule(ctx, arg)
}

func (schemaResolver) UpdateRule(ctx context.Context, arg *UpdateRuleArgs) (Rule, error) {
	if Rules == nil {
		return nil, errRulesNotImplemented
	}
	return Rules.UpdateRule(ctx, arg)
}

func (schemaResolver) DeleteRule(ctx context.Context, arg *DeleteRuleArgs) (*EmptyResponse, error) {
	if Rules == nil {
		return nil, errRulesNotImplemented
	}
	return Rules.DeleteRule(ctx, arg)
}

// RulesResolver is the interface for the rules GraphQL API.
type RulesResolver interface {
	CreateRule(context.Context, *CreateRuleArgs) (Rule, error)
	UpdateRule(context.Context, *UpdateRuleArgs) (Rule, error)
	DeleteRule(context.Context, *DeleteRuleArgs) (*EmptyResponse, error)

	// RuleByID is called by the RuleByID func but is not in the GraphQL API.
	RuleByID(context.Context, graphql.ID) (Rule, error)

	RuleContainerByID(context.Context, graphql.ID) (*ToRuleContainer, error)

	// RulesInRuleContainer is called by the RulesInRuleContainer func but is not in the GraphQL
	// API.
	RulesInRuleContainer(ctx context.Context, container graphql.ID, arg *graphqlutil.ConnectionArgs) (RuleConnection, error)
}

type NewRuleInput struct {
	Name        string
	Description *string
	Definition  JSONCString
}

type CreateRuleArgs struct {
	Input struct {
		Container graphql.ID
		Rule      NewRuleInput
	}
}

type UpdateRuleArgs struct {
	Input struct {
		ID          graphql.ID
		Name        *string
		Description *string
		Definition  *JSONCString
	}
}

type DeleteRuleArgs struct {
	Rule graphql.ID
}

// Rule is the interface for the GraphQL type Rule.
type Rule interface {
	ID() graphql.ID
	Container(context.Context) (*ToRuleContainer, error)
	Name() string
	Description() *string
	Definition() JSONC
	CreatedAt() DateTime
	UpdatedAt() DateTime
	URL(context.Context) (string, error)
	Updatable
}

// RuleConnection is the interface for the GraphQL type RuleConnection.
type RuleConnection interface {
	Nodes(context.Context) ([]Rule, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}

type ruleContainer interface {
	Rules(context.Context, *graphqlutil.ConnectionArgs) (RuleConnection, error)
}

type ToRuleContainer struct {
	Campaign Campaign
	Thread   Thread
}

func (v ToRuleContainer) RuleContainer() ruleContainer {
	switch {
	case v.Campaign != nil:
		return v.Campaign
	case v.Thread != nil:
		return v.Thread
	default:
		panic("invalid RuleContainer")
	}
}

func (v ToRuleContainer) Rules(ctx context.Context, arg *graphqlutil.ConnectionArgs) (RuleConnection, error) {
	return v.RuleContainer().Rules(ctx, arg)
}

func (v ToRuleContainer) ToCampaign() (Campaign, bool) { return v.Campaign, v.Campaign != nil }
func (v ToRuleContainer) ToThread() (Thread, bool)     { return v.Thread, v.Thread != nil }
