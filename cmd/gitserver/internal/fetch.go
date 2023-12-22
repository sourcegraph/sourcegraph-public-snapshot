package internal

import (
	"context"
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/perforce"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/urlredactor"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *Server) repoUpdate(req *protocol.RepoUpdateRequest) protocol.RepoUpdateResponse {
	logger := s.Logger.Scoped("handleRepoUpdate")
	var resp protocol.RepoUpdateResponse
	req.Repo = protocol.NormalizeRepo(req.Repo)
	dir := gitserverfs.RepoDirFromName(s.ReposDir, req.Repo)

	// despite the existence of a context on the request, we don't want to
	// cancel the git commands partway through if the request terminates.
	ctx, cancel1 := s.serverContext()
	defer cancel1()
	ctx, cancel2 := context.WithTimeout(ctx, conf.GitLongCommandTimeout())
	defer cancel2()

	if !repoCloned(dir) && !s.skipCloneForTests {
		_, err := s.CloneRepo(ctx, req.Repo, CloneOptions{Block: true})
		if err != nil {
			logger.Warn("error cloning repo", log.String("repo", string(req.Repo)), log.Error(err))
			resp.Error = err.Error()
		}
		return resp
	}

	var statusErr, updateErr error

	if debounce(req.Repo, req.Since) {
		updateErr = s.doRepoUpdate(ctx, req.Repo, "")
	}

	// attempts to acquire these values are not contingent on the success of
	// the update.
	lastFetched, err := repoLastFetched(dir)
	if err != nil {
		statusErr = err
	} else {
		resp.LastFetched = &lastFetched
	}
	lastChanged, err := repoLastChanged(dir)
	if err != nil {
		statusErr = err
	} else {
		resp.LastChanged = &lastChanged
	}
	if statusErr != nil {
		logger.Error("failed to get status of repo", log.String("repo", string(req.Repo)), log.Error(statusErr))
		// report this error in-band, but still produce a valid response with the
		// other information.
		resp.Error = statusErr.Error()
	}
	// If an error occurred during update, report it but don't actually make
	// it into an http error; we want the client to get the information cleanly.
	// An update error "wins" over a status error.
	if updateErr != nil {
		resp.Error = updateErr.Error()
	} else {
		s.Perforce.EnqueueChangelistMappingJob(perforce.NewChangelistMappingJob(req.Repo, dir))
	}

	return resp
}

func (s *Server) doRepoUpdate(ctx context.Context, repo api.RepoName, revspec string) (err error) {
	tr, ctx := trace.New(ctx, "doRepoUpdate", repo.Attr())
	defer tr.EndWithErr(&err)

	s.repoUpdateLocksMu.Lock()
	l, ok := s.repoUpdateLocks[repo]
	if !ok {
		l = &locks{
			once: new(sync.Once),
			mu:   new(sync.Mutex),
		}
		s.repoUpdateLocks[repo] = l
	}
	once := l.once
	mu := l.mu
	s.repoUpdateLocksMu.Unlock()

	// doBackgroundRepoUpdate can block longer than our context deadline. done will
	// close when its done. We can return when either done is closed or our
	// deadline has passed.
	done := make(chan struct{})
	err = errors.New("another operation is already in progress")
	go func() {
		defer close(done)
		once.Do(func() {
			mu.Lock() // Prevent multiple updates in parallel. It works fine, but it wastes resources.
			defer mu.Unlock()

			s.repoUpdateLocksMu.Lock()
			l.once = new(sync.Once) // Make new requests wait for next update.
			s.repoUpdateLocksMu.Unlock()

			err = s.doBackgroundRepoUpdate(repo, revspec)
			if err != nil {
				// We don't want to spam our logs when the rate limiter has been set to block all
				// updates
				if !errors.Is(err, ratelimit.ErrBlockAll) {
					s.Logger.Error("performing background repo update", log.Error(err))
				}

				// The repo update might have failed due to the repo being corrupt
				var gitErr *common.GitCommandError
				if errors.As(err, &gitErr) {
					s.logIfCorrupt(ctx, repo, gitserverfs.RepoDirFromName(s.ReposDir, repo), gitErr.Output)
				}
			}
			setLastErrorNonFatal(s.ctx, s.Logger, s.DB, s.Hostname, repo, err)
		})
	}()

	select {
	case <-done:
		return errors.Wrapf(err, "repo %s:", repo)
	case <-ctx.Done():
		return ctx.Err()
	}
}

var doBackgroundRepoUpdateMock func(api.RepoName) error

func (s *Server) doBackgroundRepoUpdate(repo api.RepoName, revspec string) error {
	logger := s.Logger.Scoped("backgroundRepoUpdate").With(log.String("repo", string(repo)))

	if doBackgroundRepoUpdateMock != nil {
		return doBackgroundRepoUpdateMock(repo)
	}
	// background context.
	ctx, cancel1 := s.serverContext()
	defer cancel1()

	// ensure the background update doesn't hang forever
	ctx, cancel2 := context.WithTimeout(ctx, conf.GitLongCommandTimeout())
	defer cancel2()

	// This background process should use our internal actor
	ctx = actor.WithInternalActor(ctx)

	ctx, cancel2, err := s.acquireCloneLimiter(ctx)
	if err != nil {
		return err
	}
	defer cancel2()

	if err = s.RPSLimiter.Wait(ctx); err != nil {
		return err
	}

	repo = protocol.NormalizeRepo(repo)
	dir := gitserverfs.RepoDirFromName(s.ReposDir, repo)

	remoteURL, err := s.getRemoteURL(ctx, repo)
	if err != nil {
		return errors.Wrap(err, "failed to determine Git remote URL")
	}

	syncer, err := s.GetVCSSyncer(ctx, repo)
	if err != nil {
		return errors.Wrap(err, "get VCS syncer")
	}

	// drop temporary pack files after a fetch. this function won't
	// return until this fetch has completed or definitely-failed,
	// either way they can't still be in use. we don't care exactly
	// when the cleanup happens, just that it does.
	// TODO: Should be done in janitor.
	defer git.CleanTmpPackFiles(s.Logger, dir)

	output, err := syncer.Fetch(ctx, remoteURL, repo, dir, revspec)
	// TODO: Move the redaction also into the VCSSyncer layer here, to be in line
	// with what clone does.
	redactedOutput := urlredactor.New(remoteURL).Redact(string(output))
	// best-effort update the output of the fetch
	if err := s.DB.GitserverRepos().SetLastOutput(context.Background(), repo, redactedOutput); err != nil {
		s.Logger.Warn("Setting last output in DB", log.Error(err))
	}

	if err != nil {
		if output != nil {
			return errors.Wrapf(err, "failed to fetch repo %q with output %q", repo, redactedOutput)
		} else {
			return errors.Wrapf(err, "failed to fetch repo %q", repo)
		}
	}

	return postRepoFetchActions(ctx, logger, s.DB, s.Hostname, s.RecordingCommandFactory, s.ReposDir, repo, dir, remoteURL, syncer)
}
