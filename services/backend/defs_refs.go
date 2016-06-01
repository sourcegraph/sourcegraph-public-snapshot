package backend

import (
	"path"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
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
	var repo, commitID string
	switch {
	case opt.Repo != "":
		repo = opt.Repo
	case defSpec.Repo != "":
		repo = defSpec.Repo
	default:
		return nil, grpc.Errorf(codes.InvalidArgument, "ListRefs: Repo must be specified")
	}
	switch {
	case opt.CommitID != "":
		commitID = opt.CommitID
	case defSpec.CommitID != "":
		commitID = defSpec.CommitID
	default:
		return nil, grpc.Errorf(codes.InvalidArgument, "ListRefs: CommitID must be specified")
	}
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Defs.ListRefs", repo); err != nil {
		return nil, err
	}

	repoFilters := []srcstore.RefFilter{
		srcstore.ByRepos(repo),
		srcstore.ByCommitIDs(commitID),
	}

	refFilters := []srcstore.RefFilter{
		srcstore.ByRefDef(graph.RefDefKey{
			DefRepo:     defSpec.Repo,
			DefUnitType: defSpec.UnitType,
			DefUnit:     defSpec.Unit,
			DefPath:     defSpec.Path,
		}),
		srcstore.ByCommitIDs(commitID),
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
	return store.GlobalRefsFromContext(ctx).Get(ctx, op)
}

func (s *defs) RefreshIndex(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) (*pbtypes.Void, error) {
	if op.RefreshRefLocations {
		if err := store.GlobalRefsFromContext(ctx).Update(ctx, *op.Repo); err != nil {
			return nil, err
		}
	}
	return &pbtypes.Void{}, nil
}
