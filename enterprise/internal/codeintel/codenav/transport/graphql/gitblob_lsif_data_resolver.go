package graphql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/log"
	traceLog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// gitBlobLSIFDataResolver is the main interface to bundle-related operations exposed to the GraphQL API. This
// resolver concerns itself with GraphQL/API-specific behaviors (auth, validation, marshaling, etc.).
// All code intel-specific behavior is delegated to the underlying resolver instance, which is defined
// in the parent package.
type gitBlobLSIFDataResolver struct {
	codeNavSvc      CodeNavService
	autoindexingSvc AutoIndexingService
	uploadSvc       UploadsService
	policiesSvc     PolicyService

	requestState     codenav.RequestState
	locationResolver *sharedresolvers.CachedLocationResolver
	errTracer        *observation.ErrCollector

	operations *operations
}

// NewQueryResolver creates a new QueryResolver with the given resolver that defines all code intel-specific
// behavior. A cached location resolver instance is also given to the query resolver, which should be used
// to resolve all location-related values.
func NewGitBlobLSIFDataResolver(
	codeNavSvc CodeNavService,
	autoindexSvc AutoIndexingService,
	uploadSvc UploadsService,
	policiesSvc PolicyService,
	requestState codenav.RequestState,
	errTracer *observation.ErrCollector,
	operations *operations,
) resolverstubs.GitBlobLSIFDataResolver {
	db := autoindexSvc.GetUnsafeDB()
	return &gitBlobLSIFDataResolver{
		codeNavSvc:       codeNavSvc,
		autoindexingSvc:  autoindexSvc,
		uploadSvc:        uploadSvc,
		policiesSvc:      policiesSvc,
		requestState:     requestState,
		locationResolver: sharedresolvers.NewCachedLocationResolver(db, gitserver.NewClient(db)),
		errTracer:        errTracer,
		operations:       operations,
	}
}

func (r *gitBlobLSIFDataResolver) ToGitTreeLSIFData() (resolverstubs.GitTreeLSIFDataResolver, bool) {
	return r, true
}

func (r *gitBlobLSIFDataResolver) ToGitBlobLSIFData() (resolverstubs.GitBlobLSIFDataResolver, bool) {
	return r, true
}

func (r *gitBlobLSIFDataResolver) Stencil(ctx context.Context) (_ []resolverstubs.RangeResolver, err error) {
	args := shared.RequestArgs{RepositoryID: r.requestState.RepositoryID, Commit: r.requestState.Commit, Path: r.requestState.Path}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.stencil, time.Second, getObservationArgs(args))
	defer endObservation()

	ranges, err := r.codeNavSvc.GetStencil(ctx, args, r.requestState)
	if err != nil {
		return nil, errors.Wrap(err, "svc.GetStencil")
	}

	resolvers := make([]resolverstubs.RangeResolver, 0, len(ranges))
	for _, r := range ranges {
		resolvers = append(resolvers, NewRangeResolver(convertRange(r)))
	}

	return resolvers, nil
}

// ErrIllegalBounds occurs when a negative or zero-width bound is supplied by the user.
var ErrIllegalBounds = errors.New("illegal bounds")

// Ranges returns code intelligence for the ranges that fall within the given range of lines. These
// results are partial and do not include references outside the current file, or any location that
// requires cross-linking of bundles (cross-repo or cross-root).
func (r *gitBlobLSIFDataResolver) Ranges(ctx context.Context, args *resolverstubs.LSIFRangesArgs) (_ resolverstubs.CodeIntelligenceRangeConnectionResolver, err error) {
	requestArgs := shared.RequestArgs{RepositoryID: r.requestState.RepositoryID, Commit: r.requestState.Commit, Path: r.requestState.Path}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.ranges, time.Second, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", requestArgs.RepositoryID),
			log.String("commit", requestArgs.Commit),
			log.String("path", requestArgs.Path),
			log.Int("startLine", int(args.StartLine)),
			log.Int("endLine", int(args.EndLine)),
		},
	})
	defer endObservation()

	if args.StartLine < 0 || args.EndLine < args.StartLine {
		return nil, ErrIllegalBounds
	}

	ranges, err := r.codeNavSvc.GetRanges(ctx, requestArgs, r.requestState, int(args.StartLine), int(args.EndLine))
	if err != nil {
		return nil, err
	}

	return NewCodeIntelligenceRangeConnectionResolver(ranges, r.locationResolver), nil
}

// Definitions returns the list of source locations that define the symbol at the given position.
func (r *gitBlobLSIFDataResolver) Definitions(ctx context.Context, args *resolverstubs.LSIFQueryPositionArgs) (_ resolverstubs.LocationConnectionResolver, err error) {
	requestArgs := shared.RequestArgs{RepositoryID: r.requestState.RepositoryID, Commit: r.requestState.Commit, Path: r.requestState.Path, Line: int(args.Line), Character: int(args.Character)}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.definitions, time.Second, observation.Args{
		LogFields: []traceLog.Field{
			traceLog.Int("repositoryID", requestArgs.RepositoryID),
			traceLog.String("commit", requestArgs.Commit),
			traceLog.String("path", requestArgs.Path),
			traceLog.Int("line", requestArgs.Line),
			traceLog.Int("character", requestArgs.Character),
			traceLog.Int("limit", requestArgs.Limit),
		},
	})
	defer endObservation()

	def, err := r.codeNavSvc.GetDefinitions(ctx, requestArgs, r.requestState)
	if err != nil {
		return nil, errors.Wrap(err, "codeNavSvc.GetDefinitions")
	}

	if args.Filter != nil && *args.Filter != "" {
		filtered := def[:0]
		for _, loc := range def {
			if strings.Contains(loc.Path, *args.Filter) {
				filtered = append(filtered, loc)
			}
		}
		def = filtered
	}

	return NewLocationConnectionResolver(def, nil, r.locationResolver), nil
}

const DefaultReferencesPageSize = 100

// References returns the list of source locations that reference the symbol at the given position.
func (r *gitBlobLSIFDataResolver) References(ctx context.Context, args *resolverstubs.LSIFPagedQueryPositionArgs) (_ resolverstubs.LocationConnectionResolver, err error) {
	limit := derefInt32(args.First, DefaultReferencesPageSize)
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	rawCursor, err := DecodeCursor(args.After)
	if err != nil {
		return nil, err
	}

	requestArgs := shared.RequestArgs{RepositoryID: r.requestState.RepositoryID, Commit: r.requestState.Commit, Path: r.requestState.Path, Line: int(args.Line), Character: int(args.Character), Limit: limit, RawCursor: rawCursor}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.references, time.Second, getObservationArgs(requestArgs))
	defer endObservation()

	// Decode cursor given from previous response or create a new one with default values.
	// We use the cursor state track offsets with the result set and cache initial data that
	// is used to resolve each page. This cursor will be modified in-place to become the
	// cursor used to fetch the subsequent page of results in this result set.
	var nextCursor string
	cursor, err := decodeReferencesCursor(requestArgs.RawCursor)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("invalid cursor: %q", rawCursor))
	}

	refs, refCursor, err := r.codeNavSvc.GetReferences(ctx, requestArgs, r.requestState, cursor)
	if err != nil {
		return nil, errors.Wrap(err, "svc.GetReferences")
	}

	if refCursor.Phase != "done" {
		nextCursor = encodeReferencesCursor(refCursor)
	}

	if args.Filter != nil && *args.Filter != "" {
		filtered := refs[:0]
		for _, loc := range refs {
			if strings.Contains(loc.Path, *args.Filter) {
				filtered = append(filtered, loc)
			}
		}
		refs = filtered
	}

	return NewLocationConnectionResolver(refs, strPtr(nextCursor), r.locationResolver), nil
}

// DefaultReferencesPageSize is the implementation result page size when no limit is supplied.
const DefaultImplementationsPageSize = 100

// ErrIllegalLimit occurs when the user requests less than one object per page.
var ErrIllegalLimit = errors.New("illegal limit")

func (r *gitBlobLSIFDataResolver) Implementations(ctx context.Context, args *resolverstubs.LSIFPagedQueryPositionArgs) (_ resolverstubs.LocationConnectionResolver, err error) {
	limit := derefInt32(args.First, DefaultImplementationsPageSize)
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	rawCursor, err := DecodeCursor(args.After)
	if err != nil {
		return nil, err
	}

	requestArgs := shared.RequestArgs{RepositoryID: r.requestState.RepositoryID, Commit: r.requestState.Commit, Path: r.requestState.Path, Line: int(args.Line), Character: int(args.Character), Limit: limit, RawCursor: rawCursor}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.implementations, time.Second, getObservationArgs(requestArgs))
	defer endObservation()

	// Decode cursor given from previous response or create a new one with default values.
	// We use the cursor state track offsets with the result set and cache initial data that
	// is used to resolve each page. This cursor will be modified in-place to become the
	// cursor used to fetch the subsequent page of results in this result set.
	var nextCursor string
	cursor, err := decodeImplementationsCursor(rawCursor)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("invalid cursor: %q", rawCursor))
	}

	impls, implsCursor, err := r.codeNavSvc.GetImplementations(ctx, requestArgs, r.requestState, cursor)
	if err != nil {
		return nil, errors.Wrap(err, "codeNavSvc.GetImplementations")
	}

	if implsCursor.Phase != "done" {
		nextCursor = encodeImplementationsCursor(implsCursor)
	}

	if args.Filter != nil && *args.Filter != "" {
		filtered := impls[:0]
		for _, loc := range impls {
			if strings.Contains(loc.Path, *args.Filter) {
				filtered = append(filtered, loc)
			}
		}
		impls = filtered
	}

	return NewLocationConnectionResolver(impls, strPtr(nextCursor), r.locationResolver), nil
}

// Hover returns the hover text and range for the symbol at the given position.
func (r *gitBlobLSIFDataResolver) Hover(ctx context.Context, args *resolverstubs.LSIFQueryPositionArgs) (_ resolverstubs.HoverResolver, err error) {
	requestArgs := shared.RequestArgs{RepositoryID: r.requestState.RepositoryID, Commit: r.requestState.Commit, Path: r.requestState.Path, Line: int(args.Line), Character: int(args.Character)}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.hover, time.Second, getObservationArgs(requestArgs))
	defer endObservation()

	text, rx, exists, err := r.codeNavSvc.GetHover(ctx, requestArgs, r.requestState)
	if err != nil || !exists {
		return nil, err
	}

	return NewHoverResolver(text, sharedRangeTolspRange(rx)), nil
}

// LSIFUploads returns the list of dbstore.Uploads for the store.Dumps determined to be applicable
// for answering code-intel queries.
func (r *gitBlobLSIFDataResolver) LSIFUploads(ctx context.Context) (_ []resolverstubs.LSIFUploadResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "lsifUploads"))

	cacheUploads := r.requestState.GetCacheUploads()
	ids := make([]int, 0, len(cacheUploads))
	for _, dump := range cacheUploads {
		ids = append(ids, dump.ID)
	}

	uploads, err := r.codeNavSvc.GetDumpsByIDs(ctx, ids)

	dbUploads := []types.Upload{}
	for _, u := range uploads {
		dbUploads = append(dbUploads, sharedDumpToDbstoreUpload(u))
	}

	prefetcher := sharedresolvers.NewPrefetcher(r.autoindexingSvc, r.uploadSvc)

	resolvers := make([]resolverstubs.LSIFUploadResolver, 0, len(uploads))
	for _, upload := range dbUploads {
		resolvers = append(resolvers, sharedresolvers.NewUploadResolver(r.uploadSvc, r.autoindexingSvc, r.policiesSvc, upload, prefetcher, r.errTracer))
	}

	return resolvers, nil
}

// DefaultDiagnosticsPageSize is the diagnostic result page size when no limit is supplied.
const DefaultDiagnosticsPageSize = 100

// Diagnostics returns the diagnostics for documents with the given path prefix.
func (r *gitBlobLSIFDataResolver) Diagnostics(ctx context.Context, args *resolverstubs.LSIFDiagnosticsArgs) (_ resolverstubs.DiagnosticConnectionResolver, err error) {
	limit := derefInt32(args.First, DefaultDiagnosticsPageSize)
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	requestArgs := shared.RequestArgs{RepositoryID: r.requestState.RepositoryID, Commit: r.requestState.Commit, Path: r.requestState.Path, Limit: limit}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.diagnostics, time.Second, getObservationArgs(requestArgs))
	defer endObservation()

	diagnostics, totalCount, err := r.codeNavSvc.GetDiagnostics(ctx, requestArgs, r.requestState)
	if err != nil {
		return nil, errors.Wrap(err, "codeNavSvc.GetDiagnostics")
	}

	return NewDiagnosticConnectionResolver(diagnostics, totalCount, r.locationResolver), nil
}
