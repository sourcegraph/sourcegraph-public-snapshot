package federated

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/svc"
)

// This file contains federation handlers for methods that need to be
// federated but don't have a single repo argument. These methods must
// be federated in a custom fashion.
//
// If you add/remove a custom func, method must also add/remove it
// from the list in ../gen_middleware.go's methodHasCustomFederation to prevent
// `go generate` from writing the non-custom method (which will cause
// a "Xxx redeclared in this block" compile error).

// CustomDefsList lists defs by accumulating the results of calling Defs.List on
// the Sourcegraph instances corresponding to each entry in
// opt.RepoRevs (using discovery on each repo path).
func CustomDefsList(ctx context.Context, opt *sourcegraph.DefListOptions, s sourcegraph.DefsServer) (*sourcegraph.DefList, error) {
	if len(opt.RepoRevs) == 0 {
		// No repos means all repos.
		repos, err := svc.Repos(ctx).List(ctx, &sourcegraph.RepoListOptions{ListOptions: sourcegraph.ListOptions{PerPage: 9999}})
		if err != nil {
			return nil, err
		}
		for _, repo := range repos.Repos {
			opt.RepoRevs = append(opt.RepoRevs, repo.URI+"@"+repo.DefaultBranch)
		}
	}

	// TODO(sqs): parallelize
	var defList sourcegraph.DefList
	for _, repoRev := range opt.RepoRevs {
		repo, commitID := sourcegraph.ParseRepoAndCommitID(repoRev)
		repoCtx, _, err := lookupRepo(ctx, &repo)
		if err != nil {
			return nil, err
		}
		if repoCtx == nil {
			return s.List(ctx, opt)
		}

		repoOpt := *opt
		repoOpt.RepoRevs = []string{repo + "@" + commitID}
		defs, err := svc.Defs(repoCtx).List(repoCtx, &repoOpt)
		if err != nil {
			return nil, err
		}
		defList.Defs = append(defList.Defs, defs.Defs...)
	}

	return &defList, nil
}

func CustomReposCreate(ctx context.Context, op *sourcegraph.ReposCreateOp, s sourcegraph.ReposServer) (*sourcegraph.Repo, error) {
	// Avoid federating operations for private repositories.
	if op.Private {
		return s.Create(ctx, op)
	}

	// At this time, we never federate Create anyway (but if we did, it would happen here).
	return s.Create(ctx, op)
}

// Get sets repo.Origin on repos that originated from a remote server.
func CustomReposGet(ctx context.Context, v1 *sourcegraph.RepoSpec, s sourcegraph.ReposServer) (*sourcegraph.Repo, error) {
	repoCtx, info, err := lookupRepo(ctx, &v1.URI)
	if err != nil {
		return nil, err
	}
	if repoCtx == nil {
		return s.Get(ctx, v1)
	}
	ctx = repoCtx
	repo, err := svc.Repos(ctx).Get(ctx, v1)
	if repo != nil && info != nil {
		repo.Origin = info.String()
	}
	return repo, err
}
