package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GitBlobResolver is the main interface to bundle-related operations exposed to the GraphQL API. This
// resolver consolidates the logic for bundle operations and is not itself concerned with GraphQL/API
// specifics (auth, validation, marshaling, etc.). This resolver is wrapped by a symmetrics resolver
// in this package's graphql subpackage, which is exposed directly by the API.
type GitBlobResolver interface {
	LSIFUploads(ctx context.Context) ([]types.Dump, error)
	Ranges(ctx context.Context, startLine, endLine int) ([]shared.AdjustedCodeIntelligenceRange, error)
	Stencil(ctx context.Context) ([]types.Range, error)
	Diagnostics(ctx context.Context, limit int) ([]shared.DiagnosticAtUpload, int, error)
	Hover(ctx context.Context, line, character int) (string, types.Range, bool, error)
	Definitions(ctx context.Context, line, character int) ([]types.UploadLocation, error)
	References(ctx context.Context, line, character, limit int, rawCursor string) ([]types.UploadLocation, string, error)
	Implementations(ctx context.Context, line, character, limit int, rawCursor string) ([]types.UploadLocation, string, error)
}

type gitBlobResolver struct {
	svc Service

	repositoryID int
	commit       string
	path         string

	operations *operations

	requestState codenav.RequestState
}

// NewGitBlobLSIFDataResolver create a new query resolver with the given services. The methods of this
// struct return queries for the given repository, commit, and path, and will query only the
// bundles associated with the given dump objects.
func NewGitBlobResolver(svc Service, repositoryID int, commit, path string, operations *operations, requestState codenav.RequestState) GitBlobResolver {
	return &gitBlobResolver{
		svc: svc,

		repositoryID: repositoryID,
		commit:       commit,
		path:         path,

		operations: operations,

		requestState: requestState,
	}
}

// Definitions returns the list of source locations that define the symbol at the given position.
func (r *gitBlobResolver) Definitions(ctx context.Context, line, character int) (_ []types.UploadLocation, err error) {
	args := shared.RequestArgs{RepositoryID: r.repositoryID, Commit: r.commit, Path: r.path, Line: line, Character: character}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.definitions, time.Second, getObservationArgs(args))
	defer endObservation()

	def, err := r.svc.GetDefinitions(ctx, args, r.requestState)
	if err != nil {
		return nil, errors.Wrap(err, "svc.GetDefinitions")
	}

	return def, nil
}

// Diagnostics returns the diagnostics for documents with the given path prefix.
func (r *gitBlobResolver) Diagnostics(ctx context.Context, limit int) (diagnosticsAtUploads []shared.DiagnosticAtUpload, _ int, err error) {
	args := shared.RequestArgs{RepositoryID: r.repositoryID, Commit: r.commit, Path: r.path, Limit: limit}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.diagnostics, time.Second, getObservationArgs(args))
	defer endObservation()

	diag, totalCount, err := r.svc.GetDiagnostics(ctx, args, r.requestState)
	if err != nil {
		return nil, 0, errors.Wrap(err, "svc.GetDiagnostics")
	}

	return diag, totalCount, nil
}

// Hover returns the hover text and range for the symbol at the given position.
func (r *gitBlobResolver) Hover(ctx context.Context, line, character int) (_ string, _ types.Range, _ bool, err error) {
	args := shared.RequestArgs{RepositoryID: r.repositoryID, Commit: r.commit, Path: r.path, Line: line, Character: character}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.hover, time.Second, getObservationArgs(args))
	defer endObservation()

	hover, rng, ok, err := r.svc.GetHover(ctx, args, r.requestState)
	if err != nil {
		return "", types.Range{}, false, err
	}

	return hover, rng, ok, err
}

// Implementations returns the list of source locations that define the symbol at the given position.
func (r *gitBlobResolver) Implementations(ctx context.Context, line, character int, limit int, rawCursor string) (_ []types.UploadLocation, nextCursor string, err error) {
	args := shared.RequestArgs{RepositoryID: r.repositoryID, Commit: r.commit, Path: r.path, Line: line, Character: character, Limit: limit, RawCursor: rawCursor}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.implementations, time.Second, getObservationArgs(args))
	defer endObservation()

	// Decode cursor given from previous response or create a new one with default values.
	// We use the cursor state track offsets with the result set and cache initial data that
	// is used to resolve each page. This cursor will be modified in-place to become the
	// cursor used to fetch the subsequent page of results in this result set.
	cursor, err := decodeImplementationsCursor(rawCursor)
	if err != nil {
		return nil, "", errors.Wrap(err, fmt.Sprintf("invalid cursor: %q", rawCursor))
	}

	impls, implsCursor, err := r.svc.GetImplementations(ctx, args, r.requestState, cursor)
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
func (r *gitBlobResolver) LSIFUploads(ctx context.Context) (uploads []types.Dump, err error) {
	cacheUploads := r.requestState.GetCacheUploads()
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
func (r *gitBlobResolver) Ranges(ctx context.Context, startLine, endLine int) (adjustedRanges []shared.AdjustedCodeIntelligenceRange, err error) {
	args := shared.RequestArgs{RepositoryID: r.repositoryID, Commit: r.commit, Path: r.path}
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

	rng, err := r.svc.GetRanges(ctx, args, r.requestState, startLine, endLine)
	if err != nil {
		return nil, err
	}

	return rng, nil
}

// References returns the list of source locations that reference the symbol at the given position.
func (r *gitBlobResolver) References(ctx context.Context, line, character, limit int, rawCursor string) (_ []types.UploadLocation, nextCursor string, err error) {
	args := shared.RequestArgs{RepositoryID: r.repositoryID, Commit: r.commit, Path: r.path, Line: line, Character: character, Limit: limit, RawCursor: rawCursor}
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

	refs, refCursor, err := r.svc.GetReferences(ctx, args, r.requestState, cursor)
	if err != nil {
		return nil, "", errors.Wrap(err, "svc.GetReferences")
	}

	if refCursor.Phase != "done" {
		nextCursor = encodeReferencesCursor(refCursor)
	}

	return refs, nextCursor, nil
}

// Stencil returns all ranges within a single document.
func (r *gitBlobResolver) Stencil(ctx context.Context) (adjustedRanges []types.Range, err error) {
	args := shared.RequestArgs{RepositoryID: r.repositoryID, Commit: r.commit, Path: r.path}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.stencil, time.Second, getObservationArgs(args))
	defer endObservation()

	st, err := r.svc.GetStencil(ctx, args, r.requestState)
	if err != nil {
		return nil, errors.Wrap(err, "svc.GetStencil")
	}

	return st, nil
}
