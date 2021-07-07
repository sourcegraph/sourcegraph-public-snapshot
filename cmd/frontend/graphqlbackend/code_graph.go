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
	Dependencies() []string
	Dependents() []string
}

func (r *UserResolver) CodeGraph(ctx context.Context) (CodeGraphPersonNodeResolver, error) {
	return EnterpriseResolvers.codeGraphResolver.UserCodeGraph(ctx, r)
}
