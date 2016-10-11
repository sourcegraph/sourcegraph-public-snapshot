package backend

import (
	"path"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
)

func (s *defs) DeprecatedListRefs(ctx context.Context, op *sourcegraph.DeprecatedDefsListRefsOp) (res *sourcegraph.RefList, err error) {
	if Mocks.Defs.ListRefs != nil {
		return Mocks.Defs.ListRefs(ctx, op)
	}

	ctx, done := trace(ctx, "Defs", "ListRefs", op, &err)
	defer done()

	defSpec := op.Def
	opt := op.Opt
	if opt == nil {
		opt = &sourcegraph.DeprecatedDefListRefsOptions{}
	}

	// Restrict the ref search to a single repo and commit for performance.
	if opt.Repo == 0 && defSpec.Repo != 0 {
		opt.Repo = defSpec.Repo
	}
	if opt.CommitID == "" {
		opt.CommitID = defSpec.CommitID
	}
	if opt.Repo == 0 {
		return nil, legacyerr.Errorf(legacyerr.InvalidArgument, "ListRefs: Repo must be specified")
	}
	if opt.CommitID == "" {
		return nil, legacyerr.Errorf(legacyerr.InvalidArgument, "ListRefs: CommitID must be specified")
	}

	defRepoObj, err := Repos.Get(ctx, &sourcegraph.RepoSpec{ID: defSpec.Repo})
	if err != nil {
		return nil, err
	}
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Defs.ListRefs", defRepoObj.ID); err != nil {
		return nil, err
	}

	refRepoObj, err := Repos.Get(ctx, &sourcegraph.RepoSpec{ID: opt.Repo})
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

func (s *defs) DeprecatedListRefLocations(ctx context.Context, op *sourcegraph.DeprecatedDefsListRefLocationsOp) (res *sourcegraph.DeprecatedRefLocationsList, err error) {
	if Mocks.Defs.ListRefLocations != nil {
		return Mocks.Defs.ListRefLocations(ctx, op)
	}

	ctx, done := trace(ctx, "Defs", "ListRefLocations", op, &err)
	defer done()

	return localstore.DeprecatedGlobalRefs.DeprecatedGet(ctx, op)
}

func (s *defs) RefreshIndex(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) (err error) {
	if Mocks.Defs.RefreshIndex != nil {
		return Mocks.Defs.RefreshIndex(ctx, op)
	}

	ctx, done := trace(ctx, "Defs", "RefreshIndex", op, &err)
	defer done()

	rev, err := Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: op.Repo})
	if err != nil {
		log15.Warn("Defs.RefreshIndex (Repos.ResolveRev) failed", "error", err)
		return err
	}

	// rev.CommitID will be the latest commit on the DefaultBranch
	indexOp := localstore.DeprecatedRefreshIndexOp{
		Repo:     op.Repo,
		CommitID: rev.CommitID,
	}

	// Update defs table for the exported symbols in repo.
	defsErr := localstore.Defs.Update(ctx, indexOp)

	// Update the references this repo makes to external repos
	refsErr := localstore.DeprecatedGlobalRefs.DeprecatedUpdate(ctx, indexOp)

	// We care more about defsErr, since it should be more stable. So lets
	// lean on the side of reporting it instead of refsErr. We only return
	// one error (instead of a errList or something) since it may have a
	// specific gRPC meaning.
	if defsErr != nil {
		log15.Warn("Defs.RefreshIndex (Defs.Update) failed", "error", defsErr)
		return defsErr
	}
	if refsErr != nil {
		log15.Warn("Defs.RefreshIndex (GlobalRefs.Update) failed", "error", refsErr)
		return refsErr
	}

	return nil
}
