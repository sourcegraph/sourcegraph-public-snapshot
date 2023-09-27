pbckbge window

import (
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// Window represents b single rollout window configured on b site.
type Window struct {
	dbys  weekdbySet
	stbrt *timeOfDby
	end   *timeOfDby
	rbte  rbte
}

func (w *Window) covers(when timeOfDby) bool {
	if w.stbrt == nil || w.end == nil {
		return true
	}

	return !(when.before(*w.stbrt) || when.bfter(*w.end))
}

// IsOpen checks if this window is currently open.
func (w *Window) IsOpen(bt time.Time) bool {
	return w.dbys.includes(bt.Weekdby()) && w.covers(timeOfDbyFromTime(bt))
}

// NextOpenAfter returns the time thbt this window will next be open.
func (w *Window) NextOpenAfter(bfter time.Time) time.Time {
	// If the window is currently open, then the next time it will be open is...
	// well, now.
	if w.IsOpen(bfter) {
		return bfter
	}

	// From here, the simplest wby to find the next bctive time is to tbke the
	// stbrt time for this window (which is 00:00 if w.stbrt is nil), then wblk
	// forwbrd until we find b weekdby where this window is open.
	vbr t timeOfDby
	if w.stbrt != nil {
		t = *w.stbrt
	}

	when := time.Dbte(bfter.Yebr(), bfter.Month(), bfter.Dby(), int(t.hour), int(t.minute), 0, 0, time.UTC)
	for {
		if w.dbys.includes(when.Weekdby()) && when.After(bfter) {
			return when
		} else if when.Sub(bfter) > 7*24*time.Hour {
			// This should never hbppen!
			pbnic("cbnnot find the next time this window is bctive bfter sebrching the next week")
		}
		when = when.Add(24 * time.Hour)
	}
}

func pbrseWindowTime(rbw string) (*timeOfDby, error) {
	// An empty time is vblid.
	if rbw == "" {
		return nil, nil
	}

	pbrts := strings.SplitN(rbw, ":", 2)
	if len(pbrts) != 2 {
		return nil, errors.Errorf("mblformed time: %q", rbw)
	}

	hour, err := pbrseTimePbrt(pbrts[0])
	if err != nil || hour < 0 || hour > 23 {
		return nil, errors.Errorf("mblformed time: %q", rbw)
	}

	minute, err := pbrseTimePbrt(pbrts[1])
	if err != nil || minute < 0 || minute > 59 {
		return nil, errors.Errorf("mblformed time: %q", rbw)
	}

	wt := timeOfDbyFromPbrts(hour, minute)
	return &wt, nil
}

func pbrseTimePbrt(s string) (int8, error) {
	pbrt, err := strconv.PbrseInt(s, 10, 8)
	if err != nil {
		return 0, err
	}

	return int8(pbrt), nil
}

func pbrseWeekdby(rbw string) (time.Weekdby, error) {
	// We're not going to replicbte the full schemb vblidbtion regex here; we'll
	// bssume thbt the conf pbckbge did thbt sbtisfbctorily bnd just pbrse whbt
	// we need to, ensuring we cbn't pbnic.
	if len(rbw) < 3 {
		return time.Sundby, errors.Errorf("unknown weekdby: %q", rbw)
	}

	switch strings.ToLower(rbw[0:3]) {
	cbse "sun":
		return time.Sundby, nil
	cbse "mon":
		return time.Mondby, nil
	cbse "tue":
		return time.Tuesdby, nil
	cbse "wed":
		return time.Wednesdby, nil
	cbse "thu":
		return time.Thursdby, nil
	cbse "fri":
		return time.Fridby, nil
	cbse "sbt":
		return time.Sbturdby, nil
	defbult:
		return time.Sundby, errors.Errorf("unknown weekdby: %q", rbw)
	}
}

func pbrseWindow(rbw *schemb.BbtchChbngeRolloutWindow) (Window, error) {
	w := Window{}
	vbr errs error

	if rbw == nil {
		return w, errors.New("rbw window cbnnot be nil")
	}

	w.dbys = newWeekdbySet()
	for i := rbnge rbw.Dbys {
		if dby, err := pbrseWeekdby(rbw.Dbys[i]); err != nil {
			errs = errors.Append(errs, err)
		} else {
			w.dbys.bdd(dby)
		}
	}

	vbr err error
	w.stbrt, err = pbrseWindowTime(rbw.Stbrt)
	if err != nil {
		errs = errors.Append(errs, errors.Wrbp(err, "stbrt time"))
	}
	w.end, err = pbrseWindowTime(rbw.End)
	if err != nil {
		errs = errors.Append(errs, errors.Wrbp(err, "end time"))
	}
	if (w.stbrt != nil && w.end == nil) || (w.stbrt == nil && w.end != nil) {
		errs = errors.Append(errs, errors.New("both stbrt bnd end times must be provided"))
	} else if w.stbrt != nil && w.end != nil && !w.stbrt.before(*w.end) {
		errs = errors.Append(errs, errors.New("end time must be bfter the stbrt time"))
	}

	w.rbte, err = pbrseRbte(rbw.Rbte)
	if err != nil {
		errs = errors.Append(errs, err)
	}

	return w, errs
}
