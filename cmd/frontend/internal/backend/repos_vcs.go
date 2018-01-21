package backend

import (
	"context"

	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

// ResolveRev will return the absolute commit for a commit-ish spec in a repo.
// If no rev is specified, HEAD is used.
// Error cases:
// * Repo does not exist: vcs.RepoNotExistError
// * Commit does not exist: vcs.ErrRevisionNotFound
// * Empty repository: vcs.ErrRevisionNotFound
// * The user does not have permission: db.ErrRepoNotFound
// * Other unexpected errors.
func (s *repos) ResolveRev(ctx context.Context, repo api.RepoID, rev string) (commitID api.CommitID, err error) {
	if Mocks.Repos.ResolveRev != nil {
		return Mocks.Repos.ResolveRev(ctx, repo, rev)
	}

	ctx, done := trace(ctx, "Repos", "ResolveRev", map[string]interface{}{"repo": repo, "rev": rev}, &err)
	defer done()

	vcsrepo, err := db.RepoVCS.Open(ctx, repo)
	if err != nil {
		return "", err
	}
	return vcsrepo.ResolveRevision(ctx, rev)
}

func (s *repos) GetCommit(ctx context.Context, repo api.RepoID, commitID api.CommitID) (res *vcs.Commit, err error) {
	if Mocks.Repos.GetCommit != nil {
		return Mocks.Repos.GetCommit(ctx, repo, commitID)
	}

	ctx, done := trace(ctx, "Repos", "GetCommit", map[string]interface{}{"repo": repo, "commitID": commitID}, &err)
	defer done()

	log15.Debug("svc.local.repos.GetCommit", "repo", repo, "commitID", commitID)

	if !isAbsCommitID(commitID) {
		return nil, errNotAbsCommitID
	}

	vcsrepo, err := db.RepoVCS.Open(ctx, repo)
	if err != nil {
		return nil, err
	}

	return vcsrepo.GetCommit(ctx, commitID)
}

func isAbsCommitID(commitID api.CommitID) bool { return len(commitID) == 40 }

func makeErrNotAbsCommitID(prefix string) error {
	str := "absolute commit ID required (40 hex chars)"
	if prefix != "" {
		str = prefix + ": " + str
	}
	return legacyerr.Errorf(legacyerr.InvalidArgument, str)
}

var errNotAbsCommitID = makeErrNotAbsCommitID("")
