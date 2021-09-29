package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
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

	// Maintain a map from identifers to hydrated upload records from the database. We use
	// this map as a quick lookup when constructing the resulting location set. Any additional
	// upload records pulled back from the database while processing this page will be added
	// to this map.
	uploadsByID := make(map[int]dbstore.Dump, len(r.uploads))
	for i := range r.uploads {
		uploadsByID[r.uploads[i].ID] = r.uploads[i]
	}

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

	adjustedUploads, err := r.adjustedUploadsFromCursor(ctx, line, character, uploadsByID, &cursor)
	if err != nil {
		return nil, "", err
	}

	// Gather allmonikers attached to the ranges enclosing the requested position. This data
	// may already be stashed in the cursor decoded above, in which case we don't need to hit
	// the database.

	orderedMonikers, err := r.orderedMonikersFromCursor(ctx, adjustedUploads, &cursor, "implementation")
	if err != nil {
		return nil, "", err
	}
	traceLog(
		log.Int("numMonikers", len(orderedMonikers)),
		log.String("monikers", monikersToString(orderedMonikers)),
	)

	fmt.Println("monikers:")
	for _, moniker := range orderedMonikers {
		fmt.Println("- ", moniker)
	}

	// Determine the set of uploads that define one of the ordered monikers. This may include
	// one of the adjusted indexes. This data may already be stashed in the cursor decoded above,
	// in which case we don't need to hit the database.

	// Set of dumps that cover the monikers' packages
	definitionUploadIDs, definitionUploads, err := r.definitionUploadIDsFromCursor(ctx, adjustedUploads, orderedMonikers, &cursor)
	if err != nil {
		return nil, "", err
	}
	traceLog(
		log.Int("numDefinitionUploads", len(definitionUploadIDs)),
		log.String("definitionUploads", intsToString(definitionUploadIDs)),
	)

	// If we pulled additional records back from the database, add them to the upload map. This
	// slice will be empty if the definition ids were cached on the cursor.

	for i := range definitionUploads {
		uploadsByID[definitionUploads[i].ID] = definitionUploads[i]
	}

	// Query a single page of location results
	locations, hasMore, err := r.pageReferences(ctx, "implementations", "definitions", adjustedUploads, orderedMonikers, definitionUploadIDs, uploadsByID, &cursor, limit)
	if err != nil {
		return nil, "", err
	}
	traceLog(log.Int("numLocations", len(locations)))
	fmt.Println("Implementations: len(locations)", len(locations))

	// Adjust the locations back to the appropriate range in the target commits. This adjusts
	// locations within the repository the user is browsing so that it appears all references
	// are occurring at the same commit they are looking at.

	adjustedLocations, err := r.adjustLocations(ctx, uploadsByID, locations)
	if err != nil {
		return nil, "", err
	}
	traceLog(log.Int("numAdjustedLocations", len(adjustedLocations)))

	nextCursor := ""
	if hasMore {
		nextCursor = encodeCursor(cursor)
	}

	return adjustedLocations, nextCursor, nil
}
