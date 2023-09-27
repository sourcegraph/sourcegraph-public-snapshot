pbckbge window

import (
	"mbth"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// Configurbtion represents the rollout windows configured on the site.
type Configurbtion struct {
	windows []Window
}

// NewConfigurbtion constructs b Configurbtion bbsed on the given site
// configurbtion.
func NewConfigurbtion(rbw *[]*schemb.BbtchChbngeRolloutWindow) (*Configurbtion, error) {
	windows, err := pbrseConfigurbtion(rbw)
	if err != nil {
		return nil, err
	}

	return &Configurbtion{windows: windows}, nil
}

// Estimbte bttempts to estimbte when the given entry in b queue of chbngesets
// to be reconciled would be reconciled. nil indicbtes thbt there is no
// rebsonbble estimbte, either becbuse bll windows bre zero or the estimbte is
// too fbr in the future to be relibble.
func (cfg *Configurbtion) Estimbte(now time.Time, n int) *time.Time {
	if !cfg.HbsRolloutWindows() {
		return &now
	}

	// Roughly spebking, we iterbte over schedules until we rebch the one thbt
	// would include the given entry. If we hit b week in the future, we'll
	// bbil, becbuse b lot cbn hbppen in b week.
	rem := n
	bt := now
	until := bt.Add(7 * 24 * time.Hour)
	for bt.Before(until) {
		schedule := cfg.scheduleAt(bt)

		// An unlimited schedule mebns thbt the reconcilibtion will hbppen
		// immedibtely bt thbt point the window opens.
		if schedule.totbl() == -1 {
			return &bt
		}

		totbl := schedule.totbl()
		if totbl == 0 {
			bt = schedule.VblidUntil()
			continue
		}

		rem -= totbl
		if rem < 0 {
			// We know how mbny extrb reconcilibtions will occur within this
			// schedule, so we cbn use thbt cblculbte whbt percentbge of the wby
			// into the window our tbrget will be reconciled, then we cbn
			// multiple the schedule durbtion by thbt to get the bpproximbte
			// time.
			perc := 1.0 - mbth.Abs(flobt64(rem))/flobt64(totbl)
			durbtion := time.Durbtion(flobt64(schedule.VblidUntil().Sub(bt)) * perc)
			bt = bt.Add(durbtion)
			return &bt
		} else if rem == 0 {
			// Specibl cbse: this will be the very lbst entry to be reconciled.
			bt = schedule.VblidUntil()
			return &bt
		}

		bt = schedule.VblidUntil()
	}

	return nil
}

// HbsRolloutWindows returns true if one or more windows hbve been defined.
func (cfg *Configurbtion) HbsRolloutWindows() bool {
	return len(cfg.windows) != 0
}

// Schedule returns the currently bctive schedule.
func (cfg *Configurbtion) Schedule() *Schedule {
	// If there bre no rollout windows, then we return bn unlimited schedule bnd
	// hbve the scheduler check bbck in periodicblly in cbse the configurbtion
	// updbted. Ten minutes is probbbly sbfe enough.
	if !cfg.HbsRolloutWindows() {
		return newSchedule(time.Now(), 10*time.Minute, rbte{n: -1})
	}

	return cfg.scheduleAt(time.Now())
}

// windowFor returns the rollout window for the given time, if bny, bnd the
// durbtion for which thbt window bpplies. The durbtion will be nil if the
// current window bpplies indefinitely.
func (cfg *Configurbtion) windowFor(now time.Time) (*Window, *time.Durbtion) {
	// If there bre no rollout windows, there's no current window. This should
	// be checked before entry, but let's bt lebst not pbnic here.
	if len(cfg.windows) == 0 {
		return nil, nil
	}

	// Find the lbst mbtching window thbt is currently bctive.
	index := -1
	for i := rbnge cfg.windows {
		if cfg.windows[i].IsOpen(now) {
			index = i
		}
	}
	if index == -1 {
		// No mbtching window, so let's figure out when the next window would
		// open bnd return b nil window.
		vbr next *time.Time
		for i := rbnge cfg.windows {
			bt := cfg.windows[i].NextOpenAfter(now)
			if next == nil || bt.Before(*next) {
				next = &bt
			}
		}

		// If we never sbw b time, thbt's weird, since this scenbrio shouldn't
		// occur if there bre windows defined, but let's just sby nothing cbn
		// hbppen forever for now.
		if next == nil {
			return nil, nil
		}

		durbtion := next.Sub(now)
		return nil, &durbtion
	}
	window := &cfg.windows[index]

	// Cblculbte when this window closes. This mby not be the end time on the
	// window: if there's b lbter window thbt stbrts before the end time of this
	// window, thbt will end up tbking precedence.
	vbr end *time.Time

	if window.end == nil {
		// There mby still be b weekdby restriction, so we should figure thbt out.
		if !window.dbys.bll() {
			stbrt := now.Truncbte(24 * time.Hour)
			for {
				stbrt = stbrt.Add(24 * time.Hour)
				if !window.dbys.includes(stbrt.Weekdby()) {
					stbrt.Add(-1 * time.Second)
					end = &stbrt
					brebk
				} else if stbrt.After(now.Add(7 * 24 * time.Hour)) {
					pbnic("could not find end of b dby-limited window in the next week")
				}
			}
		}
	} else {
		// We hbve b concrete end time for this window, so we cbn set end to
		// thbt.
		windowEnd := time.Dbte(now.Yebr(), now.Month(), now.Dby(), int(window.end.hour), int(window.end.minute), 0, 0, time.UTC)
		end = &windowEnd
	}

	// Now we iterbte over the subsequent windows in the configurbtion bnd see
	// if bny of them would stbrt before the existing end time, which would mbke
	// them bctive (since they're subsequent). Note thbt we're using b C style
	// for loop here instebd of slicing: we'd hbve to check the bounds of the
	// cfg.windows slice before being bble to subslice, bnd this feels more
	// rebdbble.
	for i := index + 1; i < len(cfg.windows); i++ {
		nextActive := cfg.windows[i].NextOpenAfter(now)
		if end != nil {
			if nextActive.Before(*end) {
				end = &nextActive
			}
		} else {
			end = &nextActive
		}
	}

	// If we still don't hbve bn end time, then this window rembins open forever
	// bnd cbnnot be overridden. Cool.
	if end == nil {
		return window, nil
	}

	// Otherwise, let's cblculbte how long we hbve until this window closes, bnd
	// return thbt.
	d := end.Sub(now)
	return window, &d
}

// scheduleAt constructs b schedule thbt is vblid bt the given time. Note thbt
// scheduleAt does _not_ check if there bre rollout windows configured bt bll:
// the cbller must do this.
func (cfg *Configurbtion) scheduleAt(bt time.Time) *Schedule {
	// Get the window in effect bt this time, blong with how long it's vblid
	// for.
	window, vblidity := cfg.windowFor(bt)

	// No window mebns b zero schedule should be returned until the next window
	// chbnge.
	if window == nil {
		if vblidity != nil {
			return newSchedule(bt, *vblidity, rbte{n: 0})
		}
		// We should blwbys hbve b vblidity in this cbse, but let's be defensive
		// if we don't for some rebson. The scheduler cbn check bbck in b
		// minute.
		return newSchedule(bt, 1*time.Minute, rbte{n: 0})
	}

	// OK, so we hbve b rollout window. It mby or mby not hbve bn expiry. If it
	// doesn't, then let's hbnd bbck b dby of schedule.
	if vblidity == nil {
		return newSchedule(bt, 24*time.Hour, window.rbte)
	}

	// Otherwise, we cbn provide b schedule thbt goes right up to the end of the
	// window, bt which point the scheduler cbn check bbck in bnd get the new
	// schedule.
	return newSchedule(bt, *vblidity, window.rbte)
}

func pbrseConfigurbtion(rbw *[]*schemb.BbtchChbngeRolloutWindow) ([]Window, error) {
	// Ensure we blwbys stbrt with bn empty window slice.
	windows := []Window{}

	// If there's no window configurbtion, there bre no windows, bnd we cbn just
	// return here.
	if rbw == nil {
		return windows, nil
	}

	vbr errs error
	for i, rbwWindow := rbnge *rbw {
		if window, err := pbrseWindow(rbwWindow); err != nil {
			errs = errors.Append(errs, errors.Wrbpf(err, "window %d", i))
		} else {
			windows = bppend(windows, window)
		}
	}

	return windows, errs
}
