package local

import (
	"encoding/json"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	localcli "src.sourcegraph.com/sourcegraph/server/local/cli"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
)

func (s *repos) GetCommit(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*vcs.Commit, error) {
	log15.Debug("svc.local.repos.GetCommit", "repo-rev", repoRev)
	cacheOnCommitID(ctx, repoRev.CommitID)

	// Get default branch if no revision specified
	if err := s.resolveRepoRevBranch(ctx, repoRev); err != nil {
		return nil, err
	}

	// Return cached commit if available
	if cachedCommits, err := s.listCommitsCached(ctx, *repoRev); err == nil && cachedCommits != nil {
		if len(cachedCommits.Commits) > 0 {
			return cachedCommits.Commits[0], nil
		}
	}

	if err := s.resolveRepoRev(ctx, repoRev); err != nil {
		return nil, err
	}

	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, repoRev.URI)
	if err != nil {
		return nil, err
	}

	return vcsrepo.GetCommit(vcs.CommitID(repoRev.CommitID))
}

func (s *repos) ListCommits(ctx context.Context, op *sourcegraph.ReposListCommitsOp) (*sourcegraph.CommitList, error) {
	log15.Debug("svc.local.repos.ListCommits", "op", op)
	veryShortCache(ctx)

	if op.Opt.PerPage == 0 {
		op.Opt.PerPage = 20
	}
	if op.Opt.Head == "" {
		defBr, err := s.defaultBranch(ctx, op.Repo.URI)
		if err != nil {
			return nil, err
		}
		op.Opt.Head = defBr
	}

	// Uncacheable case
	if op.Opt.Base != "" || op.Opt.Path != "" || op.Opt.Page > 1 {
		return s.listCommitsUncached(ctx, op)
	}

	// Refresh cache if RefreshCache is true
	repoRev := sourcegraph.RepoRevSpec{RepoSpec: op.Repo, Rev: op.Opt.Head}
	if op.Opt.RefreshCache {
		if _, err := s.refreshCommitsCache(ctx, repoRev); err != nil {
			return nil, err
		}
	}

	// Return cached value if it exists
	commitList, err := s.listCommitsCached(ctx, repoRev)
	if err != nil {
		return nil, err
	}
	if commitList != nil {
		if len(commitList.Commits) > int(op.Opt.PerPage) {
			commitList.Commits = commitList.Commits[:op.Opt.PerPage]
		}
		return commitList, nil
	}

	return s.listCommitsUncached(ctx, op)
}

func (s *repos) listCommitsCached(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (*sourcegraph.CommitList, error) {
	log15.Debug("svc.local.repos.listCommitsCached", "repo-rev", repoRev)

	// Don't try to use cache if it's not enabled at all.
	if localcli.Flags.CommitLogCachePeriod == 0 {
		return nil, nil
	}

	cmbStatus, err := svc.RepoStatuses(ctx).GetCombined(ctx, &repoRev)
	if err != nil {
		return nil, err
	}

	var commitList sourcegraph.CommitList
	for _, status := range cmbStatus.Statuses {
		if status.Context == "graph_data_commit" {
			if err := json.Unmarshal([]byte(status.Description), &commitList); err != nil {
				return nil, err
			}
			break
		}
	}
	if commitList.Commits == nil {
		return nil, nil
	}
	return &commitList, nil
}

func (s *repos) refreshCommitsCache(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (*sourcegraph.CommitList, error) {
	commitList, err := s.listCommitsUncached(ctx, &sourcegraph.ReposListCommitsOp{
		Repo: repoRev.RepoSpec,
		Opt: &sourcegraph.RepoListCommitsOptions{
			Head:        repoRev.Rev,
			ListOptions: sourcegraph.ListOptions{PerPage: localcli.Flags.CommitLogCacheSize},
		},
	})
	if err != nil {
		return nil, err
	}

	commitListJSON, err := json.Marshal(commitList)
	if err != nil {
		return nil, err
	}
	_, err = svc.RepoStatuses(ctx).Create(ctx, &sourcegraph.RepoStatusesCreateOp{
		Repo: repoRev,
		Status: sourcegraph.RepoStatus{
			Description: string(commitListJSON),
			Context:     "graph_data_commit",
		},
	})
	if err != nil {
		return nil, err
	}
	return commitList, nil
}

func (s *repos) listCommitsUncached(ctx context.Context, op *sourcegraph.ReposListCommitsOp) (*sourcegraph.CommitList, error) {
	log15.Debug("svc.local.repos.listCommitsUncached", "op", op)

	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, op.Repo.URI)
	if err != nil {
		return nil, err
	}

	head, err := vcsrepo.ResolveRevision(op.Opt.Head)
	if err != nil {
		return nil, err
	}

	var base vcs.CommitID
	if op.Opt.Base != "" {
		base, err = vcsrepo.ResolveRevision(op.Opt.Base)
		if err != nil {
			return nil, err
		}
	}

	n := uint(op.Opt.PerPageOrDefault()) + 1 // Request one additional commit to determine value of StreamResponse.HasMore.
	if op.Opt.PerPage == -1 {
		n = 0 // retrieve all commits
	}
	commits, _, err := vcsrepo.Commits(vcs.CommitsOptions{
		Head:    head,
		Base:    base,
		Skip:    uint(op.Opt.ListOptions.Offset()),
		N:       n,
		Path:    op.Opt.Path,
		NoTotal: true,
	})
	if err != nil {
		return nil, err
	}

	// Determine if there are more results.
	var streamResponse sourcegraph.StreamResponse
	if n != 0 && uint(len(commits)) == n {
		streamResponse.HasMore = true
		commits = commits[:len(commits)-1] // Don't include the additional commit in results, it's from next page.
	}

	return &sourcegraph.CommitList{Commits: commits, StreamResponse: streamResponse}, nil
}

func (s *repos) ListBranches(ctx context.Context, op *sourcegraph.ReposListBranchesOp) (*sourcegraph.BranchList, error) {
	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, op.Repo.URI)
	if err != nil {
		return nil, err
	}

	branches, err := vcsrepo.Branches(vcs.BranchesOptions{
		IncludeCommit:     op.Opt.IncludeCommit,
		BehindAheadBranch: op.Opt.BehindAheadBranch,
		ContainsCommit:    op.Opt.ContainsCommit,
	})
	if err != nil {
		return nil, err
	}

	cacheFor(ctx, maxMutableVCSAge)
	return &sourcegraph.BranchList{Branches: branches}, nil
}

func (s *repos) ListTags(ctx context.Context, op *sourcegraph.ReposListTagsOp) (*sourcegraph.TagList, error) {
	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, op.Repo.URI)
	if err != nil {
		return nil, err
	}

	tags, err := vcsrepo.Tags()
	if err != nil {
		return nil, err
	}

	cacheFor(ctx, maxMutableVCSAge)
	return &sourcegraph.TagList{Tags: tags}, nil
}

func (s *repos) ListCommitters(ctx context.Context, op *sourcegraph.ReposListCommittersOp) (*sourcegraph.CommitterList, error) {
	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, op.Repo.URI)
	if err != nil {
		return nil, err
	}

	var opt vcs.CommittersOptions
	if op.Opt != nil {
		opt.Rev = op.Opt.Rev
		opt.N = int(op.Opt.PerPage)
	}

	committers, err := vcsrepo.Committers(opt)
	if err != nil {
		return nil, err
	}

	cacheFor(ctx, maxMutableVCSAge)
	return &sourcegraph.CommitterList{Committers: committers}, nil
}

func isAbsCommitID(commitID string) bool { return len(commitID) == 40 }
