package resolvers

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// defaultReferencesPageSize is the reference result page size when no limit is supplied in the
// GraphQL layer. This is used as an approximation for an acceptable page size as we make multiple
// references requests to discover references _not_ in the same file as the definition for API
// docs.
const defaultReferencesPageSize = 100

// DocumentationReferences returns the list of source locations that reference the symbol found at
// the given documentation path ID, if any.
func (r *queryResolver) DocumentationReferences(ctx context.Context, pathID string, limit int, rawCursor string) (_ []AdjustedLocation, _ string, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, r.operations.documentationReferences, slowReferencesRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", r.repositoryID),
			log.String("commit", r.commit),
			log.Int("numUploads", len(r.uploads)),
			log.String("uploads", uploadIDsToString(r.uploads)),
			log.String("pathID", pathID),
		},
	})
	defer endObservation()

	// A documentation path ID is repo-local, i.e. the associated definition is always going to be
	// found in the "local" bundles. It's not possible for it to be in another repository. However,
	// references to that definition could be in other repositories and we want to include those.
	// That requires some fairly complex logic (see query_references.go) for moniker searches, etc.
	//
	// What we do here is first resolve the local definitions, then execute a standard references
	// request on the first location we find.
	for _, upload := range r.uploads {
		trace.Log(log.Int("uploadID", upload.ID))

		// Effectively replicate what we do in r.DocumentationDefinition in order to lookup the definition
		// locations.
		locations, _, err := r.lsifStore.DocumentationDefinitions(ctx, upload.ID, pathID, DefinitionsLimit, 0)
		if err != nil {
			return nil, "", errors.Wrap(err, "lsifStore.DocumentationDefinitions")
		}
		if len(locations) == 0 {
			continue
		}
		r.path = locations[0].Path
		adjustedLocations, err := r.adjustLocations(ctx, locations)
		if err != nil {
			return nil, "", err
		}

		// Now lookup references.
		location := adjustedLocations[0]
		var (
			references       = make([]AdjustedLocation, 0, limit)
			inDefinitionFile []AdjustedLocation
		)
		for len(references) < limit {
			var candidates []AdjustedLocation
			r.path = location.Path
			candidates, rawCursor, err = r.References(ctx, location.AdjustedRange.Start.Line, location.AdjustedRange.Start.Character, defaultReferencesPageSize, rawCursor)
			if err != nil {
				return nil, rawCursor, err
			}
			for _, candidate := range candidates {
				isDefinitionFile := candidate.Dump.RepositoryID == r.repositoryID && candidate.Path == location.Path
				isDefinition := isDefinitionFile && candidate.AdjustedRange == location.AdjustedRange
				if isDefinition {
					// we never want the definition itself to show up as a reference.
				} else if isDefinitionFile {
					inDefinitionFile = append(inDefinitionFile, candidate)
				} else {
					references = append(references, candidate)
				}
			}
			if len(candidates) == 0 || rawCursor == "" {
				break // no more pages
			}
		}
		if len(references) == 0 {
			// If we found no references at all, we're willing to consider references in the
			// definition file. Otherwise, we don't really want these as they make poor usage
			// examples.
			return inDefinitionFile, rawCursor, nil
		}
		return references[:limit], rawCursor, nil
	}
	return nil, "", nil
}
