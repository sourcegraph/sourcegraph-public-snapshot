package local

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	srclibstore "sourcegraph.com/sourcegraph/srclib/store"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	localcli "src.sourcegraph.com/sourcegraph/server/local/cli"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
)

func (s *repos) GetSrclibDataVersionForPath(ctx context.Context, entry *sourcegraph.TreeEntrySpec) (*sourcegraph.SrclibDataVersion, error) {
	wasAbs := isAbsCommitID(entry.RepoRev.CommitID)

	if err := s.resolveRepoRev(ctx, &entry.RepoRev); err != nil {
		return nil, err
	}

	// First, try to find an exact match.
	vers, err := store.GraphFromContext(ctx).Versions(
		srclibstore.ByRepoCommitIDs(srclibstore.Version{Repo: entry.RepoRev.URI, CommitID: entry.RepoRev.CommitID}),
	)
	if err != nil {
		return nil, err
	}
	if len(vers) == 1 {
		if wasAbs {
			veryShortCache(ctx)
		}
		return &sourcegraph.SrclibDataVersion{CommitID: vers[0].CommitID, CommitsBehind: 0}, nil
	}

	if entry.Path == "." {
		// All commits affect the root, so there is no hope of finding
		// an earlier srclib-built commit that we can use.
		return nil, grpc.Errorf(codes.NotFound, "no srclib data version found for head commit %v (can't look-back because path is root)", entry.RepoRev)
	}

	// Do expensive search backwards through history.
	info, err := s.getSrclibDataVersionForPathLookback(ctx, entry)
	if err != nil {
		return nil, err
	}
	veryShortCache(ctx)
	return info, nil
}

func (s *repos) getSrclibDataVersionForPathLookback(ctx context.Context, entry *sourcegraph.TreeEntrySpec) (*sourcegraph.SrclibDataVersion, error) {
	// Find the base commit (the farthest ancestor commit we'll
	// consider).
	//
	// If entry.Path is empty, we theoretically are OK going back as
	// far as possible. This is the intended behavior for repo-wide
	// actions (such as search), where there is no non-arbitrary point
	// to stop our lookback. However, we apply a lookback limit for
	// performance reasons.
	//
	// If entry.Path is set, then we need to find a commit equal to or
	// a descendant of the last commit that touched that
	// path. Otherwise, we'd return srclib data that applies to a
	// different version of the file.
	var base string
	if entry.Path != "" {
		lastPathCommit, err := svc.Repos(ctx).ListCommits(ctx, &sourcegraph.ReposListCommitsOp{
			Repo: entry.RepoRev.RepoSpec,
			Opt: &sourcegraph.RepoListCommitsOptions{
				Head:        entry.RepoRev.CommitID,
				Path:        entry.Path,
				ListOptions: sourcegraph.ListOptions{PerPage: 1}, // only the most recent commit needed
			},
		})
		if err != nil {
			return nil, err
		}
		if len(lastPathCommit.Commits) != 1 {
			return nil, grpc.Errorf(codes.NotFound, "no commits found for path %q in repo %v", entry.Path, entry.RepoRev)
		}
		base = string(lastPathCommit.Commits[0].ID) + "~1" // make it inclusive of the base
	}

	// TODO(beyang): move clcache flag into lookbackLimit flag
	var lookbackLimit int32 = 250
	if localcli.Flags.CommitLogCacheSize > 250 {
		lookbackLimit = localcli.Flags.CommitLogCacheSize
	}

	// List the recent commits that we'll use to check for builds.
	candidateCommits, err := svc.Repos(ctx).ListCommits(ctx,
		&sourcegraph.ReposListCommitsOp{
			Repo: entry.RepoRev.RepoSpec,
			Opt: &sourcegraph.RepoListCommitsOptions{
				Head: entry.RepoRev.CommitID,
				Base: base,
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

	// Get all srclib built data versions.
	vers, err := store.GraphFromContext(ctx).Versions(srclibstore.ByRepos(entry.RepoRev.URI))
	if err != nil {
		return nil, err
	}
	verMap := make(map[string]struct{}, len(vers))
	for _, ver := range vers {
		verMap[ver.CommitID] = struct{}{}
	}

	for i, cc := range candidateCommits.Commits {
		if _, present := verMap[string(cc.ID)]; present {
			return &sourcegraph.SrclibDataVersion{CommitID: string(cc.ID), CommitsBehind: int32(i)}, nil
		}
	}

	return nil, grpc.Errorf(codes.NotFound, "no srclib data versions found for %v (%d candidate commits, %d srclib data versions)", entry, len(candidateCommits.Commits), len(vers))
}
