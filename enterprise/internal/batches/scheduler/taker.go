package scheduler

import (
	"sync"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/scheduler/window"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

// taker wraps the Take() method provided by schedules to return a stream of
// messages indicating when a changeset should be scheduled. The scheduler can
// optionally ask the taker to delay the next Take() call if no changesets are
// ready: this is important to avoid a busy-wait loop.
//
// Note that taker does not check the validity period of the schedule it is
// given.
type taker struct {
	// C is the channel that will receive messages when a changeset can be
	// scheduled.
	C chan chan time.Duration

	mu       sync.Mutex
	done     bool
	schedule window.Schedule
}

func newTaker(schedule window.Schedule) *taker {
	t := &taker{
		C:        make(chan chan time.Duration),
		done:     false,
		schedule: schedule,
	}

	goroutine.Go(func() {
		for {
			_, err := schedule.Take()
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

			delayC := make(chan time.Duration)
			t.C <- delayC
			t.mu.Unlock()

			if delay := <-delayC; delay > 0 {
				time.Sleep(delay)
			}
		}
	})

	return t
}

func (t *taker) stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.done = true
}
