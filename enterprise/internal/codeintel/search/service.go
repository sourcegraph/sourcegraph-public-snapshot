package search

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	sgtypes "github.com/sourcegraph/sourcegraph/internal/types"
)

// Service is a shim over codeintel services for powering search over the code graph.
type Service struct {
	codeNav    *codenav.Service
	gs         *gitserver.Client
	hunkCache  codenav.HunkCache
	maxIndexes int
}

func NewService(observationCtx *observation.Context, gs *gitserver.Client, codeIntelServices *codenav.Service, maxIndexes int) (*Service, error) {
	c, err := codenav.NewHunkCache(50) // 50 is a common size used for hunk cache elsewhere
	if err != nil {
		return nil, err
	}
	return &Service{
		codeNav:    codeIntelServices,
		gs:         gs,
		hunkCache:  c,
		maxIndexes: maxIndexes,
	}, nil
}

func (s *Service) GetDefinitions(ctx context.Context, repo sgtypes.MinimalRepo, args shared.RequestArgs) (_ []types.UploadLocation, err error) {
	uploads, err := s.codeNav.GetClosestDumpsForBlob(ctx, args.RepositoryID, args.Commit, args.Path, true, "")
	if err != nil || len(uploads) == 0 {
		return nil, err
	}

	reqState := codenav.NewRequestState(uploads, authz.DefaultSubRepoPermsChecker, s.gs, repo.ToRepo(), args.Commit, args.Path, s.maxIndexes, s.hunkCache)

	locs, err := s.codeNav.GetDefinitions(ctx, args, reqState)
	return locs, err
}

func (s *Service) GetReferences(ctx context.Context, repo sgtypes.MinimalRepo, args shared.RequestArgs) (_ []types.UploadLocation, err error) {
	uploads, err := s.codeNav.GetClosestDumpsForBlob(ctx, args.RepositoryID, args.Commit, args.Path, true, "")
	if err != nil || len(uploads) == 0 {
		return nil, err
	}

	reqState := codenav.NewRequestState(uploads, authz.DefaultSubRepoPermsChecker, s.gs, repo.ToRepo(), args.Commit, args.Path, s.maxIndexes, s.hunkCache)

	// TODO: pagination, phases?
	locs, _, err := s.codeNav.GetReferences(ctx, args, reqState, shared.ReferencesCursor{
		Phase: "local",
	})
	return locs, err
}

func (s *Service) GetImplementations(ctx context.Context, repo sgtypes.MinimalRepo, args shared.RequestArgs) (_ []types.UploadLocation, err error) {
	uploads, err := s.codeNav.GetClosestDumpsForBlob(ctx, args.RepositoryID, args.Commit, args.Path, true, "")
	if err != nil || len(uploads) == 0 {
		return nil, err
	}

	reqState := codenav.NewRequestState(uploads, authz.DefaultSubRepoPermsChecker, s.gs, repo.ToRepo(), args.Commit, args.Path, s.maxIndexes, s.hunkCache)

	// TODO: pagination, phases?
	locs, _, err := s.codeNav.GetImplementations(ctx, args, reqState, shared.ImplementationsCursor{
		Phase: "local",
	})
	return locs, err
}
