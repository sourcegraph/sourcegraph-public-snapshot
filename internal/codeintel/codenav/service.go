package codenav

import (
	"context"
	"strings"

	traceLog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	codeintelgitserver "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	uploads "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/symbols"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ service = (*Service)(nil)

type service interface {
	GetDefinitions(ctx context.Context, args shared.RequestArgs, requestState RequestState) (_ []shared.UploadLocation, err error)
	GetDiagnostics(ctx context.Context, args shared.RequestArgs, requestState RequestState) (diagnosticsAtUploads []shared.DiagnosticAtUpload, _ int, err error)
	GetHover(ctx context.Context, args shared.RequestArgs, requestState RequestState) (_ string, _ shared.Range, _ bool, err error)
	GetImplementations(ctx context.Context, args shared.RequestArgs, requestState RequestState, cursor shared.ImplementationsCursor) (_ []shared.UploadLocation, nextCursor shared.ImplementationsCursor, err error)
	GetRanges(ctx context.Context, args shared.RequestArgs, requestState RequestState, startLine, endLine int) (adjustedRanges []shared.AdjustedCodeIntelligenceRange, err error)
	GetReferences(ctx context.Context, args shared.RequestArgs, requestState RequestState, cursor shared.ReferencesCursor) (_ []shared.UploadLocation, nextCursor shared.ReferencesCursor, err error)
	GetStencil(ctx context.Context, args shared.RequestArgs, requestState RequestState) (adjustedRanges []shared.Range, err error)

	GetMonikersByPosition(ctx context.Context, bundleID int, path string, line, character int) (_ [][]precise.MonikerData, err error)
	GetBulkMonikerLocations(ctx context.Context, tableName string, uploadIDs []int, monikers []precise.MonikerData, limit, offset int) (_ []shared.Location, _ int, err error)
	GetPackageInformation(ctx context.Context, bundleID int, path, packageInformationID string) (_ precise.PackageInformationData, _ bool, err error)
	GetClosestDumpsForBlob(ctx context.Context, repositoryID int, commit, path string, exactPath bool, indexer string) (_ []shared.Dump, err error)

	// Symbols client
	GetSupportedByCtags(ctx context.Context, filepath string, repoName api.RepoName) (bool, string, error)

	// Language Support
	GetLanguagesRequestedBy(ctx context.Context, userID int) (_ []string, err error)
	SetRequestLanguageSupport(ctx context.Context, userID int, language string) (err error)

	// Uploads Service
	GetDumpsByIDs(ctx context.Context, ids []int) (_ []shared.Dump, err error)
	GetUploadsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QualifiedMonikerData) (_ []shared.Dump, err error)
	GetUploadIDsWithReferences(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData, ignoreIDs []int, repositoryID int, commit string, limit int, offset int) (ids []int, recordsScanned int, totalCount int, err error)
}

type Service struct {
	store         store.Store
	lsifstore     lsifstore.LsifStore
	gitserver     GitserverClient
	symbolsClient *symbols.Client
	uploadSvc     UploadService
	operations    *operations
}

func newService(store store.Store, lsifstore lsifstore.LsifStore, uploadSvc UploadService, gitserver GitserverClient, symbolsClient *symbols.Client, observationContext *observation.Context) *Service {
	return &Service{
		store:         store,
		lsifstore:     lsifstore,
		gitserver:     gitserver,
		symbolsClient: symbolsClient,
		uploadSvc:     uploadSvc,
		operations:    newOperations(observationContext),
	}
}

// GetHover returns the set of locations defining the symbol at the given position.
func (s *Service) GetHover(ctx context.Context, args shared.RequestArgs, requestState RequestState) (_ string, _ shared.Range, _ bool, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, s.operations.getHover, serviceObserverThreshold, observation.Args{
		LogFields: []traceLog.Field{
			traceLog.Int("repositoryID", args.RepositoryID),
			traceLog.String("commit", args.Commit),
			traceLog.String("path", args.Path),
			traceLog.Int("numUploads", len(requestState.GetCacheUploads())),
			traceLog.String("uploads", uploadIDsToString(requestState.GetCacheUploads())),
			traceLog.Int("line", args.Line),
			traceLog.Int("character", args.Character),
		},
	})
	defer endObservation()

	adjustedUploads, err := s.getVisibleUploads(ctx, args.Line, args.Character, requestState)
	if err != nil {
		return "", shared.Range{}, false, err
	}

	// Keep track of each adjusted range we know about enclosing the requested position.
	//
	// If we don't have hover text within the index where the range is defined, we'll
	// have to look in the definition index and search for the text there. We don't
	// want to return the range associated with the definition, as the range is used
	// as a hint to highlight a range in the current document.
	adjustedRanges := make([]shared.Range, 0, len(adjustedUploads))

	cachedUploads := requestState.GetCacheUploads()
	for i := range adjustedUploads {
		adjustedUpload := adjustedUploads[i]
		trace.Log(traceLog.Int("uploadID", adjustedUpload.Upload.ID))

		// Fetch hover text from the index
		text, rn, exists, err := s.lsifstore.GetHover(
			ctx,
			adjustedUpload.Upload.ID,
			adjustedUpload.TargetPathWithoutRoot,
			adjustedUpload.TargetPosition.Line,
			adjustedUpload.TargetPosition.Character,
		)
		if err != nil {
			return "", shared.Range{}, false, errors.Wrap(err, "lsifStore.Hover")
		}
		if !exists {
			continue
		}

		// Adjust the highlighted range back to the appropriate range in the target commit
		_, adjustedRange, _, err := s.getSourceRange(ctx, args, requestState, cachedUploads[i].RepositoryID, cachedUploads[i].Commit, args.Path, rn)
		if err != nil {
			return "", shared.Range{}, false, err
		}
		if text != "" {
			// Text attached to source range
			return text, adjustedRange, true, nil
		}

		adjustedRanges = append(adjustedRanges, adjustedRange)
	}

	// The Slow path:
	//
	// The indexes we searched in doesn't attach hover text to externally defined symbols.
	// Each indexer is free to make that choice as it's a compromise between ease of development,
	// efficiency of indexing, index output sizes, etc. We can deal with this situation by
	// looking for hover text attached to the precise definition (if one exists).

	// The range we will end up returning is interpreted within the context of the current text
	// document, so any range inside of a remote index would be of no use. We'll return the first
	// (inner-most) range that we adjusted from the source index traversals above.
	var adjustedRange shared.Range
	if len(adjustedRanges) > 0 {
		adjustedRange = adjustedRanges[0]
	}

	// Gather all import monikers attached to the ranges enclosing the requested position
	orderedMonikers, err := s.getOrderedMonikers(ctx, adjustedUploads, "import")
	if err != nil {
		return "", shared.Range{}, false, err
	}
	trace.Log(
		traceLog.Int("numMonikers", len(orderedMonikers)),
		traceLog.String("monikers", monikersToString(orderedMonikers)),
	)

	// Determine the set of uploads over which we need to perform a moniker search. This will
	// include all all indexes which define one of the ordered monikers. This should not include
	// any of the indexes we have already performed an LSIF graph traversal in above.
	uploads, err := s.getUploadsWithDefinitionsForMonikers(ctx, orderedMonikers, requestState)
	if err != nil {
		return "", shared.Range{}, false, err
	}
	trace.Log(
		traceLog.Int("numDefinitionUploads", len(uploads)),
		traceLog.String("definitionUploads", uploadIDsToString(uploads)),
	)

	// Perform the moniker search. This returns a set of locations defining one of the monikers
	// attached to one of the source ranges.
	locations, _, err := s.getBulkMonikerLocations(ctx, uploads, orderedMonikers, "definitions", DefinitionsLimit, 0)
	if err != nil {
		return "", shared.Range{}, false, err
	}
	trace.Log(traceLog.Int("numLocations", len(locations)))

	for i := range locations {
		// Fetch hover text attached to a definition in the defining index
		text, _, exists, err := s.lsifstore.GetHover(
			ctx,
			locations[i].DumpID,
			locations[i].Path,
			locations[i].Range.Start.Line,
			locations[i].Range.Start.Character,
		)
		if err != nil {
			return "", shared.Range{}, false, errors.Wrap(err, "lsifStore.Hover")
		}
		if exists && text != "" {
			// Text attached to definition
			return text, adjustedRange, true, nil
		}
	}

	// No text available
	return "", shared.Range{}, false, nil
}

// GetReferences returns the list of source locations that reference the symbol at the given position.
func (s *Service) GetReferences(ctx context.Context, args shared.RequestArgs, requestState RequestState, cursor shared.ReferencesCursor) (_ []shared.UploadLocation, _ shared.ReferencesCursor, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, s.operations.getReferences, serviceObserverThreshold, observation.Args{
		LogFields: []traceLog.Field{
			traceLog.Int("repositoryID", args.RepositoryID),
			traceLog.String("commit", args.Commit),
			traceLog.String("path", args.Path),
			traceLog.Int("numUploads", len(requestState.GetCacheUploads())),
			traceLog.String("uploads", uploadIDsToString(requestState.GetCacheUploads())),
			traceLog.Int("line", args.Line),
			traceLog.Int("character", args.Character),
		},
	})
	defer endObservation()

	// Adjust the path and position for each visible upload based on its git difference to
	// the target commit. This data may already be stashed in the cursor decoded above, in
	// which case we don't need to hit the database.

	// References at the given file:line:character could come from multiple uploads, so we
	// need to look in all uploads and merge the results.
	adjustedUploads, cursorsToVisibleUploads, err := s.getVisibleUploadsFromCursor(ctx, args.Line, args.Character, &cursor.CursorsToVisibleUploads, requestState)
	if err != nil {
		return nil, cursor, err
	}

	// Update the cursors with the updated visible uploads.
	cursor.CursorsToVisibleUploads = cursorsToVisibleUploads

	// Gather all monikers attached to the ranges enclosing the requested position. This data
	// may already be stashed in the cursor decoded above, in which case we don't need to hit
	// the database.
	if cursor.OrderedMonikers == nil {
		if cursor.OrderedMonikers, err = s.getOrderedMonikers(ctx, adjustedUploads, "import", "export"); err != nil {
			return nil, cursor, err
		}
	}
	trace.Log(
		traceLog.Int("numMonikers", len(cursor.OrderedMonikers)),
		traceLog.String("monikers", monikersToString(cursor.OrderedMonikers)),
	)

	// Phase 1: Gather all "local" locations via LSIF graph traversal. We'll continue to request additional
	// locations until we fill an entire page (the size of which is denoted by the given limit) or there are
	// no more local results remaining.
	var locations []shared.Location
	if cursor.Phase == "local" {
		localLocations, hasMore, err := s.getPageLocalLocations(
			ctx,
			s.lsifstore.GetReferenceLocations,
			adjustedUploads,
			&cursor.LocalCursor,
			args.Limit-len(locations),
			trace,
		)
		if err != nil {
			return nil, cursor, err
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
			definitionUploads, err := s.getUploadsWithDefinitionsForMonikers(ctx, cursor.OrderedMonikers, requestState)
			if err != nil {
				return nil, cursor, err
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
			remoteLocations, hasMore, err := s.getPageRemoteLocations(ctx, "references", adjustedUploads, cursor.OrderedMonikers, &cursor.RemoteCursor, args.Limit-len(locations), trace, args, requestState)
			if err != nil {
				return nil, cursor, err
			}
			locations = append(locations, remoteLocations...)

			if !hasMore {
				cursor.Phase = "done"
				break
			}
		}
	}

	trace.Log(traceLog.Int("numLocations", len(locations)))

	// Adjust the locations back to the appropriate range in the target commits. This adjusts
	// locations within the repository the user is browsing so that it appears all references
	// are occurring at the same commit they are looking at.
	referenceLocations, err := s.getUploadLocations(ctx, args, requestState, locations)
	if err != nil {
		return nil, cursor, err
	}
	trace.Log(traceLog.Int("numReferenceLocations", len(referenceLocations)))

	return referenceLocations, cursor, nil
}

// getUploadsWithDefinitionsForMonikers returns the set of uploads that provide any of the given monikers.
// This method will not return uploads for commits which are unknown to gitserver.
func (s *Service) getUploadsWithDefinitionsForMonikers(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData, requestState RequestState) ([]shared.Dump, error) {
	uploads, err := s.GetUploadsWithDefinitionsForMonikers(ctx, orderedMonikers)
	if err != nil {
		return nil, errors.Wrap(err, "dbstore.DefinitionDumps")
	}

	requestState.dataLoader.SetUploadInCacheMap(uploads)

	uploadsWithResolvableCommits, err := s.removeUploadsWithUnknownCommits(ctx, uploads, requestState)
	if err != nil {
		return nil, err
	}

	return uploadsWithResolvableCommits, nil
}

// monikerLimit is the maximum number of monikers that can be returned from orderedMonikers.
const monikerLimit = 10

func (r *Service) getOrderedMonikers(ctx context.Context, visibleUploads []visibleUpload, kinds ...string) ([]precise.QualifiedMonikerData, error) {
	monikerSet := newQualifiedMonikerSet()

	for i := range visibleUploads {
		rangeMonikers, err := r.GetMonikersByPosition(
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

				packageInformationData, _, err := r.GetPackageInformation(
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

type getLocationsFn = func(ctx context.Context, bundleID int, path string, line int, character int, limit int, offset int) ([]shared.Location, int, error)

// getPageLocalLocations returns a slice of the (local) result set denoted by the given cursor fulfilled by
// traversing the LSIF graph. The given cursor will be adjusted to reflect the offsets required to resolve
// the next page of results. If there are no more pages left in the result set, a false-valued flag is returned.
func (s *Service) getPageLocalLocations(ctx context.Context, getLocations getLocationsFn, visibleUploads []visibleUpload, cursor *shared.LocalCursor, limit int, trace observation.TraceLogger) ([]shared.Location, bool, error) {
	var allLocations []shared.Location
	for i := range visibleUploads {
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
			visibleUploads[i].Upload.ID,
			visibleUploads[i].TargetPathWithoutRoot,
			visibleUploads[i].TargetPosition.Line,
			visibleUploads[i].TargetPosition.Character,
			limit-len(allLocations),
			cursor.LocationOffset,
		)
		if err != nil {
			return nil, false, errors.Wrap(err, "in an lsifstore locations call")
		}

		numLocations := len(locations)
		trace.Log(traceLog.Int("pageLocalLocations.numLocations", numLocations))
		cursor.LocationOffset += numLocations

		if cursor.LocationOffset >= totalCount {
			// Skip this index on next request
			cursor.LocationOffset = 0
			cursor.UploadOffset++
		}

		allLocations = append(allLocations, locations...)
	}

	return allLocations, cursor.UploadOffset < len(visibleUploads), nil
}

// getPageRemoteLocations returns a slice of the (remote) result set denoted by the given cursor fulfilled by
// performing a moniker search over a group of indexes. The given cursor will be adjusted to reflect the
// offsets required to resolve the next page of results. If there are no more pages left in the result set,
// a false-valued flag is returned.
func (s *Service) getPageRemoteLocations(
	ctx context.Context,
	lsifDataTable string,
	visibleUploads []visibleUpload,
	orderedMonikers []precise.QualifiedMonikerData,
	cursor *shared.RemoteCursor,
	limit int,
	trace observation.TraceLogger,
	args shared.RequestArgs,
	requestState RequestState,
) ([]shared.Location, bool, error) {
	for len(cursor.UploadBatchIDs) == 0 {
		if cursor.UploadOffset < 0 {
			// No more batches
			return nil, false, nil
		}

		ignoreIDs := []int{}
		for _, adjustedUpload := range visibleUploads {
			ignoreIDs = append(ignoreIDs, adjustedUpload.Upload.ID)
		}

		// Find the next batch of indexes to perform a moniker search over
		referenceUploadIDs, recordsScanned, totalRecords, err := s.GetUploadIDsWithReferences(
			ctx,
			orderedMonikers,
			ignoreIDs,
			args.RepositoryID,
			args.Commit,
			requestState.maximumIndexesPerMonikerSearch,
			cursor.UploadOffset,
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
	monikerSearchUploads, err := s.getUploadsByIDs(ctx, cursor.UploadBatchIDs, requestState)
	if err != nil {
		return nil, false, err
	}

	// Perform the moniker search
	locations, totalCount, err := s.getBulkMonikerLocations(ctx, monikerSearchUploads, orderedMonikers, lsifDataTable, limit, cursor.LocationOffset)
	if err != nil {
		return nil, false, err
	}

	numLocations := len(locations)
	trace.Log(traceLog.Int("pageLocalLocations.numLocations", numLocations))
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
		if !isSourceLocation(visibleUploads, location) {
			filtered = append(filtered, location)
		}
	}

	// We have another page if we still have results in the current batch of reference indexes, or if
	// we can query a next batch of reference indexes. We may return true here when we are actually
	// out of references. This behavior may change in the future.
	hasAnotherPage := len(cursor.UploadBatchIDs) > 0 || cursor.UploadOffset >= 0

	return filtered, hasAnotherPage, nil
}

// getUploadLocations translates a set of locations into an equivalent set of locations in the requested
// commit.
func (s *Service) getUploadLocations(ctx context.Context, args shared.RequestArgs, requestState RequestState, locations []shared.Location) ([]shared.UploadLocation, error) {
	uploadLocations := make([]shared.UploadLocation, 0, len(locations))

	checkerEnabled := authz.SubRepoEnabled(requestState.authChecker)
	var a *actor.Actor
	if checkerEnabled {
		a = actor.FromContext(ctx)
	}
	for _, location := range locations {
		upload, ok := requestState.dataLoader.GetUploadFromCacheMap(location.DumpID)
		if !ok {
			continue
		}

		adjustedLocation, err := s.getUploadLocation(ctx, args, requestState, upload, location)
		if err != nil {
			return nil, err
		}

		if !checkerEnabled {
			uploadLocations = append(uploadLocations, adjustedLocation)
		} else {
			repo := api.RepoName(adjustedLocation.Dump.RepositoryName)
			if include, err := authz.FilterActorPath(ctx, requestState.authChecker, a, repo, adjustedLocation.Path); err != nil {
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
func (s *Service) getUploadLocation(ctx context.Context, args shared.RequestArgs, requestState RequestState, dump shared.Dump, location shared.Location) (shared.UploadLocation, error) {
	adjustedCommit, adjustedRange, _, err := s.getSourceRange(ctx, args, requestState, dump.RepositoryID, dump.Commit, dump.Root+location.Path, location.Range)
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
func (s *Service) getSourceRange(ctx context.Context, args shared.RequestArgs, requestState RequestState, repositoryID int, commit, path string, rng shared.Range) (string, shared.Range, bool, error) {
	if repositoryID != args.RepositoryID {
		// No diffs between distinct repositories
		return commit, rng, true, nil
	}

	if _, sourceRange, ok, err := requestState.GitTreeTranslator.GetTargetCommitRangeFromSourceRange(ctx, commit, path, rng, true); err != nil {
		return "", shared.Range{}, false, errors.Wrap(err, "gitTreeTranslator.GetTargetCommitRangeFromSourceRange")
	} else if ok {
		return args.Commit, sourceRange, true, nil
	}

	return commit, rng, false, nil
}

// getUploadsByIDs returns a slice of uploads with the given identifiers. This method will not return a
// new upload record for a commit which is unknown to gitserver. The given upload map is used as a
// caching mechanism - uploads present in the map are not fetched again from the database.
func (s *Service) getUploadsByIDs(ctx context.Context, ids []int, requestState RequestState) ([]shared.Dump, error) {
	missingIDs := make([]int, 0, len(ids))
	existingUploads := make([]shared.Dump, 0, len(ids))

	for _, id := range ids {
		if upload, ok := requestState.dataLoader.GetUploadFromCacheMap(id); ok {
			existingUploads = append(existingUploads, upload)
		} else {
			missingIDs = append(missingIDs, id)
		}
	}

	uploads, err := s.GetDumpsByIDs(ctx, missingIDs)
	if err != nil {
		return nil, errors.Wrap(err, "service.GetDumpsByIDs")
	}

	uploadsWithResolvableCommits, err := s.removeUploadsWithUnknownCommits(ctx, uploads, requestState)
	if err != nil {
		return nil, nil
	}
	requestState.dataLoader.SetUploadInCacheMap(uploadsWithResolvableCommits)

	allUploads := append(existingUploads, uploadsWithResolvableCommits...)

	return allUploads, nil
}

// removeUploadsWithUnknownCommits removes uploads for commits which are unknown to gitserver from the given
// slice. The slice is filtered in-place and returned (to update the slice length).
func (s *Service) removeUploadsWithUnknownCommits(ctx context.Context, uploads []shared.Dump, requestState RequestState) ([]shared.Dump, error) {
	rcs := make([]codeintelgitserver.RepositoryCommit, 0, len(uploads))
	for _, upload := range uploads {
		rcs = append(rcs, codeintelgitserver.RepositoryCommit{
			RepositoryID: upload.RepositoryID,
			Commit:       upload.Commit,
		})
	}
	exists, err := requestState.commitCache.AreCommitsResolvable(ctx, rcs)
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
func (s *Service) getBulkMonikerLocations(ctx context.Context, uploads []shared.Dump, orderedMonikers []precise.QualifiedMonikerData, tableName string, limit, offset int) ([]shared.Location, int, error) {
	ids := make([]int, 0, len(uploads))
	for i := range uploads {
		ids = append(ids, uploads[i].ID)
	}

	args := make([]precise.MonikerData, 0, len(orderedMonikers))
	for _, moniker := range orderedMonikers {
		args = append(args, moniker.MonikerData)
	}

	locations, totalCount, err := s.GetBulkMonikerLocations(ctx, tableName, ids, args, limit, offset)
	if err != nil {
		return nil, 0, errors.Wrap(err, "lsifStore.GetBulkMonikerLocations")
	}

	return locations, totalCount, nil
}

// DefinitionsLimit is maximum the number of locations returned from Definitions.
const DefinitionsLimit = 100

func (s *Service) GetImplementations(ctx context.Context, args shared.RequestArgs, requestState RequestState, cursor shared.ImplementationsCursor) (_ []shared.UploadLocation, _ shared.ImplementationsCursor, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, s.operations.getImplementations, serviceObserverThreshold, observation.Args{
		LogFields: []traceLog.Field{
			traceLog.Int("repositoryID", args.RepositoryID),
			traceLog.String("commit", args.Commit),
			traceLog.String("path", args.Path),
			traceLog.Int("numUploads", len(requestState.GetCacheUploads())),
			traceLog.String("uploads", uploadIDsToString(requestState.GetCacheUploads())),
			traceLog.Int("line", args.Line),
			traceLog.Int("character", args.Character),
		},
	})
	defer endObservation()

	// Adjust the path and position for each visible upload based on its git difference to
	// the target commit. This data may already be stashed in the cursor decoded above, in
	// which case we don't need to hit the database.
	visibleUploads, cursorsToVisibleUploads, err := s.getVisibleUploadsFromCursor(ctx, args.Line, args.Character, &cursor.CursorsToVisibleUploads, requestState)
	if err != nil {
		return nil, cursor, err
	}

	// Update the cursors with the updated visible uploads.
	cursor.CursorsToVisibleUploads = cursorsToVisibleUploads

	// Gather all monikers attached to the ranges enclosing the requested position. This data
	// may already be stashed in the cursor decoded above, in which case we don't need to hit
	// the database.
	if cursor.OrderedImplementationMonikers == nil {
		if cursor.OrderedImplementationMonikers, err = s.getOrderedMonikers(ctx, visibleUploads, precise.Implementation); err != nil {
			return nil, cursor, err
		}
	}
	trace.Log(
		traceLog.Int("numImplementationMonikers", len(cursor.OrderedImplementationMonikers)),
		traceLog.String("implementationMonikers", monikersToString(cursor.OrderedImplementationMonikers)),
	)

	if cursor.OrderedExportMonikers == nil {
		if cursor.OrderedExportMonikers, err = s.getOrderedMonikers(ctx, visibleUploads, "export"); err != nil {
			return nil, cursor, err
		}
	}
	trace.Log(
		traceLog.Int("numExportMonikers", len(cursor.OrderedExportMonikers)),
		traceLog.String("exportMonikers", monikersToString(cursor.OrderedExportMonikers)),
	)

	// Phase 1: Gather all "local" locations via LSIF graph traversal. We'll continue to request additional
	// locations until we fill an entire page (the size of which is denoted by the given limit) or there are
	// no more local results remaining.
	var locations []shared.Location
	if cursor.Phase == "local" {
		for len(locations) < args.Limit {
			localLocations, hasMore, err := s.getPageLocalLocations(ctx, s.lsifstore.GetImplementationLocations, visibleUploads, &cursor.LocalCursor, args.Limit-len(locations), trace)
			if err != nil {
				return nil, cursor, err
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
		uploads, err := s.getUploadsWithDefinitionsForMonikers(ctx, cursor.OrderedImplementationMonikers, requestState)
		if err != nil {
			return nil, cursor, err
		}
		trace.Log(
			traceLog.Int("numGetUploadsWithDefinitionsForMonikers", len(uploads)),
			traceLog.String("getUploadsWithDefinitionsForMonikers", uploadIDsToString(uploads)),
		)

		definitionLocations, _, err := s.getBulkMonikerLocations(ctx, uploads, cursor.OrderedImplementationMonikers, "definitions", DefinitionsLimit, 0)
		if err != nil {
			return nil, cursor, err
		}
		locations = append(locations, definitionLocations...)

		cursor.Phase = "dependents"
	}

	// Phase 3: Gather all "remote" locations in dependents via moniker search.
	if cursor.Phase == "dependents" {
		for len(locations) < args.Limit {
			remoteLocations, hasMore, err := s.getPageRemoteLocations(ctx, "implementations", visibleUploads, cursor.OrderedExportMonikers, &cursor.RemoteCursor, args.Limit-len(locations), trace, args, requestState)
			if err != nil {
				return nil, cursor, err
			}
			locations = append(locations, remoteLocations...)

			if !hasMore {
				cursor.Phase = "done"
				break
			}
		}
	}

	trace.Log(traceLog.Int("numLocations", len(locations)))

	// Adjust the locations back to the appropriate range in the target commits. This adjusts
	// locations within the repository the user is browsing so that it appears all implementations
	// are occurring at the same commit they are looking at.

	implementationLocations, err := s.getUploadLocations(ctx, args, requestState, locations)
	if err != nil {
		return nil, cursor, err
	}
	trace.Log(traceLog.Int("numImplementationsLocations", len(implementationLocations)))

	return implementationLocations, cursor, nil
}

// GetDefinitions returns the set of locations defining the symbol at the given position.
func (s *Service) GetDefinitions(ctx context.Context, args shared.RequestArgs, requestState RequestState) (_ []shared.UploadLocation, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, s.operations.getDefinitions, serviceObserverThreshold, observation.Args{
		LogFields: []traceLog.Field{
			traceLog.Int("repositoryID", args.RepositoryID),
			traceLog.String("commit", args.Commit),
			traceLog.String("path", args.Path),
			traceLog.Int("numUploads", len(requestState.GetCacheUploads())),
			traceLog.String("uploads", uploadIDsToString(requestState.GetCacheUploads())),
			traceLog.Int("line", args.Line),
			traceLog.Int("character", args.Character),
		},
	})
	defer endObservation()

	// Adjust the path and position for each visible upload based on its git difference to
	// the target commit.
	visibleUploads, err := s.getVisibleUploads(ctx, args.Line, args.Character, requestState)
	if err != nil {
		return nil, err
	}

	// Gather the "local" reference locations that are reachable via a referenceResult vertex.
	// If the definition exists within the index, it should be reachable via an LSIF graph
	// traversal and should not require an additional moniker search in the same index.
	for i := range visibleUploads {
		trace.Log(traceLog.Int("uploadID", visibleUploads[i].Upload.ID))

		locations, _, err := s.lsifstore.GetDefinitionLocations(
			ctx,
			visibleUploads[i].Upload.ID,
			visibleUploads[i].TargetPathWithoutRoot,
			visibleUploads[i].TargetPosition.Line,
			visibleUploads[i].TargetPosition.Character,
			DefinitionsLimit,
			0,
		)
		if err != nil {
			return nil, errors.Wrap(err, "lsifStore.Definitions")
		}
		if len(locations) > 0 {
			// If we have a local definition, we won't find a better one and can exit early
			return s.getUploadLocations(ctx, args, requestState, locations)
		}
	}

	// Gather all import monikers attached to the ranges enclosing the requested position
	orderedMonikers, err := s.getOrderedMonikers(ctx, visibleUploads, "import")
	if err != nil {
		return nil, err
	}
	trace.Log(
		traceLog.Int("numMonikers", len(orderedMonikers)),
		traceLog.String("monikers", monikersToString(orderedMonikers)),
	)

	// Determine the set of uploads over which we need to perform a moniker search. This will
	// include all all indexes which define one of the ordered monikers. This should not include
	// any of the indexes we have already performed an LSIF graph traversal in above.
	uploads, err := s.getUploadsWithDefinitionsForMonikers(ctx, orderedMonikers, requestState)
	if err != nil {
		return nil, err
	}
	trace.Log(
		traceLog.Int("numXrepoDefinitionUploads", len(uploads)),
		traceLog.String("xrepoDefinitionUploads", uploadIDsToString(uploads)),
	)

	// Perform the moniker search
	locations, _, err := s.getBulkMonikerLocations(ctx, uploads, orderedMonikers, "definitions", DefinitionsLimit, 0)
	if err != nil {
		return nil, err
	}
	trace.Log(traceLog.Int("numXrepoLocations", len(locations)))

	// Adjust the locations back to the appropriate range in the target commits. This adjusts
	// locations within the repository the user is browsing so that it appears all definitions
	// are occurring at the same commit they are looking at.

	adjustedLocations, err := s.getUploadLocations(ctx, args, requestState, locations)
	if err != nil {
		return nil, err
	}
	trace.Log(traceLog.Int("numAdjustedXrepoLocations", len(adjustedLocations)))

	return adjustedLocations, nil
}

func (s *Service) GetDiagnostics(ctx context.Context, args shared.RequestArgs, requestState RequestState) (diagnosticsAtUploads []shared.DiagnosticAtUpload, _ int, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, s.operations.getDiagnostics, serviceObserverThreshold, observation.Args{
		LogFields: []traceLog.Field{
			traceLog.Int("repositoryID", args.RepositoryID),
			traceLog.String("commit", args.Commit),
			traceLog.String("path", args.Path),
			traceLog.Int("numUploads", len(requestState.GetCacheUploads())),
			traceLog.String("uploads", uploadIDsToString(requestState.GetCacheUploads())),
			traceLog.Int("limit", args.Limit),
		},
	})
	defer endObservation()

	visibleUploads, err := s.getUploadPaths(ctx, args.Path, requestState)
	if err != nil {
		return nil, 0, err
	}

	totalCount := 0

	checkerEnabled := authz.SubRepoEnabled(requestState.authChecker)
	var a *actor.Actor
	if checkerEnabled {
		a = actor.FromContext(ctx)
	}
	for i := range visibleUploads {
		trace.Log(traceLog.Int("uploadID", visibleUploads[i].Upload.ID))

		diagnostics, count, err := s.lsifstore.GetDiagnostics(
			ctx,
			visibleUploads[i].Upload.ID,
			visibleUploads[i].TargetPathWithoutRoot,
			args.Limit-len(diagnosticsAtUploads),
			0,
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, "lsifStore.Diagnostics")
		}

		for _, diagnostic := range diagnostics {
			adjustedDiagnostic, err := s.getRequestedCommitDiagnostic(ctx, args, requestState, visibleUploads[i], diagnostic)
			if err != nil {
				return nil, 0, err
			}

			if !checkerEnabled {
				diagnosticsAtUploads = append(diagnosticsAtUploads, adjustedDiagnostic)
				continue
			}

			// sub-repo checker is enabled, proceeding with check
			if include, err := authz.FilterActorPath(ctx, requestState.authChecker, a, api.RepoName(adjustedDiagnostic.Dump.RepositoryName), adjustedDiagnostic.Path); err != nil {
				return nil, 0, err
			} else if include {
				diagnosticsAtUploads = append(diagnosticsAtUploads, adjustedDiagnostic)
			}
		}

		totalCount += count
	}

	if len(diagnosticsAtUploads) > args.Limit {
		diagnosticsAtUploads = diagnosticsAtUploads[:args.Limit]
	}
	trace.Log(
		traceLog.Int("totalCount", totalCount),
		traceLog.Int("numDiagnostics", len(diagnosticsAtUploads)),
	)

	return diagnosticsAtUploads, totalCount, nil
}

// getRequestedCommitDiagnostic translates a diagnostic (relative to the indexed commit) into an equivalent diagnostic
// in the requested commit.
func (s *Service) getRequestedCommitDiagnostic(ctx context.Context, args shared.RequestArgs, requestState RequestState, adjustedUpload visibleUpload, diagnostic shared.Diagnostic) (shared.DiagnosticAtUpload, error) {
	rn := shared.Range{
		Start: shared.Position{
			Line:      diagnostic.StartLine,
			Character: diagnostic.StartCharacter,
		},
		End: shared.Position{
			Line:      diagnostic.EndLine,
			Character: diagnostic.EndCharacter,
		},
	}

	// Adjust path in diagnostic before reading it. This value is used in the adjustRange
	// call below, and is also reflected in the embedded diagnostic value in the return.
	diagnostic.Path = adjustedUpload.Upload.Root + diagnostic.Path

	adjustedCommit, adjustedRange, _, err := s.getSourceRange(
		ctx,
		args,
		requestState,
		adjustedUpload.Upload.RepositoryID,
		adjustedUpload.Upload.Commit,
		diagnostic.Path,
		rn,
	)
	if err != nil {
		return shared.DiagnosticAtUpload{}, err
	}

	return shared.DiagnosticAtUpload{
		Diagnostic:     diagnostic,
		Dump:           adjustedUpload.Upload,
		AdjustedCommit: adjustedCommit,
		AdjustedRange:  adjustedRange,
	}, nil
}

// getUploadPaths adjusts the current target path for each upload visible from the current target
// commit. If an upload cannot be adjusted, it will be omitted from the returned slice.
func (s *Service) getUploadPaths(ctx context.Context, path string, requestState RequestState) ([]visibleUpload, error) {
	cacheUploads := requestState.GetCacheUploads()
	visibleUploads := make([]visibleUpload, 0, len(cacheUploads))
	for _, cu := range cacheUploads {
		targetPath, ok, err := requestState.GitTreeTranslator.GetTargetCommitPathFromSourcePath(ctx, cu.Commit, path, false)
		if err != nil {
			return nil, errors.Wrap(err, "r.GitTreeTranslator.GetTargetCommitPathFromSourcePath")
		}
		if !ok {
			continue
		}

		visibleUploads = append(visibleUploads, visibleUpload{
			Upload:                cu,
			TargetPath:            targetPath,
			TargetPathWithoutRoot: strings.TrimPrefix(targetPath, cu.Root),
		})
	}

	return visibleUploads, nil
}

func (s *Service) GetRanges(ctx context.Context, args shared.RequestArgs, requestState RequestState, startLine, endLine int) (adjustedRanges []shared.AdjustedCodeIntelligenceRange, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, s.operations.getRanges, serviceObserverThreshold, observation.Args{
		LogFields: []traceLog.Field{
			traceLog.Int("repositoryID", args.RepositoryID),
			traceLog.String("commit", args.Commit),
			traceLog.String("path", args.Path),
			traceLog.Int("numUploads", len(requestState.GetCacheUploads())),
			traceLog.String("uploads", uploadIDsToString(requestState.GetCacheUploads())),
			traceLog.Int("startLine", startLine),
			traceLog.Int("endLine", endLine),
		},
	})
	defer endObservation()

	uploadsWithPath, err := s.getUploadPaths(ctx, args.Path, requestState)
	if err != nil {
		return nil, err
	}

	for i := range uploadsWithPath {
		trace.Log(traceLog.Int("uploadID", uploadsWithPath[i].Upload.ID))

		ranges, err := s.lsifstore.GetRanges(
			ctx,
			uploadsWithPath[i].Upload.ID,
			uploadsWithPath[i].TargetPathWithoutRoot,
			startLine,
			endLine,
		)
		if err != nil {
			return nil, errors.Wrap(err, "lsifStore.Ranges")
		}

		for _, rn := range ranges {
			adjustedRange, ok, err := s.getCodeIntelligenceRange(ctx, args, requestState, uploadsWithPath[i], rn)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}

			adjustedRanges = append(adjustedRanges, adjustedRange)
		}
	}
	trace.Log(traceLog.Int("numRanges", len(adjustedRanges)))

	return adjustedRanges, nil
}

// getCodeIntelligenceRange translates a range summary (relative to the indexed commit) into an
// equivalent range summary in the requested commit. If the translation fails, a false-valued flag
// is returned.
func (s *Service) getCodeIntelligenceRange(ctx context.Context, args shared.RequestArgs, requestState RequestState, upload visibleUpload, rn shared.CodeIntelligenceRange) (shared.AdjustedCodeIntelligenceRange, bool, error) {
	_, adjustedRange, ok, err := s.getSourceRange(ctx, args, requestState, upload.Upload.RepositoryID, upload.Upload.Commit, upload.TargetPath, rn.Range)
	if err != nil || !ok {
		return shared.AdjustedCodeIntelligenceRange{}, false, err
	}

	definitions, err := s.getUploadLocations(ctx, args, requestState, rn.Definitions)
	if err != nil {
		return shared.AdjustedCodeIntelligenceRange{}, false, err
	}

	references, err := s.getUploadLocations(ctx, args, requestState, rn.References)
	if err != nil {
		return shared.AdjustedCodeIntelligenceRange{}, false, err
	}

	implementations, err := s.getUploadLocations(ctx, args, requestState, rn.Implementations)
	if err != nil {
		return shared.AdjustedCodeIntelligenceRange{}, false, err
	}

	return shared.AdjustedCodeIntelligenceRange{
		Range:           adjustedRange,
		Definitions:     definitions,
		References:      references,
		Implementations: implementations,
		HoverText:       rn.HoverText,
	}, true, nil
}

// GetStencil returns the set of locations defining the symbol at the given position.
func (s *Service) GetStencil(ctx context.Context, args shared.RequestArgs, requestState RequestState) (adjustedRanges []shared.Range, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, s.operations.getStencil, serviceObserverThreshold, observation.Args{
		LogFields: []traceLog.Field{
			traceLog.Int("repositoryID", args.RepositoryID),
			traceLog.String("commit", args.Commit),
			traceLog.String("path", args.Path),
			traceLog.Int("numUploads", len(requestState.GetCacheUploads())),
			traceLog.String("uploads", uploadIDsToString(requestState.GetCacheUploads())),
		},
	})
	defer endObservation()

	adjustedUploads, err := s.getUploadPaths(ctx, args.Path, requestState)
	if err != nil {
		return nil, err
	}

	for i := range adjustedUploads {
		trace.Log(traceLog.Int("uploadID", adjustedUploads[i].Upload.ID))

		ranges, err := s.lsifstore.GetStencil(
			ctx,
			adjustedUploads[i].Upload.ID,
			adjustedUploads[i].TargetPathWithoutRoot,
		)
		if err != nil {
			return nil, errors.Wrap(err, "lsifStore.Stencil")
		}

		for i, rn := range ranges {
			// FIXME: change this at it expects an empty shared.Dump{}
			cu := requestState.GetCacheUploadsAtIndex(i)
			// Adjust the highlighted range back to the appropriate range in the target commit
			_, adjustedRange, _, err := s.getSourceRange(ctx, args, requestState, cu.RepositoryID, cu.Commit, args.Path, rn)
			if err != nil {
				return nil, err
			}

			adjustedRanges = append(adjustedRanges, adjustedRange)
		}
	}
	trace.Log(traceLog.Int("numRanges", len(adjustedRanges)))

	return sortRanges(adjustedRanges), nil
}

func (s *Service) GetMonikersByPosition(ctx context.Context, bundleID int, path string, line, character int) (_ [][]precise.MonikerData, err error) {
	ctx, _, endObservation := s.operations.getMonikersByPosition.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.lsifstore.GetMonikersByPosition(ctx, bundleID, path, line, character)
}

func (s *Service) GetBulkMonikerLocations(ctx context.Context, tableName string, uploadIDs []int, monikers []precise.MonikerData, limit, offset int) (_ []shared.Location, _ int, err error) {
	ctx, _, endObservation := s.operations.getBulkMonikerLocations.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.lsifstore.GetBulkMonikerLocations(ctx, tableName, uploadIDs, monikers, limit, offset)
}

func (s *Service) GetUploadsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QualifiedMonikerData) (_ []shared.Dump, err error) {
	ctx, _, endObservation := s.operations.getUploadsWithDefinitionsForMonikers.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	uploadDumps, err := s.uploadSvc.GetDumpsWithDefinitionsForMonikers(ctx, monikers)
	if err != nil {
		return nil, err
	}
	dumps := updateSvcDumpToSharedDump(uploadDumps)

	return dumps, nil
}

func (s *Service) GetDumpsByIDs(ctx context.Context, ids []int) (_ []shared.Dump, err error) {
	ctx, _, endObservation := s.operations.getDumpsByIDs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	uploadDumps, err := s.uploadSvc.GetDumpsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	dumps := updateSvcDumpToSharedDump(uploadDumps)

	return dumps, nil
}

func (s *Service) GetUploadIDsWithReferences(
	ctx context.Context,
	orderedMonikers []precise.QualifiedMonikerData,
	ignoreIDs []int,
	repositoryID int,
	commit string,
	limit int,
	offset int,
) (ids []int, recordsScanned int, totalCount int, err error) {
	ctx, _, endObservation := s.operations.getUploadIDsWithReferences.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.uploadSvc.GetUploadIDsWithReferences(ctx, orderedMonikers, ignoreIDs, repositoryID, commit, limit, offset)
}

func (s *Service) GetPackageInformation(ctx context.Context, bundleID int, path, packageInformationID string) (_ precise.PackageInformationData, _ bool, err error) {
	ctx, _, endObservation := s.operations.getPackageInformation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.lsifstore.GetPackageInformation(ctx, bundleID, path, packageInformationID)
}

func (s *Service) GetClosestDumpsForBlob(ctx context.Context, repositoryID int, commit, path string, exactPath bool, indexer string) (_ []shared.Dump, err error) {
	ctx, trace, endObservation := s.operations.getClosestDumpsForBlob.With(ctx, &err, observation.Args{
		LogFields: []traceLog.Field{
			traceLog.Int("repositoryID", repositoryID),
			traceLog.String("commit", commit),
			traceLog.String("path", path),
			traceLog.Bool("exactPath", exactPath),
			traceLog.String("indexer", indexer),
		},
	})
	defer endObservation(1, observation.Args{})

	candidates, err := s.uploadSvc.InferClosestUploads(ctx, repositoryID, commit, path, exactPath, indexer)
	if err != nil {
		return nil, err
	}

	uploadCandidates := updateSvcDumpToSharedDump(candidates)
	trace.Log(
		traceLog.Int("numCandidates", len(candidates)),
		traceLog.String("candidates", uploadIDsToString(uploadCandidates)),
	)

	commitChecker := NewCommitCache(s.gitserver)
	commitChecker.SetResolvableCommit(repositoryID, commit)

	candidatesWithCommits, err := filterUploadsWithCommits(ctx, commitChecker, uploadCandidates)
	if err != nil {
		return nil, err
	}
	trace.Log(
		traceLog.Int("numCandidatesWithCommits", len(candidatesWithCommits)),
		traceLog.String("candidatesWithCommits", uploadIDsToString(candidatesWithCommits)),
	)

	// Filter in-place
	filtered := candidatesWithCommits[:0]

	for i := range candidatesWithCommits {
		if exactPath {
			// TODO - this breaks if the file was renamed in git diff
			pathExists, err := s.lsifstore.GetPathExists(ctx, candidates[i].ID, strings.TrimPrefix(path, candidates[i].Root))
			if err != nil {
				return nil, errors.Wrap(err, "lsifStore.Exists")
			}
			if !pathExists {
				continue
			}
		} else { //nolint:staticcheck
			// TODO(efritz) - ensure there's a valid document path for this condition as well
		}

		filtered = append(filtered, uploadCandidates[i])
	}
	trace.Log(
		traceLog.Int("numFiltered", len(filtered)),
		traceLog.String("filtered", uploadIDsToString(filtered)),
	)

	return filtered, nil
}

func (s *Service) GetSupportedByCtags(ctx context.Context, filepath string, repoName api.RepoName) (bool, string, error) {
	mappings, err := s.symbolsClient.ListLanguageMappings(ctx, repoName)
	if err != nil {
		return false, "", err
	}

	for language, globs := range mappings {
		for _, glob := range globs {
			if glob.Match(filepath) {
				return true, language, nil
			}
		}
	}

	return false, "", nil
}

func (s *Service) SetRequestLanguageSupport(ctx context.Context, userID int, language string) (err error) {
	ctx, _, endObservation := s.operations.setRequestLanguageSupport.With(ctx, &err, observation.Args{
		LogFields: []traceLog.Field{traceLog.Int("userID", userID), traceLog.String("language", language)},
	})
	defer endObservation(1, observation.Args{})

	return s.store.SetRequestLanguageSupport(ctx, userID, language)
}

func (s *Service) GetLanguagesRequestedBy(ctx context.Context, userID int) (_ []string, err error) {
	ctx, _, endObservation := s.operations.getLanguagesRequestedBy.With(ctx, &err, observation.Args{
		LogFields: []traceLog.Field{traceLog.Int("userID", userID)},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetLanguagesRequestedBy(ctx, userID)
}

// filterUploadsWithCommits removes the uploads for commits which are unknown to gitserver from the given
// slice. The slice is filtered in-place and returned (to update the slice length).
func filterUploadsWithCommits(ctx context.Context, commitCache CommitCache, uploads []shared.Dump) ([]shared.Dump, error) {
	rcs := make([]codeintelgitserver.RepositoryCommit, 0, len(uploads))
	for _, upload := range uploads {
		rcs = append(rcs, codeintelgitserver.RepositoryCommit{
			RepositoryID: upload.RepositoryID,
			Commit:       upload.Commit,
		})
	}
	exists, err := commitCache.ExistsBatch(ctx, rcs)
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

func updateSvcDumpToSharedDump(uploadDumps []uploads.Dump) []shared.Dump {
	dumps := make([]shared.Dump, 0, len(uploadDumps))
	for _, d := range uploadDumps {
		dumps = append(dumps, shared.Dump{
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
func (s *Service) getVisibleUploadsFromCursor(ctx context.Context, line, character int, cursorsToVisibleUploads *[]shared.CursorToVisibleUpload, r RequestState) ([]visibleUpload, []shared.CursorToVisibleUpload, error) {
	if *cursorsToVisibleUploads != nil {
		visibleUploads := make([]visibleUpload, 0, len(*cursorsToVisibleUploads))
		for _, u := range *cursorsToVisibleUploads {
			upload, ok := r.dataLoader.GetUploadFromCacheMap(u.DumpID)
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

	visibleUploads, err := s.getVisibleUploads(ctx, line, character, r)
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
func (s *Service) getVisibleUploads(ctx context.Context, line, character int, r RequestState) ([]visibleUpload, error) {
	visibleUploads := make([]visibleUpload, 0, len(r.dataLoader.uploads))
	for i := range r.dataLoader.uploads {
		adjustedUpload, ok, err := s.getVisibleUpload(ctx, line, character, r.dataLoader.uploads[i], r)
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
func (s *Service) getVisibleUpload(ctx context.Context, line, character int, upload shared.Dump, r RequestState) (visibleUpload, bool, error) {
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
