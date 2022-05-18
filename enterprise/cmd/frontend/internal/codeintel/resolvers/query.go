package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/authz"
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
	Stencil(ctx context.Context) ([]lsifstore.Range, error)
	Ranges(ctx context.Context, startLine, endLine int) ([]AdjustedCodeIntelligenceRange, error)
	Definitions(ctx context.Context, line, character int) ([]AdjustedLocation, error)
	References(ctx context.Context, line, character, limit int, rawCursor string) ([]AdjustedLocation, string, error)
	Implementations(ctx context.Context, line, character, limit int, rawCursor string) ([]AdjustedLocation, string, error)
	Hover(ctx context.Context, line, character int) (string, lsifstore.Range, bool, error)
	Diagnostics(ctx context.Context, limit int) ([]AdjustedDiagnostic, int, error)
}

type queryResolver struct {
	db                  database.DB
	dbStore             DBStore
	lsifStore           LSIFStore
	cachedCommitChecker *cachedCommitChecker
	positionAdjuster    PositionAdjuster
	repositoryID        int
	commit              string
	path                string
	uploads             []store.Dump
	uploadCache         map[int]store.Dump
	operations          *operations
	checker             authz.SubRepoPermissionChecker

	// maximumIndexesPerMonikerSearch configures the maximum number of reference upload identifiers
	// that can be passed to a single moniker search query. Previously this limit was meant to keep
	// the number of SQLite files we'd have to open within a single call relatively low. Since we've
	// migrated to Postgres this limit is not a concern. Now we only want to limit these values
	// based on the number of elements we can pass to an IN () clause in the codeintel-db, as well
	// as the size required to encode them in a user-facing pagination cursor.
	maximumIndexesPerMonikerSearch int
}

// NewQueryResolver create a new query resolver with the given services. The methods of this
// struct return queries for the given repository, commit, and path, and will query only the
// bundles associated with the given dump objects.
func NewQueryResolver(
	db database.DB,
	dbStore DBStore,
	lsifStore LSIFStore,
	cachedCommitChecker *cachedCommitChecker,
	positionAdjuster PositionAdjuster,
	repositoryID int,
	commit string,
	path string,
	uploads []store.Dump,
	operations *operations,
	checker authz.SubRepoPermissionChecker,
	maximumIndexesPerMonikerSearch int,
) QueryResolver {
	return newQueryResolver(db, dbStore, lsifStore, cachedCommitChecker, positionAdjuster,
		repositoryID, commit, path, uploads, operations, checker, maximumIndexesPerMonikerSearch)
}

func newQueryResolver(
	db database.DB,
	dbStore DBStore,
	lsifStore LSIFStore,
	cachedCommitChecker *cachedCommitChecker,
	positionAdjuster PositionAdjuster,
	repositoryID int,
	commit string,
	path string,
	uploads []store.Dump,
	operations *operations,
	checker authz.SubRepoPermissionChecker,
	maximumIndexesPerMonikerSearch int,
) *queryResolver {
	// Maintain a map from identifers to hydrated upload records from the database. We use
	// this map as a quick lookup when constructing the resulting location set. Any additional
	// upload records pulled back from the database while processing this page will be added
	// to this map.
	uploadCache := make(map[int]store.Dump, len(uploads))
	for i := range uploads {
		uploadCache[uploads[i].ID] = uploads[i]
	}

	return &queryResolver{
		db:                             db,
		dbStore:                        dbStore,
		lsifStore:                      lsifStore,
		cachedCommitChecker:            cachedCommitChecker,
		positionAdjuster:               positionAdjuster,
		operations:                     operations,
		repositoryID:                   repositoryID,
		commit:                         commit,
		path:                           path,
		uploads:                        uploads,
		uploadCache:                    uploadCache,
		checker:                        checker,
		maximumIndexesPerMonikerSearch: maximumIndexesPerMonikerSearch,
	}
}
