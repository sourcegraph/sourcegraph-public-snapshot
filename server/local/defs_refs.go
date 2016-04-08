package local

import (
	"path"

	"gopkg.in/inconshreveable/log15.v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/server/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
)

func (s *defs) ListRefs(ctx context.Context, op *sourcegraph.DefsListRefsOp) (*sourcegraph.RefList, error) {
	defSpec := op.Def
	opt := op.Opt
	if opt == nil {
		opt = &sourcegraph.DefListRefsOptions{}
	}

	var repoFilters []srcstore.RefFilter
	if opt.Repo != "" {
		if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Defs.ListRefs", opt.Repo); err != nil {
			return nil, err
		}
		repoFilters = []srcstore.RefFilter{
			srcstore.ByRepos(opt.Repo),
		}
	} else {
		if defSpec.CommitID == "" {
			return nil, grpc.Errorf(codes.InvalidArgument, "ListRefs: CommitID is empty")
		}
		if defSpec.Repo == "" {
			return nil, grpc.Errorf(codes.InvalidArgument, "ListRefs: Repo is empty")
		}
		if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Defs.ListRefs", defSpec.Repo); err != nil {
			return nil, err
		}
		repoFilters = []srcstore.RefFilter{
			// TODO(sqs): don't restrict to same-commit
			srcstore.ByRepos(defSpec.Repo),
			srcstore.ByCommitIDs(defSpec.CommitID),
		}
	}

	refFilters := []srcstore.RefFilter{
		srcstore.ByRefDef(graph.RefDefKey{
			DefRepo:     defSpec.Repo,
			DefUnitType: defSpec.UnitType,
			DefUnit:     defSpec.Unit,
			DefPath:     defSpec.Path,
		}),
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
	if defSpec.CommitID != "" {
		refFilters = append(refFilters, srcstore.ByCommitIDs(defSpec.CommitID))
	}

	filters := append(repoFilters, refFilters...)
	bareRefs, err := store.GraphFromContext(ctx).Refs(filters...)
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

func (s *defs) ListRefLocations(ctx context.Context, op *sourcegraph.DefsListRefLocationsOp) (*sourcegraph.RefLocationsList, error) {
	refLocations, err := store.GlobalRefsFromContext(ctx).Get(ctx, op)
	if err != nil {
		// Temporarily log and ignore error in querying the global ref store.
		// TODO: fail with error here once the rollout of global ref store is complete.
		log15.Error("error fetching ref locations", "def", op.Def, "error", err)
		return &sourcegraph.RefLocationsList{}, nil
	}
	return refLocations, nil
}
