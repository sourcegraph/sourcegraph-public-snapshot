package backend

import (
	"context"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ResolveRev will return the absolute commit for a commit-ish spec in a repo.
// If no rev is specified, HEAD is used.
// Error cases:
// * Repo does not exist: gitdomain.RepoNotExistError
// * Commit does not exist: gitdomain.RevisionNotFoundError
// * Empty repository: gitdomain.RevisionNotFoundError
// * The user does not have permission: errcode.IsNotFound
// * Other unexpected errors.
func (s *repos) ResolveRev(ctx context.Context, repo *types.Repo, rev string) (commitID api.CommitID, err error) {
	if Mocks.Repos.ResolveRev != nil {
		return Mocks.Repos.ResolveRev(ctx, repo, rev)
	}

	ctx, done := trace(ctx, "Repos", "ResolveRev", map[string]any{"repo": repo.Name, "rev": rev}, &err)
	defer done()

	return gitserver.NewClient(s.db).ResolveRevision(ctx, repo.Name, rev, gitserver.ResolveRevisionOptions{})
}

func (s *repos) GetCommit(ctx context.Context, repo *types.Repo, commitID api.CommitID) (res *gitdomain.Commit, err error) {
	if Mocks.Repos.GetCommit != nil {
		return Mocks.Repos.GetCommit(ctx, repo, commitID)
	}

	ctx, done := trace(ctx, "Repos", "GetCommit", map[string]any{"repo": repo.Name, "commitID": commitID}, &err)
	defer done()

	s.logger.Debug("GetCommit", log.String("repo", string(repo.Name)), log.String("commitID", string(commitID)))

	if !gitserver.IsAbsoluteRevision(string(commitID)) {
		return nil, errors.Errorf("non-absolute CommitID for Repos.GetCommit: %v", commitID)
	}

	return gitserver.NewClient(s.db).GetCommit(ctx, repo.Name, commitID, gitserver.ResolveRevisionOptions{}, authz.DefaultSubRepoPermsChecker)
}
