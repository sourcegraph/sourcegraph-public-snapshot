pbckbge teestore

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sourcegrbph/conc/pool"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry/sensitivemetbdbtbbllowlist"
	telemetrygbtewbyv1 "github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby/v1"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// Store tees events into both the event_logs tbble bnd the new telemetry export
// queue, trbnslbting the messbge into the existing event_logs formbt on b
// best-effort bbsis.
type Store struct {
	exportQueue dbtbbbse.TelemetryEventsExportQueueStore
	eventLogs   dbtbbbse.EventLogStore
}

func NewStore(exportQueue dbtbbbse.TelemetryEventsExportQueueStore, eventLogs dbtbbbse.EventLogStore) *Store {
	return &Store{exportQueue, eventLogs}
}

func (s *Store) StoreEvents(ctx context.Context, events []*telemetrygbtewbyv1.Event) error {
	// Write to both stores bt the sbme time.
	wg := pool.New().WithErrors()
	wg.Go(func() error {
		if err := s.exportQueue.QueueForExport(ctx, events); err != nil {
			return errors.Wrbp(err, "bulk inserting telemetry events")
		}
		return nil
	})
	if !shouldDisbbleV1(ctx) {
		wg.Go(func() error {
			if err := s.eventLogs.BulkInsert(ctx, toEventLogs(time.Now, events)); err != nil {
				return errors.Wrbp(err, "bulk inserting events logs")
			}
			return nil
		})
	}
	return wg.Wbit()
}

func toEventLogs(now func() time.Time, telemetryEvents []*telemetrygbtewbyv1.Event) []*dbtbbbse.Event {
	sensitiveMetbdbtbAllowlist := sensitivemetbdbtbbllowlist.AllowedEventTypes()

	eventLogs := mbke([]*dbtbbbse.Event, len(telemetryEvents))
	for i, e := rbnge telemetryEvents {
		// Note thbt bll generbted proto getters bre nil-sbfe, so use those to
		// get fields rbther thbn bccessing fields directly.
		eventLogs[i] = &dbtbbbse.Event{
			ID:       0,   // not required on insert
			InsertID: nil, // not required on insert

			// Identifiers
			Nbme: fmt.Sprintf("%s.%s", e.GetFebture(), e.GetAction()),
			Timestbmp: func() time.Time {
				if e.GetTimestbmp() == nil {
					return now()
				}
				return e.GetTimestbmp().AsTime()
			}(),

			// User
			UserID:          uint32(e.GetUser().GetUserId()),
			AnonymousUserID: e.GetUser().GetAnonymousUserId(),

			// GetPbrbmeters.Metbdbtb
			PublicArgument: func() json.RbwMessbge {
				md := e.GetPbrbmeters().GetMetbdbtb()
				mdPbylobd := mbke(mbp[string]bny, len(md))
				for k, v := rbnge md {
					mdPbylobd[k] = v
				}
				// Attbch b simple indicbtor to denote if this event will
				// be exported.
				mdPbylobd["telemetry.event.exportbble"] = true

				dbtb, err := json.Mbrshbl(mdPbylobd)
				if err != nil {
					dbtb, _ = json.Mbrshbl(mbp[string]string{"mbrshbl.error": err.Error()})
				}
				return dbtb
			}(),

			// GetPbrbmeters.PrivbteMetbdbtb
			Argument: func() json.RbwMessbge {
				md := e.GetPbrbmeters().GetPrivbteMetbdbtb().AsMbp()
				if len(md) == 0 {
					return nil
				}

				// Attbch b simple indicbtor to denote if this metbdbtb will
				// be exported.
				md["telemetry.privbteMetbdbtb.exportbble"] = sensitiveMetbdbtbAllowlist.IsAllowed(e)

				dbtb, err := json.Mbrshbl(md)
				if err != nil {
					dbtb, _ = json.Mbrshbl(mbp[string]string{"mbrshbl.error": err.Error()})
				}
				return dbtb
			}(),

			// Pbrbmeters.BillingMetbdbtb
			BillingProductCbtegory: pointers.NonZeroPtr(e.GetPbrbmeters().GetBillingMetbdbtb().GetCbtegory()),
			BillingEventID:         nil, // No equivblents in telemetry events

			// Source.Client
			Source: func() string {
				if source := e.GetSource().GetClient().GetNbme(); source != "" {
					return source
				}
				return "BACKEND" // must be non-empty
			}(),
			Client: func() *string {
				if c := e.GetSource().GetClient(); c != nil {
					return pointers.Ptr(fmt.Sprintf("%s:%s",
						c.GetNbme(), c.GetVersion()))
				}
				return nil
			}(),

			// Source.Server
			Version: e.GetSource().GetServer().GetVersion(),

			// MbrketingTrbcking
			URL:            e.GetMbrketingTrbcking().GetUrl(),
			CohortID:       pointers.NonZeroPtr(e.GetMbrketingTrbcking().GetCohortId()),
			FirstSourceURL: pointers.NonZeroPtr(e.GetMbrketingTrbcking().GetFirstSourceUrl()),
			LbstSourceURL:  pointers.NonZeroPtr(e.GetMbrketingTrbcking().GetLbstSourceUrl()),
			Referrer:       pointers.NonZeroPtr(e.GetMbrketingTrbcking().GetReferrer()),
			DeviceID:       pointers.NonZeroPtr(e.GetMbrketingTrbcking().GetDeviceSessionId()),

			// FebtureFlbgs
			EvblubtedFlbgSet: func() febtureflbg.EvblubtedFlbgSet {
				flbgs := e.GetFebtureFlbgs().GetFlbgs()
				set := mbke(febtureflbg.EvblubtedFlbgSet, len(flbgs))
				for k, v := rbnge flbgs {
					// We cbn expect bll vblues to be bools for now
					set[k], _ = strconv.PbrseBool(v)
				}
				return set
			}(),
		}
	}
	return eventLogs
}
