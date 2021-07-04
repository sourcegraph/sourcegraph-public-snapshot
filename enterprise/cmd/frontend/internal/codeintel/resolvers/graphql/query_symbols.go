package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
)

// func (r *QueryResolver) Symbols(ctx context.Context, args *gql.LSIFSymbolsArgs) (gql.SymbolConnectionResolver, error) {
// 	limit := derefInt32(args.First, DefaultReferencesPageSize*10 /* TODO(sqs) */)
// 	if limit <= 0 {
// 		return nil, ErrIllegalLimit
// 	}

// 	symbols, totalCount, err := r.resolver.Symbols(ctx, args.Filters, limit)
// 	if err != nil {
// 		return nil, err
// 	}

// 	newQueryResolver := func(ctx context.Context, path string) (*QueryResolver, error) {
// 		type tmpWithPath interface {
// 			TmpWithPath(path string) resolvers.QueryResolver
// 		}
// 		// TODO(sqs): hacky
// 		return NewQueryResolver(r.resolver.(tmpWithPath).TmpWithPath(path), r.locationResolver).(*QueryResolver), nil
// 	}
// 	return NewSymbolConnectionResolver(symbols, nil, totalCount, r.locationResolver, newQueryResolver), nil
// }

func (r *QueryResolver) Symbol(ctx context.Context, args *gql.LSIFSymbolArgs) (gql.SymbolResolver, error) {
	rootSymbol, treePath, err := r.resolver.Symbol(ctx, args.Moniker.Scheme, args.Moniker.Identifier)
	if rootSymbol == nil || err != nil {
		return nil, err
	}

	var symbol resolvers.AdjustedSymbol
	if len(treePath) == 0 || true /* TODO(sqs): always returns rootSymbol for now */ {
		symbol = *rootSymbol
		rootSymbol = nil
	} else {
		symbol = rootSymbol.Descendant(treePath)
	}

	newQueryResolver := func(ctx context.Context, path string) (*QueryResolver, error) {
		type tmpWithPath interface {
			TmpWithPath(path string) resolvers.QueryResolver
		}
		// TODO(sqs): hacky
		return NewQueryResolver(r.resolver.(tmpWithPath).TmpWithPath(path), r.locationResolver).(*QueryResolver), nil
	}
	return NewSymbolResolver(symbol, rootSymbol, r.locationResolver, newQueryResolver), nil
}
