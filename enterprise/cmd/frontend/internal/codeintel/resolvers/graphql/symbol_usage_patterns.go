package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
)

var sampleUsagePatternDescriptions = []string{
	"Most popular",
	"`nil` first arg",
	"in `for` loop",
}

func (r *symbolUsageResolver) UsagePatterns(ctx context.Context) ([]gql.SymbolUsagePatternResolver, error) {
	locations, _, err := r.symbol.references(ctx)
	if err != nil {
		return nil, err
	}
	if len(locations) == 0 {
		return nil, nil
	}
	locations = locations[1:] // remove "definition" location

	const maxLocations = 25 // TODO(sqs)
	if len(locations) > maxLocations {
		locations = locations[:maxLocations]
	}

	// TODO(sqs): break them up into arbitrary patterns
	numPatterns := 1 + len(locations)/7
	const maxPatterns = 1 // TODO(sqs)
	if numPatterns > maxPatterns {
		numPatterns = maxPatterns
	}
	patterns := make([]symbolUsagePattern, numPatterns)
	for i, loc := range locations {
		p := &patterns[i%numPatterns]
		p.symbol = r.symbol.symbol
		p.description = sampleUsagePatternDescriptions[i%numPatterns]
		p.exampleLocations = append(p.exampleLocations, symbolUsagePatternExampleLocation{
			symbol:   r.symbol.symbol,
			location: loc,
		})
	}

	// Sort and rank.
	for _, pattern := range patterns {
		exampleLocations, err := sortAndRankExampleLocations(ctx, r.locationResolver, pattern.exampleLocations)
		if err != nil {
			return nil, err
		}
		pattern.exampleLocations = exampleLocations
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
	symbol           resolvers.AdjustedSymbol
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
	symbol      resolvers.AdjustedSymbol
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
