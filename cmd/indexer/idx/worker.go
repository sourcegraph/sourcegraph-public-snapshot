package idx

import (
	"context"
	"sync"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

type Worker struct {
	Ctx context.Context

	currentJobs   map[qitem]struct{}
	currentJobsMu sync.Mutex

	primary   workQueue    // primary queue from which to draw repositories to index
	secondary <-chan qitem // secondary queue from which to draw repositories to index
}

func (w *Worker) Enqueue(repo api.RepoURI, rev string) {
	w.primary.Enqueue(repo, rev)
}

func NewWorker(ctx context.Context, primary workQueue, secondary <-chan qitem) *Worker {
	return &Worker{
		Ctx:         ctx,
		currentJobs: make(map[qitem]struct{}),
		primary:     primary,
		secondary:   secondary,
	}
}

func (w *Worker) Work() {
	for {
		var (
			isPrimary bool
			repoRev   qitem
		)

		// Select a job from the primary queue if possible. Otherwise,
		// select a job from the secondary queue or wait for whichever
		// one opens up first.
		c := make(chan qitem)
		select {
		case w.primary.dequeue <- c:
			repoRev = <-c
			isPrimary = true
		default:
			select {
			case w.primary.dequeue <- c:
				repoRev = <-c
				isPrimary = true
			case repoRev = <-w.secondary:
			}
		}

		{ // detect duplicate jobs
			w.currentJobsMu.Lock()
			if _, ok := w.currentJobs[repoRev]; ok {
				w.currentJobsMu.Unlock()
				continue // in progress, discard
			}
			w.currentJobs[repoRev] = struct{}{}
			w.currentJobsMu.Unlock()
		}

		start := time.Now()
		log15.Debug("Index started", "repo", repoRev.repo, "rev", repoRev.rev, "isPrimary", isPrimary)
		if err := w.index(repoRev.repo, repoRev.rev, isPrimary); err == nil {
			log15.Debug("Index finished", "repo", repoRev.repo, "rev", repoRev.rev, "isPrimary", isPrimary, "duration", time.Since(start))
		} else if vcs.IsCloneInProgress(err) {
			if isPrimary {
				log15.Debug("Index postponed (clone in progress)", "repo", repoRev.repo, "rev", repoRev.rev)
			}
		} else if !git.IsRevisionNotFound(err) {
			log15.Error("Index failed", "repo", repoRev.repo, "rev", repoRev.rev, "isPrimary", isPrimary, "error", err)
		}

		{
			w.currentJobsMu.Lock()
			delete(w.currentJobs, repoRev)
			w.currentJobsMu.Unlock()
		}
	}
}
