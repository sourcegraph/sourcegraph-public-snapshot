package server

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

// acquireStoreCloning is a helper around managing the state transition from
// cloning to cloned or not_cloned. It will set the state of the repository to
// cloning and return a function which should be called when done
// cloning. When called the state will then reflect if the repo is cloned or
// not.
func (s *Server) withStoreCloning(ctx context.Context, repo api.RepoName) func() {
	store := s.Store
	if store == nil {
		return func() {}
	}

	// TODO we do clones on repositories that exist. Do not mangle the state
	// in that case by setting the state to cloning.

	if err := store.SetState(ctx, repo, "cloning"); err != nil {
		log15.Warn("failed to set repo state in store", "state", "cloning", "error", err)
		metricStoreStateFailed.Inc()
		return func() {}
	}

	return func() {
		state := "cloned"
		if !repoCloned(s.dir(repo)) {
			state = "cloning"
		}
		if err := store.SetState(ctx, repo, state); err != nil {
			metricStoreStateFailed.Inc()
			log15.Warn("failed to set repo state in store", "state", state, "error", err)
		}
	}
}

var (
	metricStoreStateFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_store_state_failed_total",
		Help: "The total number of times setting a repo state in the store failed.",
	})
)
