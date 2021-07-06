package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
)

func (r *symbolUsageResolver) Patterns(ctx context.Context) ([]gql.SymbolUsagePatternResolver, error) {
	locations, _, err := r.symbol.references(ctx)
	if err != nil {
		return nil, err
	}

	// TODO(sqs): break them up into arbitrary patterns
	numPatterns := 1 + len(locations)/7
	const maxPatterns = 3
	if numPatterns > maxPatterns {
		numPatterns = 3
	}
	patterns := make([]symbolUsagePattern, numPatterns)
	for i, loc := range locations {
		p := &patterns[i%numPatterns]
		p.exampleLocations = append(p.exampleLocations, symbolUsagePatternExampleLocation{
			description: "foo",
			location:    loc,
		})
	}

	resolvers := make([]gql.SymbolUsagePatternResolver, len(patterns))
	for i, pattern := range patterns {
		resolvers[i] = &symbolUsagePatternResolver{
			symbolUsagePattern: pattern,
			locationResolver:   r.locationResolver,
		}
	}
	return resolvers, nil
}

type symbolUsagePattern struct {
	description      string
	exampleLocations []symbolUsagePatternExampleLocation
}

type symbolUsagePatternResolver struct {
	symbolUsagePattern

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
