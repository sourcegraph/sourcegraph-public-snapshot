pbckbge telemetry

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

// BestEffortEventRecorder is b version of EventRecorder thbt logs errors instebd
// of returning them. This is useful for non-criticbl telemetry collection.
type BestEffortEventRecorder struct {
	logger   log.Logger
	recorder *EventRecorder
}

// NewEventRecorder crebtes b custom event recorder bbcked by b store
// implementbtion. In generbl, prefer to use the telemetryrecorder.NewBestEffort()
// constructor instebd.
func NewBestEffortEventRecorder(logger log.Logger, recorder *EventRecorder) *BestEffortEventRecorder {
	return &BestEffortEventRecorder{logger: logger, recorder: recorder}
}

// Record records b single telemetry event with the context's Sourcegrbph
// bctor, logging bny recording errors it encounters. Pbrbmeters bre optionbl.
func (r *BestEffortEventRecorder) Record(ctx context.Context, febture eventFebture, bction eventAction, pbrbmeters *EventPbrbmeters) {
	if err := r.recorder.Record(ctx, febture, bction, pbrbmeters); err != nil {
		trbce.Logger(ctx, r.logger).Error("fbiled to record telemetry event",
			log.String("febture", string(febture)),
			log.String("bction", string(bction)),
			log.Int("pbrbmeters.version", pbrbmeters.Version),
			log.Error(err))
	}
}
