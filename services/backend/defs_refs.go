package backend

import (
	"context"
	"path"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log15 "gopkg.in/inconshreveable/log15.v2"

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

func (s *defs) DeprecatedTotalRefs(ctx context.Context, repoURI string) (res int, err error) {
	ctx, done := trace(ctx, "Defs", "DeprecatedTotalRefs", repoURI, &err)
	defer done()
	return localstore.DeprecatedGlobalRefs.DeprecatedTotalRefs(ctx, repoURI)
}

func (s *defs) TotalRefs(ctx context.Context, source string) (res int, err error) {
	ctx, done := trace(ctx, "Defs", "TotalRefs", source, &err)
	defer done()
	return localstore.GlobalRefs.TotalRefs(ctx, source)
}

func (s *defs) TopDefs(ctx context.Context, op sourcegraph.TopDefsOptions) (res *sourcegraph.TopDefs, err error) {
	ctx, done := trace(ctx, "Defs", "TopDefs", op, &err)
	defer done()
	return localstore.GlobalRefs.TopDefs(ctx, op)
}

func (s *defs) RefLocations(ctx context.Context, op sourcegraph.RefLocationsOptions) (res *sourcegraph.RefLocations, err error) {
	ctx, done := trace(ctx, "Defs", "RefLocations", op, &err)
	defer done()
	return localstore.GlobalRefs.RefLocations(ctx, op)
}

var indexDuration = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "src",
	Name:      "index_duration_seconds",
	Help:      "Duration of time that indexing a repository takes.",
})

func init() {
	prometheus.MustRegister(indexDuration)
}

func (s *defs) RefreshIndex(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) (err error) {
	start := time.Now()
	defer func() {
		indexDuration.Set(time.Since(start).Seconds())
	}()

	if Mocks.Defs.RefreshIndex != nil {
		return Mocks.Defs.RefreshIndex(ctx, op)
	}

	ctx, done := trace(ctx, "Defs", "RefreshIndex", op, &err)
	defer done()

	// Refuse to index private repositories. For the time being, we do not. We
	// must decide on an approach, and there are serious implications to both
	// approaches.
	repo, err := Repos.Get(ctx, &sourcegraph.RepoSpec{ID: op.Repo})
	if err != nil {
		return err
	}
	if repo.Private {
		log15.Warn("Refusing to index private repository", "repo", repo.URI)
		return nil
	}

	rev, err := Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: op.Repo})
	if err != nil {
		return err
	}

	// Refresh global references indexes.
	return localstore.GlobalRefs.RefreshIndex(ctx, repo.URI, rev.CommitID)
}
