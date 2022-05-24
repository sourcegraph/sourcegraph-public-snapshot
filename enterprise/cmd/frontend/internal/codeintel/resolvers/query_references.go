package resolvers

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/opentracing/opentracing-go/log"

	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const slowReferencesRequestThreshold = time.Second

// References returns the list of source locations that reference the symbol at the given position.
func (r *queryResolver) References(ctx context.Context, line, character, limit int, rawCursor string) (_ []AdjustedLocation, _ string, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, r.operations.references, slowReferencesRequestThreshold, observation.Args{
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
	cursor, err := decodeReferencesCursor(rawCursor)
	if err != nil {
		return nil, "", errors.Wrap(err, fmt.Sprintf("invalid cursor: %q", rawCursor))
	}

	// Adjust the path and position for each visible upload based on its git difference to
	// the target commit. This data may already be stashed in the cursor decoded above, in
	// which case we don't need to hit the database.

	// References at the given file:line:character could come from multiple uploads, so we
	// need to look in all uploads and merge the results.

	adjustedUploads, err := r.adjustedUploadsFromCursor(ctx, line, character, &cursor.AdjustedUploads)
	if err != nil {
		return nil, "", err
	}

	// Gather all monikers attached to the ranges enclosing the requested position. This data
	// may already be stashed in the cursor decoded above, in which case we don't need to hit
	// the database.

	if cursor.OrderedMonikers == nil {
		if cursor.OrderedMonikers, err = r.orderedMonikers(ctx, adjustedUploads, "import", "export"); err != nil {
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
	var locations []lsifstore.Location
	if cursor.Phase == "local" {
		localLocations, hasMore, err := r.pageLocalLocations(
			ctx,
			r.lsifStore.References,
			adjustedUploads,
			&cursor.LocalCursor,
			limit-len(locations),
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
			definitionUploads, err := r.definitionUploads(ctx, cursor.OrderedMonikers)
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

		for len(locations) < limit {
			remoteLocations, hasMore, err := r.pageRemoteLocations(ctx, "references", adjustedUploads, cursor.OrderedMonikers, &cursor.RemoteCursor, limit-len(locations), trace)
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

	adjustedLocations, err := r.adjustLocations(ctx, locations)
	if err != nil {
		return nil, "", err
	}
	trace.Log(log.Int("numAdjustedLocations", len(adjustedLocations)))

	nextCursor := ""
	if cursor.Phase != "done" {
		nextCursor = encodeReferencesCursor(cursor)
	}

	return adjustedLocations, nextCursor, nil
}

// ErrConcurrentModification occurs when a page of a references request cannot be resolved as
// the set of visible uploads have changed since the previous request for the same result set.
var ErrConcurrentModification = errors.New("result set changed while paginating")

// adjustedUploadsFromCursor adjusts the current target path and the given position for each upload
// visible from the current target commit. If an upload cannot be adjusted, it will be omitted from
// the returned slice. The returned slice will be cached on the given cursor. If this data is already
// stashed on the given cursor, the result is recalculated from the cursor data/resolver context, and
// we don't need to hit the database.
//
// An error is returned if the set of visible uploads has changed since the previous request of this
// result set (specifically if an index becomes invisible). This behavior may change in the future.
func (r *queryResolver) adjustedUploadsFromCursor(ctx context.Context, line, character int, cursorAdjustedUploads *[]cursorAdjustedUpload) ([]adjustedUpload, error) {
	if *cursorAdjustedUploads != nil {
		adjustedUploads := make([]adjustedUpload, 0, len(*cursorAdjustedUploads))
		for _, u := range *cursorAdjustedUploads {
			upload, ok := r.uploadCache[u.DumpID]
			if !ok {
				return nil, ErrConcurrentModification
			}

			adjustedUploads = append(adjustedUploads, adjustedUpload{
				Upload:               upload,
				AdjustedPath:         u.AdjustedPath,
				AdjustedPosition:     u.AdjustedPosition,
				AdjustedPathInBundle: u.AdjustedPathInBundle,
			})
		}

		return adjustedUploads, nil
	}

	adjustedUploads, err := r.adjustUploads(ctx, line, character)
	if err != nil {
		return nil, err
	}

	*cursorAdjustedUploads = make([]cursorAdjustedUpload, 0, len(adjustedUploads))
	for i := range adjustedUploads {
		*cursorAdjustedUploads = append(*cursorAdjustedUploads, cursorAdjustedUpload{
			DumpID:               adjustedUploads[i].Upload.ID,
			AdjustedPath:         adjustedUploads[i].AdjustedPath,
			AdjustedPosition:     adjustedUploads[i].AdjustedPosition,
			AdjustedPathInBundle: adjustedUploads[i].AdjustedPathInBundle,
		})
	}

	return adjustedUploads, nil
}

type getLocationsFn = func(ctx context.Context, bundleID int, path string, line int, character int, limit int, offset int) ([]lsifstore.Location, int, error)

// pageLocalLocations returns a slice of the (local) result set denoted by the given cursor fulfilled by
// traversing the LSIF graph. The given cursor will be adjusted to reflect the offsets required to resolve
// the next page of results. If there are no more pages left in the result set, a false-valued flag is
// returned.
func (r *queryResolver) pageLocalLocations(
	ctx context.Context,
	getLocations getLocationsFn,
	adjustedUploads []adjustedUpload,
	cursor *localCursor,
	limit int,
	trace observation.TraceLogger,
) ([]lsifstore.Location, bool, error) {
	var allLocations []lsifstore.Location
	for i := range adjustedUploads {
		if len(allLocations) >= limit {
			// We've filled the page
			break
		}
		if i < cursor.UploadOffset {
			// Skip indexes we've searched completely
			continue
		}

		locations, totalCount, err := getLocations(
			ctx,
			adjustedUploads[i].Upload.ID,
			adjustedUploads[i].AdjustedPathInBundle,
			adjustedUploads[i].AdjustedPosition.Line,
			adjustedUploads[i].AdjustedPosition.Character,
			limit-len(allLocations),
			cursor.LocationOffset,
		)
		if err != nil {
			return nil, false, errors.Wrap(err, "in an lsifstore locations call")
		}

		numLocations := len(locations)
		trace.Log(log.Int("pageLocalLocations.numLocations", numLocations))
		cursor.LocationOffset += numLocations

		if cursor.LocationOffset >= totalCount {
			// Skip this index on next request
			cursor.LocationOffset = 0
			cursor.UploadOffset++
		}

		allLocations = append(allLocations, locations...)
	}

	return allLocations, cursor.UploadOffset < len(adjustedUploads), nil
}

// pageRemoteLocations returns a slice of the (remote) result set denoted by the given cursor fulfilled by
// performing a moniker search over a group of indexes. The given cursor will be adjusted to reflect the
// offsets required to resolve the next page of results. If there are no more pages left in the result set,
// a false-valued flag is returned.
func (r *queryResolver) pageRemoteLocations(
	ctx context.Context,
	lsifDataTable string,
	adjustedUploads []adjustedUpload,
	orderedMonikers []precise.QualifiedMonikerData,
	cursor *remoteCursor,
	limit int,
	trace observation.TraceLogger,
) ([]lsifstore.Location, bool, error) {
	for len(cursor.UploadBatchIDs) == 0 {
		if cursor.UploadOffset < 0 {
			// No more batches
			return nil, false, nil
		}

		ignoreIDs := []int{}
		for _, adjustedUpload := range adjustedUploads {
			ignoreIDs = append(ignoreIDs, adjustedUpload.Upload.ID)
		}

		// Find the next batch of indexes to perform a moniker search over
		referenceUploadIDs, recordsScanned, totalRecords, err := r.uploadIDsWithReferences(
			ctx,
			orderedMonikers,
			ignoreIDs,
			r.maximumIndexesPerMonikerSearch,
			cursor.UploadOffset,
			trace,
		)
		if err != nil {
			return nil, false, err
		}

		cursor.UploadBatchIDs = referenceUploadIDs
		cursor.UploadOffset += recordsScanned

		if cursor.UploadOffset >= totalRecords {
			// Signal no batches remaining
			cursor.UploadOffset = -1
		}
	}

	// Fetch the upload records we don't currently have hydrated and insert them into the map
	monikerSearchUploads, err := r.uploadsByIDs(ctx, cursor.UploadBatchIDs)
	if err != nil {
		return nil, false, err
	}

	// Perform the moniker search
	locations, totalCount, err := r.monikerLocations(ctx, monikerSearchUploads, orderedMonikers, lsifDataTable, limit, cursor.LocationOffset)
	if err != nil {
		return nil, false, err
	}

	numLocations := len(locations)
	trace.Log(log.Int("pageLocalLocations.numLocations", numLocations))
	cursor.LocationOffset += numLocations

	if cursor.LocationOffset >= totalCount {
		// Require a new batch on next page
		cursor.LocationOffset = 0
		cursor.UploadBatchIDs = []int{}
	}

	// Perform an in-place filter to remove specific duplicate locations. Ranges that enclose the
	// target position will be returned by both an LSIF graph traversal as well as a moniker search.
	// We remove the latter instances.

	filtered := locations[:0]

	for _, location := range locations {
		if !isSourceLocation(adjustedUploads, location) {
			filtered = append(filtered, location)
		}
	}

	// We have another page if we still have results in the current batch of reference indexes, or if
	// we can query a next batch of reference indexes. We may return true here when we are actually
	// out of references. This behavior may change in the future.
	return filtered, len(cursor.UploadBatchIDs) > 0 || cursor.UploadOffset >= 0, nil
}

// isSourceLocation returns true if the given location encloses the source position within one of the visible uploads.
func isSourceLocation(adjustedUploads []adjustedUpload, location lsifstore.Location) bool {
	for i := range adjustedUploads {
		if location.DumpID == adjustedUploads[i].Upload.ID && location.Path == adjustedUploads[i].AdjustedPath {
			if rangeContainsPosition(location.Range, adjustedUploads[i].AdjustedPosition) {
				return true
			}
		}
	}

	return false
}

// rangeContainsPosition returns true if the given range encloses the given position.
func rangeContainsPosition(r lsifstore.Range, pos lsifstore.Position) bool {
	if pos.Line < r.Start.Line {
		return false
	}

	if pos.Line > r.End.Line {
		return false
	}

	if pos.Line == r.Start.Line && pos.Character < r.Start.Character {
		return false
	}

	if pos.Line == r.End.Line && pos.Character > r.End.Character {
		return false
	}

	return true
}

// uploadIDsWithReferences returns uploads that probably contain an import
// or implementation moniker whose identifier matches any of the given monikers' identifiers. This method
// will not return uploads for commits which are unknown to gitserver, nor will it return uploads which
// are listed in the given ignored identifier slice. This method also returns the number of records
// scanned (but possibly filtered out from the return slice) from the database (the offset for the
// subsequent request) and the total number of records in the database.
func (r *queryResolver) uploadIDsWithReferences(
	ctx context.Context,
	orderedMonikers []precise.QualifiedMonikerData,
	ignoreIDs []int,
	limit int,
	offset int,
	trace observation.TraceLogger,
) (ids []int, recordsScanned int, totalCount int, err error) {
	scanner, totalCount, err := r.dbStore.ReferenceIDs(ctx, r.repositoryID, r.commit, orderedMonikers, limit, offset)
	if err != nil {
		return nil, 0, 0, errors.Wrap(err, "dbstore.ReferenceIDs")
	}

	defer func() {
		if closeErr := scanner.Close(); closeErr != nil {
			err = errors.Append(err, errors.Wrap(closeErr, "dbstore.ReferenceIDs.Close"))
		}
	}()

	ignoreIDsMap := map[int]struct{}{}
	for _, id := range ignoreIDs {
		ignoreIDsMap[id] = struct{}{}
	}

	filtered := map[int]struct{}{}

	for len(filtered) < limit {
		packageReference, exists, err := scanner.Next()
		if err != nil {
			return nil, 0, 0, errors.Wrap(err, "dbstore.ReferenceIDs.Next")
		}
		if !exists {
			break
		}
		recordsScanned++

		if _, ok := filtered[packageReference.DumpID]; ok {
			// This index includes a definition so we can skip testing the filters here. The index
			// will be included in the moniker search regardless if it contains additional references.
			continue
		}

		if _, ok := ignoreIDsMap[packageReference.DumpID]; ok {
			// Ignore this dump
			continue
		}

		filtered[packageReference.DumpID] = struct{}{}
	}

	trace.Log(
		log.Int("uploadIDsWithReferences.numFiltered", len(filtered)),
		log.Int("uploadIDsWithReferences.numRecordsScanned", recordsScanned),
	)

	flattened := make([]int, 0, len(filtered))
	for k := range filtered {
		flattened = append(flattened, k)
	}
	sort.Ints(flattened)

	return flattened, recordsScanned, totalCount, nil
}

// uploadsByIDs returns a slice of uploads with the given identifiers. This method will not return a
// new upload record for a commit which is unknown to gitserver. The given upload map is used as a
// caching mechanism - uploads present in the map are not fetched again from the database.
func (r *queryResolver) uploadsByIDs(ctx context.Context, ids []int) ([]store.Dump, error) {
	missingIDs := make([]int, 0, len(ids))
	existingUploads := make([]store.Dump, 0, len(ids))

	for _, id := range ids {
		if upload, ok := r.uploadCache[id]; ok {
			existingUploads = append(existingUploads, upload)
		} else {
			missingIDs = append(missingIDs, id)
		}
	}

	uploads, err := r.dbStore.GetDumpsByIDs(ctx, missingIDs)
	if err != nil {
		return nil, errors.Wrap(err, "dbstore.GetDumpsByIDs")
	}

	newUploads, err := filterUploadsWithCommits(ctx, r.cachedCommitChecker, uploads)
	if err != nil {
		return nil, nil
	}

	for i := range newUploads {
		r.uploadCache[newUploads[i].ID] = newUploads[i]
	}

	return append(existingUploads, newUploads...), nil
}
