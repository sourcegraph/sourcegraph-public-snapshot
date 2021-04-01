package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/scheduler/window"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
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
	return &Scheduler{
		cfg:   window.NewConfiguration(),
		ctx:   ctx,
		done:  make(chan struct{}),
		store: bstore,
	}
}

func (s *Scheduler) Start() {
	goroutine.Go(func() {
		log15.Debug("starting batch change scheduler")

		for {
			schedule := s.cfg.Schedule()
			taker := newTaker(schedule)
			timer := time.NewTimer(time.Until(schedule.ValidUntil()))

			log15.Debug("applying batch change schedule", "schedule", schedule, "until", schedule.ValidUntil())

			select {
			case <-taker.C:
				log15.Debug("would apply the next changeset")
				// TODO: grab the next scheduled changeset and queue it.
			case <-timer.C:
				log15.Debug("current batch change schedule is outdated; looping")
				continue
			case <-s.done:
				log15.Debug("stopping the batch change scheduler")
				return
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
