package database

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

func TestTelemetryEventsExportQueueLifecycle(t *testing.T) {
	// Context with FF enabled.
	ff := featureflag.NewMemoryStore(
		nil, nil, map[string]bool{FeatureFlagTelemetryExport: true})
	ctx := featureflag.WithFlags(context.Background(), ff)

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	store := TelemetryEventsExportQueueWith(logger, db)

	events := []*telemetrygatewayv1.Event{{
		Id:        "1",
		Feature:   "Feature",
		Action:    "View",
		Timestamp: timestamppb.New(time.Date(2022, 11, 3, 1, 0, 0, 0, time.UTC)),
		Parameters: &telemetrygatewayv1.EventParameters{
			Metadata: map[string]int64{"public": 1},
			PrivateMetadata: &structpb.Struct{
				Fields: map[string]*structpb.Value{"sensitive": structpb.NewStringValue("sensitive")},
			},
		},
	}, {
		Id:        "2",
		Feature:   "Feature",
		Action:    "Click",
		Timestamp: timestamppb.New(time.Date(2022, 11, 3, 2, 0, 0, 0, time.UTC)),
	}, {
		Id:        "3",
		Feature:   "Feature",
		Action:    "Show",
		Timestamp: timestamppb.New(time.Date(2022, 11, 3, 3, 0, 0, 0, time.UTC)),
	}}
	eventsToExport := []string{"1", "2"}

	t.Run("feature flag off", func(t *testing.T) {
		require.NoError(t, store.QueueForExport(context.Background(), events))
		export, err := store.ListForExport(ctx, 100)
		require.NoError(t, err)
		assert.Len(t, export, 0)
	})

	t.Run("QueueForExport", func(t *testing.T) {
		require.NoError(t, store.QueueForExport(ctx, events))
	})

	t.Run("ListForExport", func(t *testing.T) {
		limit := len(events) - 1
		export, err := store.ListForExport(ctx, limit)
		require.NoError(t, err)
		assert.Len(t, export, limit)

		// Check integrity of first item
		original, err := proto.Marshal(events[0])
		require.NoError(t, err)
		got, err := proto.Marshal(export[0])
		require.NoError(t, err)
		require.Equal(t, string(original), string(got))
	})

	t.Run("before export: DeleteExported", func(t *testing.T) {
		affected, err := store.DeletedExported(ctx, time.Now())
		require.NoError(t, err)
		assert.Zero(t, affected)
	})

	t.Run("MarkAsExported", func(t *testing.T) {
		require.NoError(t, store.MarkAsExported(ctx, eventsToExport))
	})

	t.Run("after export: QueueForExport", func(t *testing.T) {
		export, err := store.ListForExport(ctx, len(events))
		require.NoError(t, err)
		assert.Len(t, export, len(events)-len(eventsToExport))
	})

	t.Run("after export: DeleteExported", func(t *testing.T) {
		affected, err := store.DeletedExported(ctx, time.Now())
		require.NoError(t, err)
		assert.Equal(t, int(affected), len(eventsToExport))
	})
}
