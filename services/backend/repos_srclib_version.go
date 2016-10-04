package backend

import (
	"fmt"

	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/universe"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	srclibstore "sourcegraph.com/sourcegraph/srclib/store"
)

func (s *repos) GetSrclibDataVersionForPath(ctx context.Context, entry *sourcegraph.TreeEntrySpec) (res *sourcegraph.SrclibDataVersion, err error) {
	if Mocks.Repos.GetSrclibDataVersionForPath != nil {
		return Mocks.Repos.GetSrclibDataVersionForPath(ctx, entry)
	}

	ctx, done := trace(ctx, "Repos", "GetSrclibDataVersionForPath", entry, &err)
	defer done()

	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Repos.GetSrclibDataVersionForPath", entry.RepoRev.Repo); err != nil {
		return nil, err
	}

	if !isAbsCommitID(entry.RepoRev.CommitID) {
		return nil, errNotAbsCommitID
	}

	repo, err := localstore.Repos.Get(ctx, entry.RepoRev.Repo)
	if err != nil {
		return nil, err
	}
	if feature.Features.NoSrclib && universe.EnabledFile(entry.Path) {
		return &sourcegraph.SrclibDataVersion{
			CommitID:      entry.RepoRev.CommitID,
			CommitsBehind: 0,
		}, nil
	}

	// First, try to find an exact match.
	vers, err := localstore.Graph.Versions(
		srclibstore.ByRepoCommitIDs(srclibstore.Version{Repo: repo.URI, CommitID: entry.RepoRev.CommitID}),
	)
	if err != nil {
		return nil, err
	}
	if len(vers) == 1 {
		log15.Debug("svc.local.repos.GetSrclibDataVersionForPath", "entry", entry, "result", "exact match")
		return &sourcegraph.SrclibDataVersion{CommitID: vers[0].CommitID, CommitsBehind: 0}, nil
	}

	if entry.Path == "." && len(entry.RepoRev.CommitID) == 40 {
		// All commits affect the root, so there is no hope of finding
		// an earlier srclib-built commit that we can use.
		log15.Debug("svc.local.repos.GetSrclibDataVersionForPath", "entry", entry, "result", "no version for root")
		return nil, grpc.Errorf(codes.NotFound, "no srclib data version found for head commit %v (can't look-back because path is root)", entry.RepoRev)
	}

	// Do expensive search backwards through history.
	info, err := s.getSrclibDataVersionForPathLookback(ctx, entry, repo.URI)
	if err != nil {
		if errcode.GRPC(err) == codes.NotFound {
			log15.Debug("svc.local.repos.GetSrclibDataVersionForPath", "entry", entry, "result", "not found: "+err.Error())
		}
		return nil, err
	}
	log15.Debug("svc.local.repos.GetSrclibDataVersionForPath", "entry", entry, "result", fmt.Sprintf("lookback match %+v", info))
	return info, nil
}

func (s *repos) getSrclibDataVersionForPathLookback(ctx context.Context, entry *sourcegraph.TreeEntrySpec, repo string) (*sourcegraph.SrclibDataVersion, error) {
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
		lastPathCommit, err := Repos.ListCommits(ctx, &sourcegraph.ReposListCommitsOp{
			Repo: entry.RepoRev.Repo,
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
		lastPathCommitID := string(lastPathCommit.Commits[0].ID)
		if entry.RepoRev.CommitID == lastPathCommitID {
			// We have already looked checked if we have a build
			// for entry.RepoRev.CommitID, so there is no hope to
			// finding an earlier srclib-built commit that we can
			// use.
			return nil, grpc.Errorf(codes.NotFound, "no srclib data version found for head commit %v (can't look-back because path  was last modified by head commit)", entry.RepoRev)

		}
		base = lastPathCommitID
	}

	const lookbackLimit = 250

	// List the recent commits that we'll use to check for builds.
	candidateCommits, err := Repos.ListCommits(ctx,
		&sourcegraph.ReposListCommitsOp{
			Repo: entry.RepoRev.Repo,
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

	// If we had a base, then the candidateCommits list will exclude
	// the base (by the specification of git ranges; see `man
	// gitrevision`). But we want to include the base commit when
	// searching for versions, so add it back.
	if base != "" {
		candidateCommits.Commits = append(candidateCommits.Commits, &vcs.Commit{ID: vcs.CommitID(base)})
	}

	candidateCommitIDs := make([]string, len(candidateCommits.Commits))
	for i, c := range candidateCommits.Commits {
		candidateCommitIDs[i] = string(c.ID)
	}

	// Get all srclib built data versions.
	vers, err := localstore.Graph.Versions(
		srclibstore.ByRepos(repo),
		srclibstore.ByCommitIDs(candidateCommitIDs...),
	)
	if err != nil {
		return nil, err
	}
	verMap := make(map[string]struct{}, len(vers))
	for _, ver := range vers {
		verMap[ver.CommitID] = struct{}{}
	}

	for i, cc := range candidateCommitIDs {
		if _, present := verMap[cc]; present {
			return &sourcegraph.SrclibDataVersion{CommitID: cc, CommitsBehind: int32(i)}, nil
		}
	}

	return nil, grpc.Errorf(codes.NotFound, "no srclib data versions found for %v (%d candidate commits, %d srclib data versions)", entry, len(candidateCommits.Commits), len(vers))
}
