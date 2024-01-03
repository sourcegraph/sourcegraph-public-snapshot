package scheduler

import (
	"time"

	"github.com/inconshreveable/log15" //nolint:logging // TODO move all logging to sourcegraph/log

	"github.com/sourcegraph/sourcegraph/internal/batches/types/scheduler/window"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

// ticker wraps the Take() method provided by schedules to return a stream of
// messages indicating when a changeset should be scheduled, in essentially the
// same way a time.Ticker periodically sends messages over a channel. The
// scheduler can optionally ask the ticker to delay the next Take() call if no
// changesets were ready when consuming the last message: this is important to
// avoid a busy-wait loop.
//
// Note that ticker does not check the validity period of the schedule it is
// given; the caller should do this and stop the ticker if the schedule expires
// or the configuration updates.
//
// It is important that the caller calls stop() when the ticker is no longer in
// use, otherwise a goroutine, channel, and probably a rate limiter will be
// leaked.
type ticker struct {
	// C is the channel that will receive messages when a changeset can be
	// scheduled. The receiver must respond on the channel embedded in the
	// message to indicate if the next tick should be delayed: if so, the
	// duration must be that value, otherwise a zero duration must be sent.
	//
	// If nil is sent over this channel, an error occurred, and this ticker must
	// be stopped and discarded immediately.
	C chan chan time.Duration

	done     chan struct{}
	schedule *window.Schedule
}

func newTicker(schedule *window.Schedule) *ticker {
	t := &ticker{
		C:        make(chan chan time.Duration),
		done:     make(chan struct{}),
		schedule: schedule,
	}

	goroutine.Go(func() {
		for {
			// Check if we received a done signal after sleeping on the previous
			// iteration.
			select {
			case <-t.done:
				close(t.C)
				return
			default:
			}

			if _, err := schedule.Take(); err == window.ErrZeroSchedule {
				// With a zero schedule, we never want to send anything over C.
				// We'll wait for the ticker to be stopped, and then close C.
				<-t.done
				close(t.C)
				return
			} else if err != nil {
				log15.Warn("error taking from schedule", "schedule", t.schedule, "err", err)
				close(t.C)
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

func (t *ticker) stop() {
	close(t.done)
}
