package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const slowImplementationsRequestThreshold = time.Second

// ImplementationsLimit is maximum the number of locations returned from Implementations.
const ImplementationsLimit = 100

// Implementations returns the list of source locations that define the symbol at the given position.
func (r *queryResolver) Implementations(ctx context.Context, line, character int, limit int, rawCursor string) (_ []AdjustedLocation, _ string, err error) {
	ctx, traceLog, endObservation := observeResolver(ctx, &err, "Implementations", r.operations.implementations, slowImplementationsRequestThreshold, observation.Args{
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
	cursor, err := decodeCursor(rawCursor)
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

	if cursor.OrderedMonikers == nil {
		if cursor.OrderedMonikers, err = r.orderedMonikers(ctx, adjustedUploads, "implementation"); err != nil {
			return nil, "", err
		}
	}
	traceLog(
		log.Int("numMonikers", len(cursor.OrderedMonikers)),
		log.String("monikers", monikersToString(cursor.OrderedMonikers)),
	)

	fmt.Println("monikers:")
	for _, moniker := range cursor.OrderedMonikers {
		fmt.Println("- ", moniker)
	}

	// Determine the set of uploads that define one of the ordered monikers. This may include
	// one of the adjusted indexes. This data may already be stashed in the cursor decoded above,
	// in which case we don't need to hit the database.

	// Set of dumps that cover the monikers' packages
	definitionUploadIDs, err := r.definitionUploadIDsFromCursor(ctx, adjustedUploads, cursor.OrderedMonikers, &cursor)
	if err != nil {
		return nil, "", err
	}
	traceLog(
		log.Int("numDefinitionUploads", len(definitionUploadIDs)),
		log.String("definitionUploads", intsToString(definitionUploadIDs)),
	)

	// If we pulled additional records back from the database, add them to the upload map. This
	// slice will be empty if the definition ids were cached on the cursor.

	// Query a single page of location results
	locations, err := r.pageReferences(ctx, "implementations", "definitions", adjustedUploads, cursor.OrderedMonikers, definitionUploadIDs, &cursor, limit)
	if err != nil {
		return nil, "", err
	}
	traceLog(log.Int("numLocations", len(locations)))
	fmt.Println("Implementations: len(locations)", len(locations))

	// Adjust the locations back to the appropriate range in the target commits. This adjusts
	// locations within the repository the user is browsing so that it appears all references
	// are occurring at the same commit they are looking at.

	adjustedLocations, err := r.adjustLocations(ctx, locations)
	if err != nil {
		return nil, "", err
	}
	traceLog(log.Int("numAdjustedLocations", len(adjustedLocations)))

	nextCursor := ""
	if cursor.Phase != "done" {
		nextCursor = encodeCursor(cursor)
	}

	return adjustedLocations, nextCursor, nil
}
