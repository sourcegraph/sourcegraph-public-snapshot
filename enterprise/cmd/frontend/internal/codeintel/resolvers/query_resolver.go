package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
)

// QueryResolver is the main interface to bundle-related operations exposed to the GraphQL API. This
// resolver consolidates the logic for bundle operations and is not itself concerned with GraphQL/API
// specifics (auth, validation, marshaling, etc.). This resolver is wrapped by a symmetrics resolver
// in this package's graphql subpackage, which is exposed directly by the API.
type QueryResolver interface {
	LSIFUploads(ctx context.Context) ([]store.Upload, error)
	Ranges(ctx context.Context, startLine, endLine int) ([]AdjustedCodeIntelligenceRange, error)
	Stencil(ctx context.Context) ([]lsifstore.Range, error)
	Diagnostics(ctx context.Context, limit int) ([]AdjustedDiagnostic, int, error)
	Hover(ctx context.Context, line, character int) (string, lsifstore.Range, bool, error)
	Definitions(ctx context.Context, line, character int) ([]AdjustedLocation, error)
	References(ctx context.Context, line, character, limit int, rawCursor string) ([]AdjustedLocation, string, error)
	Implementations(ctx context.Context, line, character, limit int, rawCursor string) ([]AdjustedLocation, string, error)
}

type queryResolver struct {
	repositoryID int
	commit       string
	path         string

	operations *operations

	codenavResolver CodeNavResolver
	requestState    codenav.RequestState
}

// NewQueryResolver create a new query resolver with the given services. The methods of this
// struct return queries for the given repository, commit, and path, and will query only the
// bundles associated with the given dump objects.
func NewQueryResolver(repositoryID int, commit string, path string, operations *operations, codenavResolver CodeNavResolver, requestState codenav.RequestState) QueryResolver {
	return &queryResolver{
		operations:      operations,
		repositoryID:    repositoryID,
		commit:          commit,
		path:            path,
		codenavResolver: codenavResolver,
		requestState:    requestState,
	}
}

// LSIFUploads returns the list of dbstore.Uploads for the store.Dumps determined to be applicable
// for answering code-intel queries.
func (r *queryResolver) LSIFUploads(ctx context.Context) ([]dbstore.Upload, error) {
	uploads, err := r.codenavResolver.LSIFUploads(ctx, r.requestState)
	if err != nil {
		return []dbstore.Upload{}, err
	}

	dbUploads := []dbstore.Upload{}
	for _, u := range uploads {
		dbUploads = append(dbUploads, sharedDumpToDbstoreUpload(u))
	}

	return dbUploads, err
}

// Ranges returns code intelligence for the ranges that fall within the given range of lines. These
// results are partial and do not include references outside the current file, or any location that
// requires cross-linking of bundles (cross-repo or cross-root).
func (r *queryResolver) Ranges(ctx context.Context, startLine, endLine int) (adjustedRanges []AdjustedCodeIntelligenceRange, err error) {
	args := shared.RequestArgs{
		RepositoryID: r.repositoryID,
		Commit:       r.commit,
		Path:         r.path,
	}
	rngs, err := r.codenavResolver.Ranges(ctx, args, r.requestState, startLine, endLine)
	if err != nil {
		return nil, err
	}

	adjustedRanges = sharedRangeToAdjustedRange(rngs)

	return adjustedRanges, nil
}

// Stencil returns all ranges within a single document.
func (r *queryResolver) Stencil(ctx context.Context) (adjustedRanges []lsifstore.Range, err error) {
	args := shared.RequestArgs{
		RepositoryID: r.repositoryID,
		Commit:       r.commit,
		Path:         r.path,
	}
	ranges, err := r.codenavResolver.Stencil(ctx, args, r.requestState)
	for _, r := range ranges {
		adjustedRanges = append(adjustedRanges, sharedRangeTolsifstoreRange(r))
	}
	return adjustedRanges, err
}

// Diagnostics returns the diagnostics for documents with the given path prefix.
func (r *queryResolver) Diagnostics(ctx context.Context, limit int) (adjustedDiagnostics []AdjustedDiagnostic, _ int, err error) {
	args := shared.RequestArgs{
		RepositoryID: r.repositoryID,
		Commit:       r.commit,
		Path:         r.path,
		Limit:        limit,
	}
	diag, cursor, err := r.codenavResolver.Diagnostics(ctx, args, r.requestState)
	if err != nil {
		return nil, 0, err
	}

	adjustedDiag := sharedDiagnosticAtUploadToAdjustedDiagnostic(diag)

	return adjustedDiag, cursor, nil
}

// Hover returns the hover text and range for the symbol at the given position.
func (r *queryResolver) Hover(ctx context.Context, line, character int) (_ string, _ lsifstore.Range, _ bool, err error) {
	args := shared.RequestArgs{
		RepositoryID: r.repositoryID,
		Commit:       r.commit,
		Path:         r.path,
		Line:         line,
		Character:    character,
	}
	text, rnge, ok, err := r.codenavResolver.Hover(ctx, args, r.requestState)
	return text, sharedRangeTolsifstoreRange(rnge), ok, err
}

// Definitions returns the list of source locations that define the symbol at the given position.
func (r *queryResolver) Definitions(ctx context.Context, line, character int) (_ []AdjustedLocation, err error) {
	args := shared.RequestArgs{
		RepositoryID: r.repositoryID,
		Commit:       r.commit,
		Path:         r.path,
		Line:         line,
		Character:    character,
	}
	defs, err := r.codenavResolver.Definitions(ctx, args, r.requestState)
	if err != nil {
		return nil, err
	}

	return uploadLocationToAdjustedLocations(defs), nil
}

// References returns the list of source locations that reference the symbol at the given position.
func (r *queryResolver) References(ctx context.Context, line, character, limit int, rawCursor string) (_ []AdjustedLocation, _ string, err error) {
	args := shared.RequestArgs{
		RepositoryID: r.repositoryID,
		Commit:       r.commit,
		Path:         r.path,
		Line:         line,
		Character:    character,
		Limit:        limit,
		RawCursor:    rawCursor,
	}
	refs, cursor, err := r.codenavResolver.References(ctx, args, r.requestState)
	if err != nil {
		return nil, "", err
	}

	adjstedLoc := uploadLocationToAdjustedLocations(refs)

	return adjstedLoc, cursor, nil
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
	impl, cursor, err := r.codenavResolver.Implementations(ctx, args, r.requestState)
	if err != nil {
		return nil, "", err
	}

	adjustedLoc := uploadLocationToAdjustedLocations(impl)

	return adjustedLoc, cursor, nil
}
