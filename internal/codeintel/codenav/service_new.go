package codenav

import (
	"context"
	"strings"

	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *Service) GetDefinitions(
	ctx context.Context,
	args OccurrenceRequestArgs,
	requestState RequestState,
	cursor PreciseCursor,
) (_ []shared.UploadUsage, nextCursor PreciseCursor, err error) {
	return s.gatherLocations(
		ctx, args, requestState, cursor,
		s.operations.getDefinitions, // operation
		shared.UsageKindDefinition,
		false, // includeReferencingIndexes
		LocationExtractorNamedFunc{shared.UsageKindDefinition, s.lsifstore.ExtractDefinitionLocationsFromPosition},
	)
}

func (s *Service) GetReferences(
	ctx context.Context,
	args OccurrenceRequestArgs,
	requestState RequestState,
	cursor PreciseCursor,
) (_ []shared.UploadUsage, nextCursor PreciseCursor, err error) {
	return s.gatherLocations(
		ctx, args, requestState, cursor,
		s.operations.getReferences, // operation
		shared.UsageKindReference,
		true, // includeReferencingIndexes
		LocationExtractorNamedFunc{shared.UsageKindReference, s.lsifstore.ExtractReferenceLocationsFromPosition},
	)
}

func (s *Service) GetImplementations(
	ctx context.Context,
	args OccurrenceRequestArgs,
	requestState RequestState,
	cursor PreciseCursor,
) (_ []shared.UploadUsage, nextCursor PreciseCursor, err error) {
	return s.gatherLocations(
		ctx, args, requestState, cursor,
		s.operations.getImplementations, // operation
		shared.UsageKindImplementation,
		true, // includeReferencingIndexes
		LocationExtractorNamedFunc{shared.UsageKindImplementation, s.lsifstore.ExtractImplementationLocationsFromPosition},
	)
}

func (s *Service) GetPrototypes(
	ctx context.Context,
	args OccurrenceRequestArgs,
	requestState RequestState,
	cursor PreciseCursor,
) (_ []shared.UploadUsage, nextCursor PreciseCursor, err error) {
	return s.gatherLocations(
		ctx, args, requestState, cursor,
		s.operations.getPrototypes, // operation
		shared.UsageKindSuper,
		false, // includeReferencingIndexes
		LocationExtractorNamedFunc{shared.UsageKindSuper, s.lsifstore.ExtractPrototypeLocationsFromPosition},
	)
}

type LocationExtractor interface {
	// Extract converts a key (an object that can match a small part of a Document) into a
	// set of Usages within _that specific document_ related to the symbol at that position, as well
	// as the set of related symbol names that should be searched in other indexes for a complete result
	// set.
	//
	// The relationship between symbols is implementation specific.
	//
	// The return usages will all have Path == key.Path and UploadID == key.UploadID
	Extract(ctx context.Context, key lsifstore.FindUsagesKey) ([]shared.Usage, []string, error)
}

type LocationExtractorNamedFunc struct {
	Kind shared.UsageKind
	Func func(context.Context, lsifstore.FindUsagesKey) ([]shared.UsageBuilder, []string, error)
}

func (f LocationExtractorNamedFunc) Extract(ctx context.Context, key lsifstore.FindUsagesKey) ([]shared.Usage, []string, error) {
	builders, symbols, err := f.Func(ctx, key)
	return shared.BuildUsages(builders, key.UploadID, key.Path, f.Kind), symbols, err
}

func (s *Service) gatherLocations(
	ctx context.Context,
	args OccurrenceRequestArgs,
	requestState RequestState,
	cursor PreciseCursor,
	operation *observation.Operation,
	usageKind shared.UsageKind,
	includeReferencingIndexes bool,
	extractor LocationExtractor,
) (allOccurrences []shared.UploadUsage, _ PreciseCursor, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, operation, serviceObserverThreshold,
		observation.Args{Attrs: observation.MergeAttributes(args.Attrs(), requestState.Attrs()...)})
	defer endObservation()

	if cursor.Phase == "" {
		cursor.Phase = "local"
	}

	// First, we determine the set of SCIP indexes that can act as one of our "roots" for the
	// following traversal. We see which SCIP indexes cover the particular query position and
	// stash this metadata on the cursor for subsequent queries.

	var visibleUploads []visibleUpload

	// N.B.: cursor is purposefully re-assigned here
	visibleUploads, cursor, err = s.getVisibleUploadsFromCursor(
		ctx,
		args,
		requestState,
		cursor,
	)
	if err != nil {
		return nil, PreciseCursor{}, err
	}

	var visibleUploadIDs []int
	for _, upload := range visibleUploads {
		visibleUploadIDs = append(visibleUploadIDs, upload.Upload.ID)
	}
	trace.AddEvent("VisibleUploads", attribute.IntSlice("visibleUploadIDs", visibleUploadIDs))

	// The following loop calls local and remote location resolution phases in alternation. As
	// each phase controls whether or not it should execute, this is safe.
	//
	// Such a loop exists as each invocation of either phase may produce fewer results than the
	// requested page size. For example, the local phase may have a small number of results but
	// the remote phase has additional results that could fit on the first page. Similarly, if
	// there are many references to a symbol over a large number of indexes but each index has
	// only a small number of locations, they can all be combined into a single page. Running
	// each phase multiple times and combining the results will create a full page, if the
	// result set was not exhausted), on each round-trip call to this service's method.

outer:
	for cursor.Phase != "done" {
		for _, gatherLocations := range []gatherLocationsFunc{s.gatherLocalLocations, s.gatherRemoteLocationsShim} {
			trace.AddEvent("Gather", attribute.String("phase", cursor.Phase), attribute.Int("numLocationsGathered", len(allOccurrences)))

			if len(allOccurrences) >= args.Limit {
				// we've filled our page, exit with current results
				break outer
			}

			var occurrences []shared.UploadUsage

			// N.B.: cursor is purposefully re-assigned here
			occurrences, cursor, err = gatherLocations(
				ctx,
				trace,
				args.RequestArgs(),
				requestState,
				usageKind,
				includeReferencingIndexes,
				cursor,
				args.Limit-len(allOccurrences), // remaining space in the page
				extractor,
				visibleUploads,
			)
			if err != nil {
				return nil, PreciseCursor{}, err
			}
			allOccurrences = append(allOccurrences, occurrences...)
		}
	}

	return allOccurrences, cursor, nil
}

func (s *Service) getVisibleUploadsFromCursor(
	ctx context.Context,
	args OccurrenceRequestArgs,
	requestState RequestState,
	cursor PreciseCursor,
) ([]visibleUpload, PreciseCursor, error) {

	if cursor.VisibleUploads != nil {
		visibleUploads := make([]visibleUpload, 0, len(cursor.VisibleUploads))
		for _, u := range cursor.VisibleUploads {
			upload, ok := requestState.dataLoader.GetUploadFromCacheMap(u.UploadID)
			if !ok {
				return nil, PreciseCursor{}, ErrConcurrentModification
			}

			// OK to use Unchecked functions at ~serialization boundary for simplicity.
			visibleUploads = append(visibleUploads, visibleUpload{
				Upload:        upload,
				TargetPath:    core.NewRepoRelPathUnchecked(u.TargetPath),
				TargetMatcher: u.TargetMatcher.ToShared(),
			})
		}

		return visibleUploads, cursor, nil
	}

	visibleUploads, err := s.getVisibleUploads(ctx, args.Matcher, requestState)
	if err != nil {
		return nil, PreciseCursor{}, err
	}

	cursorVisibleUpload := make([]CursorVisibleUpload, 0, len(visibleUploads))
	for i := range visibleUploads {
		cursorVisibleUpload = append(cursorVisibleUpload, CursorVisibleUpload{
			UploadID:      visibleUploads[i].Upload.ID,
			TargetPath:    visibleUploads[i].TargetPath.RawValue(),
			TargetMatcher: NewCursorMatcher(visibleUploads[i].TargetMatcher),
		})
	}

	cursor.VisibleUploads = cursorVisibleUpload
	return visibleUploads, cursor, nil
}

type gatherLocationsFunc func(
	ctx context.Context,
	trace observation.TraceLogger,
	args RequestArgs,
	requestState RequestState,
	usageKind shared.UsageKind,
	includeReferencingIndexes bool,
	cursor PreciseCursor,
	limit int,
	extractor LocationExtractor,
	visibleUploads []visibleUpload,
) ([]shared.UploadUsage, PreciseCursor, error)

const skipPrefix = "lsif ."

func (s *Service) gatherLocalLocations(
	ctx context.Context,
	trace observation.TraceLogger,
	args RequestArgs,
	requestState RequestState,
	_ shared.UsageKind,
	includeReferencingIndexes bool,
	cursor PreciseCursor,
	limit int,
	extractor LocationExtractor,
	visibleUploads []visibleUpload,
) (allLocations []shared.UploadUsage, _ PreciseCursor, _ error) {
	if cursor.Phase != "local" {
		// not our turn
		return nil, cursor, nil
	}
	if cursor.LocalUploadOffset >= len(visibleUploads) {
		// nothing left to do
		cursor.Phase = "remote"
		return nil, cursor, nil
	}
	unconsumedVisibleUploads := visibleUploads[cursor.LocalUploadOffset:]

	var unconsumedVisibleUploadIDs []int
	for _, u := range unconsumedVisibleUploads {
		unconsumedVisibleUploadIDs = append(unconsumedVisibleUploadIDs, u.Upload.ID)
	}
	trace.AddEvent("GatherLocalLocations", attribute.IntSlice("unconsumedVisibleUploadIDs", unconsumedVisibleUploadIDs))

	// Create local copy of mutable cursor scope and normalize it before use.
	// We will re-assign these values back to the response cursor before the
	// function exits.
	allSymbolNames := collections.NewSet(cursor.SymbolNames...)
	skipPathsByUploadID := cursor.SkipPathsByUploadID

	if skipPathsByUploadID == nil {
		// prevent writes to nil map
		skipPathsByUploadID = map[int]string{}
	}

	for _, visibleUpload := range unconsumedVisibleUploads {
		if len(allLocations) >= limit {
			// break if we've already hit our page maximum
			break
		}

		// Gather response locations directly from the document containing the
		// target position. This may also return relevant symbol names that we
		// collect for a remote search.
		usages, symbolNames, err := extractor.Extract(
			ctx,
			lsifstore.FindUsagesKey{
				UploadID: visibleUpload.Upload.ID,
				Path:     visibleUpload.TargetPathWithoutRoot(),
				Matcher:  visibleUpload.TargetMatcher,
			},
		)
		// Attach the upload ID, TargetPath and def/ref information here?

		if err != nil {
			return nil, PreciseCursor{}, err
		}
		trace.AddEvent("ReadDocument", attribute.Int("numLocations", len(usages)), attribute.Int("numSymbolNames", len(symbolNames)))

		// remaining space in the page
		pageLimit := limit - len(allLocations)

		// Perform pagination on this level instead of in lsifstore; we bring back the
		// raw SCIP document payload anyway, so there's no reason to hide behind the API
		// that it's doing that amount of work.
		totalCount := len(usages)
		usages = pageSlice(usages, pageLimit, cursor.LocalLocationOffset)

		// adjust cursor offset for next page
		cursor = cursor.BumpLocalLocationOffset(len(usages), totalCount)

		// consume locations
		if len(usages) > 0 {
			adjustedLocations, err := s.getUploadLocations(
				ctx,
				args,
				requestState,
				usages,
				true,
			)
			if err != nil {
				return nil, PreciseCursor{}, err
			}
			allLocations = append(allLocations, adjustedLocations...)

			// Stash paths with non-empty locations in the cursor so we can prevent
			// local and "remote" searches from returning duplicate sets of of target
			// ranges.
			skipPathsByUploadID[visibleUpload.Upload.ID] = visibleUpload.TargetPathWithoutRoot().RawValue()
		}

		// stash relevant symbol names in cursor
		for _, symbolName := range symbolNames {
			if !strings.HasPrefix(symbolName, skipPrefix) {
				allSymbolNames.Add(symbolName)
			}
		}
	}

	// re-assign mutable cursor scope to response cursor
	cursor.SymbolNames = collections.SortedSetValues(allSymbolNames)
	cursor.SkipPathsByUploadID = skipPathsByUploadID

	return allLocations, cursor, nil
}

func (s *Service) gatherRemoteLocationsShim(
	ctx context.Context,
	trace observation.TraceLogger,
	args RequestArgs,
	requestState RequestState,
	usageKind shared.UsageKind,
	includeReferencingIndexes bool,
	cursor PreciseCursor,
	limit int,
	_ LocationExtractor,
	_ []visibleUpload,
) ([]shared.UploadUsage, PreciseCursor, error) {
	return s.gatherRemoteLocations(
		ctx,
		trace,
		args,
		requestState,
		cursor,
		usageKind,
		includeReferencingIndexes,
		limit,
	)
}

func (s *Service) gatherRemoteLocations(
	ctx context.Context,
	trace observation.TraceLogger,
	args RequestArgs,
	requestState RequestState,
	cursor PreciseCursor,
	usageKind shared.UsageKind,
	includeReferencingIndexes bool,
	limit int,
) ([]shared.UploadUsage, PreciseCursor, error) {
	if cursor.Phase != "remote" {
		// not our turn
		return nil, cursor, nil
	}
	trace.AddEvent("GatherRemoteLocations", attribute.StringSlice("symbolNames", cursor.SymbolNames))

	monikers, err := symbolsToMonikers(cursor.SymbolNames)
	if err != nil {
		return nil, PreciseCursor{}, err
	}
	if len(monikers) == 0 {
		// no symbol names from local phase
		return nil, exhaustedCursor, nil
	}

	// N.B.: cursor is purposefully re-assigned here
	var includeFallbackLocations bool
	cursor, includeFallbackLocations, err = s.prepareCandidateUploads(
		ctx,
		trace,
		args,
		requestState,
		cursor,
		includeReferencingIndexes,
		monikers,
	)
	if err != nil {
		return nil, PreciseCursor{}, err
	}

	// If we have no upload ids stashed in our cursor at this point then there are no more
	// uploads to search in, and we've reached the end of our result set. Congratulations!
	if len(cursor.UploadIDs) == 0 {
		return nil, exhaustedCursor, nil
	}
	trace.AddEvent("RemoteSymbolSearch", attribute.IntSlice("uploadIDs", cursor.UploadIDs))

	// Finally, query time!
	// Fetch indexed ranges of the given symbols within the given uploads.

	globalSymbolNames := genslices.Map(monikers, func(m precise.QualifiedMonikerData) string { return m.Identifier })
	usages, totalUsageCount, err := s.lsifstore.GetSymbolUsages(ctx, lsifstore.SymbolUsagesOptions{
		UsageKind:           usageKind,
		UploadIDs:           cursor.UploadIDs,
		SkipPathsByUploadID: cursor.SkipPathsByUploadID,
		LookupSymbols:       globalSymbolNames,
		Limit:               limit,
		Offset:              cursor.RemoteLocationOffset,
	})
	if err != nil {
		return nil, PreciseCursor{}, err
	}

	// adjust cursor offset for next page
	cursor = cursor.BumpRemoteUsageOffset(len(usages), totalUsageCount)

	// Adjust locations back to target commit
	adjustedLocations, err := s.getUploadLocations(ctx, args, requestState, usages, includeFallbackLocations)
	if err != nil {
		return nil, PreciseCursor{}, err
	}

	return adjustedLocations, cursor, nil
}

// prepareCandidateUploads returns a bunch of upload IDs (via cursor.UploadIDs) which
// can be used to search for symbol definitions/references/etc.
//
//  1. If the uploads containing the definitions of the monikers are not known,
//     it identifies them and adds them to the returned cursor's DefinitionIDs and UploadIDs.
//  2. If referencing indexes are also needed (e.g. for triggering Find references
//     or for Find implementations), it will get the next page of UploadsIDs if the current
//     page is exhausted.
//
// Post-condition: The upload IDs identified are guaranteed to be loaded in
// the request data loader.
func (s *Service) prepareCandidateUploads(
	ctx context.Context,
	trace observation.TraceLogger,
	args RequestArgs,
	requestState RequestState,
	cursor PreciseCursor,
	includeReferencingIndexes bool,
	monikers []precise.QualifiedMonikerData,
) (_ PreciseCursor, fallback bool, _ error) {
	fallback = true // TODO - document

	// We always want to look into the uploads that define one of the symbols for our
	// "remote" phase. We'll conditionally also look at uploads that contain only a
	// reference (see below). We deal with the former set of uploads first in the
	// cursor.

	if len(cursor.DefinitionIDs) == 0 && len(cursor.UploadIDs) == 0 && cursor.RemoteUploadOffset == 0 {
		// N.B.: We only end up in in this branch on the first time it's invoked while
		// in the remote phase. If there truly are no definitions, we'll either have a
		// non-empty set of upload ids, or a non-zero remote upload offset on the next
		// invocation. If there are neither definitions nor an upload batch, we'll end
		// up returning an exhausted cursor from _this_ invocation.

		uploads, err := s.getUploadsWithDefinitionsForMonikers(ctx, monikers, requestState)
		if err != nil {
			return PreciseCursor{}, false, err
		}

		idSet := collections.NewSet[int]()
		for _, upload := range cursor.VisibleUploads {
			idSet.Add(upload.UploadID)
		}
		for _, upload := range uploads {
			idSet.Add(upload.ID)
		}
		ids := collections.SortedSetValues(idSet)

		fallback = false
		cursor.UploadIDs = ids
		cursor.DefinitionIDs = ids
		trace.AddEvent("Loaded indexes with definitions of symbols", attribute.IntSlice("ids", ids))
	}

	// TODO - redocument
	// This traversal isn't looking in uploads without definitions to one of the symbols
	if includeReferencingIndexes {
		// If we have no upload ids stashed in our cursor, then we'll try to fetch the next
		// batch of uploads in which we'll search for symbol names. If our remote upload offset
		// is set to -1 here, then it indicates the end of the set of relevant upload records.

		if len(cursor.UploadIDs) == 0 && cursor.RemoteUploadOffset != -1 {
			uploadIDs, _, totalCount, err := s.uploadSvc.GetUploadIDsWithReferences(
				ctx,
				monikers,
				cursor.DefinitionIDs,
				int(args.RepositoryID),
				string(args.Commit),
				requestState.maximumIndexesPerMonikerSearch, // limit
				cursor.RemoteUploadOffset,                   // offset
			)
			if err != nil {
				return PreciseCursor{}, false, err
			}

			cursor.UploadIDs = uploadIDs
			trace.AddEvent("Loaded batch of indexes with references to symbols", attribute.IntSlice("ids", uploadIDs))

			// adjust cursor offset for next page
			cursor = cursor.BumpRemoteUploadOffset(len(uploadIDs), totalCount)
		}
	}

	// Hydrate upload records into the request state data loader. This must be called prior
	// to the invocation of getUploadUsage, which will silently throw out records belonging
	// to uploads that have not yet fetched from the database. We've assumed that the data loader
	// is consistently up-to-date with any extant upload identifier reference.
	//
	// FIXME: That's a dangerous design assumption we should get rid of.
	if _, err := s.getUploadsByIDs(ctx, cursor.UploadIDs, requestState); err != nil {
		return PreciseCursor{}, false, err
	}

	return cursor, fallback, nil
}

func symbolsToMonikers(symbolNames []string) ([]precise.QualifiedMonikerData, error) {
	var monikers []precise.QualifiedMonikerData
	for _, symbolName := range symbolNames {
		parsedSymbol, err := scip.ParseSymbol(symbolName)
		if err != nil {
			return nil, err
		}
		if parsedSymbol.Package == nil {
			continue
		}

		monikers = append(monikers, precise.QualifiedMonikerData{
			MonikerData: precise.MonikerData{
				Scheme:     parsedSymbol.Scheme,
				Identifier: symbolName,
			},
			PackageInformationData: precise.PackageInformationData{
				Manager: parsedSymbol.Package.Manager,
				Name:    parsedSymbol.Package.Name,
				Version: parsedSymbol.Package.Version,
			},
		})
	}

	return monikers, nil
}

func pageSlice[T any](s []T, limit, offset int) []T {
	if offset < len(s) {
		s = s[offset:]
	} else {
		s = []T{}
	}

	if len(s) > limit {
		s = s[:limit]
	}

	return s
}

func (s *Service) PreciseUsages(ctx context.Context, requestState RequestState, args UsagesForSymbolResolvedArgs) (returnUsages []shared.UploadUsage, _ core.Option[UsagesCursor], err error) {
	ctx, trace, endObservation := s.operations.preciseUsages.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repoId", int(args.Repo.ID)),
		attribute.String("commit", string(args.CommitID)),
		attribute.String("path", args.Path.RawValue()),
		attribute.String("range", args.Range.String()),
	}})
	defer endObservation(1, observation.Args{})

	remainingCount := int(args.RemainingCount)
	noCursor := core.None[UsagesCursor]()
	currentCursor := args.Cursor
	lookupSymbol := ""
	if args.Symbol != nil {
		if args.Symbol.EqualsSymbol.Scheme == uploadsshared.SyntacticIndexer {
			return nil, noCursor, nil
		}
		lookupSymbol = args.Symbol.EqualsName
	}

	// Loop invariant: At the end of an iteration, either
	//    (1) remainingCount has decreased
	// OR (2) currentCursor.CursorType has been advanced
	for remainingCount > 0 {
		requestArgs := OccurrenceRequestArgs{
			RepositoryID: args.Repo.ID,
			Commit:       args.CommitID,
			Limit:        remainingCount,
			RawCursor:    currentCursor.PreciseCursor.Encode(),
			Path:         args.Path,
			Matcher:      shared.NewSCIPBasedMatcher(args.Range, lookupSymbol),
		}

		var err error
		var nextUsages []shared.UploadUsage
		var nextPreciseCursor PreciseCursor
		switch currentCursor.CursorType {
		case CursorTypeDefinitions:
			nextUsages, nextPreciseCursor, err = s.GetDefinitions(ctx, requestArgs, requestState, currentCursor.PreciseCursor)
		case CursorTypeImplementations:
			nextUsages, nextPreciseCursor, err = s.GetImplementations(ctx, requestArgs, requestState, currentCursor.PreciseCursor)
		case CursorTypePrototypes:
			nextUsages, nextPreciseCursor, err = s.GetPrototypes(ctx, requestArgs, requestState, currentCursor.PreciseCursor)
		case CursorTypeReferences:
			nextUsages, nextPreciseCursor, err = s.GetReferences(ctx, requestArgs, requestState, currentCursor.PreciseCursor)
		default:
			return nil, noCursor, errors.New("Non-precise cursor type in PreciseUsages")
		}

		if err != nil {
			return nil, noCursor, err
		}

		if len(nextUsages) > remainingCount {
			trace.Warn("sub-operation returned usages that exceeded limit",
				log.Int("numNextUsages", len(nextUsages)),
				log.Int("limit", remainingCount),
				log.String("cursorType", string(currentCursor.CursorType)))
		}
		returnUsages = append(returnUsages, nextUsages...)
		remainingCount -= min(remainingCount, len(nextUsages))

		currentCursor.PreciseCursor = nextPreciseCursor
		if len(nextUsages) == 0 || currentCursor.PreciseCursor.Phase == "done" {
			// Switching types requires zero-ing the precise cursor
			// as the old Service API code is meant to be used with separate
			// cursors per usage type.
			switch currentCursor.CursorType {
			case CursorTypeDefinitions:
				currentCursor = UsagesCursor{
					PreciseCursor: PreciseCursor{},
					CursorType:    CursorTypeImplementations,
				}
			case CursorTypeImplementations:
				currentCursor = UsagesCursor{
					PreciseCursor: PreciseCursor{},
					CursorType:    CursorTypePrototypes,
				}
			case CursorTypePrototypes:
				currentCursor = UsagesCursor{
					PreciseCursor: PreciseCursor{},
					CursorType:    CursorTypeReferences,
				}
			case CursorTypeReferences:
				return returnUsages, noCursor, nil
			default:
				return nil, noCursor, errors.New("Non-precise cursor type in PreciseUsages")
			}
		}
	}

	return returnUsages, core.Some(currentCursor), nil
}
