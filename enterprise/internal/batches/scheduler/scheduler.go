package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/scheduler/window"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/batches"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type Scheduler struct {
	cfg   *window.Configuration
	ctx   context.Context
	done  chan struct{}
	store *store.Store
}

var _ goroutine.BackgroundRoutine = &Scheduler{}

func NewScheduler(ctx context.Context, bstore *store.Store) *Scheduler {
	log15.Info("creating a batch change scheduler")
	return &Scheduler{
		cfg:   window.NewConfiguration(),
		ctx:   ctx,
		done:  make(chan struct{}),
		store: bstore,
	}
}

func (s *Scheduler) Start() {
	goroutine.Go(func() {
		log15.Info("starting batch change scheduler")

		for {
			schedule := s.cfg.Schedule()
			taker := newTaker(schedule)
			timer := time.NewTimer(time.Until(schedule.ValidUntil()))

			log15.Info("applying batch change schedule", "schedule", schedule, "until", schedule.ValidUntil())

			// TODO: don't busy-wait until we actually _have_ something to
			// schedule.

			for {
				select {
				case <-taker.C:
					if cs, err := s.store.GetNextScheduledChangeset(s.ctx); err == store.ErrNoResults {
						// TODO: be smarter about the busy-wait.
						log15.Info("no scheduled changeset waiting to be queued")
						time.Sleep(schedule.Sleep())
					} else if err != nil {
						log15.Warn("error retrieving the next scheduled changeset", "err", err)
					} else {
						log15.Info("queueing changeset", "changeset", cs)
						cs.ReconcilerState = batches.ReconcilerStateQueued
						if err := s.store.UpsertChangeset(s.ctx, cs); err != nil {
							log15.Warn("error updating the next scheduled changeset", "err", err, "changeset", cs)
						}
					}
				case <-timer.C:
					log15.Info("current batch change schedule is outdated; looping")
					continue
				case <-s.done:
					log15.Info("stopping the batch change scheduler")
					return
				}
			}
		}
	})
}

func (s *Scheduler) Stop() {
	s.done <- struct{}{}
}

// taker calls Take on a schedule until it is told not to. Note that this type
// does not enforce the validity period of the schedule; this must be done
// outside the taker and signalled in via stop().
type taker struct {
	C chan time.Time

	mu       sync.Mutex
	done     bool
	schedule window.Schedule
}

func newTaker(schedule window.Schedule) *taker {
	t := &taker{
		C:        make(chan time.Time),
		done:     false,
		schedule: schedule,
	}

	goroutine.Go(func() {
		for {
			at, err := schedule.Take()
			if err != nil {
				log15.Warn("error taking from schedule", "schedule", t.schedule, "err", err)
				return
			}

			t.mu.Lock()
			// We want to check this _after_ the Take call potentially blocks so
			// that we don't send down a channel that's no longer listening.
			if t.done {
				t.mu.Unlock()
				return
			}

			t.C <- at
			t.mu.Unlock()
		}
	})

	return t
}

func (t *taker) stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.done = true
}
