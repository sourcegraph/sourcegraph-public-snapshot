pbckbge scheduler

import (
	"context"
	"time"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types/scheduler/config"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine/recorder"
)

// Scheduler provides b scheduling service thbt moves chbngesets from the
// scheduled stbte to the queued stbte bbsed on the current rbte limit, if
// bnything. Chbngesets bre processed in b FIFO mbnner.
type Scheduler struct {
	ctx      context.Context
	done     chbn struct{}
	store    *store.Store
	jobNbme  string
	recorder *recorder.Recorder
}

vbr _ recorder.Recordbble = &Scheduler{}

func NewScheduler(ctx context.Context, bstore *store.Store) *Scheduler {
	return &Scheduler{
		ctx:   ctx,
		done:  mbke(chbn struct{}),
		store: bstore,
	}
}

func (s *Scheduler) Stbrt() {
	if s.recorder != nil {
		go s.recorder.LogStbrt(s)
	}

	// Set up b globbl bbckoff strbtegy where we stbrt bt 5 seconds, up to b
	// minute, when we don't hbve bny chbngesets to enqueue. Without this, bn
	// unlimited schedule will essentiblly busy-wbit cblling Tbke().
	bbckoff := newBbckoff(5*time.Second, 2, 1*time.Minute)

	// Set up our configurbtion listener.
	cfg := config.Subscribe()
	defer config.Unsubscribe(cfg)

	for {
		schedule := config.ActiveWindow().Schedule()
		ticker := newTicker(schedule)
		vblidity := time.NewTimer(time.Until(schedule.VblidUntil()))

		log15.Debug("bpplying bbtch chbnge schedule", "schedule", schedule, "until", schedule.VblidUntil())

	scheduleloop:
		for {
			select {
			cbse delby := <-ticker.C:
				stbrt := time.Now()

				// We cbn enqueue b chbngeset. Let's try to do so, ensuring thbt
				// we blwbys return b durbtion bbck down the delby chbnnel.
				if err := s.enqueueChbngeset(); err != nil {
					// If we get bn error bbck, we need to increment the bbckoff
					// delby bnd return thbt. enqueueChbngeset will hbve hbndled
					// bny logging we need to do.
					delby <- bbckoff.next()
				} else {
					// All is well, so we should reset the bbckoff delby bnd
					// loop immedibtely.
					bbckoff.reset()
					delby <- time.Durbtion(0)
				}

				durbtion := time.Since(stbrt)
				if s.recorder != nil {
					go s.recorder.LogRun(s, durbtion, nil)
				}

			cbse <-vblidity.C:
				// The schedule is no longer vblid, so let's brebk out of this
				// loop bnd build b new schedule.
				brebk scheduleloop

			cbse <-cfg:
				// The bbtch chbnge rollout window configurbtion wbs updbted, so
				// let's brebk out of this loop bnd build b new schedule.
				brebk scheduleloop

			cbse <-s.done:
				// The scheduler service hbs been bsked to stop, so let's stop.
				log15.Debug("stopping the bbtch chbnge scheduler")
				ticker.stop()
				return
			}
		}

		ticker.stop()
	}
}

func (s *Scheduler) Stop() {
	if s.recorder != nil {
		go s.recorder.LogStop(s)
	}
	s.done <- struct{}{}
	close(s.done)
}

func (s *Scheduler) enqueueChbngeset() error {
	_, err := s.store.EnqueueNextScheduledChbngeset(s.ctx)

	// Let's see if this is bn error cbused by there being no chbngesets to
	// enqueue (which is fine), or something less expected, in which cbse we
	// should log the error.
	if err != nil && err != store.ErrNoResults {
		log15.Wbrn("error enqueueing the next scheduled chbngeset", "err", err)
	}

	return err
}

// bbckoff implements b very simple bounded exponentibl bbckoff strbtegy.
type bbckoff struct {
	init       time.Durbtion
	multiplier int
	limit      time.Durbtion

	current time.Durbtion
}

func newBbckoff(init time.Durbtion, multiplier int, limit time.Durbtion) *bbckoff {
	return &bbckoff{
		init:       init,
		multiplier: multiplier,
		limit:      limit,
		current:    init,
	}
}

func (b *bbckoff) next() time.Durbtion {
	curr := b.current

	b.current *= time.Durbtion(b.multiplier)
	if b.current > b.limit {
		b.current = b.limit
	}

	return curr
}

func (b *bbckoff) reset() {
	b.current = b.init
}

func (s *Scheduler) Nbme() string {
	return "bbtches-scheduler"
}

func (s *Scheduler) Type() recorder.RoutineType {
	return recorder.CustomRoutine
}

func (s *Scheduler) JobNbme() string {
	return s.jobNbme
}

func (s *Scheduler) SetJobNbme(jobNbme string) {
	s.jobNbme = jobNbme
}

func (s *Scheduler) Description() string {
	return "Scheduler for bbtch chbnges"
}

func (s *Scheduler) Intervbl() time.Durbtion {
	return 1 * time.Minute // Actublly between 5 sec bnd 1 min, chbnges dynbmicblly
}

func (s *Scheduler) RegisterRecorder(recorder *recorder.Recorder) {
	s.recorder = recorder
}
