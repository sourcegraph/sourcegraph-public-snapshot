package graphql

import (
	"context"

	"github.com/sourcegraph/go-lsp"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

// convertRange creates an LSP range from a bundle range.
func convertRange(r types.Range) lsp.Range {
	return lsp.Range{Start: convertPosition(r.Start.Line, r.Start.Character), End: convertPosition(r.End.Line, r.End.Character)}
}

func sharedRangeTolspRange(r types.Range) lsp.Range {
	return lsp.Range{Start: convertPosition(r.Start.Line, r.Start.Character), End: convertPosition(r.End.Line, r.End.Character)}
}

func convertPosition(line, character int) lsp.Position {
	return lsp.Position{Line: line, Character: character}
}

// resolveLocations creates a slide of LocationResolvers for the given list of adjusted locations. The
// resulting list may be smaller than the input list as any locations with a commit not known by
// gitserver will be skipped.
func resolveLocations(ctx context.Context, locationResolver *sharedresolvers.CachedLocationResolver, locations []types.UploadLocation) ([]resolverstubs.LocationResolver, error) {
	resolvedLocations := make([]resolverstubs.LocationResolver, 0, len(locations))
	for i := range locations {
		resolver, err := resolveLocation(ctx, locationResolver, locations[i])
		if err != nil {
			return nil, err
		}
		if resolver == nil {
			continue
		}

		resolvedLocations = append(resolvedLocations, resolver)
	}

	return resolvedLocations, nil
}

// resolveLocation creates a LocationResolver for the given adjusted location. This function may return a
// nil resolver if the location's commit is not known by gitserver.
func resolveLocation(ctx context.Context, locationResolver *sharedresolvers.CachedLocationResolver, location types.UploadLocation) (resolverstubs.LocationResolver, error) {
	treeResolver, err := locationResolver.Path(ctx, api.RepoID(location.Dump.RepositoryID), location.TargetCommit, location.Path, false)
	if err != nil || treeResolver == nil {
		return nil, err
	}

	lspRange := convertRange(location.TargetRange)
	return NewLocationResolver(treeResolver, &lspRange), nil
}
