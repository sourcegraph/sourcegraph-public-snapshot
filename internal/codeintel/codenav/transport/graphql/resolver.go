package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Resolver interface {
	Definitions(ctx context.Context, args shared.RequestArgs) (_ []shared.UploadLocation, err error)
	Diagnostics(ctx context.Context, args shared.RequestArgs) (diagnosticsAtUploads []shared.DiagnosticAtUpload, _ int, err error)
	Hover(ctx context.Context, args shared.RequestArgs) (_ string, _ shared.Range, _ bool, err error)
	Implementations(ctx context.Context, args shared.RequestArgs) (_ []shared.UploadLocation, _ string, err error)
	Ranges(ctx context.Context, args shared.RequestArgs, startLine, endLine int) (adjustedRanges []shared.AdjustedCodeIntelligenceRange, err error)
	References(ctx context.Context, args shared.RequestArgs) (_ []shared.UploadLocation, _ string, err error)
	Stencil(ctx context.Context, args shared.RequestArgs) (adjustedRanges []shared.Range, err error)

	GetHunkCacheSize() int
}

type resolver struct {
	svc Service

	// Local Request Caches
	hunkCacheSize int

	// Metrics
	operations *operations
}

func New(svc Service, hunkCacheSize int, observationContext *observation.Context) *resolver {
	return &resolver{
		svc:           svc,
		operations:    newOperations(observationContext),
		hunkCacheSize: hunkCacheSize,
	}
}

func (r *resolver) GetHunkCacheSize() int {
	return r.hunkCacheSize
}

// Definitions returns the list of source locations that define the symbol at the given position.
func (r *resolver) Definitions(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (_ []shared.UploadLocation, err error) {
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.definitions, time.Second, getObservationArgs(args))
	defer endObservation()

	def, err := r.svc.GetDefinitions(ctx, args, requestState)
	if err != nil {
		return nil, errors.Wrap(err, "svc.GetDefinitions")
	}

	return def, nil
}

// Diagnostics returns the diagnostics for documents with the given path prefix.
func (r *resolver) Diagnostics(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (diagnosticsAtUploads []shared.DiagnosticAtUpload, _ int, err error) {
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.diagnostics, time.Second, getObservationArgs(args))
	defer endObservation()

	diag, totalCount, err := r.svc.GetDiagnostics(ctx, args, requestState)
	if err != nil {
		return nil, 0, errors.Wrap(err, "svc.GetDiagnostics")
	}

	return diag, totalCount, nil
}

// Hover returns the hover text and range for the symbol at the given position.
func (r *resolver) Hover(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (_ string, _ shared.Range, _ bool, err error) {
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.hover, time.Second, getObservationArgs(args))
	defer endObservation()

	hover, rng, ok, err := r.svc.GetHover(ctx, args, requestState)
	if err != nil {
		return "", shared.Range{}, false, err
	}

	return hover, rng, ok, err
}

// Implementations returns the list of source locations that define the symbol at the given position.
func (r *resolver) Implementations(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (_ []shared.UploadLocation, nextCursor string, err error) {
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.implementations, time.Second, getObservationArgs(args))
	defer endObservation()

	// Decode cursor given from previous response or create a new one with default values.
	// We use the cursor state track offsets with the result set and cache initial data that
	// is used to resolve each page. This cursor will be modified in-place to become the
	// cursor used to fetch the subsequent page of results in this result set.
	cursor, err := decodeImplementationsCursor(args.RawCursor)
	if err != nil {
		return nil, "", errors.Wrap(err, fmt.Sprintf("invalid cursor: %q", args.RawCursor))
	}

	impls, implsCursor, err := r.svc.GetImplementations(ctx, args, requestState, cursor)
	if err != nil {
		return nil, "", errors.Wrap(err, "svc.GetImplementations")
	}

	if implsCursor.Phase != "done" {
		nextCursor = encodeImplementationsCursor(implsCursor)
	}

	return impls, nextCursor, nil
}

// LSIFUploads returns the list of dbstore.Uploads for the store.Dumps determined to be applicable
// for answering code-intel queries.
func (r *resolver) LSIFUploads(ctx context.Context, requestState codenav.RequestState) (uploads []shared.Dump, err error) {
	cacheUploads := requestState.GetCacheUploads()
	ids := make([]int, 0, len(cacheUploads))
	for _, dump := range cacheUploads {
		ids = append(ids, dump.ID)
	}

	dumps, err := r.svc.GetDumpsByIDs(ctx, ids)

	return dumps, err
}

// Ranges returns code intelligence for the ranges that fall within the given range of lines. These
// results are partial and do not include references outside the current file, or any location that
// requires cross-linking of bundles (cross-repo or cross-root).
func (r *resolver) Ranges(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState, startLine, endLine int) (adjustedRanges []shared.AdjustedCodeIntelligenceRange, err error) {
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.ranges, time.Second, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", args.RepositoryID),
			log.String("commit", args.Commit),
			log.String("path", args.Path),
			log.Int("startLine", startLine),
			log.Int("endLine", endLine),
		},
	})
	defer endObservation()

	rng, err := r.svc.GetRanges(ctx, args, requestState, startLine, endLine)
	if err != nil {
		return nil, err
	}

	return rng, nil
}

// References returns the list of source locations that reference the symbol at the given position.
func (r *resolver) References(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (_ []shared.UploadLocation, nextCursor string, err error) {
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.references, time.Second, getObservationArgs(args))
	defer endObservation()

	// Decode cursor given from previous response or create a new one with default values.
	// We use the cursor state track offsets with the result set and cache initial data that
	// is used to resolve each page. This cursor will be modified in-place to become the
	// cursor used to fetch the subsequent page of results in this result set.
	cursor, err := decodeReferencesCursor(args.RawCursor)
	if err != nil {
		return nil, "", errors.Wrap(err, fmt.Sprintf("invalid cursor: %q", args.RawCursor))
	}

	refs, refCursor, err := r.svc.GetReferences(ctx, args, requestState, cursor)
	if err != nil {
		return nil, "", errors.Wrap(err, "svc.GetReferences")
	}

	if refCursor.Phase != "done" {
		nextCursor = encodeReferencesCursor(refCursor)
	}

	return refs, nextCursor, nil
}

// Stencil returns all ranges within a single document.
func (r *resolver) Stencil(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (adjustedRanges []shared.Range, err error) {
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.stencil, time.Second, getObservationArgs(args))
	defer endObservation()

	st, err := r.svc.GetStencil(ctx, args, requestState)
	if err != nil {
		return nil, errors.Wrap(err, "svc.GetStencil")
	}

	return st, nil
}
