package app

import (
	"sync"
	"time"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/ext"
	"src.sourcegraph.com/sourcegraph/util"
)

const (
	repoUpdaterQueueDepth       = 10
	repoUpdaterDebounceDuration = 10 * time.Second
)

// RepoUpdater is the app repo updater worker. Repo update requests can be enqueued, with debouncing taken care of.
var RepoUpdater = &repoUpdater{
	recent: make(map[sourcegraph.RepoSpec]time.Time),
	queue:  make(chan *sourcegraph.Repo, repoUpdaterQueueDepth),
}

type repoUpdater struct {
	mu     sync.Mutex
	recent map[sourcegraph.RepoSpec]time.Time // Map of recently scheduled repo updates. Value is last updated time.

	queue chan *sourcegraph.Repo // Queue of scheduled repo updates.
}

// Start one background repo updater worker with the given context.
func (ru *repoUpdater) Start(ctx context.Context) {
	go ru.run(ctx)
}

// enqueue the given repo to be updated.
//
// If the same repo spec has been recently enqueued (within repoUpdaterDebounceDuration), it is ignored.
// If the backlog for repos to be updated is too large (reaches repoUpdaterQueueDepth), it is also ignored.
func (ru *repoUpdater) enqueue(repo *sourcegraph.Repo) {
	ru.mu.Lock()
	defer ru.mu.Unlock()

	now := time.Now()

	// Clear repos that were updated long ago from recent map.
	for rs, lastUpdated := range ru.recent {
		if lastUpdated.Before(now.Add(-repoUpdaterDebounceDuration)) {
			delete(ru.recent, rs)
		}
	}

	// Skip if recently updated.
	if _, recent := ru.recent[repo.RepoSpec()]; recent {
		return
	}

	select {
	case ru.queue <- repo:
		ru.recent[repo.RepoSpec()] = now
	default:
		// Skip since queue is full.
	}
}

func (ru *repoUpdater) run(ctx context.Context) {
	apiclient := sourcegraph.NewClientFromContext(ctx)

	for repo := range ru.queue {
		op := &sourcegraph.MirrorReposRefreshVCSOp{
			Repo: repo.RepoSpec(),
		}
		// For private repos, supply auth.
		if repo.Private {
			host := util.RepoURIHost(repo.URI)
			tokenStore := ext.AccessTokens{}
			token, err := tokenStore.Get(ctx, host)
			if err != nil {
				log15.Warn("repoUpdater: could not fetch credentials", "host", host, "error", err)
				continue
			}

			// Setting credentials will perform this operation locally (non-federated).
			op.Credentials = &sourcegraph.VCSCredentials{
				Pass: token,
			}
		}

		if _, err := apiclient.MirrorRepos.RefreshVCS(ctx, op); err != nil {
			log15.Warn("repoUpdater: RefreshVCS:", "error", err)
			continue
		}
	}
}
