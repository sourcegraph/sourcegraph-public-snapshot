package backend

import (
	"context"

	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/services/repoupdater"
)

func (s *repos) ResolveRev(ctx context.Context, op *sourcegraph.ReposResolveRevOp) (res *sourcegraph.ResolvedRev, err error) {
	if Mocks.Repos.ResolveRev != nil {
		return Mocks.Repos.ResolveRev(ctx, op)
	}

	ctx, done := trace(ctx, "Repos", "ResolveRev", op, &err)
	defer done()

	commitID, err := resolveRepoRev(ctx, op.Repo, op.Rev)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.ResolvedRev{CommitID: string(commitID)}, nil
}

// resolveRepoRev resolves the repo's rev to an absolute commit ID (by
// consulting its VCS data). If no rev is specified, the repo's
// default branch is used.
func resolveRepoRev(ctx context.Context, repo int32, rev string) (vcs.CommitID, error) {
	repoObj, err := Repos.Get(ctx, &sourcegraph.RepoSpec{ID: repo})
	if err != nil {
		return "", err
	}

	if rev == "" {
		if repoObj.DefaultBranch == "" {
			return "", legacyerr.Errorf(legacyerr.FailedPrecondition, "repo %d has no default branch", repo)
		}
		rev = repoObj.DefaultBranch
	}

	vcsrepo, err := localstore.RepoVCS.Open(ctx, repo)
	if err != nil {
		return "", err
	}
	commitID, err := vcsrepo.ResolveRevision(ctx, rev)
	if err != nil {
		enqueueUpdate := false
		if notExistError, ok := err.(vcs.RepoNotExistError); ok {
			// Attempt to clone repo if its VCS repository doesn't exist.
			// Do it in the background, return 202 so that frontend can display cloning interstitual.
			if !notExistError.CloneInProgress {
				enqueueUpdate = true
			}
			return "", vcs.RepoNotExistError{CloneInProgress: true}
		}
		if err == vcs.ErrRevisionNotFound {
			// Attempt to update the VCS repo if the revision wasn't
			// found, for the when (e.g.) a specific commit ID or
			// branch is requested that we don't yet know about. This
			// request will still fail, but subsequent requests for
			// that rev will succeed after the update is complete.
			enqueueUpdate = true
		}
		if enqueueUpdate {
			var asUser *sourcegraph.UserSpec
			if actor := authpkg.ActorFromContext(ctx); actor.UID != "" {
				asUser = actor.UserSpec()
			}
			repoupdater.Enqueue(repo, asUser)
		}
		return "", err
	}
	return commitID, nil
}

func (s *repos) GetCommit(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (res *vcs.Commit, err error) {
	if Mocks.Repos.GetCommit != nil {
		return Mocks.Repos.GetCommit(ctx, repoRev)
	}

	ctx, done := trace(ctx, "Repos", "GetCommit", repoRev, &err)
	defer done()

	log15.Debug("svc.local.repos.GetCommit", "repo-rev", repoRev)

	if !isAbsCommitID(repoRev.CommitID) {
		return nil, errNotAbsCommitID
	}

	vcsrepo, err := localstore.RepoVCS.Open(ctx, repoRev.Repo)
	if err != nil {
		return nil, err
	}

	return vcsrepo.GetCommit(ctx, vcs.CommitID(repoRev.CommitID))
}

func (s *repos) ListCommits(ctx context.Context, op *sourcegraph.ReposListCommitsOp) (res *sourcegraph.CommitList, err error) {
	if Mocks.Repos.ListCommits != nil {
		return Mocks.Repos.ListCommits(ctx, op)
	}

	ctx, done := trace(ctx, "Repos", "ListCommits", op, &err)
	defer done()

	log15.Debug("svc.local.repos.ListCommits", "op", op)

	repo, err := Repos.Get(ctx, &sourcegraph.RepoSpec{ID: op.Repo})
	if err != nil {
		return nil, err
	}

	if op.Opt == nil {
		op.Opt = &sourcegraph.RepoListCommitsOptions{}
	}
	if op.Opt.PerPage == 0 {
		op.Opt.PerPage = 20
	}
	if op.Opt.Head == "" {
		return nil, legacyerr.Errorf(legacyerr.InvalidArgument, "Head (revision specifier) is required")
	}

	vcsrepo, err := localstore.RepoVCS.Open(ctx, repo.ID)
	if err != nil {
		return nil, err
	}

	head, err := vcsrepo.ResolveRevision(ctx, op.Opt.Head)
	if err != nil {
		return nil, err
	}

	var base vcs.CommitID
	if op.Opt.Base != "" {
		base, err = vcsrepo.ResolveRevision(ctx, op.Opt.Base)
		if err != nil {
			return nil, err
		}
	}

	n := uint(op.Opt.PerPageOrDefault()) + 1 // Request one additional commit to determine value of StreamResponse.HasMore.
	if op.Opt.PerPage == -1 {
		n = 0 // retrieve all commits
	}
	commits, _, err := vcsrepo.Commits(ctx, vcs.CommitsOptions{
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

func (s *repos) ListBranches(ctx context.Context, op *sourcegraph.ReposListBranchesOp) (res *sourcegraph.BranchList, err error) {
	if Mocks.Repos.ListBranches != nil {
		return Mocks.Repos.ListBranches(ctx, op)
	}

	ctx, done := trace(ctx, "Repos", "ListBranches", op, &err)
	defer done()

	repo, err := s.Get(ctx, &sourcegraph.RepoSpec{ID: op.Repo})
	if err != nil {
		return nil, err
	}

	vcsrepo, err := localstore.RepoVCS.Open(ctx, repo.ID)
	if err != nil {
		return nil, err
	}

	branches, err := vcsrepo.Branches(ctx, vcs.BranchesOptions{
		IncludeCommit:     op.Opt.IncludeCommit,
		BehindAheadBranch: op.Opt.BehindAheadBranch,
		ContainsCommit:    op.Opt.ContainsCommit,
	})
	if err != nil {
		return nil, err
	}

	return &sourcegraph.BranchList{Branches: branches}, nil
}

func (s *repos) ListTags(ctx context.Context, op *sourcegraph.ReposListTagsOp) (res *sourcegraph.TagList, err error) {
	if Mocks.Repos.ListTags != nil {
		return Mocks.Repos.ListTags(ctx, op)
	}

	ctx, done := trace(ctx, "Repos", "ListTags", op, &err)
	defer done()

	repo, err := s.Get(ctx, &sourcegraph.RepoSpec{ID: op.Repo})
	if err != nil {
		return nil, err
	}

	vcsrepo, err := localstore.RepoVCS.Open(ctx, repo.ID)
	if err != nil {
		return nil, err
	}

	tags, err := vcsrepo.Tags(ctx)
	if err != nil {
		return nil, err
	}

	return &sourcegraph.TagList{Tags: tags}, nil
}

func (s *repos) ListCommitters(ctx context.Context, op *sourcegraph.ReposListCommittersOp) (res *sourcegraph.CommitterList, err error) {
	if Mocks.Repos.ListCommitters != nil {
		return Mocks.Repos.ListCommitters(ctx, op)
	}

	ctx, done := trace(ctx, "Repos", "ListCommitters", op, &err)
	defer done()

	repo, err := s.Get(ctx, &sourcegraph.RepoSpec{ID: op.Repo})
	if err != nil {
		return nil, err
	}

	vcsrepo, err := localstore.RepoVCS.Open(ctx, repo.ID)
	if err != nil {
		return nil, err
	}

	var opt vcs.CommittersOptions
	if op.Opt != nil {
		opt.Rev = op.Opt.Rev
		opt.N = int(op.Opt.PerPage)
	}

	committers, err := vcsrepo.Committers(ctx, opt)
	if err != nil {
		return nil, err
	}

	return &sourcegraph.CommitterList{Committers: committers}, nil
}

func isAbsCommitID(commitID string) bool { return len(commitID) == 40 }

func makeErrNotAbsCommitID(prefix string) error {
	str := "absolute commit ID required (40 hex chars)"
	if prefix != "" {
		str = prefix + ": " + str
	}
	return legacyerr.Errorf(legacyerr.InvalidArgument, str)
}

var errNotAbsCommitID = makeErrNotAbsCommitID("")
