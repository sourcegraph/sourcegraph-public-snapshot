package resolvers

import (
	"context"
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
	// TODO pagination (see query_references.go)
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

	// Adjust the path and position for each visible upload based on its git difference to
	// the target commit.

	adjustedUploads, err := r.adjustUploads(ctx, line, character)
	if err != nil {
		return nil, "", err
	}

	// Gather the "local" implementation locations that are reachable via an implementationResult vertex.

	for i := range adjustedUploads {
		traceLog(log.Int("uploadID", adjustedUploads[i].Upload.ID))

		locations, _, err := r.lsifStore.Definitions(
			ctx,
			adjustedUploads[i].Upload.ID,
			adjustedUploads[i].AdjustedPathInBundle,
			adjustedUploads[i].AdjustedPosition.Line,
			adjustedUploads[i].AdjustedPosition.Character,
			DefinitionsLimit,
			0,
		)
		if err != nil {
			return nil, "", errors.Wrap(err, "lsifStore.Definitions")
		}
		if len(locations) > 0 {
			uploadsByID := map[int]dbstore.Dump{
				adjustedUploads[i].Upload.ID: adjustedUploads[i].Upload,
			}

			// If we have a local definition, we won't find a better one and can exit early
			adjustedLocations, err := r.adjustLocations(ctx, uploadsByID, locations)
			return adjustedLocations, "", err
		}
	}

	// TODO implement monikers
	return []AdjustedLocation{}, "", nil

	// -------------- The rest of this func is unmodified from query_definitions.go, except for the addition of a cursor in the return value -----------

	// Gather all import monikers attached to the ranges enclosing the requested position
	orderedMonikers, err := r.orderedMonikers(ctx, adjustedUploads, "import")
	if err != nil {
		return nil, "", err
	}
	traceLog(
		log.Int("numMonikers", len(orderedMonikers)),
		log.String("monikers", monikersToString(orderedMonikers)),
	)

	// Determine the set of uploads over which we need to perform a moniker search. This will
	// include all all indexes which define one of the ordered monikers. This should not include
	// any of the indexes we have already performed an LSIF graph traversal in above.
	uploads, err := r.definitionUploads(ctx, orderedMonikers)
	if err != nil {
		return nil, "", err
	}
	traceLog(
		log.Int("numDefinitionUploads", len(uploads)),
		log.String("definitionUploads", uploadIDsToString(uploads)),
	)

	// Perform the moniker search
	locations, _, err := r.monikerLocations(ctx, uploads, orderedMonikers, "definitions", DefinitionsLimit, 0)
	if err != nil {
		return nil, "", err
	}
	traceLog(log.Int("numLocations", len(locations)))

	// Adjust the locations back to the appropriate range in the target commits. This adjusts
	// locations within the repository the user is browsing so that it appears all definitions
	// are occurring at the same commit they are looking at.

	uploadsByID := make(map[int]dbstore.Dump, len(uploads))
	for i := range uploads {
		uploadsByID[uploads[i].ID] = uploads[i]
	}

	adjustedLocations, err := r.adjustLocations(ctx, uploadsByID, locations)
	if err != nil {
		return nil, "", err
	}
	traceLog(log.Int("numAdjustedLocations", len(adjustedLocations)))

	return adjustedLocations, "", nil
}
