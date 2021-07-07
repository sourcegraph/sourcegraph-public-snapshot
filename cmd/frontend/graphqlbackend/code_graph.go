package graphqlbackend

import (
	"context"
)

// This file just contains stub GraphQL resolvers and data types for the code graph that return an
// error if not running in enterprise mode. The actual resolvers are in
// enterprise/internal/codegraph/resolvers.

type CodeGraphResolver interface {
	UserCodeGraph(context.Context, *UserResolver) (CodeGraphPersonNodeResolver, error)
}

type CodeGraphPersonNodeResolver interface {
	Symbols(context.Context) ([]string, error)

	Dependencies(context.Context) ([]*RepositoryResolver, error)
	Authors(context.Context) ([]*PersonResolver, error)

	Dependents(context.Context) ([]*RepositoryResolver, error)
	Callers(context.Context) ([]*PersonResolver, error)
}

func (r *UserResolver) CodeGraph(ctx context.Context) (CodeGraphPersonNodeResolver, error) {
	return EnterpriseResolvers.codeGraphResolver.UserCodeGraph(ctx, r)
}
