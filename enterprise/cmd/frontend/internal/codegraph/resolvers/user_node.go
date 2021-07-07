package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *Resolver) UserCodeGraph(ctx context.Context, user *graphqlbackend.UserResolver) (graphqlbackend.CodeGraphPersonNodeResolver, error) {
	return &CodeGraphPersonNodeResolver{
		user: user,
	}, nil
}

type CodeGraphPersonNodeResolver struct {
	user *graphqlbackend.UserResolver

	userSymbolsResolver
}

func (CodeGraphPersonNodeResolver) Dependencies() []string {
	return []string{"mydependency1", "mydependency2", "mydependency3"}
}

func (CodeGraphPersonNodeResolver) Dependents() []string {
	return []string{"mydependent1", "mydependent2"}
}
