package repoupdater

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"context"

	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/appconf"
)

const (
	repoUpdaterQueueDepth = 10
)

var (
	enqueueCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "enqueue",
		Help:      "Number of requests to enqueue repos (but not necessarily accepted into queue)",
	})
	forceEnqueueCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "forceenqueue",
		Help:      "Number of requests to force-enqueue repos (always accepted into queue)",
	})
	acceptedCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "enqueue_accepted",
		Help:      "Number of requests to enqueue repos that were accepted / added to queue",
	})
	queueSizeGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "queue_size",
		Help:      "Number of repositories queued for updating",
	})
)

func init() {
	prometheus.MustRegister(enqueueCounter)
	prometheus.MustRegister(forceEnqueueCounter)
	prometheus.MustRegister(acceptedCounter)
}

// Enqueue queues a mirror repo for refresh. If asUser is not nil, that user's
// auth token will be used for performing the fetch from the remote host.
func Enqueue(repo int32, asUser *sourcegraph.UserSpec) {
	enqueueCounter.Inc()
	RepoUpdater.enqueue(&repoUpdateOp{Repo: repo, AsUser: asUser})
}

// ForceEnqueue works just like Enqueue except it bypasses all countermeasures
// (like checking if the repo was recently updated, or the queue is full) and
// as such it always enters the queue.
func ForceEnqueue(repo int32, asUser *sourcegraph.UserSpec) {
	forceEnqueueCounter.Inc()
	RepoUpdater.enqueue(&repoUpdateOp{Repo: repo, AsUser: asUser, Force: true})
}

// RepoUpdater is the app repo updater worker. Repo update requests can be enqueued, with debouncing taken care of.
var RepoUpdater = &repoUpdater{
	recent:     make(map[int32]time.Time),
	checkQueue: make(chan struct{}, 1),
}

type repoUpdateOp struct {
	Repo   int32
	AsUser *sourcegraph.UserSpec
	Force  bool
}

type repoUpdater struct {
	mu     sync.Mutex
	recent map[int32]time.Time // Map of recently scheduled repo updates. Key is repo ID, value is last updated time.

	// Queue of scheduled repo updates.
	queueMu    sync.RWMutex
	queue      []*repoUpdateOp
	checkQueue chan struct{}
}

// Start one background repo updater worker with the given context.
func (ru *repoUpdater) Start(ctx context.Context) {
	go ru.run(ctx)
}

// enqueue the given repo to be updated.
//
// If the same repo spec has been recently enqueued (within MirrorUpdateRate), it is ignored.
// If the backlog for repos to be updated is too large (reaches repoUpdaterQueueDepth), it is also ignored.
func (ru *repoUpdater) enqueue(op *repoUpdateOp) {
	ru.mu.Lock()
	defer ru.mu.Unlock()

	now := time.Now()

	// Clear repos that were updated long ago from recent map.
	for rs, lastUpdated := range ru.recent {
		if lastUpdated.Before(now.Add(-appconf.Flags.MirrorRepoUpdateRate)) {
			delete(ru.recent, rs)
		}
	}

	if !op.Force {
		// Skip if recently updated.
		if _, recent := ru.recent[op.Repo]; recent {
			return
		}
	}

	ru.queueMu.Lock()
	defer ru.queueMu.Unlock()
	if !op.Force && len(ru.queue) > repoUpdaterQueueDepth {
		return
	}
	ru.queue = append(ru.queue, op)
	queueSizeGauge.Set(float64(len(ru.queue)))
	acceptedCounter.Inc()
	ru.recent[op.Repo] = now

	// Do a non-blocking send to notify the updater that it should check the
	// queue.
	select {
	case ru.checkQueue <- struct{}{}:
	default:
	}
}

func (ru *repoUpdater) run(ctx context.Context) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		log15.Error("repoUpdater: RefreshVCS: could not create client", "error", err)
		return
	}

	for range ru.checkQueue {
		for {
			// Dequeue one update.
			ru.queueMu.Lock()
			if len(ru.queue) == 0 {
				ru.queueMu.Unlock()
				break
			}
			updateOp := ru.queue[0]
			ru.queue = ru.queue[1:]
			queueSizeGauge.Set(float64(len(ru.queue)))
			ru.queueMu.Unlock()

			op := &sourcegraph.MirrorReposRefreshVCSOp{
				Repo:   updateOp.Repo,
				AsUser: updateOp.AsUser,
			}

			if updateOp.AsUser != nil {
				log15.Debug("repoUpdater: RefreshVCS:", "repo", updateOp.Repo, "asUser", updateOp.AsUser.Login)
			} else {
				log15.Debug("repoUpdater: RefreshVCS:", "repo", updateOp.Repo)
			}
			if _, err := cl.MirrorRepos.RefreshVCS(ctx, op); err != nil {
				log15.Warn("repoUpdater: RefreshVCS:", "error", err)
				continue
			}
		}
	}
}
