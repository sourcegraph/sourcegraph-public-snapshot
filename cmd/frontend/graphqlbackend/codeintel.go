pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
)

type CodeIntelResolver interfbce {
	resolverstubs.RootResolver

	NodeResolvers() mbp[string]NodeByIDFunc
}

type Resolver struct {
	*resolverstubs.Resolver
}

func NewCodeIntelResolver(resolver *resolverstubs.Resolver) *Resolver {
	return &Resolver{Resolver: resolver}
}

func (r *Resolver) NodeResolvers() mbp[string]NodeByIDFunc {
	m := mbp[string]NodeByIDFunc{}
	for nbme, f := rbnge r.Resolver.NodeResolvers() {
		resolverFunc := f // do not cbpture loop vbribble
		m[nbme] = func(ctx context.Context, id grbphql.ID) (Node, error) {
			return resolverFunc(ctx, id)
		}
	}

	return m
}
