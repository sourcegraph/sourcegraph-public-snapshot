package internal

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func repoCloneProgress(reposDir string, locker RepositoryLocker, repo api.RepoName) *protocol.RepoCloneProgress {
	dir := gitserverfs.RepoDirFromName(reposDir, repo)
	resp := protocol.RepoCloneProgress{
		Cloned: repoCloned(dir),
	}
	resp.CloneProgress, resp.CloneInProgress = locker.Status(dir)
	if isAlwaysCloningTest(repo) {
		resp.CloneInProgress = true
		resp.CloneProgress = "This will never finish cloning"
	}
	return &resp
}

func deleteRepo(
	ctx context.Context,
	logger log.Logger,
	db database.DB,
	shardID string,
	reposDir string,
	repo api.RepoName,
) error {
	// The repo may be deleted in the database, in this case we need to get the
	// original name in order to find it on disk
	err := gitserverfs.RemoveRepoDirectory(ctx, logger, db, shardID, reposDir, gitserverfs.RepoDirFromName(reposDir, api.UndeletedRepoName(repo)), true)
	if err != nil {
		return errors.Wrap(err, "removing repo directory")
	}
	err = db.GitserverRepos().SetCloneStatus(ctx, repo, types.CloneStatusNotCloned, shardID)
	if err != nil {
		return errors.Wrap(err, "setting clone status after delete")
	}
	return nil
}
