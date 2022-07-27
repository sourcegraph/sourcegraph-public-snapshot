package graphql

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	codeintelgitserver "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const slowImplementationsRequestThreshold = time.Second

// Implementations returns the list of source locations that define the symbol at the given position.
func (r *resolver) Implementations(ctx context.Context, args shared.RequestArgs) (_ []shared.UploadLocation, nextCursor string, err error) {
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.references, slowImplementationsRequestThreshold, observation.Args{
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
	cursor, err := decodeImplementationsCursor(args.RawCursor)
	if err != nil {
		return nil, "", errors.Wrap(err, fmt.Sprintf("invalid cursor: %q", args.RawCursor))
	}

	impls, implsCursor, err := r.svc.GetImplementations(ctx, args, r.requestState, cursor)
	if err != nil {
		return nil, "", errors.Wrap(err, "svc.GetImplementations")
	}

	if cursor.Phase != "done" {
		nextCursor = encodeImplementationsCursor(implsCursor)
	}

	return impls, nextCursor, nil
}

// ErrConcurrentModification occurs when a page of a references request cannot be resolved as
// the set of visible uploads have changed since the previous request for the same result set.
var ErrConcurrentModification = errors.New("result set changed while paginating")

// getVisibleUploadsFromCursor returns the current target path and the given position for each upload
// visible from the current target commit. If an upload cannot be adjusted, it will be omitted from
// the returned slice. The returned slice will be cached on the given cursor. If this data is already
// stashed on the given cursor, the result is recalculated from the cursor data/resolver context, and
// we don't need to hit the database.
//
// An error is returned if the set of visible uploads has changed since the previous request of this
// result set (specifically if an index becomes invisible). This behavior may change in the future.
func (r *resolver) getVisibleUploadsFromCursor(ctx context.Context, line, character int, cursorsToVisibleUploads *[]shared.CursorToVisibleUpload) ([]visibleUpload, []shared.CursorToVisibleUpload, error) {
	if *cursorsToVisibleUploads != nil {
		visibleUploads := make([]visibleUpload, 0, len(*cursorsToVisibleUploads))
		for _, u := range *cursorsToVisibleUploads {
			upload, ok := r.dataLoader.getUploadFromCacheMap(u.DumpID)
			if !ok {
				return nil, nil, ErrConcurrentModification
			}

			visibleUploads = append(visibleUploads, visibleUpload{
				Upload:                upload,
				TargetPath:            u.TargetPath,
				TargetPosition:        u.TargetPosition,
				TargetPathWithoutRoot: u.TargetPathWithoutRoot,
			})
		}

		return visibleUploads, *cursorsToVisibleUploads, nil
	}

	visibleUploads, err := r.getVisibleUploads(ctx, line, character)
	if err != nil {
		return nil, nil, err
	}

	updatedCursorsToVisibleUploads := make([]shared.CursorToVisibleUpload, 0, len(visibleUploads))
	for i := range visibleUploads {
		updatedCursorsToVisibleUploads = append(updatedCursorsToVisibleUploads, shared.CursorToVisibleUpload{
			DumpID:                visibleUploads[i].Upload.ID,
			TargetPath:            visibleUploads[i].TargetPath,
			TargetPosition:        visibleUploads[i].TargetPosition,
			TargetPathWithoutRoot: visibleUploads[i].TargetPathWithoutRoot,
		})
	}

	return visibleUploads, updatedCursorsToVisibleUploads, nil
}

// getVisibleUploads adjusts the current target path and the given position for each upload visible
// from the current target commit. If an upload cannot be adjusted, it will be omitted from the
// returned slice.
func (r *resolver) getVisibleUploads(ctx context.Context, line, character int) ([]visibleUpload, error) {
	visibleUploads := make([]visibleUpload, 0, len(r.dataLoader.uploads))
	for i := range r.dataLoader.uploads {
		adjustedUpload, ok, err := r.getVisibleUpload(ctx, line, character, r.dataLoader.uploads[i])
		if err != nil {
			return nil, err
		}
		if ok {
			visibleUploads = append(visibleUploads, adjustedUpload)
		}
	}

	return visibleUploads, nil
}

// getVisibleUpload returns the current target path and the given position for the given upload. If
// the upload cannot be adjusted, a false-valued flag is returned.
func (r *resolver) getVisibleUpload(ctx context.Context, line, character int, upload shared.Dump) (visibleUpload, bool, error) {
	position := shared.Position{
		Line:      line,
		Character: character,
	}

	targetPath, targetPosition, ok, err := r.GitTreeTranslator.GetTargetCommitPositionFromSourcePosition(ctx, upload.Commit, position, false)
	if err != nil || !ok {
		return visibleUpload{}, false, errors.Wrap(err, "gitTreeTranslator.GetTargetCommitPositionFromSourcePosition")
	}

	return visibleUpload{
		Upload:                upload,
		TargetPath:            targetPath,
		TargetPosition:        targetPosition,
		TargetPathWithoutRoot: strings.TrimPrefix(targetPath, upload.Root),
	}, true, nil
}

type qualifiedMonikerSet struct {
	monikers       []precise.QualifiedMonikerData
	monikerHashMap map[string]struct{}
}

func newQualifiedMonikerSet() *qualifiedMonikerSet {
	return &qualifiedMonikerSet{
		monikerHashMap: map[string]struct{}{},
	}
}

// add the given qualified moniker to the set if it is distinct from all elements
// currently in the set.
func (s *qualifiedMonikerSet) add(qualifiedMoniker precise.QualifiedMonikerData) {
	monikerHash := strings.Join([]string{
		qualifiedMoniker.PackageInformationData.Name,
		qualifiedMoniker.PackageInformationData.Version,
		qualifiedMoniker.MonikerData.Scheme,
		qualifiedMoniker.MonikerData.Identifier,
	}, ":")

	if _, ok := s.monikerHashMap[monikerHash]; ok {
		return
	}

	s.monikerHashMap[monikerHash] = struct{}{}
	s.monikers = append(s.monikers, qualifiedMoniker)
}

// monikerLimit is the maximum number of monikers that can be returned from orderedMonikers.
const monikerLimit = 10

// getOrderedMonikers returns the set of monikers of the given kind(s) attached to the ranges specified by
// the given upload list.
//
// The return slice is ordered by visible upload, then by specificity, i.e., monikers attached to
// enclosed ranges before before monikers attached to enclosing ranges. Monikers are de-duplicated, such
// that the second (third, ...) occurrences are removed.
func (r *resolver) getOrderedMonikers(ctx context.Context, visibleUploads []visibleUpload, kinds ...string) ([]precise.QualifiedMonikerData, error) {
	monikerSet := newQualifiedMonikerSet()

	for i := range visibleUploads {
		rangeMonikers, err := r.svc.GetMonikersByPosition(
			ctx,
			visibleUploads[i].Upload.ID,
			visibleUploads[i].TargetPathWithoutRoot,
			visibleUploads[i].TargetPosition.Line,
			visibleUploads[i].TargetPosition.Character,
		)
		if err != nil {
			return nil, errors.Wrap(err, "lsifStore.MonikersByPosition")
		}

		for _, monikers := range rangeMonikers {
			for _, moniker := range monikers {
				if moniker.PackageInformationID == "" || !sliceContains(kinds, moniker.Kind) {
					continue
				}

				packageInformationData, _, err := r.svc.GetPackageInformation(
					ctx,
					visibleUploads[i].Upload.ID,
					visibleUploads[i].TargetPathWithoutRoot,
					string(moniker.PackageInformationID),
				)
				if err != nil {
					return nil, errors.Wrap(err, "lsifStore.PackageInformation")
				}

				monikerSet.add(precise.QualifiedMonikerData{
					MonikerData:            moniker,
					PackageInformationData: packageInformationData,
				})

				if len(monikerSet.monikers) >= monikerLimit {
					return monikerSet.monikers, nil
				}
			}
		}
	}

	return monikerSet.monikers, nil
}

// type getLocationsFn = func(ctx context.Context, bundleID int, path string, line int, character int, limit int, offset int) ([]shared.Location, int, error)

// // getPageLocalLocations returns a slice of the (local) result set denoted by the given cursor fulfilled by
// // traversing the LSIF graph. The given cursor will be adjusted to reflect the offsets required to resolve
// // the next page of results. If there are no more pages left in the result set, a false-valued flag is returned.
// func (r *resolver) getPageLocalLocations(ctx context.Context, getLocations getLocationsFn, visibleUploads []visibleUpload, cursor *localCursor, limit int, trace observation.TraceLogger) ([]shared.Location, bool, error) {
// 	var allLocations []shared.Location
// 	for i := range visibleUploads {
// 		if len(allLocations) >= limit {
// 			// We've filled the page
// 			break
// 		}
// 		if i < cursor.UploadOffset {
// 			// Skip indexes we've searched completely
// 			continue
// 		}

// 		locations, totalCount, err := getLocations(
// 			ctx,
// 			visibleUploads[i].Upload.ID,
// 			visibleUploads[i].TargetPathWithoutRoot,
// 			visibleUploads[i].TargetPosition.Line,
// 			visibleUploads[i].TargetPosition.Character,
// 			limit-len(allLocations),
// 			cursor.LocationOffset,
// 		)
// 		if err != nil {
// 			return nil, false, errors.Wrap(err, "in an lsifstore locations call")
// 		}

// 		numLocations := len(locations)
// 		trace.Log(log.Int("pageLocalLocations.numLocations", numLocations))
// 		cursor.LocationOffset += numLocations

// 		if cursor.LocationOffset >= totalCount {
// 			// Skip this index on next request
// 			cursor.LocationOffset = 0
// 			cursor.UploadOffset++
// 		}

// 		allLocations = append(allLocations, locations...)
// 	}

// 	return allLocations, cursor.UploadOffset < len(visibleUploads), nil
// }

// // getPageRemoteLocations returns a slice of the (remote) result set denoted by the given cursor fulfilled by
// // performing a moniker search over a group of indexes. The given cursor will be adjusted to reflect the
// // offsets required to resolve the next page of results. If there are no more pages left in the result set,
// // a false-valued flag is returned.
// func (r *resolver) getPageRemoteLocations(
// 	ctx context.Context,
// 	lsifDataTable string,
// 	visibleUploads []visibleUpload,
// 	orderedMonikers []precise.QualifiedMonikerData,
// 	cursor *remoteCursor,
// 	limit int,
// 	trace observation.TraceLogger,
// ) ([]shared.Location, bool, error) {
// 	for len(cursor.UploadBatchIDs) == 0 {
// 		if cursor.UploadOffset < 0 {
// 			// No more batches
// 			return nil, false, nil
// 		}

// 		ignoreIDs := []int{}
// 		for _, adjustedUpload := range visibleUploads {
// 			ignoreIDs = append(ignoreIDs, adjustedUpload.Upload.ID)
// 		}

// 		// Find the next batch of indexes to perform a moniker search over
// 		referenceUploadIDs, recordsScanned, totalRecords, err := r.svc.GetUploadIDsWithReferences(
// 			ctx,
// 			orderedMonikers,
// 			ignoreIDs,
// 			r.requestArgs.GetRepoID(),
// 			r.requestArgs.commit,
// 			r.maximumIndexesPerMonikerSearch,
// 			cursor.UploadOffset,
// 		)
// 		if err != nil {
// 			return nil, false, err
// 		}

// 		cursor.UploadBatchIDs = referenceUploadIDs
// 		cursor.UploadOffset += recordsScanned

// 		if cursor.UploadOffset >= totalRecords {
// 			// Signal no batches remaining
// 			cursor.UploadOffset = -1
// 		}
// 	}

// 	// Fetch the upload records we don't currently have hydrated and insert them into the map
// 	monikerSearchUploads, err := r.getUploadsByIDs(ctx, cursor.UploadBatchIDs)
// 	if err != nil {
// 		return nil, false, err
// 	}

// 	// Perform the moniker search
// 	locations, totalCount, err := r.getBulkMonikerLocations(ctx, monikerSearchUploads, orderedMonikers, lsifDataTable, limit, cursor.LocationOffset)
// 	if err != nil {
// 		return nil, false, err
// 	}

// 	numLocations := len(locations)
// 	trace.Log(log.Int("pageLocalLocations.numLocations", numLocations))
// 	cursor.LocationOffset += numLocations

// 	if cursor.LocationOffset >= totalCount {
// 		// Require a new batch on next page
// 		cursor.LocationOffset = 0
// 		cursor.UploadBatchIDs = []int{}
// 	}

// 	// Perform an in-place filter to remove specific duplicate locations. Ranges that enclose the
// 	// target position will be returned by both an LSIF graph traversal as well as a moniker search.
// 	// We remove the latter instances.

// 	filtered := locations[:0]

// 	for _, location := range locations {
// 		if !isSourceLocation(visibleUploads, location) {
// 			filtered = append(filtered, location)
// 		}
// 	}

// 	// We have another page if we still have results in the current batch of reference indexes, or if
// 	// we can query a next batch of reference indexes. We may return true here when we are actually
// 	// out of references. This behavior may change in the future.
// 	hasAnotherPage := len(cursor.UploadBatchIDs) > 0 || cursor.UploadOffset >= 0

// 	return filtered, hasAnotherPage, nil
// }

// getUploadsWithDefinitionsForMonikers returns the set of uploads that provide any of the given monikers.
// This method will not return uploads for commits which are unknown to gitserver.
func (r *resolver) getUploadsWithDefinitionsForMonikers(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData) ([]shared.Dump, error) {
	uploads, err := r.svc.GetUploadsWithDefinitionsForMonikers(ctx, orderedMonikers)
	if err != nil {
		return nil, errors.Wrap(err, "dbstore.DefinitionDumps")
	}

	r.dataLoader.setUploadInCacheMap(uploads)

	uploadsWithResolvableCommits, err := r.removeUploadsWithUnknownCommits(ctx, uploads)
	if err != nil {
		return nil, err
	}

	return uploadsWithResolvableCommits, nil
}

// removeUploadsWithUnknownCommits removes uploads for commits which are unknown to gitserver from the given
// slice. The slice is filtered in-place and returned (to update the slice length).
func (r *resolver) removeUploadsWithUnknownCommits(ctx context.Context, uploads []shared.Dump) ([]shared.Dump, error) {
	rcs := make([]codeintelgitserver.RepositoryCommit, 0, len(uploads))
	for _, upload := range uploads {
		rcs = append(rcs, codeintelgitserver.RepositoryCommit{
			RepositoryID: upload.RepositoryID,
			Commit:       upload.Commit,
		})
	}
	exists, err := r.commitCache.AreCommitsResolvable(ctx, rcs)
	if err != nil {
		return nil, err
	}

	filtered := uploads[:0]
	for i, upload := range uploads {
		if exists[i] {
			filtered = append(filtered, upload)
		}
	}

	return filtered, nil
}

// getBulkMonikerLocations returns the set of locations (within the given uploads) with an attached moniker
// whose scheme+identifier matches any of the given monikers.
func (r *resolver) getBulkMonikerLocations(ctx context.Context, uploads []shared.Dump, orderedMonikers []precise.QualifiedMonikerData, tableName string, limit, offset int) ([]shared.Location, int, error) {
	ids := make([]int, 0, len(uploads))
	for i := range uploads {
		ids = append(ids, uploads[i].ID)
	}

	args := make([]precise.MonikerData, 0, len(orderedMonikers))
	for _, moniker := range orderedMonikers {
		args = append(args, moniker.MonikerData)
	}

	locations, totalCount, err := r.svc.GetBulkMonikerLocations(ctx, tableName, ids, args, limit, offset)
	if err != nil {
		return nil, 0, errors.Wrap(err, "lsifStore.GetBulkMonikerLocations")
	}

	return locations, totalCount, nil
}

// getUploadsByIDs returns a slice of uploads with the given identifiers. This method will not return a
// new upload record for a commit which is unknown to gitserver. The given upload map is used as a
// caching mechanism - uploads present in the map are not fetched again from the database.
func (r *resolver) getUploadsByIDs(ctx context.Context, ids []int) ([]shared.Dump, error) {
	missingIDs := make([]int, 0, len(ids))
	existingUploads := make([]shared.Dump, 0, len(ids))

	for _, id := range ids {
		if upload, ok := r.dataLoader.getUploadFromCacheMap(id); ok {
			existingUploads = append(existingUploads, upload)
		} else {
			missingIDs = append(missingIDs, id)
		}
	}

	uploads, err := r.svc.GetDumpsByIDs(ctx, missingIDs)
	if err != nil {
		return nil, errors.Wrap(err, "service.GetDumpsByIDs")
	}

	uploadsWithResolvableCommits, err := r.removeUploadsWithUnknownCommits(ctx, uploads)
	if err != nil {
		return nil, nil
	}
	r.dataLoader.setUploadInCacheMap(uploadsWithResolvableCommits)

	allUploads := append(existingUploads, uploadsWithResolvableCommits...)

	return allUploads, nil
}

// getUploadLocations translates a set of locations into an equivalent set of locations in the requested
// commit.
func (r *resolver) getUploadLocations(ctx context.Context, locations []shared.Location) ([]shared.UploadLocation, error) {
	uploadLocations := make([]shared.UploadLocation, 0, len(locations))

	checkerEnabled := authz.SubRepoEnabled(r.authChecker)
	var a *actor.Actor
	if checkerEnabled {
		a = actor.FromContext(ctx)
	}
	for _, location := range locations {
		upload, ok := r.dataLoader.getUploadFromCacheMap(location.DumpID)
		if !ok {
			continue
		}

		adjustedLocation, err := r.getUploadLocation(ctx, upload, location)
		if err != nil {
			return nil, err
		}

		if !checkerEnabled {
			uploadLocations = append(uploadLocations, adjustedLocation)
		} else {
			repo := api.RepoName(adjustedLocation.Dump.RepositoryName)
			if include, err := authz.FilterActorPath(ctx, r.authChecker, a, repo, adjustedLocation.Path); err != nil {
				return nil, err
			} else if include {
				uploadLocations = append(uploadLocations, adjustedLocation)
			}
		}
	}

	return uploadLocations, nil
}

// getUploadLocation translates a location (relative to the indexed commit) into an equivalent location in
// the requested commit. If the translation fails, then the original commit and range are used as the
// commit and range of the adjusted location.
func (r *resolver) getUploadLocation(ctx context.Context, dump shared.Dump, location shared.Location) (shared.UploadLocation, error) {
	adjustedCommit, adjustedRange, _, err := r.getSourceRange(ctx, dump.RepositoryID, dump.Commit, dump.Root+location.Path, location.Range)
	if err != nil {
		return shared.UploadLocation{}, err
	}

	return shared.UploadLocation{
		Dump:         dump,
		Path:         dump.Root + location.Path,
		TargetCommit: adjustedCommit,
		TargetRange:  adjustedRange,
	}, nil
}

// getSourceRange translates a range (relative to the indexed commit) into an equivalent range in the requested
// commit. If the translation fails, then the original commit and range are returned along with a false-valued
// flag.
func (r *resolver) getSourceRange(ctx context.Context, repositoryID int, commit, path string, rng shared.Range) (string, shared.Range, bool, error) {
	if repositoryID != r.requestArgs.GetRepoID() {
		// No diffs between distinct repositories
		return commit, rng, true, nil
	}

	if _, sourceRange, ok, err := r.GitTreeTranslator.GetTargetCommitRangeFromSourceRange(ctx, commit, path, rng, true); err != nil {
		return "", shared.Range{}, false, errors.Wrap(err, "gitTreeTranslator.GetTargetCommitRangeFromSourceRange")
	} else if ok {
		return r.requestArgs.commit, sourceRange, true, nil
	}

	return commit, rng, false, nil
}

func uploadIDsToString(vs []shared.Dump) string {
	ids := make([]string, 0, len(vs))
	for _, v := range vs {
		ids = append(ids, strconv.Itoa(v.ID))
	}

	return strings.Join(ids, ", ")
}

func sliceContains(slice []string, str string) bool {
	for _, el := range slice {
		if el == str {
			return true
		}
	}
	return false
}

func monikersToString(vs []precise.QualifiedMonikerData) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, fmt.Sprintf("%s:%s:%s:%s", v.Kind, v.Scheme, v.Identifier, v.Version))
	}

	return strings.Join(strs, ", ")
}

// isSourceLocation returns true if the given location encloses the source position within one of the visible uploads.
func isSourceLocation(visibleUploads []visibleUpload, location shared.Location) bool {
	for i := range visibleUploads {
		if location.DumpID == visibleUploads[i].Upload.ID && location.Path == visibleUploads[i].TargetPath {
			if rangeContainsPosition(location.Range, visibleUploads[i].TargetPosition) {
				return true
			}
		}
	}

	return false
}

// rangeContainsPosition returns true if the given range encloses the given position.
func rangeContainsPosition(r shared.Range, pos shared.Position) bool {
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
