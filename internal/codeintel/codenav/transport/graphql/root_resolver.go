package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type RootResolver interface {
	GitBlobLSIFData(ctx context.Context, args *GitBlobLSIFDataArgs) (_ GitBlobLSIFDataResolver, err error)
}

type rootResolver struct {
	svc                            Service
	autoindexingSvc                AutoIndexingService
	uploadSvc                      UploadsService
	policiesSvc                    PolicyService
	gitserver                      GitserverClient
	maximumIndexesPerMonikerSearch int
	hunkCacheSize                  int

	// Metrics
	operations *operations
}

func NewRootResolver(svc Service, autoindexingSvc AutoIndexingService, uploadSvc UploadsService, policiesSvc PolicyService, gitserver GitserverClient, maxIndexSearch, hunkCacheSize int, observationContext *observation.Context) RootResolver {
	return &rootResolver{
		svc:                            svc,
		autoindexingSvc:                autoindexingSvc,
		uploadSvc:                      uploadSvc,
		policiesSvc:                    policiesSvc,
		gitserver:                      gitserver,
		operations:                     newOperations(observationContext),
		hunkCacheSize:                  hunkCacheSize,
		maximumIndexesPerMonikerSearch: maxIndexSearch,
	}
}

type GitBlobLSIFDataArgs struct {
	Repo      *types.Repo
	Commit    api.CommitID
	Path      string
	ExactPath bool
	ToolName  string
}

// 🚨 SECURITY: dbstore layer handles authz for query resolution
func (r *rootResolver) GitBlobLSIFData(ctx context.Context, args *GitBlobLSIFDataArgs) (_ GitBlobLSIFDataResolver, err error) {
	ctx, errTracer, endObservation := r.operations.gitBlobLsifData.WithErrors(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	uploads, err := r.svc.GetClosestDumpsForBlob(ctx, int(args.Repo.ID), string(args.Commit), args.Path, args.ExactPath, args.ToolName)
	if err != nil || len(uploads) == 0 {
		return nil, err
	}

	reqState := codenav.NewRequestState(uploads, authz.DefaultSubRepoPermsChecker, r.gitserver, args.Repo, string(args.Commit), args.Path, r.maximumIndexesPerMonikerSearch, r.hunkCacheSize)
	gbr := NewGitBlobResolver(r.svc, int(args.Repo.ID), string(args.Commit), args.Path, r.operations, reqState)

	return NewGitBlobLSIFDataResolverQueryResolver(r.autoindexingSvc, r.uploadSvc, r.policiesSvc, r.gitserver, gbr, errTracer), nil
}
