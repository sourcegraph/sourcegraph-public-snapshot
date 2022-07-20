package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	symbols "github.com/sourcegraph/sourcegraph/internal/codeintel/symbols"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Resolver interface {
	SetUploadsDataLoader(uploads []dbstore.Dump)
	SetLocalGitTreeTranslator(client gitserver.Client, repo *types.Repo, commit, path string) error
	SetLocalCommitCache(client shared.GitserverClient)
	SetMaximumIndexesPerMonikerSearch(maxNumber int)

	References(ctx context.Context, args shared.RequestArgs) (_ []shared.UploadLocation, _ string, err error)
	Implementations(ctx context.Context, args shared.RequestArgs) (_ []shared.UploadLocation, _ string, err error)

	// temporarily needed until we move all the methods to the new resolver
	GetUploadsWithDefinitionsForMonikers(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData) ([]shared.Dump, error)
}

type resolver struct {
	svc                            *symbols.Service
	requestArgs                    *requestArgs
	gitTreeTranslator              GitTreeTranslator
	maximumIndexesPerMonikerSearch int

	// Local Request Caches
	dataLoader    *UploadsDataLoader
	hunkCacheSize int
	commitCache   CommitCache

	// Metrics
	operations *operations
}

func New(svc *symbols.Service, hunkCacheSize int, observationContext *observation.Context) Resolver {
	return &resolver{
		svc:           svc,
		operations:    newOperations(observationContext),
		dataLoader:    NewUploadsDataLoader(),
		hunkCacheSize: hunkCacheSize,
	}
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
	r.gitTreeTranslator = NewGitTreeTranslator(client, args, hunkCache)

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
