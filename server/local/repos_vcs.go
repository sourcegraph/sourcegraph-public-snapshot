package local

import (
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
)

func (s *repos) GetCommit(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*vcs.Commit, error) {
	log15.Debug("svc.local.repos.GetCommit", "repo-rev", repoRev)

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

	if op.Opt == nil {
		op.Opt = &sourcegraph.RepoListCommitsOptions{}
	}
	if op.Opt.PerPage == 0 {
		op.Opt.PerPage = 20
	}
	if op.Opt.Head == "" {
		defBr, err := defaultBranch(ctx, op.Repo.URI)
		if err != nil {
			return nil, err
		}
		op.Opt.Head = defBr
	}

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

	return &sourcegraph.CommitterList{Committers: committers}, nil
}

func isAbsCommitID(commitID string) bool { return len(commitID) == 40 }
