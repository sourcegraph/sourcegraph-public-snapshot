package internal

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

var (
	ensureRevisionCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_gitserver_ensure_revision",
		Help: "A request triggered ensureRevision in Gitserver",
	}, []string{"status"})
)

func (s *Server) EnsureRevision(ctx context.Context, repo api.RepoName, rev string) (didUpdate bool) {
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

	// Revision not found, update before returning.
	lastFetched, lastChanged, err := s.FetchRepository(ctx, repo)
	if err != nil {
		if ctx.Err() == nil {
			ensureRevisionCounter.WithLabelValues("update_failed").Inc()
		}
		s.logger.Warn("failed to perform background repo update", log.Error(err), log.String("repo", string(repo)), log.String("rev", rev))
		// TODO: Shouldn't we return false here?
	} else {
		if err := s.db.GitserverRepos().SetLastFetched(ctx, repo, database.GitserverFetchData{
			LastFetched: lastFetched,
			LastChanged: lastChanged,
		}); err != nil {
			s.logger.Error("failed to store repo update timestamps", log.Error(err))
		}
		ensureRevisionCounter.WithLabelValues("updated").Inc()
	}
	return true
}
