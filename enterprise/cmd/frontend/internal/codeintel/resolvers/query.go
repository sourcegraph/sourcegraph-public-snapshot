package resolvers

import (
	"context"

	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// AdjustedLocation is a path and range pair from within a particular upload. The adjusted commit
// denotes the target commit for which the location was adjusted (the originally requested commit).
type AdjustedLocation struct {
	Dump           store.Dump
	Path           string
	AdjustedCommit string
	AdjustedRange  lsifstore.Range
}

// AdjustedDiagnostic is a diagnostic from within a particular upload. The adjusted commit denotes
// the target commit for which the location was adjusted (the originally requested commit).
type AdjustedDiagnostic struct {
	lsifstore.Diagnostic
	Dump           store.Dump
	AdjustedCommit string
	AdjustedRange  lsifstore.Range
}

// AdjustedCodeIntelligenceRange stores definition, reference, and hover information for all ranges
// within a block of lines. The definition and reference locations have been adjusted to fit the
// target (originally requested) commit.
type AdjustedCodeIntelligenceRange struct {
	Range           lsifstore.Range
	Definitions     []AdjustedLocation
	References      []AdjustedLocation
	Implementations []AdjustedLocation
	HoverText       string
}

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
	db        database.DB
	dbStore   DBStore
	lsifStore LSIFStore

	cachedCommitChecker *cachedCommitChecker

	repositoryID int
	commit       string
	path         string

	operations *operations

	symbolsResolver SymbolsResolver
}

// NewQueryResolver create a new query resolver with the given services. The methods of this
// struct return queries for the given repository, commit, and path, and will query only the
// bundles associated with the given dump objects.
func NewQueryResolver(
	db database.DB,
	dbStore DBStore,
	lsifStore LSIFStore,
	cachedCommitChecker *cachedCommitChecker,
	repositoryID int,
	commit string,
	path string,
	operations *operations,
	symbolsResolver SymbolsResolver,
) QueryResolver {
	return &queryResolver{
		db:                  db,
		dbStore:             dbStore,
		lsifStore:           lsifStore,
		cachedCommitChecker: cachedCommitChecker,
		operations:          operations,
		repositoryID:        repositoryID,
		commit:              commit,
		path:                path,
		symbolsResolver:     symbolsResolver,
	}
}
