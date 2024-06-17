package codenav

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	searcher "github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/languages"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Service struct {
	repoStore    database.RepoStore
	lsifstore    lsifstore.LsifStore
	gitserver    gitserver.Client
	uploadSvc    UploadService
	searchClient searcher.SearchClient
	operations   *operations
	logger       log.Logger
}

func newService(
	observationCtx *observation.Context,
	repoStore database.RepoStore,
	lsifstore lsifstore.LsifStore,
	uploadSvc UploadService,
	gitserver gitserver.Client,
	searchClient searcher.SearchClient,
	logger log.Logger,
) *Service {
	return &Service{
		repoStore:    repoStore,
		lsifstore:    lsifstore,
		gitserver:    gitserver,
		uploadSvc:    uploadSvc,
		searchClient: searchClient,
		operations:   newOperations(observationCtx),
		logger:       logger,
	}
}

// GetHover returns the set of locations defining the symbol at the given position.
func (s *Service) GetHover(ctx context.Context, args PositionalRequestArgs, requestState RequestState) (_ string, _ shared.Range, _ bool, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, s.operations.getHover, serviceObserverThreshold,
		observation.Args{Attrs: observation.MergeAttributes(args.Attrs(), requestState.Attrs()...)})
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
		trace.AddEvent("TODO Domain Owner", attribute.Int("uploadID", adjustedUpload.Upload.ID))

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
		_, adjustedRange, _, err := s.getSourceRange(ctx, args.RequestArgs, requestState, cachedUploads[i].RepositoryID, cachedUploads[i].Commit, args.Path, rn)
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
	trace.AddEvent("TODO Domain Owner",
		attribute.Int("numMonikers", len(orderedMonikers)),
		attribute.String("monikers", monikersToString(orderedMonikers)))

	// Determine the set of uploads over which we need to perform a moniker search. This will
	// include all all indexes which define one of the ordered monikers. This should not include
	// any of the indexes we have already performed an LSIF graph traversal in above.
	uploads, err := s.getUploadsWithDefinitionsForMonikers(ctx, orderedMonikers, requestState)
	if err != nil {
		return "", shared.Range{}, false, err
	}
	trace.AddEvent("TODO Domain Owner",
		attribute.Int("numDefinitionUploads", len(uploads)),
		attribute.String("definitionUploads", uploadIDsToString(uploads)))

	// Perform the moniker search. This returns a set of locations defining one of the monikers
	// attached to one of the source ranges.
	locations, _, err := s.getBulkMonikerLocations(ctx, uploads, orderedMonikers, "definitions", DefinitionsLimit, 0)
	if err != nil {
		return "", shared.Range{}, false, err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("numLocations", len(locations)))

	for i := range locations {
		// Fetch hover text attached to a definition in the defining index
		text, _, exists, err := s.lsifstore.GetHover(
			ctx,
			locations[i].UploadID,
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

// getUploadsWithDefinitionsForMonikers returns the set of uploads that provide any of the given monikers.
// This method will not return uploads for commits which are unknown to gitserver.
func (s *Service) getUploadsWithDefinitionsForMonikers(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData, requestState RequestState) ([]uploadsshared.CompletedUpload, error) {
	uploads, err := s.uploadSvc.GetCompletedUploadsWithDefinitionsForMonikers(ctx, orderedMonikers)
	if err != nil {
		return nil, errors.Wrap(err, "dbstore.DefinitionDumps")
	}

	uploadsCopy := copyUploads(uploads)
	requestState.dataLoader.SetUploadInCacheMap(uploadsCopy)

	uploadsWithResolvableCommits, err := filterUploadsWithCommits(ctx, requestState.commitCache, uploadsCopy)
	if err != nil {
		return nil, err
	}

	return uploadsWithResolvableCommits, nil
}

// monikerLimit is the maximum number of monikers that can be returned from orderedMonikers.
const monikerLimit = 10

func (s *Service) getOrderedMonikers(ctx context.Context, visibleUploads []visibleUpload, kinds ...string) ([]precise.QualifiedMonikerData, error) {
	monikerSet := newQualifiedMonikerSet()

	for i := range visibleUploads {
		rangeMonikers, err := s.lsifstore.GetMonikersByPosition(
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

				packageInformationData, _, err := s.lsifstore.GetPackageInformation(
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

// getUploadLocations translates a set of locations into an equivalent set of locations in the requested
// commit. If includeFallbackLocations is true, then any range in the indexed commit that cannot be translated
// will use the indexed location. Otherwise, such location are dropped.
func (s *Service) getUploadLocations(ctx context.Context, args RequestArgs, requestState RequestState, locations []shared.Location, includeFallbackLocations bool) ([]shared.UploadLocation, error) {
	uploadLocations := make([]shared.UploadLocation, 0, len(locations))

	checkerEnabled := authz.SubRepoEnabled(requestState.authChecker)
	var a *actor.Actor
	if checkerEnabled {
		a = actor.FromContext(ctx)
	}
	for _, location := range locations {
		upload, ok := requestState.dataLoader.GetUploadFromCacheMap(location.UploadID)
		if !ok {
			continue
		}

		adjustedLocation, ok, err := s.getUploadLocation(ctx, args, requestState, upload, location)
		if err != nil {
			return nil, err
		}
		if !includeFallbackLocations && !ok {
			continue
		}

		if !checkerEnabled {
			uploadLocations = append(uploadLocations, adjustedLocation)
		} else {
			repo := api.RepoName(adjustedLocation.Upload.RepositoryName)
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
// commit and range of the adjusted location and a false flag is returned.
func (s *Service) getUploadLocation(ctx context.Context, args RequestArgs, requestState RequestState, upload uploadsshared.CompletedUpload, location shared.Location) (shared.UploadLocation, bool, error) {
	adjustedCommit, adjustedRange, ok, err := s.getSourceRange(ctx, args, requestState, upload.RepositoryID, upload.Commit, upload.Root+location.Path, location.Range)
	if err != nil {
		return shared.UploadLocation{}, ok, err
	}

	return shared.UploadLocation{
		Upload:       upload,
		Path:         upload.Root + location.Path,
		TargetCommit: adjustedCommit,
		TargetRange:  adjustedRange,
	}, ok, nil
}

// getSourceRange translates a range (relative to the indexed commit) into an equivalent range in the requested
// commit. If the translation fails, then the original commit and range are returned along with a false-valued
// flag.
func (s *Service) getSourceRange(ctx context.Context, args RequestArgs, requestState RequestState, repositoryID int, commit, path string, rng shared.Range) (string, shared.Range, bool, error) {
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
func (s *Service) getUploadsByIDs(ctx context.Context, ids []int, requestState RequestState) ([]uploadsshared.CompletedUpload, error) {
	missingIDs := make([]int, 0, len(ids))
	existingUploads := make([]uploadsshared.CompletedUpload, 0, len(ids))

	for _, id := range ids {
		if upload, ok := requestState.dataLoader.GetUploadFromCacheMap(id); ok {
			existingUploads = append(existingUploads, upload)
		} else {
			missingIDs = append(missingIDs, id)
		}
	}

	uploads, err := s.uploadSvc.GetCompletedUploadsByIDs(ctx, missingIDs)
	if err != nil {
		return nil, errors.Wrap(err, "service.GetCompletedUploadsByIDs")
	}

	uploadsWithResolvableCommits, err := filterUploadsWithCommits(ctx, requestState.commitCache, uploads)
	if err != nil {
		return nil, nil
	}
	requestState.dataLoader.SetUploadInCacheMap(uploadsWithResolvableCommits)

	allUploads := append(existingUploads, uploadsWithResolvableCommits...)

	return allUploads, nil
}

// getBulkMonikerLocations returns the set of locations (within the given uploads) with an attached moniker
// whose scheme+identifier matches any of the given monikers.
func (s *Service) getBulkMonikerLocations(ctx context.Context, uploads []uploadsshared.CompletedUpload, orderedMonikers []precise.QualifiedMonikerData, tableName string, limit, offset int) ([]shared.Location, int, error) {
	ids := make([]int, 0, len(uploads))
	for i := range uploads {
		ids = append(ids, uploads[i].ID)
	}

	args := make([]precise.MonikerData, 0, len(orderedMonikers))
	for _, moniker := range orderedMonikers {
		args = append(args, moniker.MonikerData)
	}

	locations, totalCount, err := s.lsifstore.GetBulkMonikerLocations(ctx, tableName, ids, args, limit, offset)
	if err != nil {
		return nil, 0, errors.Wrap(err, "lsifStore.GetBulkMonikerLocations")
	}

	return locations, totalCount, nil
}

// DefinitionsLimit is maximum the number of locations returned from Definitions.
const DefinitionsLimit = 100

func (s *Service) GetDiagnostics(ctx context.Context, args PositionalRequestArgs, requestState RequestState) (diagnosticsAtUploads []DiagnosticAtUpload, _ int, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, s.operations.getDiagnostics, serviceObserverThreshold,
		observation.Args{Attrs: observation.MergeAttributes(args.Attrs(), requestState.Attrs()...)})
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
		trace.AddEvent("TODO Domain Owner", attribute.Int("uploadID", visibleUploads[i].Upload.ID))

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
			adjustedDiagnostic, err := s.getRequestedCommitDiagnostic(ctx, args.RequestArgs, requestState, visibleUploads[i], diagnostic)
			if err != nil {
				return nil, 0, err
			}

			if !checkerEnabled {
				diagnosticsAtUploads = append(diagnosticsAtUploads, adjustedDiagnostic)
				continue
			}

			// sub-repo checker is enabled, proceeding with check
			if include, err := authz.FilterActorPath(ctx, requestState.authChecker, a, api.RepoName(adjustedDiagnostic.Upload.RepositoryName), adjustedDiagnostic.Path); err != nil {
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
	trace.AddEvent("TODO Domain Owner",
		attribute.Int("totalCount", totalCount),
		attribute.Int("numDiagnostics", len(diagnosticsAtUploads)))

	return diagnosticsAtUploads, totalCount, nil
}

// getRequestedCommitDiagnostic translates a diagnostic (relative to the indexed commit) into an equivalent diagnostic
// in the requested commit.
func (s *Service) getRequestedCommitDiagnostic(ctx context.Context, args RequestArgs, requestState RequestState, adjustedUpload visibleUpload, diagnostic shared.Diagnostic) (DiagnosticAtUpload, error) {
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
		return DiagnosticAtUpload{}, err
	}

	return DiagnosticAtUpload{
		Diagnostic:     diagnostic,
		Upload:         adjustedUpload.Upload,
		AdjustedCommit: adjustedCommit,
		AdjustedRange:  adjustedRange,
	}, nil
}

func (s *Service) VisibleUploadsForPath(ctx context.Context, requestState RequestState) (uploads []uploadsshared.CompletedUpload, err error) {
	ctx, _, endObservation := s.operations.visibleUploadsForPath.With(ctx, &err, observation.Args{Attrs: requestState.Attrs()})
	defer func() {
		endObservation(1, observation.Args{Attrs: []attribute.KeyValue{
			attribute.Int("numUploads", len(uploads)),
		}})
	}()

	visibleUploads, err := s.getUploadPaths(ctx, requestState.Path, requestState)
	if err != nil {
		return nil, err
	}

	for _, upload := range visibleUploads {
		uploads = append(uploads, upload.Upload)
	}

	return
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

func (s *Service) GetRanges(ctx context.Context, args PositionalRequestArgs, requestState RequestState, startLine, endLine int) (adjustedRanges []AdjustedCodeIntelligenceRange, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, s.operations.getRanges, serviceObserverThreshold,
		observation.Args{Attrs: append(requestState.Attrs(),
			attribute.Int("startLine", startLine),
			attribute.Int("endLine", endLine))},
	)
	defer endObservation()

	uploadsWithPath, err := s.getUploadPaths(ctx, args.Path, requestState)
	if err != nil {
		return nil, err
	}

	for i := range uploadsWithPath {
		trace.AddEvent("TODO Domain Owner", attribute.Int("uploadID", uploadsWithPath[i].Upload.ID))

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
			adjustedRange, ok, err := s.getCodeIntelligenceRange(ctx, args.RequestArgs, requestState, uploadsWithPath[i], rn)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}

			adjustedRanges = append(adjustedRanges, adjustedRange)
		}
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("numRanges", len(adjustedRanges)))

	return adjustedRanges, nil
}

// getCodeIntelligenceRange translates a range summary (relative to the indexed commit) into an
// equivalent range summary in the requested commit. If the translation fails, a false-valued flag
// is returned.
func (s *Service) getCodeIntelligenceRange(ctx context.Context, args RequestArgs, requestState RequestState, upload visibleUpload, rn shared.CodeIntelligenceRange) (AdjustedCodeIntelligenceRange, bool, error) {
	_, adjustedRange, ok, err := s.getSourceRange(ctx, args, requestState, upload.Upload.RepositoryID, upload.Upload.Commit, upload.TargetPath, rn.Range)
	if err != nil || !ok {
		return AdjustedCodeIntelligenceRange{}, false, err
	}

	definitions, err := s.getUploadLocations(ctx, args, requestState, rn.Definitions, false)
	if err != nil {
		return AdjustedCodeIntelligenceRange{}, false, err
	}

	references, err := s.getUploadLocations(ctx, args, requestState, rn.References, false)
	if err != nil {
		return AdjustedCodeIntelligenceRange{}, false, err
	}

	implementations, err := s.getUploadLocations(ctx, args, requestState, rn.Implementations, false)
	if err != nil {
		return AdjustedCodeIntelligenceRange{}, false, err
	}

	return AdjustedCodeIntelligenceRange{
		Range:           adjustedRange,
		Definitions:     definitions,
		References:      references,
		Implementations: implementations,
		HoverText:       rn.HoverText,
	}, true, nil
}

// GetStencil returns the set of locations defining the symbol at the given position.
func (s *Service) GetStencil(ctx context.Context, args PositionalRequestArgs, requestState RequestState) (adjustedRanges []shared.Range, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, s.operations.getStencil, serviceObserverThreshold, observation.Args{Attrs: requestState.Attrs()})
	defer endObservation()

	adjustedUploads, err := s.getUploadPaths(ctx, args.Path, requestState)
	if err != nil {
		return nil, err
	}

	for i := range adjustedUploads {
		trace.AddEvent("TODO Domain Owner", attribute.Int("uploadID", adjustedUploads[i].Upload.ID))

		ranges, err := s.lsifstore.GetStencil(
			ctx,
			adjustedUploads[i].Upload.ID,
			adjustedUploads[i].TargetPathWithoutRoot,
		)
		if err != nil {
			return nil, errors.Wrap(err, "lsifStore.Stencil")
		}

		for i, rn := range ranges {
			// FIXME: change this at it expects an empty uploadsshared.CompletedUpload{}
			cu := requestState.GetCacheUploadsAtIndex(i)
			// Adjust the highlighted range back to the appropriate range in the target commit
			_, adjustedRange, _, err := s.getSourceRange(ctx, args.RequestArgs, requestState, cu.RepositoryID, cu.Commit, args.Path, rn)
			if err != nil {
				return nil, err
			}

			adjustedRanges = append(adjustedRanges, adjustedRange)
		}
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("numRanges", len(adjustedRanges)))

	sortedRanges := sortRanges(adjustedRanges)
	return dedupeRanges(sortedRanges), nil
}

func (s *Service) GetClosestCompletedUploadsForBlob(ctx context.Context, opts uploadsshared.UploadMatchingOptions) (_ []uploadsshared.CompletedUpload, err error) {
	ctx, trace, endObservation := s.operations.getClosestCompletedUploadsForBlob.With(ctx, &err, observation.Args{Attrs: opts.Attrs()})
	defer endObservation(1, observation.Args{})

	candidates, err := s.uploadSvc.InferClosestUploads(ctx, opts)
	if err != nil {
		return nil, err
	}

	trace.AddEvent("InferClosestUploads",
		attribute.Int("numCandidates", len(candidates)),
		attribute.String("candidates", uploadIDsToString(candidates)))

	commitChecker := NewCommitCache(s.repoStore, s.gitserver)
	commitChecker.SetResolvableCommit(opts.RepositoryID, opts.Commit)

	candidatesWithExistingCommits, err := filterUploadsWithCommits(ctx, commitChecker, candidates)
	if err != nil {
		return nil, err
	}
	trace.AddEvent("filterUploadsWithCommits",
		attribute.Int("numCandidatesWithExistingCommits", len(candidatesWithExistingCommits)),
		attribute.String("candidatesWithExistingCommits", uploadIDsToString(candidatesWithExistingCommits)))

	candidatesWithExistingCommitsAndPaths, err := filterUploadsWithPaths(ctx, s.lsifstore, opts, candidatesWithExistingCommits)
	if err != nil {
		return nil, errors.Wrap(err, "filtering uploads based on paths")
	}
	trace.AddEvent("filterUploadsWithPaths",
		attribute.Int("numFiltered", len(candidatesWithExistingCommitsAndPaths)),
		attribute.String("filtered", uploadIDsToString(candidatesWithExistingCommitsAndPaths)))

	return candidatesWithExistingCommitsAndPaths, nil
}

// filterUploadsWithCommits only keeps the uploads for commits which are known to gitserver.
// A fresh slice is returned without modifying the original slice.
func filterUploadsWithCommits(ctx context.Context, commitCache CommitCache, uploads []uploadsshared.CompletedUpload) ([]uploadsshared.CompletedUpload, error) {
	rcs := make([]RepositoryCommit, 0, len(uploads))
	for _, upload := range uploads {
		rcs = append(rcs, RepositoryCommit{
			RepositoryID: upload.RepositoryID,
			Commit:       upload.Commit,
		})
	}
	exists, err := commitCache.ExistsBatch(ctx, rcs)
	if err != nil {
		return nil, err
	}

	filtered := make([]uploadsshared.CompletedUpload, 0, len(uploads))
	for i, upload := range uploads {
		if exists[i] {
			filtered = append(filtered, upload)
		}
	}

	return filtered, nil
}

func filterUploadsWithPaths(
	ctx context.Context,
	lsifstore lsifstore.LsifStore,
	opts uploadsshared.UploadMatchingOptions,
	candidates []uploadsshared.CompletedUpload,
) ([]uploadsshared.CompletedUpload, error) {
	filtered := make([]uploadsshared.CompletedUpload, 0, len(candidates))
	for _, candidate := range candidates {
		switch opts.RootToPathMatching {
		case uploadsshared.RootMustEnclosePath:
			// TODO - this breaks if the file was renamed in git diff
			pathExists, err := lsifstore.GetPathExists(ctx, candidate.ID, strings.TrimPrefix(opts.Path, candidate.Root))
			if err != nil {
				return nil, errors.Wrap(err, "lsifStore.Exists")
			}
			if !pathExists {
				continue
			}
		case uploadsshared.RootEnclosesPathOrPathEnclosesRoot:
			// TODO(efritz) - ensure there's a valid document path for this condition as well
		}
		filtered = append(filtered, candidate)
	}
	return filtered, nil
}

func copyUploads(uploads []uploadsshared.CompletedUpload) []uploadsshared.CompletedUpload {
	ud := make([]uploadsshared.CompletedUpload, len(uploads))
	copy(ud, uploads)
	return ud
}

// ErrConcurrentModification occurs when a page of a references request cannot be resolved as
// the set of visible uploads have changed since the previous request for the same result set.
var ErrConcurrentModification = errors.New("result set changed while paginating")

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
func (s *Service) getVisibleUpload(ctx context.Context, line, character int, upload uploadsshared.CompletedUpload, r RequestState) (visibleUpload, bool, error) {
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

func (s *Service) SnapshotForDocument(ctx context.Context, repositoryID int, commit, path string, uploadID int) (data []shared.SnapshotData, err error) {
	ctx, _, endObservation := s.operations.snapshotForDocument.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repoID", repositoryID),
		attribute.String("commit", commit),
		attribute.String("path", path),
		attribute.Int("uploadID", uploadID),
	}})
	defer func() {
		endObservation(1, observation.Args{Attrs: []attribute.KeyValue{
			attribute.Int("snapshotSymbols", len(data)),
		}})
	}()

	uploads, err := s.uploadSvc.GetCompletedUploadsByIDs(ctx, []int{uploadID})
	if err != nil {
		return nil, err
	}

	if len(uploads) == 0 {
		return nil, nil
	}

	upload := uploads[0]

	document, err := s.lsifstore.SCIPDocument(ctx, upload.ID, strings.TrimPrefix(path, upload.Root))
	if err != nil || document == nil {
		return nil, err
	}

	r, err := s.gitserver.NewFileReader(ctx, api.RepoName(upload.RepositoryName), api.CommitID(upload.Commit), path)
	if err != nil {
		return nil, err
	}
	file, err := io.ReadAll(r)
	r.Close()
	if err != nil {
		return nil, err
	}

	// client-side normalizes the file to LF, so normalize CRLF files to that so the offsets are correct
	file = bytes.ReplaceAll(file, []byte("\r\n"), []byte("\n"))

	repo, err := s.repoStore.Get(ctx, api.RepoID(upload.RepositoryID))
	if err != nil {
		return nil, err
	}

	// cache is keyed by repoID:sourceCommit:targetCommit:path, so we only need a size of 1
	hunkcache, err := NewHunkCache(1)
	if err != nil {
		return nil, err
	}
	gittranslator := NewGitTreeTranslator(s.gitserver, &requestArgs{
		repo:   repo,
		commit: commit,
		path:   path,
	}, hunkcache)

	linemap := newLinemap(string(file))
	formatter := scip.LenientVerboseSymbolFormatter
	symtab := document.SymbolTable()

	for _, occ := range document.Occurrences {
		var snapshotData shared.SnapshotData

		formatted, err := formatter.Format(occ.Symbol)
		if err != nil {
			formatted = fmt.Sprintf("error formatting %q", occ.Symbol)
		}

		originalRange := scip.NewRangeUnchecked(occ.Range)

		lineOffset := int32(linemap.positions[originalRange.Start.Line])
		line := file[lineOffset : lineOffset+originalRange.Start.Character]

		tabCount := bytes.Count(line, []byte("\t"))

		var snap strings.Builder
		snap.WriteString(strings.Repeat(" ", (int(originalRange.Start.Character)-tabCount)+(tabCount*4)))
		snap.WriteString(strings.Repeat("^", int(originalRange.End.Character-originalRange.Start.Character)))
		snap.WriteRune(' ')

		isDefinition := occ.SymbolRoles&int32(scip.SymbolRole_Definition) > 0
		if isDefinition {
			snap.WriteString("definition")
		} else {
			snap.WriteString("reference")
		}
		snap.WriteRune(' ')
		snap.WriteString(formatted)

		snapshotData.Symbol = snap.String()

		// hasOverrideDocumentation := len(occ.OverrideDocumentation) > 0
		// if hasOverrideDocumentation {
		// 	documentation := occ.OverrideDocumentation[0]
		// 	writeDocumentation(&b, documentation, prefix, true)
		// }

		if info, ok := symtab[occ.Symbol]; ok && isDefinition {
			// for _, documentation := range info.Documentation {
			// 	// At least get the first line of documentation if there is leading whitespace
			// 	documentation = strings.TrimSpace(documentation)
			// 	writeDocumentation(&b, documentation, prefix, false)
			// }
			slices.SortFunc(info.Relationships, func(a, b *scip.Relationship) int {
				return cmp.Compare(a.Symbol, b.Symbol)
			})
			for _, relationship := range info.Relationships {
				var b strings.Builder
				b.WriteString(strings.Repeat(" ", (int(originalRange.Start.Character)-tabCount)+(tabCount*4)))
				b.WriteString(strings.Repeat("^", int(originalRange.End.Character-originalRange.Start.Character)))
				b.WriteString(" relationship ")

				formatted, err = formatter.Format(relationship.Symbol)
				if err != nil {
					formatted = fmt.Sprintf("error formatting %q", occ.Symbol)
				}

				b.WriteString(formatted)
				if relationship.IsImplementation {
					b.WriteString(" implementation")
				}
				if relationship.IsReference {
					b.WriteString(" reference")
				}
				if relationship.IsTypeDefinition {
					b.WriteString(" type_definition")
				}

				snapshotData.AdditionalData = append(snapshotData.AdditionalData, b.String())
			}
		}

		_, newRange, ok, err := gittranslator.GetTargetCommitPositionFromSourcePosition(ctx, upload.Commit, shared.Position{
			Line:      int(originalRange.Start.Line),
			Character: int(originalRange.Start.Character),
		}, false)
		if err != nil {
			return nil, err
		}
		// if the line was changed, then we're not providing precise codeintel for this line, so skip it
		if !ok {
			continue
		}

		snapshotData.DocumentOffset = linemap.positions[newRange.Line+1]

		data = append(data, snapshotData)
	}

	return
}

func (s *Service) SCIPDocument(ctx context.Context, uploadID int, path string) (*scip.Document, error) {
	return s.lsifstore.SCIPDocument(ctx, uploadID, path)
}

type SyntacticUsagesErrorCode int

const (
	SU_NoSyntacticIndex SyntacticUsagesErrorCode = iota
	SU_NoSymbolAtRequestedRange
	SU_FailedToSearch
	SU_Fatal
)

type SyntacticUsagesError struct {
	Code            SyntacticUsagesErrorCode
	UnderlyingError error
}

var _ error = SyntacticUsagesError{}

func (e SyntacticUsagesError) Error() string {
	msg := ""
	switch e.Code {
	case SU_NoSyntacticIndex:
		msg = "No syntactic index"
	case SU_NoSymbolAtRequestedRange:
		msg = "No symbol at requested range"
	case SU_FailedToSearch:
		msg = "Failed to get candidate matches via searcher"
	case SU_Fatal:
		msg = "fatal error"
	}
	if e.UnderlyingError == nil {
		return msg
	}
	return fmt.Sprintf("%s: %s", msg, e.UnderlyingError)
}

func (s *Service) getSyntacticUpload(ctx context.Context, repo types.Repo, commit api.CommitID, path string) (uploadsshared.CompletedUpload, *SyntacticUsagesError) {
	uploads, err := s.GetClosestCompletedUploadsForBlob(ctx, uploadsshared.UploadMatchingOptions{
		RepositoryID:       int(repo.ID),
		Commit:             string(commit),
		Indexer:            uploadsshared.SyntacticIndexer,
		Path:               path,
		RootToPathMatching: uploadsshared.RootMustEnclosePath,
	})

	if err != nil || len(uploads) == 0 {
		return uploadsshared.CompletedUpload{}, &SyntacticUsagesError{
			Code:            SU_NoSyntacticIndex,
			UnderlyingError: err,
		}
	}

	if len(uploads) > 1 {
		s.logger.Warn(
			"Multiple syntactic uploads found, picking the first one",
			log.String("repo", repo.URI),
			log.String("commit", commit.Short()),
			log.String("path", path),
		)
	}
	return uploads[0], nil
}

// getSyntacticSymbolsAtRange tries to look up the symbols at the given coordinates
// in a syntactic upload. If this function returns an error you should most likely
// log and handle it instead of rethrowing, as this could fail for a myriad of reasons
// (some broken invariant internally, network issue etc.)
func (s *Service) getSyntacticSymbolsAtRange(
	ctx context.Context,
	repo types.Repo,
	commit api.CommitID,
	path string,
	symbolRange scip.Range,
) (symbols []*scip.Symbol, uploadId int, err *SyntacticUsagesError) {
	syntacticUpload, err := s.getSyntacticUpload(ctx, repo, commit, path)
	if err != nil {
		return nil, 0, err
	}

	doc, docErr := s.SCIPDocument(ctx, syntacticUpload.ID, path)
	if docErr != nil {
		return nil, 0, &SyntacticUsagesError{
			Code:            SU_NoSyntacticIndex,
			UnderlyingError: docErr,
		}
	}

	// TODO: Adjust symbolRange based on revision vs syntacticUpload.Commit

	symbols = make([]*scip.Symbol, 0)
	var parseFail *scip.Occurrence = nil

	// FIXME(issue: GRAPH-674): Properly handle different text encodings here.
	for _, occurrence := range findOccurrencesWithEqualRange(doc.Occurrences, symbolRange) {
		parsedSymbol, err := scip.ParseSymbol(occurrence.Symbol)
		if err != nil {
			parseFail = occurrence
			continue
		}
		symbols = append(symbols, parsedSymbol)
	}

	if parseFail != nil {
		s.logger.Warn("getSyntacticSymbolsAtRange: Failed to parse symbol", log.String("symbol", parseFail.Symbol))
	}

	return symbols, syntacticUpload.ID, nil
}

func (s *Service) findSyntacticMatchesForCandidateFile(
	ctx context.Context,
	uploadId int,
	filePath string,
	candidateFile candidateFile,
) []SyntacticMatch {
	results := make([]SyntacticMatch, 0)

	document, docErr := s.SCIPDocument(ctx, uploadId, filePath)
	if docErr != nil {
		return results
	}

	// TODO: We can optimize this further by continuously slicing the occurrences array
	// as both these arrays are sorted
	for _, candidateRange := range candidateFile.matches {
		for _, occ := range findOccurrencesWithEqualRange(document.Occurrences, candidateRange) {
			if !scip.IsLocalSymbol(occ.Symbol) {
				results = append(results, SyntacticMatch{
					Path:       filePath,
					Occurrence: occ,
				})
			}
		}
	}

	return results
}

type SyntacticMatch struct {
	Path       string
	Occurrence *scip.Occurrence
}

func (m *SyntacticMatch) Range() scip.Range {
	return scip.NewRangeUnchecked(m.Occurrence.Range)
}

func (s *Service) SyntacticUsages(
	ctx context.Context,
	repo types.Repo,
	commit api.CommitID,
	path string,
	symbolRange scip.Range,
) ([]SyntacticMatch, *SyntacticUsagesError) {
	// The `nil` in the second argument is here, because `With` does not work with custom error types.
	ctx, trace, endObservation := s.operations.syntacticUsages.With(ctx, nil, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repoId", int(repo.ID)),
		attribute.String("commit", string(commit)),
		attribute.String("path", path),
		attribute.String("symbolRange", symbolRange.String()),
	}})
	defer endObservation(1, observation.Args{})

	symbolsAtRange, uploadId, err := s.getSyntacticSymbolsAtRange(ctx, repo, commit, path, symbolRange)
	if err != nil {
		return nil, err
	}

	if len(symbolsAtRange) == 0 {
		s.logger.Warn("getSyntacticSymbolsAtRange: No symbols found at requested range")
		return nil, &SyntacticUsagesError{
			Code:            SU_NoSymbolAtRequestedRange,
			UnderlyingError: nil,
		}
	}
	// Overlapping symbolsAtRange should lead to the same display name, but be scored separately.
	// (Meaning we just need a single Searcher/Zoekt search)
	searchSymbol := symbolsAtRange[0]

	langs, _ := languages.GetLanguages(path, nil)
	if len(langs) != 1 {
		langErr := errors.New("Unknown language")
		if len(langs) > 1 {
			langErr = errors.New("Ambiguous language")
		}
		return nil, &SyntacticUsagesError{
			Code:            SU_FailedToSearch,
			UnderlyingError: langErr,
		}
	}

	candidateMatches, matchCount, searchErr := findCandidateOccurrencesViaSearch(
		ctx, s.searchClient, s.logger,
		repo, commit, searchSymbol, langs[0],
	)
	if searchErr != nil {
		return nil, &SyntacticUsagesError{
			Code:            SU_FailedToSearch,
			UnderlyingError: searchErr,
		}
	}
	trace.AddEvent("findCandidateOccurrencesViaSearch", attribute.Int("matchCount", matchCount))

	results := make([][]SyntacticMatch, candidateMatches.Len())

	for pair := candidateMatches.Oldest(); pair != nil; pair = pair.Next() {
		syntacticMatches := s.findSyntacticMatchesForCandidateFile(ctx, uploadId, pair.Key, pair.Value)
		results = append(results, syntacticMatches)
	}
	return slices.Concat(results...), nil
}
