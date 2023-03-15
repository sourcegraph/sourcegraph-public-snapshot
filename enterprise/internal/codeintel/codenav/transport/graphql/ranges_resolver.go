package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type codeIntelligenceRangeResolver struct {
	r                shared.AdjustedCodeIntelligenceRange
	locationResolver *sharedresolvers.CachedLocationResolver
}

func NewCodeIntelligenceRangeResolver(r shared.AdjustedCodeIntelligenceRange, locationResolver *sharedresolvers.CachedLocationResolver) resolverstubs.CodeIntelligenceRangeResolver {
	return &codeIntelligenceRangeResolver{
		r:                r,
		locationResolver: locationResolver,
	}
}

func (r *codeIntelligenceRangeResolver) Range(ctx context.Context) (resolverstubs.RangeResolver, error) {
	return NewRangeResolver(convertRange(r.r.Range)), nil
}

func (r *codeIntelligenceRangeResolver) Definitions(ctx context.Context) (resolverstubs.LocationConnectionResolver, error) {
	return NewLocationConnectionResolver(r.r.Definitions, nil, r.locationResolver), nil
}

func (r *codeIntelligenceRangeResolver) References(ctx context.Context) (resolverstubs.LocationConnectionResolver, error) {
	return NewLocationConnectionResolver(r.r.References, nil, r.locationResolver), nil
}

func (r *codeIntelligenceRangeResolver) Implementations(ctx context.Context) (resolverstubs.LocationConnectionResolver, error) {
	return NewLocationConnectionResolver(r.r.Implementations, nil, r.locationResolver), nil
}

func (r *codeIntelligenceRangeResolver) Hover(ctx context.Context) (resolverstubs.HoverResolver, error) {
	return NewHoverResolver(r.r.HoverText, convertRange(r.r.Range)), nil
}
