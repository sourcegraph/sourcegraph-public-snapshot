package telemetry_test

import (
	"context"
	"testing"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/teestore"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetrytest"
)

func TestRecorder(t *testing.T) {
	store := telemetrytest.NewMockEventsStore()
	recorder := telemetry.NewEventRecorder(store)

	err := recorder.Record(context.Background(), "Feature", "Action", nil)
	require.NoError(t, err)

	// stored once
	require.Len(t, store.StoreEventsFunc.History(), 1)
	// called with 1 event
	require.Len(t, store.StoreEventsFunc.History()[0].Arg1, 1)
	// stored event has 1 event
	require.Equal(t, "Feature", store.StoreEventsFunc.History()[0].Arg1[0].Feature)
}

func TestRecorderEndToEnd(t *testing.T) {
	var userID int32 = 123
	ctx := actor.WithActor(context.Background(), actor.FromMockUser(userID))

	logger := logtest.ScopedWith(t, logtest.LoggerOptions{
		Level: log.LevelDebug,
	})
	db := database.NewDB(logger, dbtest.NewDB(t))

	// Set a mock mode to ensure we are testing enabled exports
	exportStore := db.TelemetryEventsExportQueue()
	exportStore.(database.MockExportModeSetterTelemetryEventsExportQueueStore).
		SetMockExportMode(licensing.TelemetryEventsExportAll)

	recorder := telemetry.NewEventRecorder(teestore.NewStore(exportStore, db.EventLogs()))

	wantEvents := 3
	t.Run("Record and BatchRecord", func(t *testing.T) {
		assert.NoError(t, recorder.Record(ctx,
			"Test", "Action1",
			&telemetry.EventParameters{
				Metadata: telemetry.EventMetadata{
					"metadata": 1,
				},
				PrivateMetadata: map[string]any{
					"private": "sensitive",
				},
			}))
		assert.NoError(t, recorder.BatchRecord(ctx,
			telemetry.Event{
				Feature: "Test",
				Action:  "Action2",
			},
			telemetry.Event{
				Feature: "Test",
				Action:  "Action3",
			}))
	})

	t.Run("tee to EventLogs", func(t *testing.T) {
		eventLogs, err := db.EventLogs().ListAll(ctx, database.EventLogsListOptions{UserID: userID})
		require.NoError(t, err)
		assert.Len(t, eventLogs, wantEvents)
	})

	t.Run("tee to TelemetryEvents", func(t *testing.T) {
		telemetryEvents, err := db.TelemetryEventsExportQueue().ListForExport(ctx, 999)
		require.NoError(t, err)
		assert.Len(t, telemetryEvents, wantEvents)
	})

	t.Run("record without v1", func(t *testing.T) {
		ctx := teestore.WithoutV1(ctx)
		assert.NoError(t, recorder.Record(ctx, "Test", "Action1", &telemetry.EventParameters{}))

		telemetryEvents, err := db.TelemetryEventsExportQueue().ListForExport(ctx, 999)
		require.NoError(t, err)
		assert.Len(t, telemetryEvents, wantEvents+1)

		eventLogs, err := db.EventLogs().ListAll(ctx, database.EventLogsListOptions{UserID: userID})
		require.NoError(t, err)
		assert.Len(t, eventLogs, wantEvents) // v1 unchanged
	})
}
