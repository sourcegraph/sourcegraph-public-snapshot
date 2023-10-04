package internal

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

// maybeStartClone checks if a given repository is cloned on disk. If not, it starts
// cloning the repository in the background and returns a NotFound error, if no current
// clone operation is running for that repo yet. If it is already cloning, a NotFound
// error with CloneInProgress: true is returned.
// Note: If disableAutoGitUpdates is set in the site config, no operation is taken and
// a NotFound error is returned.
func (s *Server) maybeStartClone(ctx context.Context, logger log.Logger, repo api.RepoName) (notFound *protocol.NotFoundPayload, cloned bool) {
	dir := gitserverfs.RepoDirFromName(s.ReposDir, repo)
	if repoCloned(dir) {
		return nil, true
	}

	if conf.Get().DisableAutoGitUpdates {
		logger.Debug("not cloning on demand as DisableAutoGitUpdates is set")
		return &protocol.NotFoundPayload{}, false
	}

	cloneProgress, cloneInProgress := s.Locker.Status(dir)
	if cloneInProgress {
		return &protocol.NotFoundPayload{
			CloneInProgress: true,
			CloneProgress:   cloneProgress,
		}, false
	}

	cloneProgress, err := s.CloneRepo(ctx, repo, CloneOptions{})
	if err != nil {
		logger.Debug("error starting repo clone", log.String("repo", string(repo)), log.Error(err))
		return &protocol.NotFoundPayload{CloneInProgress: false}, false
	}

	return &protocol.NotFoundPayload{
		CloneInProgress: true,
		CloneProgress:   cloneProgress,
	}, false
}
