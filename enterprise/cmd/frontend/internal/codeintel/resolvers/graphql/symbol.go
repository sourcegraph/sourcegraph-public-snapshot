package graphql

import (
	"context"
	"path"
	"strings"

	"github.com/sourcegraph/go-lsp"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
)

type newQueryResolver func(ctx context.Context, path string) (*QueryResolver, error)

type SymbolResolver struct {
	symbol resolvers.AdjustedSymbol
	root   *resolvers.AdjustedSymbol

	locationResolver *CachedLocationResolver
	newQueryResolver newQueryResolver
}

func NewSymbolResolver(symbol resolvers.AdjustedSymbol, root *resolvers.AdjustedSymbol, locationResolver *CachedLocationResolver, newQueryResolver newQueryResolver) gql.SymbolResolver {
	return &SymbolResolver{
		symbol:           symbol,
		root:             root,
		locationResolver: locationResolver,
		newQueryResolver: newQueryResolver,
	}
}

func (r *SymbolResolver) Text() string {
	return r.symbol.Text
}

func (r *SymbolResolver) Detail() *string {
	if v := r.symbol.Detail; v != "" {
		return &v
	}
	return nil
}

func (r *SymbolResolver) Kind() string /* enum SymbolKind */ {
	return strings.ToUpper(lsp.SymbolKind(r.symbol.Kind).String())
}

func (r *SymbolResolver) Tags() []string /* enum SymbolTag */ {
	tags := make([]string, len(r.symbol.Tags))
	for i, tag := range r.symbol.Tags {
		tags[i] = strings.ToUpper(tag.String())
	}
	return tags
}

func (r *SymbolResolver) Monikers() []gql.MonikerResolver {
	var monikers []gql.MonikerResolver
	for _, m := range r.symbol.Monikers {
		monikers = append(monikers, NewMonikerResolver(m))
	}
	return monikers
}

func (r *SymbolResolver) Definitions(ctx context.Context) (gql.LocationConnectionResolver, error) {
	return r.definitions(ctx)
}

func (r *SymbolResolver) DefinitionsFullRanges(ctx context.Context) (gql.LocationConnectionResolver, error) {
	// TODO(sqs): use & adjust full range
	return r.definitions(ctx)
}

func (r *SymbolResolver) definitions(ctx context.Context) (gql.LocationConnectionResolver, error) {
	// TODO(sqs): workspace symbols have locations, but document symbols dont (currently based on
	// how this is all implemented).
	adjustedLocations := []resolvers.AdjustedLocation{r.symbol.AdjustedLocation}
	return NewLocationConnectionResolver(adjustedLocations, nil, r.locationResolver), nil
}

func (r *SymbolResolver) References(ctx context.Context) (gql.LocationConnectionResolver, error) {
	// if len(r.symbol.Locations) == 0 {
	// 	// TODO(sqs): instead, look up by moniker
	// 	return NewLocationConnectionResolver(nil, nil, nil), nil
	// }
	queryResolver, err := r.newQueryResolver(ctx, path.Clean(r.symbol.AdjustedLocation.Path))
	if err != nil {
		return nil, err
	}
	return queryResolver.References(ctx, &gql.LSIFPagedQueryPositionArgs{
		LSIFQueryPositionArgs: gql.LSIFQueryPositionArgs{
			Line:      int32(r.symbol.AdjustedLocation.AdjustedRange.Start.Line),
			Character: int32(r.symbol.AdjustedLocation.AdjustedRange.Start.Character),
		},
	})
}

func (r *SymbolResolver) Hover(ctx context.Context) (gql.HoverResolver, error) {
	// if len(r.symbol.Locations) == 0 {
	// 	// TODO(sqs): instead, look up by moniker
	// 	return NewLocationConnectionResolver(nil, nil, nil), nil
	// }
	queryResolver, err := r.newQueryResolver(ctx, path.Clean(r.symbol.AdjustedLocation.Path))
	if err != nil {
		return nil, err
	}
	return queryResolver.Hover(ctx, &gql.LSIFQueryPositionArgs{
		Line:      int32(r.symbol.AdjustedLocation.AdjustedRange.Start.Line),
		Character: int32(r.symbol.AdjustedLocation.AdjustedRange.Start.Character),
	})
}

func (r *SymbolResolver) RootAncestor() gql.SymbolResolver {
	if r.root == nil {
		return r
	}
	return NewSymbolResolver(*r.root, nil, r.locationResolver, r.newQueryResolver)
}

func (r *SymbolResolver) Children() []gql.SymbolResolver {
	children := make([]gql.SymbolResolver, len(r.symbol.Children))
	for i, childSymbol := range r.symbol.Children {
		children[i] = NewSymbolResolver(childSymbol, r.root, r.locationResolver, r.newQueryResolver)
	}
	return children
}

func (r *SymbolResolver) Location() (path string, line, end int) {
	// if len(r.symbol.Locations) == 0 {
	// 	return "", 0, 0
	// }
	//
	// TODO(sqs): use full range
	return r.symbol.AdjustedLocation.Path, r.symbol.AdjustedLocation.AdjustedRange.Start.Line, r.symbol.AdjustedLocation.AdjustedRange.End.Line
}
