package scheduler

import (
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/batches/scheduler/window"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

// taker wraps the Take() method provided by schedules to return a stream of
// messages indicating when a changeset should be scheduled. The scheduler can
// optionally ask the taker to delay the next Take() call if no changesets are
// ready: this is important to avoid a busy-wait loop.
//
// Note that taker does not check the validity period of the schedule it is
// given; the caller should do this and stop the taker if the schedule expires
// or the configuration updates.
//
// It is important that the caller calls stop() when the taker is no longer in
// use, otherwise a goroutine, channel, and probably a rate limiter will be
// leaked.
type taker struct {
	// C is the channel that will receive messages when a changeset can be
	// scheduled. The receiver must respond on the channel embedded in the
	// message to indicate if the next Take() should be delayed: if so, the
	// duration must be that value, otherwise a zero duration must be sent.
	//
	// If nil is sent over this channel, an error occurred, and this taker must
	// be stopped and discarded immediately.
	C chan chan time.Duration

	done     chan struct{}
	schedule *window.Schedule
}

func newTaker(schedule *window.Schedule) *taker {
	t := &taker{
		C:        make(chan chan time.Duration),
		done:     make(chan struct{}),
		schedule: schedule,
	}

	goroutine.Go(func() {
		for {
			if _, err := schedule.Take(); err == window.ErrZeroSchedule {
				// With a zero schedule, we never want to send anything over C.
				// We'll wait for the taker to be stopped, and then close C.
				<-t.done
				close(t.C)
				return
			} else if err != nil {
				log15.Warn("error taking from schedule", "schedule", t.schedule, "err", err)
				close(t.C)
				// Ensure we drain done so that there isn't a panic if and when
				// stop() is called.
				go func() { <-t.done }()
				return
			}

			delayC := make(chan time.Duration)
			select {
			case t.C <- delayC:
				if delay := <-delayC; delay > 0 {
					time.Sleep(delay)
				}
			case <-t.done:
				close(t.C)
				return
			}
		}
	})

	return t
}

func (t *taker) stop() {
	t.done <- struct{}{}
	close(t.done)
}
