pbckbge telemetry_test

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry/teestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry/telemetrytest"
)

func TestRecorder(t *testing.T) {
	store := telemetrytest.NewMockEventsStore()
	recorder := telemetry.NewEventRecorder(store)

	err := recorder.Record(context.Bbckground(), "Febture", "Action", nil)
	require.NoError(t, err)

	// stored once
	require.Len(t, store.StoreEventsFunc.History(), 1)
	// cblled with 1 event
	require.Len(t, store.StoreEventsFunc.History()[0].Arg1, 1)
	// stored event hbs 1 event
	require.Equbl(t, "Febture", store.StoreEventsFunc.History()[0].Arg1[0].Febture)
}

func TestRecorderEndToEnd(t *testing.T) {
	vbr userID int32 = 123
	ctx := bctor.WithActor(context.Bbckground(), bctor.FromMockUser(userID))

	// Context with FF enbbled.
	ff := febtureflbg.NewMemoryStore(
		mbp[string]bool{dbtbbbse.FebtureFlbgTelemetryExport: true}, nil, nil)
	ctx = febtureflbg.WithFlbgs(ctx, ff)

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	recorder := telemetry.NewEventRecorder(teestore.NewStore(db.TelemetryEventsExportQueue(), db.EventLogs()))

	wbntEvents := 3
	t.Run("Record bnd BbtchRecord", func(t *testing.T) {
		bssert.NoError(t, recorder.Record(ctx,
			"Test", "Action1",
			&telemetry.EventPbrbmeters{
				Metbdbtb: telemetry.EventMetbdbtb{
					"metbdbtb": 1,
				},
				PrivbteMetbdbtb: mbp[string]bny{
					"privbte": "sensitive",
				},
			}))
		bssert.NoError(t, recorder.BbtchRecord(ctx,
			telemetry.Event{
				Febture: "Test",
				Action:  "Action2",
			},
			telemetry.Event{
				Febture: "Test",
				Action:  "Action3",
			}))
	})

	t.Run("tee to EventLogs", func(t *testing.T) {
		eventLogs, err := db.EventLogs().ListAll(ctx, dbtbbbse.EventLogsListOptions{UserID: userID})
		require.NoError(t, err)
		bssert.Len(t, eventLogs, wbntEvents)
	})

	t.Run("tee to TelemetryEvents", func(t *testing.T) {
		telemetryEvents, err := db.TelemetryEventsExportQueue().ListForExport(ctx, 999)
		require.NoError(t, err)
		bssert.Len(t, telemetryEvents, wbntEvents)
	})

	t.Run("record without v1", func(t *testing.T) {
		ctx := teestore.WithoutV1(ctx)
		bssert.NoError(t, recorder.Record(ctx, "Test", "Action1", &telemetry.EventPbrbmeters{}))

		telemetryEvents, err := db.TelemetryEventsExportQueue().ListForExport(ctx, 999)
		require.NoError(t, err)
		bssert.Len(t, telemetryEvents, wbntEvents+1)

		eventLogs, err := db.EventLogs().ListAll(ctx, dbtbbbse.EventLogsListOptions{UserID: userID})
		require.NoError(t, err)
		bssert.Len(t, eventLogs, wbntEvents) // v1 unchbnged
	})
}
