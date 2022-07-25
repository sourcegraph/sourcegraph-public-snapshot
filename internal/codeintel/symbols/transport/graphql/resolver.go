package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Resolver interface {
	SetUploadsDataLoader(uploads []dbstore.Dump)
	SetLocalGitTreeTranslator(client gitserver.Client, repo *types.Repo, commit, path string) error
	SetLocalCommitCache(client shared.GitserverClient)
	SetMaximumIndexesPerMonikerSearch(maxNumber int)
	SetAuthChecker(authChecker authz.SubRepoPermissionChecker)

	Hover(ctx context.Context, args shared.RequestArgs) (_ string, _ shared.Range, _ bool, err error)
	Definitions(ctx context.Context, args shared.RequestArgs) (_ []shared.UploadLocation, err error)
	References(ctx context.Context, args shared.RequestArgs) (_ []shared.UploadLocation, _ string, err error)
	Implementations(ctx context.Context, args shared.RequestArgs) (_ []shared.UploadLocation, _ string, err error)
	Diagnostics(ctx context.Context, args shared.RequestArgs) (diagnosticsAtUploads []shared.DiagnosticAtUpload, _ int, err error)
	Stencil(ctx context.Context, args shared.RequestArgs) (adjustedRanges []shared.Range, err error)
	Ranges(ctx context.Context, args shared.RequestArgs, startLine, endLine int) (adjustedRanges []shared.AdjustedCodeIntelligenceRange, err error)
}

type resolver struct {
	svc               Service
	requestArgs       *requestArgs
	GitTreeTranslator GitTreeTranslator

	authChecker authz.SubRepoPermissionChecker

	// maximumIndexesPerMonikerSearch configures the maximum number of reference upload identifiers
	// that can be passed to a single moniker search query. Previously this limit was meant to keep
	// the number of SQLite files we'd have to open within a single call relatively low. Since we've
	// migrated to Postgres this limit is not a concern. Now we only want to limit these values
	// based on the number of elements we can pass to an IN () clause in the codeintel-db, as well
	// as the size required to encode them in a user-facing pagination cursor.
	maximumIndexesPerMonikerSearch int

	// Local Request Caches
	dataLoader    *UploadsDataLoader
	hunkCacheSize int
	commitCache   CommitCache

	// Metrics
	operations *operations
}

func New(svc Service, hunkCacheSize int, observationContext *observation.Context) *resolver {
	return &resolver{
		svc:           svc,
		operations:    newOperations(observationContext),
		dataLoader:    NewUploadsDataLoader(),
		hunkCacheSize: hunkCacheSize,
	}
}

func (r *resolver) SetAuthChecker(authChecker authz.SubRepoPermissionChecker) {
	r.authChecker = authChecker
}

func (r *resolver) SetUploadsDataLoader(uploads []dbstore.Dump) {
	for _, upload := range uploads {
		r.dataLoader.AddUpload(upload)
	}
}

func (r *resolver) SetLocalGitTreeTranslator(client gitserver.Client, repo *types.Repo, commit, path string) error {
	hunkCache, err := NewHunkCache(r.hunkCacheSize)
	if err != nil {
		return err
	}

	args := &requestArgs{
		repo:   repo,
		commit: commit,
		path:   path,
	}

	r.requestArgs = args
	r.GitTreeTranslator = NewGitTreeTranslator(client, args, hunkCache)

	return nil
}

func (r *resolver) SetLocalCommitCache(client shared.GitserverClient) {
	r.commitCache = newCommitCache(client)
}

func (r *resolver) SetMaximumIndexesPerMonikerSearch(maxNumber int) {
	r.maximumIndexesPerMonikerSearch = maxNumber
}

func (r *resolver) Symbol(ctx context.Context, args struct{}) (_ any, err error) {
	ctx, _, endObservation := r.operations.symbol.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33374
	_, _ = ctx, args
	return nil, errors.New("unimplemented: Symbol")
}
