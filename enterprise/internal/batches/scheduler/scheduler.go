package scheduler

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/scheduler/window"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/batches"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type Scheduler struct {
	ctx   context.Context
	done  chan struct{}
	store *store.Store

	cfg chan *window.Configuration
}

var _ goroutine.BackgroundRoutine = &Scheduler{}

func NewScheduler(ctx context.Context, bstore *store.Store) *Scheduler {
	log15.Info("creating a batch change scheduler")
	s := &Scheduler{
		ctx:   ctx,
		done:  make(chan struct{}),
		store: bstore,
		cfg:   make(chan *window.Configuration),
	}

	// Set up a configuration watcher so that we update the window
	// configuration when the configuration updates.
	goroutine.Go(func() {
		// FIXME: this gets leaked if the scheduler is dropped, since we can't
		// hook directly into the configuration client's watcher array.
		conf.Watch(func() {
			// TODO: check if the configuration actually changed.
			cfg, err := window.NewConfiguration(conf.Get().BatchChangesRolloutWindows)
			if err != nil {
				log15.Warn("invalid batch change rollout window configuration", "err", err)
				cfg, _ = window.NewConfiguration(nil)
			}

			s.cfg <- cfg
		})
	})

	return s
}

func (s *Scheduler) Start() {
	goroutine.Go(func() {
		log15.Info("starting batch change scheduler")

		// Set up a global backoff strategy where we start at 5 seconds, up to a
		// minute.
		backoff := newBackoff(5*time.Second, 2, 1*time.Minute)

		// Retrieve the initial configuration.
		cfg := <-s.cfg

		for {
			schedule := cfg.Schedule()
			taker := newTaker(schedule)
			timer := time.NewTimer(time.Until(schedule.ValidUntil()))

			log15.Info("applying batch change schedule", "schedule", schedule, "until", schedule.ValidUntil())

			for {
				select {
				case delay := <-taker.C:
					if cs, err := s.store.GetNextScheduledChangeset(s.ctx); err == store.ErrNoResults {
						log15.Info("no scheduled changeset waiting to be queued")
						delay <- backoff.next()
					} else if err != nil {
						log15.Warn("error retrieving the next scheduled changeset", "err", err)
						delay <- backoff.next()
					} else {
						log15.Info("queueing changeset", "changeset", cs)
						cs.ReconcilerState = batches.ReconcilerStateQueued
						if err := s.store.UpsertChangeset(s.ctx, cs); err != nil {
							log15.Warn("error updating the next scheduled changeset", "err", err, "changeset", cs)
						}
						backoff.reset()
						delay <- 0
					}
				case <-timer.C:
					log15.Info("current batch change schedule is outdated; looping")
					continue
				case cfg = <-s.cfg:
					log15.Info("batch change configuration updated; looping")
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

type backoff struct {
	init       time.Duration
	multiplier int
	limit      time.Duration

	current time.Duration
}

func newBackoff(init time.Duration, multiplier int, limit time.Duration) *backoff {
	return &backoff{
		init:       init,
		multiplier: multiplier,
		limit:      limit,
		current:    init,
	}
}

func (b *backoff) next() time.Duration {
	b.current *= time.Duration(b.multiplier)
	if b.current > b.limit {
		b.current = b.limit
	}

	return b.current
}

func (b *backoff) reset() {
	b.current = b.init
}
