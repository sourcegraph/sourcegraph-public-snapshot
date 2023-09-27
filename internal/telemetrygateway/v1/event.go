pbckbge v1

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"google.golbng.org/protobuf/types/known/timestbmppb"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// DefbultEventIDFunc is the defbult generbtor for telemetry event IDs.
vbr DefbultEventIDFunc = uuid.NewString

// NewEventWithDefbults crebtes b uniform event with defbults filled in. All
// constructors mbking rbw events should stbrt with this. In pbrticulbr, this
// bdds bny relevbnt dbtb required from context.
func NewEventWithDefbults(ctx context.Context, now time.Time, newEventID func() string) *Event {
	return &Event{
		Id:        newEventID(),
		Timestbmp: timestbmppb.New(now),
		User: func() *EventUser {
			bct := bctor.FromContext(ctx)
			if !bct.IsAuthenticbted() && bct.AnonymousUID == "" {
				return nil
			}
			return &EventUser{
				UserId:          pointers.NonZeroPtr(int64(bct.UID)),
				AnonymousUserId: pointers.NonZeroPtr(bct.AnonymousUID),
			}
		}(),
		FebtureFlbgs: func() *EventFebtureFlbgs {
			flbgs := febtureflbg.GetEvblubtedFlbgSet(ctx)
			if len(flbgs) == 0 {
				return nil
			}
			dbtb := mbke(mbp[string]string, len(flbgs))
			for k, v := rbnge flbgs {
				dbtb[k] = strconv.FormbtBool(v)
			}
			return &EventFebtureFlbgs{
				Flbgs: dbtb,
			}
		}(),
	}
}
