package backend

import (
	"path"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sqs/pbtypes"
)

func (s *defs) ListRefs(ctx context.Context, op *sourcegraph.DefsListRefsOp) (*sourcegraph.RefList, error) {
	defSpec := op.Def
	opt := op.Opt
	if opt == nil {
		opt = &sourcegraph.DefListRefsOptions{}
	}

	// Restrict the ref search to a single repo and commit for performance.
	if opt.Repo == 0 && defSpec.Repo != 0 {
		opt.Repo = defSpec.Repo
	}
	if opt.CommitID == "" {
		opt.CommitID = defSpec.CommitID
	}
	if opt.Repo == 0 {
		return nil, grpc.Errorf(codes.InvalidArgument, "ListRefs: Repo must be specified")
	}
	if opt.CommitID == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "ListRefs: CommitID must be specified")
	}

	defRepoObj, err := svc.Repos(ctx).Get(ctx, &sourcegraph.RepoSpec{ID: defSpec.Repo})
	if err != nil {
		return nil, err
	}
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Defs.ListRefs", defRepoObj.ID); err != nil {
		return nil, err
	}

	refRepoObj, err := svc.Repos(ctx).Get(ctx, &sourcegraph.RepoSpec{ID: opt.Repo})
	if err != nil {
		return nil, err
	}
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Defs.ListRefs", refRepoObj.ID); err != nil {
		return nil, err
	}

	repoFilters := []srcstore.RefFilter{
		srcstore.ByRepos(refRepoObj.URI),
		srcstore.ByCommitIDs(opt.CommitID),
	}

	refFilters := []srcstore.RefFilter{
		srcstore.ByRefDef(graph.RefDefKey{
			DefRepo:     defRepoObj.URI,
			DefUnitType: defSpec.UnitType,
			DefUnit:     defSpec.Unit,
			DefPath:     defSpec.Path,
		}),
		srcstore.ByCommitIDs(opt.CommitID),
		srcstore.RefFilterFunc(func(ref *graph.Ref) bool { return !ref.Def }),
		srcstore.Limit(opt.Offset()+opt.Limit()+1, 0),
	}

	if len(opt.Files) > 0 {
		for i, f := range opt.Files {
			// Files need to be clean or else graphstore will panic.
			opt.Files[i] = path.Clean(f)
		}
		refFilters = append(refFilters, srcstore.ByFiles(false, opt.Files...))
	}

	filters := append(repoFilters, refFilters...)
	bareRefs, err := localstore.Graph.Refs(filters...)
	if err != nil {
		return nil, err
	}

	// Convert to sourcegraph.Ref and file bareRefs.
	refs := make([]*graph.Ref, 0, opt.Limit())
	for i, bareRef := range bareRefs {
		if i >= opt.Offset() && i < (opt.Offset()+opt.Limit()) {
			refs = append(refs, bareRef)
		}
	}
	hasMore := len(bareRefs) > opt.Offset()+opt.Limit()

	return &sourcegraph.RefList{
		Refs:           refs,
		StreamResponse: sourcegraph.StreamResponse{HasMore: hasMore},
	}, nil
}

// TODO(slimsag): Remove this in the future.
var lastKickoff = &struct {
	mu sync.Mutex
	t  time.Time
}{
	t: time.Now(),
}

// TODO(slimsag): Remove this in the future.
func (s *defs) srclibMigrate(ctx context.Context, refLocations *sourcegraph.RefLocationsList) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "srclib migrate",
		opentracing.FollowsFrom(opentracing.SpanFromContext(ctx).Context()),
	)
	defer span.Finish()

	// We kick off a refresh for repositories, but only if the last refresh
	// kicked off happened >10s ago. This is to do the migration gradually,
	// instead of spawning thousands of goroutines
	var shouldRefresh bool
	lastKickoff.mu.Lock()
	if time.Since(lastKickoff.t) > 10*time.Second {
		shouldRefresh = true
		lastKickoff.t = time.Now()
	}
	lastKickoff.mu.Unlock()

	if shouldRefresh {
		return
	}

	// Determine which repositories in the results have not been refreshed
	// since we switched off of srclib. i.e. which repos are still resolving
	// ref location file positions via srclib store.
	//
	// This is just a stop gap solution for a nice, slow, migration. To see
	// how many repos are left to be migrated:
	//
	// 	select count(distinct repo) from global_refs_new where positions IS NULL;
	//
	// Or to list them all:
	//
	// 	select distinct repo from global_refs_new where positions IS NULL;
	//
	repos := make(map[string]struct{})
	for _, r := range refLocations.RepoRefs {
		for _, f := range r.Files {
			if len(f.Positions) != 0 {
				continue
			}
			repos[r.Repo] = struct{}{}
		}
	}

	for r := range repos {
		log15.Info("[srclib migration] refreshing", "repo", r)
		repo, err := svc.Repos(ctx).Resolve(ctx, &sourcegraph.RepoResolveOp{
			Path:   r,
			Remote: true,
		})
		if err != nil {
			log15.Warn("[srclib migration] failed to kickoff migration", "error", err)
			continue
		}
		_, err = svc.Async(ctx).RefreshIndexes(ctx, &sourcegraph.AsyncRefreshIndexesOp{
			Repo:   repo.Repo,
			Source: "migration",
			Force:  true,
		})
		if err != nil {
			log15.Warn("[srclib migration] failed to submit async migration", "error", err)
			continue
		}
	}
}

func (s *defs) ListRefLocations(ctx context.Context, op *sourcegraph.DefsListRefLocationsOp) (*sourcegraph.RefLocationsList, error) {
	refLocations, err := localstore.GlobalRefs.Get(ctx, op)
	s.srclibMigrate(ctx, refLocations)
	return refLocations, err
}

func (s *defs) RefreshIndex(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) (*pbtypes.Void, error) {
	rev, err := svc.Repos(ctx).ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: op.Repo})
	if err != nil {
		return nil, err
	}

	// rev.CommitID will be the latest commit on the DefaultBranch
	indexOp := store.RefreshIndexOp{
		Repo:     op.Repo,
		CommitID: rev.CommitID,
	}

	// Update defs table for the exported symbols in repo.
	defsErr := localstore.Defs.Update(ctx, indexOp)

	// Update the references this repo makes to external repos
	refsErr := localstore.GlobalRefs.Update(ctx, indexOp)

	// We care more about defsErr, since it should be more stable. So lets
	// lean on the side of reporting it instead of refsErr. We only return
	// one error (instead of a errList or something) since it may have a
	// specific gRPC meaning.
	if defsErr != nil {
		return nil, defsErr
	}
	if refsErr != nil {
		return nil, refsErr
	}

	return &pbtypes.Void{}, nil
}
