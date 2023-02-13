package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/graph"
	sgtypes "github.com/sourcegraph/sourcegraph/internal/types"
)

// graphSearchCodeIntelStore is a shim over codeintel services for powering search over
// the code graph. It must not use any enterprise types in its interface, and should be
// registered in package internal/search/graph on enterprise setup.
type graphSearchCodeIntelStore struct {
	svc        codeintel.Services
	gs         *gitserver.Client
	hunkCache  codenav.HunkCache
	maxIndexes int
}

func newGraphSearchCodeIntelStore(db database.DB, codeIntelServices codeintel.Services, maxIndexes int) (graph.CodeIntelStore, error) {
	c, err := codenav.NewHunkCache(50) // 50 is a common size used for hunk cache elsewhere
	if err != nil {
		return nil, err
	}
	return &graphSearchCodeIntelStore{
		svc:        codeIntelServices,
		gs:         gitserver.New(&observation.TestContext, db),
		hunkCache:  c,
		maxIndexes: maxIndexes,
	}, nil
}

func (s *graphSearchCodeIntelStore) GetDefinitions(ctx context.Context, repo sgtypes.MinimalRepo, args sgtypes.CodeIntelRequestArgs) (_ []sgtypes.CodeIntelLocation, err error) {
	uploads, err := s.svc.CodenavService.GetClosestDumpsForBlob(ctx, args.RepositoryID, args.Commit, args.Path, true, "")
	if err != nil || len(uploads) == 0 {
		return nil, err
	}

	reqState := codenav.NewRequestState(uploads, authz.DefaultSubRepoPermsChecker, s.gs, repo.ToRepo(), args.Commit, args.Path, s.maxIndexes, s.hunkCache)

	locs, err := s.svc.CodenavService.GetDefinitions(ctx, shared.RequestArgs(args), reqState)
	return toCodeIntelLocations(locs), err
}

// TODO: Make this capable of cross-repo (phase?) and pagination
func (s *graphSearchCodeIntelStore) GetReferences(ctx context.Context, repo sgtypes.MinimalRepo, args sgtypes.CodeIntelRequestArgs) (_ []sgtypes.CodeIntelLocation, err error) {
	uploads, err := s.svc.CodenavService.GetClosestDumpsForBlob(ctx, args.RepositoryID, args.Commit, args.Path, true, "")
	if err != nil || len(uploads) == 0 {
		return nil, err
	}

	reqState := codenav.NewRequestState(uploads, authz.DefaultSubRepoPermsChecker, s.gs, repo.ToRepo(), args.Commit, args.Path, s.maxIndexes, s.hunkCache)

	locs, _, err := s.svc.CodenavService.GetReferences(ctx, shared.RequestArgs(args), reqState, shared.ReferencesCursor{
		Phase: "local",
	})
	return toCodeIntelLocations(locs), err
}

// TODO: Make this capable of cross-repo (phase?) and pagination
func (s *graphSearchCodeIntelStore) GetImplementations(ctx context.Context, repo sgtypes.MinimalRepo, args sgtypes.CodeIntelRequestArgs) (_ []sgtypes.CodeIntelLocation, err error) {
	uploads, err := s.svc.CodenavService.GetClosestDumpsForBlob(ctx, args.RepositoryID, args.Commit, args.Path, true, "")
	if err != nil || len(uploads) == 0 {
		return nil, err
	}

	reqState := codenav.NewRequestState(uploads, authz.DefaultSubRepoPermsChecker, s.gs, repo.ToRepo(), args.Commit, args.Path, s.maxIndexes, s.hunkCache)

	locs, _, err := s.svc.CodenavService.GetImplementations(ctx, shared.RequestArgs(args), reqState, shared.ImplementationsCursor{
		Phase: "local",
	})
	return toCodeIntelLocations(locs), err
}

func toCodeIntelLocations(uls []types.UploadLocation) []sgtypes.CodeIntelLocation {
	ls := make([]sgtypes.CodeIntelLocation, len(uls))
	for i, l := range uls {
		ls[i] = sgtypes.CodeIntelLocation{
			Path:         l.Path,
			TargetCommit: l.TargetCommit,
			TargetRange: sgtypes.CodeIntelRange{
				Start: sgtypes.CodeIntelPosition(l.TargetRange.Start),
				End:   sgtypes.CodeIntelPosition(l.TargetRange.End),
			},
		}
	}
	return ls
}
