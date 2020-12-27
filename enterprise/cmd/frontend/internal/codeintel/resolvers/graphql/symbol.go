package graphql

import (
	"context"
	"path"
	"strings"

	"github.com/sourcegraph/go-lsp"
	protocol "github.com/sourcegraph/lsif-protocol"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
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
	return r.definitions(ctx, func(loc protocol.SymbolLocation) protocol.RangeData {
		if loc.Range != nil {
			return *loc.Range
		}
		return loc.FullRange
	})
}

func (r *SymbolResolver) DefinitionsFullRanges(ctx context.Context) (gql.LocationConnectionResolver, error) {
	return r.definitions(ctx, func(loc protocol.SymbolLocation) protocol.RangeData {
		return loc.FullRange
	})
}

func (r *SymbolResolver) definitions(ctx context.Context, rangeFn func(protocol.SymbolLocation) protocol.RangeData) (gql.LocationConnectionResolver, error) {
	// TODO(sqs): workspace symbols have locations, but document symbols dont (currently based on
	// how this is all implemented).
	var adjustedLocations []resolvers.AdjustedLocation
	for _, loc := range r.symbol.Locations {
		rng := rangeFn(loc)
		adjustedLocations = append(adjustedLocations, resolvers.AdjustedLocation{
			Dump:           r.symbol.Dump,
			Path:           path.Clean(loc.URI),
			AdjustedCommit: r.symbol.Dump.Commit,
			AdjustedRange: lsifstore.Range{
				Start: lsifstore.Position{Line: rng.Start.Line, Character: rng.Start.Character},
				End:   lsifstore.Position{Line: rng.End.Line, Character: rng.End.Character},
			},
		})
	}
	return NewLocationConnectionResolver(adjustedLocations, nil, r.locationResolver), nil
}

func (r *SymbolResolver) References(ctx context.Context) (gql.LocationConnectionResolver, error) {
	if len(r.symbol.Locations) == 0 {
		// TODO(sqs): instead, look up by moniker
		return NewLocationConnectionResolver(nil, nil, nil), nil
	}
	queryResolver, err := r.newQueryResolver(ctx, path.Clean(r.symbol.Locations[0].URI))
	if err != nil {
		return nil, err
	}
	return queryResolver.References(ctx, &gql.LSIFPagedQueryPositionArgs{
		LSIFQueryPositionArgs: gql.LSIFQueryPositionArgs{
			Line:      int32(r.symbol.Locations[0].Range.Start.Line),
			Character: int32(r.symbol.Locations[0].Range.Start.Character),
		},
	})
}

func (r *SymbolResolver) Hover(ctx context.Context) (gql.HoverResolver, error) {
	// TODO(sqs): lookup hover by moniker if needed
	if len(r.symbol.Locations) == 0 {
		return nil, nil
	}
	queryResolver, err := r.newQueryResolver(ctx, path.Clean(r.symbol.Locations[0].URI))
	if err != nil {
		return nil, err
	}
	return queryResolver.Hover(ctx, &gql.LSIFQueryPositionArgs{
		Line:      int32(r.symbol.Locations[0].Range.Start.Line),
		Character: int32(r.symbol.Locations[0].Range.Start.Character),
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
	if len(r.symbol.Locations) == 0 {
		return "", 0, 0
	}
	return r.symbol.Locations[0].URI, r.symbol.Locations[0].FullRange.Start.Line, r.symbol.Locations[0].FullRange.End.Line
}
