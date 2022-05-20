package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const slowImplementationsRequestThreshold = time.Second

// ImplementationsLimit is maximum the number of locations returned from Implementations.
const ImplementationsLimit = 100

// Implementations returns the list of source locations that define the symbol at the given position.
func (r *queryResolver) Implementations(ctx context.Context, line, character int, limit int, rawCursor string) (_ []AdjustedLocation, _ string, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, r.operations.implementations, slowImplementationsRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", r.repositoryID),
			log.String("commit", r.commit),
			log.String("path", r.path),
			log.Int("numUploads", len(r.uploads)),
			log.String("uploads", uploadIDsToString(r.uploads)),
			log.Int("line", line),
			log.Int("character", character),
		},
	})
	defer endObservation()

	// Decode cursor given from previous response or create a new one with default values.
	// We use the cursor state track offsets with the result set and cache initial data that
	// is used to resolve each page. This cursor will be modified in-place to become the
	// cursor used to fetch the subsequent page of results in this result set.
	cursor, err := decodeImplementationsCursor(rawCursor)
	if err != nil {
		return nil, "", errors.Wrap(err, fmt.Sprintf("invalid cursor: %q", rawCursor))
	}

	// Adjust the path and position for each visible upload based on its git difference to
	// the target commit. This data may already be stashed in the cursor decoded above, in
	// which case we don't need to hit the database.

	adjustedUploads, err := r.adjustedUploadsFromCursor(ctx, line, character, &cursor.AdjustedUploads)
	if err != nil {
		return nil, "", err
	}

	// Gather all monikers attached to the ranges enclosing the requested position. This data
	// may already be stashed in the cursor decoded above, in which case we don't need to hit
	// the database.

	if cursor.OrderedImplementationMonikers == nil {
		if cursor.OrderedImplementationMonikers, err = r.orderedMonikers(ctx, adjustedUploads, "implementation"); err != nil {
			return nil, "", err
		}
	}
	trace.Log(
		log.Int("numImplementationMonikers", len(cursor.OrderedImplementationMonikers)),
		log.String("implementationMonikers", monikersToString(cursor.OrderedImplementationMonikers)),
	)

	if cursor.OrderedExportMonikers == nil {
		if cursor.OrderedExportMonikers, err = r.orderedMonikers(ctx, adjustedUploads, "export"); err != nil {
			return nil, "", err
		}
	}
	trace.Log(
		log.Int("numExportMonikers", len(cursor.OrderedExportMonikers)),
		log.String("exportMonikers", monikersToString(cursor.OrderedExportMonikers)),
	)

	// Phase 1: Gather all "local" locations via LSIF graph traversal. We'll continue to request additional
	// locations until we fill an entire page (the size of which is denoted by the given limit) or there are
	// no more local results remaining.
	var locations []lsifstore.Location
	if cursor.Phase == "local" {
		for len(locations) < limit {
			localLocations, hasMore, err := r.pageLocalLocations(ctx, r.lsifStore.Implementations, adjustedUploads, &cursor.LocalCursor, limit-len(locations), trace)
			if err != nil {
				return nil, "", err
			}
			locations = append(locations, localLocations...)

			if !hasMore {
				cursor.Phase = "dependencies"
				break
			}
		}
	}

	// Phase 2: Gather all "remote" locations in dependencies via moniker search. We only do this if
	// there are no more local results. We'll continue to request additional locations until we fill an
	// entire page or there are no more local results remaining, just as we did above.
	if cursor.Phase == "dependencies" {
		uploads, err := r.definitionUploads(ctx, cursor.OrderedImplementationMonikers)
		if err != nil {
			return nil, "", err
		}
		trace.Log(
			log.Int("numDefinitionUploads", len(uploads)),
			log.String("definitionUploads", uploadIDsToString(uploads)),
		)

		definitionLocations, _, err := r.monikerLocations(ctx, uploads, cursor.OrderedImplementationMonikers, "definitions", DefinitionsLimit, 0)
		if err != nil {
			return nil, "", err
		}
		locations = append(locations, definitionLocations...)

		cursor.Phase = "dependents"
	}

	// Phase 3: Gather all "remote" locations in dependents via moniker search.
	if cursor.Phase == "dependents" {
		for len(locations) < limit {
			remoteLocations, hasMore, err := r.pageRemoteLocations(ctx, "implementations", adjustedUploads, cursor.OrderedExportMonikers, &cursor.RemoteCursor, limit-len(locations), trace)
			if err != nil {
				return nil, "", err
			}
			locations = append(locations, remoteLocations...)

			if !hasMore {
				cursor.Phase = "done"
				break
			}
		}
	}

	trace.Log(log.Int("numLocations", len(locations)))

	// Adjust the locations back to the appropriate range in the target commits. This adjusts
	// locations within the repository the user is browsing so that it appears all implementations
	// are occurring at the same commit they are looking at.

	adjustedLocations, err := r.adjustLocations(ctx, locations)
	if err != nil {
		return nil, "", err
	}
	trace.Log(log.Int("numAdjustedLocations", len(adjustedLocations)))

	nextCursor := ""
	if cursor.Phase != "done" {
		nextCursor = encodeImplementationsCursor(cursor)
	}

	return adjustedLocations, nextCursor, nil
}
