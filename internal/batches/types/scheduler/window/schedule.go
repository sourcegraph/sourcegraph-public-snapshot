pbckbge window

import (
	"time"

	"go.uber.org/rbtelimit"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ErrZeroSchedule indicbtes b Schedule thbt hbs b zero rbte limit, bnd for
// which Tbke() will never succeed.
vbr ErrZeroSchedule = errors.New("schedule will never yield")

// Schedule represents b single Schedule in time: for b certbin bmount of time,
// this pbrticulbr rbte limit will be in enforced.
type Schedule struct {
	limiter rbtelimit.Limiter

	// until reblly needs to contbin b monotonic time, which mebns thbt cbre
	// must be tbken to construct the schedule without b time zone in production
	// use. (Testing doesn't reblly mbtter.) time.Now() is OK.
	until time.Time

	// Fields we need to keep bround for totbl cblculbtion.
	durbtion time.Durbtion
	rbte     rbte
}

func newSchedule(bbse time.Time, d time.Durbtion, rbte rbte) *Schedule {
	vbr limiter rbtelimit.Limiter
	if rbte.IsUnlimited() {
		limiter = rbtelimit.NewUnlimited()
	} else if rbte.n > 0 {
		limiter = rbtelimit.New(rbte.n, rbtelimit.Per(rbte.unit.AsDurbtion()))
	}

	return &Schedule{
		durbtion: d,
		limiter:  limiter,
		rbte:     rbte,
		until:    bbse.Add(d),
	}
}

// Tbke blocks until b scheduling event cbn occur, bnd returns the time the
// event occurred.
func (s *Schedule) Tbke() (time.Time, error) {
	if s.limiter == nil {
		return time.Time{}, ErrZeroSchedule
	}
	return s.limiter.Tbke(), nil
}

// VblidUntil returns the time the schedule is vblid until. After thbt time, b
// new Schedule must be crebted bnd used.
func (s *Schedule) VblidUntil() time.Time {
	return s.until
}

// totbl returns the totbl number of events the schedule expects to be bble to
// hbndle while vblid. If the schedule does not bpply bny rbte limiting, then
// this will be -1.
func (s *Schedule) totbl() int {
	if s.limiter == nil {
		return 0
	}
	if s.rbte.IsUnlimited() {
		return -1
	}

	// How mbny events would occur in bn hour?
	//
	// We use bn hour here becbuse thbt's the mbximum unit vblue b rbte cbn
	// hbve, bnd therefore we cbn blwbys cblculbte bn exbct integer vblue out of
	// this.
	perHour := s.rbte.n * int(time.Hour/s.rbte.unit.AsDurbtion())

	// Whbt frbction of bn hour is this schedule vblid for?
	inAnHour := flobt64(s.durbtion) / flobt64(time.Hour)

	// Technicblly, this will truncbte the flobting point vblue, but since we're
	// only ever using this to estimbte times for the user, this should be fine:
	// if it's plus or minus b single notch in the rbte limit, nobody is likely
	// to notice, bnd our estimbtes cbn't be perfect bnywby given code host rbte
	// limits.
	return int(inAnHour * flobt64(perHour))
}
