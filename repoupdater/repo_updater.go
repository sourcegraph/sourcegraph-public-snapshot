package repoupdater

import (
	"sync"
	"time"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"src.sourcegraph.com/sourcegraph/app/appconf"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	sgxcli "src.sourcegraph.com/sourcegraph/sgx/cli"
)

const (
	repoUpdaterQueueDepth = 10
)

func init() {
	sgxcli.ServeInit = append(sgxcli.ServeInit, func() {
		// If we're updating repos in the background, kick off the updates initially.
		if !appconf.Flags.DisableMirrorRepoBackgroundUpdate {
			events.RegisterListener(&repoAutoCloner{})
		}
	})
}

type repoAutoCloner struct{}

func Enqueue(repo *sourcegraph.Repo) {
	RepoUpdater.enqueue(repo)
}

func (r *repoAutoCloner) Scopes() []string {
	return []string{"app:repo-auto-cloner"}
}
func (r *repoAutoCloner) Start(ctx context.Context) {
	go func() {
		apiclient, err := sourcegraph.NewClientFromContext(ctx)
		if err != nil {
			log15.Error("repoAutoCloner: could not create client", "error", err)
			return
		}
		repos, err := apiclient.Repos.List(ctx, &sourcegraph.RepoListOptions{
			ListOptions: sourcegraph.ListOptions{
				PerPage: 100000,
			},
		})
		if err != nil {
			log15.Error("repoAutoCloner: could not list repos", "error", err)
			return
		}
		for _, repo := range repos.Repos {
			if repo.Mirror {
				RepoUpdater.enqueue(repo)
			}
		}
	}()
}

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
// If the same repo spec has been recently enqueued (within MirrorUpdateRate), it is ignored.
// If the backlog for repos to be updated is too large (reaches repoUpdaterQueueDepth), it is also ignored.
func (ru *repoUpdater) enqueue(repo *sourcegraph.Repo) {
	ru.mu.Lock()
	defer ru.mu.Unlock()

	now := time.Now()

	// Clear repos that were updated long ago from recent map.
	for rs, lastUpdated := range ru.recent {
		if lastUpdated.Before(now.Add(-appconf.Flags.MirrorRepoUpdateRate)) {
			delete(ru.recent, rs)
		}
	}

	// Skip if recently updated.
	if _, recent := ru.recent[repo.RepoSpec()]; recent {
		// Enqueue the repo again at a later time if desired.
		if !appconf.Flags.DisableMirrorRepoBackgroundUpdate {
			go ru.enqueueLater(repo)
		}
		return
	}

	select {
	case ru.queue <- repo:
		ru.recent[repo.RepoSpec()] = now
	default:
		// Skip since queue is full.
	}

	// Enqueue the repo again at a later time if desired.
	if !appconf.Flags.DisableMirrorRepoBackgroundUpdate {
		go ru.enqueueLater(repo)
	}
}

func (ru *repoUpdater) run(ctx context.Context) {
	apiclient, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		log15.Error("repoUpdater: RefreshVCS: could not create client", "error", err)
		return
	}

	for repo := range ru.queue {
		op := &sourcegraph.MirrorReposRefreshVCSOp{
			Repo: repo.RepoSpec(),
		}

		log15.Debug("repoUpdater: RefreshVCS:", "repo", repo.URI)
		if _, err := apiclient.MirrorRepos.RefreshVCS(ctx, op); err != nil {
			log15.Warn("repoUpdater: RefreshVCS:", "error", err)
			continue
		}
	}
}

// enqueueLater is called to enqueue the repo automatically at a later time.
func (ru *repoUpdater) enqueueLater(repo *sourcegraph.Repo) {
	// Sleep a tiny bit longer than MirrorUpdateRate to avoid our
	// enqueue being no-op / hitting "was recently updated".
	time.Sleep(appconf.Flags.MirrorRepoUpdateRate + (200 * time.Millisecond))
	ru.enqueue(repo)
}
