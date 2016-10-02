package backend

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/services/repoupdater"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
)

func (s *repos) ResolveRev(ctx context.Context, op *sourcegraph.ReposResolveRevOp) (*sourcegraph.ResolvedRev, error) {
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
	repoObj, err := svc.Repos(ctx).Get(ctx, &sourcegraph.RepoSpec{ID: repo})
	if err != nil {
		return "", err
	}

	if rev == "" {
		if repoObj.DefaultBranch == "" {
			return "", grpc.Errorf(codes.FailedPrecondition, "repo %d has no default branch", repo)
		}
		rev = repoObj.DefaultBranch
	}

	vcsrepo, err := localstore.RepoVCS.Open(ctx, repo)
	if err != nil {
		return "", err
	}
	commitID, err := vcsrepo.ResolveRevision(ctx, rev)
	if err != nil {
		// Attempt to reclone repo if its VCS repository doesn't exist.
		// Do it in the background, return 202 so that frontend can display cloning interstitual.
		if notExistError, ok := err.(vcs.RepoNotExistError); ok {
			if !notExistError.CloneInProgress {
				var asUser *sourcegraph.UserSpec
				if actor := authpkg.ActorFromContext(ctx); actor.UID != "" {
					asUser = actor.UserSpec()
				}
				repoupdater.Enqueue(repo, asUser)
			}
			return "", vcs.RepoNotExistError{CloneInProgress: true}
		}
		commitID, err = vcsrepo.ResolveRevision(ctx, rev)
		if err != nil {
			return "", err
		}
	}
	return commitID, nil
}

func (s *repos) GetCommit(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*vcs.Commit, error) {
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

func (s *repos) ListCommits(ctx context.Context, op *sourcegraph.ReposListCommitsOp) (*sourcegraph.CommitList, error) {
	log15.Debug("svc.local.repos.ListCommits", "op", op)

	repo, err := svc.Repos(ctx).Get(ctx, &sourcegraph.RepoSpec{ID: op.Repo})
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
		return nil, grpc.Errorf(codes.InvalidArgument, "Head (revision specifier) is required")
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

func (s *repos) ListBranches(ctx context.Context, op *sourcegraph.ReposListBranchesOp) (*sourcegraph.BranchList, error) {
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

func (s *repos) ListTags(ctx context.Context, op *sourcegraph.ReposListTagsOp) (*sourcegraph.TagList, error) {
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

func (s *repos) ListCommitters(ctx context.Context, op *sourcegraph.ReposListCommittersOp) (*sourcegraph.CommitterList, error) {
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
	return grpc.Errorf(codes.InvalidArgument, str)
}

var errNotAbsCommitID = makeErrNotAbsCommitID("")
