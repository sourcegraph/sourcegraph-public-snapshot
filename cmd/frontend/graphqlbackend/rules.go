package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Rules is the implementation of the GraphQL type RulesMutation. If it is not set at runtime, a
// "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var Rules RulesResolver

// RuleByID is called to look up a Rule given its GraphQL ID.
func RuleByID(ctx context.Context, id graphql.ID) (Rule, error) {
	if Rules == nil {
		return nil, errors.New("rules is not implemented")
	}
	return Rules.RuleByID(ctx, id)
}

// RulesDefinedIn returns an instance of the GraphQL RuleConnection type with the list of rules
// defined in a project.
func RulesDefinedIn(ctx context.Context, project graphql.ID, arg *graphqlutil.ConnectionArgs) (RuleConnection, error) {
	if Rules == nil {
		return nil, errors.New("rules is not implemented")
	}
	return Rules.RulesDefinedIn(ctx, project, arg)
}

func (schemaResolver) Rules() (RulesResolver, error) {
	if Rules == nil {
		return nil, errors.New("rules is not implemented")
	}
	return Rules, nil
}

// RulesResolver is the interface for the GraphQL type RulesMutation.
type RulesResolver interface {
	CreateRule(context.Context, *CreateRuleArgs) (Rule, error)
	UpdateRule(context.Context, *UpdateRuleArgs) (Rule, error)
	DeleteRule(context.Context, *DeleteRuleArgs) (*EmptyResponse, error)

	// RuleByID is called by the RuleByID func but is not in the GraphQL API.
	RuleByID(context.Context, graphql.ID) (Rule, error)

	// RulesDefinedIn is called by the RulesDefinedIn func but is not in the GraphQL API.
	RulesDefinedIn(ctx context.Context, project graphql.ID, arg *graphqlutil.ConnectionArgs) (RuleConnection, error)
}

type CreateRuleArgs struct {
	Input struct {
		Project     graphql.ID
		Name        string
		Description *string
		Settings    *string
	}
}

type UpdateRuleArgs struct {
	Input struct {
		ID          graphql.ID
		Name        *string
		Description *string
		Settings    *string
	}
}

type DeleteRuleArgs struct {
	Rule graphql.ID
}

// Rule is the interface for the GraphQL type Rule.
type Rule interface {
	ID() graphql.ID
	Project(context.Context) (Project, error)
	Name() string
	Description() *string
	Settings() string
}

// RuleConnection is the interface for the GraphQL type RuleConnection.
type RuleConnection interface {
	Nodes(context.Context) ([]Rule, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
