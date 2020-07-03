package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
)

type AggregateCodeIntelligenceResolver struct {
	r                resolvers.AdjustedAggregateCodeIntelligence
	locationResolver *CachedLocationResolver
}

func (r *AggregateCodeIntelligenceResolver) Range(ctx context.Context) (gql.RangeResolver, error) {
	return gql.NewRangeResolver(convertRange(r.r.Range)), nil
}

func (r *AggregateCodeIntelligenceResolver) Definitions(ctx context.Context) (gql.LocationConnectionResolver, error) {
	return NewLocationConnectionResolver(r.r.Definitions, nil, r.locationResolver), nil
}

func (r *AggregateCodeIntelligenceResolver) References(ctx context.Context) (gql.LocationConnectionResolver, error) {
	return NewLocationConnectionResolver(r.r.References, nil, r.locationResolver), nil
}

func (r *AggregateCodeIntelligenceResolver) Hover(ctx context.Context) (gql.HoverResolver, error) {
	return NewHoverResolver(r.r.HoverText, convertRange(r.r.Range)), nil
}
