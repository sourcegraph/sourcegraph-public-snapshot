package graphql

import (
	"context"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type ContextService interface {
	FindMostRelevantSCIPSymbols(ctx context.Context, args *resolverstubs.FindMostRelevantSCIPSymbolsArgs) (string, error)
}
