package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type codeIntelligenceRangeConnectionResolver struct {
	ranges           []shared.AdjustedCodeIntelligenceRange
	locationResolver *sharedresolvers.CachedLocationResolver
}

func NewCodeIntelligenceRangeConnectionResolver(ranges []shared.AdjustedCodeIntelligenceRange, locationResolver *sharedresolvers.CachedLocationResolver) resolverstubs.CodeIntelligenceRangeConnectionResolver {
	return &codeIntelligenceRangeConnectionResolver{
		ranges:           ranges,
		locationResolver: locationResolver,
	}
}

func (r *codeIntelligenceRangeConnectionResolver) Nodes(ctx context.Context) ([]resolverstubs.CodeIntelligenceRangeResolver, error) {
	var resolvers []resolverstubs.CodeIntelligenceRangeResolver
	for _, rn := range r.ranges {
		resolvers = append(resolvers, &codeIntelligenceRangeResolver{
			r:                rn,
			locationResolver: r.locationResolver,
		})
	}

	return resolvers, nil
}
