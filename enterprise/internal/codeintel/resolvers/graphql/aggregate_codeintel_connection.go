package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
)

type AggregateCodeIntelligenceConnectionResolver struct {
	ranges           []resolvers.AdjustedAggregateCodeIntelligence
	locationResolver *CachedLocationResolver
}

func (r *AggregateCodeIntelligenceConnectionResolver) Nodes(ctx context.Context) ([]gql.AggregateCodeIntelligenceResolver, error) {
	var resolvers []gql.AggregateCodeIntelligenceResolver
	for _, rn := range r.ranges {
		resolvers = append(resolvers, &AggregateCodeIntelligenceResolver{
			r:                rn,
			locationResolver: r.locationResolver,
		})
	}

	return resolvers, nil
}
