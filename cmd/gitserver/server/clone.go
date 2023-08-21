package server

import (
	"context"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func (s *Server) maybeStartClone(ctx context.Context, logger log.Logger, repoName api.RepoName) (notFound *protocol.NotFoundPayload, cloned bool) {
	dir := s.dir(repoName)
	if repoCloned(dir) {
		return nil, true
	}

	if conf.Get().DisableAutoGitUpdates {
		logger.Debug("not cloning on demand as DisableAutoGitUpdates is set")
		return &protocol.NotFoundPayload{}, false
	}

	cloneProgress, cloneInProgress := RepoCloningStatus(ctx, s.DB, repoName)
	if cloneInProgress {
		return &protocol.NotFoundPayload{
			CloneInProgress: true,
			CloneProgress:   cloneProgress,
		}, false
	}

	err := ScheduleRepoClone(ctx, s.DB, repoName, CloneOptions{Priority: types.HighPriorityRepoUpdate})
	if err != nil {
		logger.Debug("error scheduling a repo clone", log.String("repo", string(repoName)), log.Error(err))
		return &protocol.NotFoundPayload{CloneInProgress: false}, false
	}

	return &protocol.NotFoundPayload{
		CloneInProgress: true,
		CloneProgress:   cloneProgress,
	}, false
}

// ScheduleRepoClone is used to schedule a high-priority repo clone job which
// will be processed by the RepoUpdateWorker.
func ScheduleRepoClone(ctx context.Context, db database.DB, repoName api.RepoName, cloneOpts CloneOptions) error {
	opts := database.CreateRepoUpdateJobOpts{RepoName: repoName, Clone: true, Priority: cloneOpts.Priority, OverwriteClone: cloneOpts.Overwrite}
	_, _, err := db.RepoUpdateJobs().Create(ctx, opts)
	if err != nil {
		return err
	}
	return nil
}

// ScheduleRepoUpdate is used to schedule a high-priority repo update/fetch job
// which will be processed by the RepoUpdateWorker.
func ScheduleRepoUpdate(ctx context.Context, db database.DB, repoName api.RepoName, revision string) error {
	opts := database.CreateRepoUpdateJobOpts{RepoName: repoName, Clone: false, Priority: types.HighPriorityRepoUpdate, FetchRevision: revision}
	_, _, err := db.RepoUpdateJobs().Create(ctx, opts)
	if err != nil {
		return err
	}
	return nil
}
