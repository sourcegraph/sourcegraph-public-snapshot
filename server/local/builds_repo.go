package local

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	localcli "src.sourcegraph.com/sourcegraph/server/local/cli"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
)

func (s *builds) GetRepoBuildInfo(ctx context.Context, op *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error) {
	repoRevSpec := op.Repo

	buildStore := store.BuildsFromContext(ctx)
	if op.Opt == nil {
		op.Opt = &sourcegraph.BuildsGetRepoBuildInfoOptions{}
	}

	var info sourcegraph.RepoBuildInfo

	var commit *vcs.Commit
	if repoRevSpec.CommitID == "" {
		// Resolve the revspec to a full commit ID.
		var err error
		commit, err = svc.Repos(ctx).GetCommit(ctx, &repoRevSpec)
		if err != nil {
			return nil, err
		}
		repoRevSpec.CommitID = string(commit.ID)
	}

	// First, try to find an exact match.
	exact, _, err := buildStore.GetFirstInCommitOrder(ctx, repoRevSpec.URI, []string{repoRevSpec.CommitID}, false)
	if err != nil {
		return nil, err
	}
	if exact != nil {
		info.Exact = exact
		if info.Exact.Success {
			info.LastSuccessful = info.Exact
			info.LastSuccessfulCommit = commit
			shortCache(ctx)
			return &info, nil
		}
	}

	// Short-circuit for exact builds
	if op.Opt.Exact {
		if info.Exact == nil {
			return nil, grpc.Errorf(codes.NotFound, "no exact match build found for %v", repoRevSpec)
		}
		return &info, nil
	}

	// Do expensive search backward through history
	info_, err := s.getRepoBuildInfoInexact(ctx, op)
	if err != nil {
		return nil, err
	}
	info = *info_

	if info.Exact == nil && info.LastSuccessful == nil {
		return nil, grpc.Errorf(codes.NotFound, "no matching build found for %v", repoRevSpec)
	}

	veryShortCache(ctx)
	return &info, nil
}

func (s *builds) getRepoBuildInfoInexact(ctx context.Context, op *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error) {
	repoRevSpec := op.Repo

	rev := repoRevSpec.Rev
	if rev == "" {
		rev = repoRevSpec.CommitID
	}

	// TODO(beyang): move clcache flag into lookbackLimit flag
	var lookbackLimit int32 = 250
	if localcli.Flags.CommitLogCacheSize > 250 {
		lookbackLimit = localcli.Flags.CommitLogCacheSize
	}

	// Find the last successful commit.
	// List the recent commits that we'll use to check for builds.
	lastCommits, err := svc.Repos(ctx).ListCommits(ctx,
		&sourcegraph.ReposListCommitsOp{
			Repo: repoRevSpec.RepoSpec,
			Opt: &sourcegraph.RepoListCommitsOptions{
				Head: rev,
				ListOptions: sourcegraph.ListOptions{
					// Note: if cached, lookback is limited to the
					// number of commits in the commit log that are
					// cached.
					PerPage: lookbackLimit,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	lastCommitIDs := make([]string, len(lastCommits.Commits))
	for i, c := range lastCommits.Commits {
		lastCommitIDs[i] = string(c.ID)
	}

	// Find the newest (i.e., fewest commits away from the revspec) commit that has a successful build.
	build, nth, err := store.BuildsFromContext(ctx).GetFirstInCommitOrder(ctx, repoRevSpec.URI, lastCommitIDs, true)
	if err != nil {
		return nil, err
	}
	var info sourcegraph.RepoBuildInfo
	if build != nil && build.Success {
		if nth == 0 {
			info.Exact = build
		}
		info.LastSuccessful = build
		info.LastSuccessfulCommit = lastCommits.Commits[nth]
		info.CommitsBehind = int32(nth)
	}
	return &info, nil
}
