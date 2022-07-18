package resolvers

import (
	"context"
	"time"

	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
)

const slowImplementationsRequestThreshold = time.Second

// ImplementationsLimit is maximum the number of locations returned from Implementations.
const ImplementationsLimit = 100

func symbolDumpToStoreDump(symbolDumps []shared.Dump) []store.Dump {
	dumps := make([]store.Dump, 0, len(symbolDumps))
	for _, d := range symbolDumps {
		dumps = append(dumps, store.Dump{
			ID:                d.ID,
			Commit:            d.Commit,
			Root:              d.Root,
			VisibleAtTip:      d.VisibleAtTip,
			UploadedAt:        d.UploadedAt,
			State:             d.State,
			FailureMessage:    d.FailureMessage,
			StartedAt:         d.StartedAt,
			FinishedAt:        d.FinishedAt,
			ProcessAfter:      d.ProcessAfter,
			NumResets:         d.NumResets,
			NumFailures:       d.NumFailures,
			RepositoryID:      d.RepositoryID,
			RepositoryName:    d.RepositoryName,
			Indexer:           d.Indexer,
			IndexerVersion:    d.IndexerVersion,
			AssociatedIndexID: d.AssociatedIndexID,
		})
	}

	return dumps
}

func adjustedLocationToUploadLocations(location []shared.UploadLocation) []AdjustedLocation {
	uploadLocation := make([]AdjustedLocation, 0, len(location))
	for _, loc := range location {
		dump := store.Dump{
			ID:                loc.Dump.ID,
			Commit:            loc.Dump.Commit,
			Root:              loc.Dump.Root,
			VisibleAtTip:      loc.Dump.VisibleAtTip,
			UploadedAt:        loc.Dump.UploadedAt,
			State:             loc.Dump.State,
			FailureMessage:    loc.Dump.FailureMessage,
			StartedAt:         loc.Dump.StartedAt,
			FinishedAt:        loc.Dump.FinishedAt,
			ProcessAfter:      loc.Dump.ProcessAfter,
			NumResets:         loc.Dump.NumResets,
			NumFailures:       loc.Dump.NumFailures,
			RepositoryID:      loc.Dump.RepositoryID,
			RepositoryName:    loc.Dump.RepositoryName,
			Indexer:           loc.Dump.Indexer,
			IndexerVersion:    loc.Dump.IndexerVersion,
			AssociatedIndexID: loc.Dump.AssociatedIndexID,
		}

		adjustedRange := lsifstore.Range{
			Start: lsifstore.Position{
				Line:      loc.TargetRange.Start.Line,
				Character: loc.TargetRange.Start.Character,
			},
			End: lsifstore.Position{
				Line:      loc.TargetRange.End.Line,
				Character: loc.TargetRange.End.Character,
			},
		}

		uploadLocation = append(uploadLocation, AdjustedLocation{
			Dump:           dump,
			Path:           loc.Path,
			AdjustedCommit: loc.TargetCommit,
			AdjustedRange:  adjustedRange,
		})
	}

	return uploadLocation
}

// Implementations returns the list of source locations that define the symbol at the given position.
func (r *queryResolver) Implementations(ctx context.Context, line, character int, limit int, rawCursor string) (_ []AdjustedLocation, _ string, err error) {
	args := shared.RequestArgs{
		RepositoryID: r.repositoryID,
		Commit:       r.commit,
		Path:         r.path,
		Line:         line,
		Character:    character,
		Limit:        limit,
		RawCursor:    rawCursor,
	}
	impl, cursor, err := r.symbolsResolver.Implementations(ctx, args)
	if err != nil {
		return nil, "", err
	}

	adjustedLoc := adjustedLocationToUploadLocations(impl)

	return adjustedLoc, cursor, nil

	// ctx, trace, endObservation := observeResolver(ctx, &err, r.operations.implementations, slowImplementationsRequestThreshold, observation.Args{
	// 	LogFields: []log.Field{
	// 		log.Int("repositoryID", r.repositoryID),
	// 		log.String("commit", r.commit),
	// 		log.String("path", r.path),
	// 		log.Int("numUploads", len(r.inMemoryUploads)),
	// 		log.String("uploads", uploadIDsToString(r.inMemoryUploads)),
	// 		log.Int("line", line),
	// 		log.Int("character", character),
	// 	},
	// })
	// defer endObservation()

	// Decode cursor given from previous response or create a new one with default values.
	// We use the cursor state track offsets with the result set and cache initial data that
	// is used to resolve each page. This cursor will be modified in-place to become the
	// cursor used to fetch the subsequent page of results in this result set.
	// cursor, err := decodeImplementationsCursor(rawCursor)
	// if err != nil {
	// 	return nil, "", errors.Wrap(err, fmt.Sprintf("invalid cursor: %q", rawCursor))
	// }

	// // Adjust the path and position for each visible upload based on its git difference to
	// // the target commit. This data may already be stashed in the cursor decoded above, in
	// // which case we don't need to hit the database.

	// adjustedUploads, err := r.adjustedUploadsFromCursor(ctx, line, character, &cursor.AdjustedUploads)
	// if err != nil {
	// 	return nil, "", err
	// }

	// // Gather all monikers attached to the ranges enclosing the requested position. This data
	// // may already be stashed in the cursor decoded above, in which case we don't need to hit
	// // the database.

	// if cursor.OrderedImplementationMonikers == nil {
	// 	if cursor.OrderedImplementationMonikers, err = r.orderedMonikers(ctx, adjustedUploads, "implementation"); err != nil {
	// 		return nil, "", err
	// 	}
	// }
	// trace.Log(
	// 	log.Int("numImplementationMonikers", len(cursor.OrderedImplementationMonikers)),
	// 	log.String("implementationMonikers", monikersToString(cursor.OrderedImplementationMonikers)),
	// )

	// if cursor.OrderedExportMonikers == nil {
	// 	if cursor.OrderedExportMonikers, err = r.orderedMonikers(ctx, adjustedUploads, "export"); err != nil {
	// 		return nil, "", err
	// 	}
	// }
	// trace.Log(
	// 	log.Int("numExportMonikers", len(cursor.OrderedExportMonikers)),
	// 	log.String("exportMonikers", monikersToString(cursor.OrderedExportMonikers)),
	// )

	// // Phase 1: Gather all "local" locations via LSIF graph traversal. We'll continue to request additional
	// // locations until we fill an entire page (the size of which is denoted by the given limit) or there are
	// // no more local results remaining.
	// var locations []lsifstore.Location
	// if cursor.Phase == "local" {
	// 	for len(locations) < limit {
	// 		localLocations, hasMore, err := r.pageLocalLocations(ctx, r.lsifStore.Implementations, adjustedUploads, &cursor.LocalCursor, limit-len(locations), trace)
	// 		if err != nil {
	// 			return nil, "", err
	// 		}
	// 		locations = append(locations, localLocations...)

	// 		if !hasMore {
	// 			cursor.Phase = "dependencies"
	// 			break
	// 		}
	// 	}
	// }

	// // Phase 2: Gather all "remote" locations in dependencies via moniker search. We only do this if
	// // there are no more local results. We'll continue to request additional locations until we fill an
	// // entire page or there are no more local results remaining, just as we did above.
	// if cursor.Phase == "dependencies" {
	// 	uploads, err := r.definitionUploads(ctx, cursor.OrderedImplementationMonikers)
	// 	if err != nil {
	// 		return nil, "", err
	// 	}
	// 	trace.Log(
	// 		log.Int("numDefinitionUploads", len(uploads)),
	// 		log.String("definitionUploads", uploadIDsToString(uploads)),
	// 	)

	// 	definitionLocations, _, err := r.monikerLocations(ctx, uploads, cursor.OrderedImplementationMonikers, "definitions", DefinitionsLimit, 0)
	// 	if err != nil {
	// 		return nil, "", err
	// 	}
	// 	locations = append(locations, definitionLocations...)

	// 	cursor.Phase = "dependents"
	// }

	// // Phase 3: Gather all "remote" locations in dependents via moniker search.
	// if cursor.Phase == "dependents" {
	// 	for len(locations) < limit {
	// 		remoteLocations, hasMore, err := r.pageRemoteLocations(ctx, "implementations", adjustedUploads, cursor.OrderedExportMonikers, &cursor.RemoteCursor, limit-len(locations), trace)
	// 		if err != nil {
	// 			return nil, "", err
	// 		}
	// 		locations = append(locations, remoteLocations...)

	// 		if !hasMore {
	// 			cursor.Phase = "done"
	// 			break
	// 		}
	// 	}
	// }

	// trace.Log(log.Int("numLocations", len(locations)))

	// // Adjust the locations back to the appropriate range in the target commits. This adjusts
	// // locations within the repository the user is browsing so that it appears all implementations
	// // are occurring at the same commit they are looking at.

	// adjustedLocations, err := r.adjustLocations(ctx, locations)
	// if err != nil {
	// 	return nil, "", err
	// }
	// trace.Log(log.Int("numAdjustedLocations", len(adjustedLocations)))

	// nextCursor := ""
	// if cursor.Phase != "done" {
	// 	nextCursor = encodeImplementationsCursor(cursor)
	// }

	// return adjustedLocations, nextCursor, nil
}
