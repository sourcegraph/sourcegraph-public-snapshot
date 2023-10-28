package internal

import (
	"context"
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

var (
	ensureRevisionCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_gitserver_ensure_revision",
		Help: "A request triggered ensureRevision in Gitserver",
	}, []string{"status"})
)

func (s *Server) ensureRevision(ctx context.Context, repo api.RepoName, rev string, repoDir common.GitDir) (didUpdate bool) {
	if rev == "" || rev == "HEAD" {
		ensureRevisionCounter.WithLabelValues("HEAD").Inc()
		return false
	}
	if conf.Get().DisableAutoGitUpdates {
		ensureRevisionCounter.WithLabelValues("disabled").Inc()
		// ensureRevision may kick off a git fetch operation which we don't want if we've
		// configured DisableAutoGitUpdates.
		return false
	}

	// rev-parse on an OID does not check if the commit actually exists, so it always
	// works. So we append ^0 to force the check
	if gitdomain.IsAbsoluteRevision(rev) {
		rev = rev + "^0"
	}

	if err := git.CheckSpecArgSafety(rev); err != nil {
		ensureRevisionCounter.WithLabelValues("invalid_rev").Inc()
		return false
	}

	cmd := exec.Command("git", "rev-parse", rev, "--")
	repoDir.Set(cmd)
	// TODO: Check here that it's actually been a rev-parse error, and not something else.
	if err := cmd.Run(); err == nil {
		ensureRevisionCounter.WithLabelValues("exists").Inc()
		return false
	}
	// Revision not found, update before returning.
	err := s.doRepoUpdate(ctx, repo, rev)
	if err != nil {
		ensureRevisionCounter.WithLabelValues("update_failed").Inc()
		s.Logger.Warn("failed to perform background repo update", log.Error(err), log.String("repo", string(repo)), log.String("rev", rev))
		// TODO: Shouldn't we return false here?
	} else {
		ensureRevisionCounter.WithLabelValues("updated").Inc()
	}
	return true
}
