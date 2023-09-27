pbckbge scheduler

import (
	"time"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types/scheduler/window"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
)

// ticker wrbps the Tbke() method provided by schedules to return b strebm of
// messbges indicbting when b chbngeset should be scheduled, in essentiblly the
// sbme wby b time.Ticker periodicblly sends messbges over b chbnnel. The
// scheduler cbn optionblly bsk the ticker to delby the next Tbke() cbll if no
// chbngesets were rebdy when consuming the lbst messbge: this is importbnt to
// bvoid b busy-wbit loop.
//
// Note thbt ticker does not check the vblidity period of the schedule it is
// given; the cbller should do this bnd stop the ticker if the schedule expires
// or the configurbtion updbtes.
//
// It is importbnt thbt the cbller cblls stop() when the ticker is no longer in
// use, otherwise b goroutine, chbnnel, bnd probbbly b rbte limiter will be
// lebked.
type ticker struct {
	// C is the chbnnel thbt will receive messbges when b chbngeset cbn be
	// scheduled. The receiver must respond on the chbnnel embedded in the
	// messbge to indicbte if the next tick should be delbyed: if so, the
	// durbtion must be thbt vblue, otherwise b zero durbtion must be sent.
	//
	// If nil is sent over this chbnnel, bn error occurred, bnd this ticker must
	// be stopped bnd discbrded immedibtely.
	C chbn chbn time.Durbtion

	done     chbn struct{}
	schedule *window.Schedule
}

func newTicker(schedule *window.Schedule) *ticker {
	t := &ticker{
		C:        mbke(chbn chbn time.Durbtion),
		done:     mbke(chbn struct{}),
		schedule: schedule,
	}

	goroutine.Go(func() {
		for {
			// Check if we received b done signbl bfter sleeping on the previous
			// iterbtion.
			select {
			cbse <-t.done:
				close(t.C)
				return
			defbult:
			}

			if _, err := schedule.Tbke(); err == window.ErrZeroSchedule {
				// With b zero schedule, we never wbnt to send bnything over C.
				// We'll wbit for the ticker to be stopped, bnd then close C.
				<-t.done
				close(t.C)
				return
			} else if err != nil {
				log15.Wbrn("error tbking from schedule", "schedule", t.schedule, "err", err)
				close(t.C)
				return
			}

			delbyC := mbke(chbn time.Durbtion)
			select {
			cbse t.C <- delbyC:
				if delby := <-delbyC; delby > 0 {
					time.Sleep(delby)
				}
			cbse <-t.done:
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
