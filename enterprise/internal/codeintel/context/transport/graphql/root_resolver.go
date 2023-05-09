package graphql

import (
	"context"
	"fmt"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	svc ContextService
}

func NewRootResolver(observationCtx *observation.Context, svc ContextService) resolverstubs.ContextServiceResolver {
	return &rootResolver{
		svc: svc,
	}
}

func (r *rootResolver) FindMostRelevantSCIPSymbols(ctx context.Context, args *resolverstubs.FindMostRelevantSCIPSymbolsArgs) (string, error) {
	fmt.Println("GetSCIPSymbolArgs: ", args)
	return "foo", nil
}
