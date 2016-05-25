package backend

import (
	"path"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rogpeppe/rog-go/parallel"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
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
	refLocs, err := store.GlobalRefsFromContext(ctx).Get(ctx, op)
	if err != nil {
		return nil, err
	}

	// For the rest of the function we are just filtering out results
	start := time.Now()
	defer func() {
		trackedRepo := repotrackutil.GetTrackedRepo(op.Def.Repo)
		defAccessDuration.WithLabelValues(trackedRepo).Observe(time.Since(start).Seconds())
	}()

	// HACK: set hard limit on # of repos returned for one def, to avoid making excessive number
	// of GitHub Repos.Get calls in the accesscontrol check below.
	// TODO: remove this limit once we properly cache GitHub API responses.
	repoRefs := refLocs.RepoRefs
	if len(repoRefs) > 100 {
		repoRefs = repoRefs[:100]
	}

	// Filter out repos that the user does not have access to.
	hasAccess := make([]bool, len(repoRefs))
	par := parallel.NewRun(30)
	var mu sync.Mutex
	for i_, r_ := range repoRefs {
		i, r := i_, r_
		par.Do(func() error {
			if err := accesscontrol.VerifyUserHasReadAccess(ctx, "GlobalRefs.Get", r.Repo); err == nil {
				mu.Lock()
				hasAccess[i] = true
				mu.Unlock()
			}
			return nil
		})
	}
	if err := par.Wait(); err != nil {
		return nil, err
	}

	refLocs.RepoRefs = make([]*sourcegraph.DefRepoRef, 0, len(repoRefs))
	for i, r := range repoRefs {
		if hasAccess[i] {
			refLocs.RepoRefs = append(refLocs.RepoRefs, r)
		}
	}
	return refLocs, nil
}

var defAccessDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: "src",
	Subsystem: "defs",
	Name:      "list_ref_locations_access_duration_seconds",
	Help:      "Duration for doing access checks on Def.ListRefLocations",
}, []string{"repo"})

func init() {
	prometheus.MustRegister(defAccessDuration)
}
