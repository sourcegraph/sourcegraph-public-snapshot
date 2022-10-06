package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
)

type CodeIntelligenceRangeResolver interface {
	Range(ctx context.Context) (RangeResolver, error)
	Definitions(ctx context.Context) (LocationConnectionResolver, error)
	References(ctx context.Context) (LocationConnectionResolver, error)
	Implementations(ctx context.Context) (LocationConnectionResolver, error)
	Hover(ctx context.Context) (HoverResolver, error)
}

type codeIntelligenceRangeResolver struct {
	r                shared.AdjustedCodeIntelligenceRange
	locationResolver *sharedresolvers.CachedLocationResolver
}

func NewCodeIntelligenceRangeResolver(r shared.AdjustedCodeIntelligenceRange, locationResolver *sharedresolvers.CachedLocationResolver) CodeIntelligenceRangeResolver {
	return &codeIntelligenceRangeResolver{
		r:                r,
		locationResolver: locationResolver,
	}
}

func (r *codeIntelligenceRangeResolver) Range(ctx context.Context) (RangeResolver, error) {
	return NewRangeResolver(convertRange(r.r.Range)), nil
}

func (r *codeIntelligenceRangeResolver) Definitions(ctx context.Context) (LocationConnectionResolver, error) {
	return NewLocationConnectionResolver(r.r.Definitions, nil, r.locationResolver), nil
}

func (r *codeIntelligenceRangeResolver) References(ctx context.Context) (LocationConnectionResolver, error) {
	return NewLocationConnectionResolver(r.r.References, nil, r.locationResolver), nil
}

func (r *codeIntelligenceRangeResolver) Implementations(ctx context.Context) (LocationConnectionResolver, error) {
	return NewLocationConnectionResolver(r.r.Implementations, nil, r.locationResolver), nil
}

func (r *codeIntelligenceRangeResolver) Hover(ctx context.Context) (HoverResolver, error) {
	return NewHoverResolver(r.r.HoverText, convertRange(r.r.Range)), nil
}
