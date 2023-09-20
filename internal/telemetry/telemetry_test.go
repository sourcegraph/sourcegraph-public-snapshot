package telemetry_test

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/teestore"
)

func TestRecorderEndToEnd(t *testing.T) {
	var userID int32 = 123
	ctx := actor.WithActor(context.Background(), actor.FromMockUser(userID))

	// Context with FF enabled.
	ff := featureflag.NewMemoryStore(
		map[string]bool{database.FeatureFlagTelemetryExport: true}, nil, nil)
	ctx = featureflag.WithFlags(ctx, ff)

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	recorder := telemetry.NewEventRecorder(teestore.NewStore(db.TelemetryEventsExportQueue(), db.EventLogs()))

	wantEvents := 3
	t.Run("Record and BatchRecord", func(t *testing.T) {
		assert.NoError(t, recorder.Record(ctx,
			"Test", "Action1",
			telemetry.EventParameters{
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
		assert.NoError(t, recorder.Record(ctx, "Test", "Action1", telemetry.EventParameters{}))

		telemetryEvents, err := db.TelemetryEventsExportQueue().ListForExport(ctx, 999)
		require.NoError(t, err)
		assert.Len(t, telemetryEvents, wantEvents+1)

		eventLogs, err := db.EventLogs().ListAll(ctx, database.EventLogsListOptions{UserID: userID})
		require.NoError(t, err)
		assert.Len(t, eventLogs, wantEvents) // v1 unchanged
	})
}
