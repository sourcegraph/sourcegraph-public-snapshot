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

		locations, _, err := r.lsifStore.Implementations(
			ctx,
			adjustedUploads[i].Upload.ID,
			adjustedUploads[i].AdjustedPathInBundle,
			adjustedUploads[i].AdjustedPosition.Line,
			adjustedUploads[i].AdjustedPosition.Character,
			ImplementationsLimit,
			0,
		)
		if err != nil {
			return nil, "", errors.Wrap(err, "lsifStore.Implementations")
		}
		if len(locations) > 0 {
			uploadsByID := map[int]dbstore.Dump{
				adjustedUploads[i].Upload.ID: adjustedUploads[i].Upload,
			}

			// If we have local implementations, we won't find a better one and can exit early
			adjustedLocations, err := r.adjustLocations(ctx, uploadsByID, locations)
			// TODO collect more implementations, don't exit early
			return adjustedLocations, "", err
		}
	}

	// TODO implement monikers
	return []AdjustedLocation{}, "", nil
}
