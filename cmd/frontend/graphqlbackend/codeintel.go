package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type CodeIntelResolver interface {
	resolverstubs.RootResolver

	NodeResolvers() map[string]NodeByIDFunc
}

type Resolver struct {
	*resolverstubs.Resolver
}

func NewCodeIntelResolver(resolver *resolverstubs.Resolver) *Resolver {
	return &Resolver{Resolver: resolver}
}

func (r *Resolver) NodeResolvers() map[string]NodeByIDFunc {
	m := map[string]NodeByIDFunc{}
	for name, resolverFunc := range r.Resolver.NodeResolvers() {
		m[name] = func(ctx context.Context, id graphql.ID) (Node, error) {
			return resolverFunc(ctx, id)
		}
	}

	return m
}
