package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
)

type CodeIntelligenceRangeConnectionResolver interface {
	Nodes(ctx context.Context) ([]CodeIntelligenceRangeResolver, error)
}

type codeIntelligenceRangeConnectionResolver struct {
	ranges           []shared.AdjustedCodeIntelligenceRange
	locationResolver *sharedresolvers.CachedLocationResolver
}

func NewCodeIntelligenceRangeConnectionResolver(ranges []shared.AdjustedCodeIntelligenceRange, locationResolver *sharedresolvers.CachedLocationResolver) CodeIntelligenceRangeConnectionResolver {
	return &codeIntelligenceRangeConnectionResolver{
		ranges:           ranges,
		locationResolver: locationResolver,
	}
}

func (r *codeIntelligenceRangeConnectionResolver) Nodes(ctx context.Context) ([]CodeIntelligenceRangeResolver, error) {
	var resolvers []CodeIntelligenceRangeResolver
	for _, rn := range r.ranges {
		resolvers = append(resolvers, &codeIntelligenceRangeResolver{
			r:                rn,
			locationResolver: r.locationResolver,
		})
	}

	return resolvers, nil
}
