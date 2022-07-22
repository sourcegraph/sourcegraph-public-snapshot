package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const slowReferencesRequestThreshold = time.Second

func (r *resolver) References(ctx context.Context, args shared.RequestArgs) (_ []shared.UploadLocation, _ string, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, r.operations.references, slowReferencesRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", args.RepositoryID),
			log.String("commit", args.Commit),
			log.String("path", args.Path),
			log.Int("numUploads", len(r.dataLoader.uploads)),
			log.String("uploads", uploadIDsToString(r.dataLoader.uploads)),
			log.Int("line", args.Line),
			log.Int("character", args.Character),
		},
	})
	defer endObservation()

	// Decode cursor given from previous response or create a new one with default values.
	// We use the cursor state track offsets with the result set and cache initial data that
	// is used to resolve each page. This cursor will be modified in-place to become the
	// cursor used to fetch the subsequent page of results in this result set.
	cursor, err := decodeReferencesCursor(args.RawCursor)
	if err != nil {
		return nil, "", errors.Wrap(err, fmt.Sprintf("invalid cursor: %q", args.RawCursor))
	}

	// Adjust the path and position for each visible upload based on its git difference to
	// the target commit. This data may already be stashed in the cursor decoded above, in
	// which case we don't need to hit the database.

	// References at the given file:line:character could come from multiple uploads, so we
	// need to look in all uploads and merge the results.

	adjustedUploads, cursorsToVisibleUploads, err := r.getVisibleUploadsFromCursor(ctx, args.Line, args.Character, &cursor.CursorsToVisibleUploads)
	if err != nil {
		return nil, "", err
	}

	// Update the cursors with the updated visible uploads.
	cursor.CursorsToVisibleUploads = cursorsToVisibleUploads

	// Gather all monikers attached to the ranges enclosing the requested position. This data
	// may already be stashed in the cursor decoded above, in which case we don't need to hit
	// the database.
	if cursor.OrderedMonikers == nil {
		if cursor.OrderedMonikers, err = r.getOrderedMonikers(ctx, adjustedUploads, "import", "export"); err != nil {
			return nil, "", err
		}
	}
	trace.Log(
		log.Int("numMonikers", len(cursor.OrderedMonikers)),
		log.String("monikers", monikersToString(cursor.OrderedMonikers)),
	)

	// Phase 1: Gather all "local" locations via LSIF graph traversal. We'll continue to request additional
	// locations until we fill an entire page (the size of which is denoted by the given limit) or there are
	// no more local results remaining.
	var locations []shared.Location
	if cursor.Phase == "local" {
		localLocations, hasMore, err := r.getPageLocalLocations(
			ctx,
			r.svc.GetReferences,
			adjustedUploads,
			&cursor.LocalCursor,
			args.Limit-len(locations),
			trace,
		)
		if err != nil {
			return nil, "", err
		}
		locations = append(locations, localLocations...)

		if !hasMore {
			// No more local results, move on to phase 2
			cursor.Phase = "remote"
		}
	}

	// Phase 2: Gather all "remote" locations via moniker search. We only do this if there are no more local
	// results. We'll continue to request additional locations until we fill an entire page or there are no
	// more local results remaining, just as we did above.
	if cursor.Phase == "remote" {
		if cursor.RemoteCursor.UploadBatchIDs == nil {
			cursor.RemoteCursor.UploadBatchIDs = []int{}
			definitionUploads, err := r.getUploadsWithDefinitionsForMonikers(ctx, cursor.OrderedMonikers)
			if err != nil {
				return nil, "", err
			}
			for i := range definitionUploads {
				found := false
				for j := range adjustedUploads {
					if definitionUploads[i].ID == adjustedUploads[j].Upload.ID {
						found = true
						break
					}
				}
				if !found {
					cursor.RemoteCursor.UploadBatchIDs = append(cursor.RemoteCursor.UploadBatchIDs, definitionUploads[i].ID)
				}
			}
		}

		for len(locations) < args.Limit {
			remoteLocations, hasMore, err := r.getPageRemoteLocations(ctx, "references", adjustedUploads, cursor.OrderedMonikers, &cursor.RemoteCursor, args.Limit-len(locations), trace)
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
	// locations within the repository the user is browsing so that it appears all references
	// are occurring at the same commit they are looking at.

	referenceLocations, err := r.getUploadLocations(ctx, locations)
	if err != nil {
		return nil, "", err
	}
	trace.Log(log.Int("numReferenceLocations", len(referenceLocations)))

	nextCursor := ""
	if cursor.Phase != "done" {
		nextCursor = encodeReferencesCursor(cursor)
	}

	return referenceLocations, nextCursor, nil
}
