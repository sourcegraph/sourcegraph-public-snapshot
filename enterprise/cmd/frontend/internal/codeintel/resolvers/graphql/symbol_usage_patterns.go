package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
)

func (r *symbolUsageResolver) Patterns(ctx context.Context) ([]gql.SymbolUsagePatternResolver, error) {
}

type symbolUsagePatternResolver struct {
	description      string
	exampleLocations []symbolUsagePatternExampleLocation

	locationResolver *CachedLocationResolver
}

func (r *symbolUsagePatternResolver) Description() string { return r.description }

func (r *symbolUsagePatternResolver) ExampleLocations() []gql.SymbolUsagePatternExampleLocationEdgeResolver {
	edges := make([]gql.SymbolUsagePatternExampleLocationEdgeResolver, len(r.exampleLocations))
	for i, exampleLocation := range r.exampleLocations {
		edges[i] = &symbolUsagePatternExampleLocationEdgeResolver{
			exampleLocation:  exampleLocation,
			locationResolver: r.locationResolver,
		}
	}
	return edges
}

type symbolUsagePatternExampleLocation struct {
	description string
	location    resolvers.AdjustedLocation
}

type symbolUsagePatternExampleLocationEdgeResolver struct {
	exampleLocation symbolUsagePatternExampleLocation

	locationResolver *CachedLocationResolver
}

func (r *symbolUsagePatternExampleLocationEdgeResolver) Description() string {
	return r.exampleLocation.description
}

func (r *symbolUsagePatternExampleLocationEdgeResolver) Location(ctx context.Context) (gql.LocationResolver, error) {
	return resolveLocation(ctx, r.locationResolver, r.exampleLocation.location)
}
