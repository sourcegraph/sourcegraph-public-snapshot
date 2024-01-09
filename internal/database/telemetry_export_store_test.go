package database

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

func TestTelemetryEventsExportQueue_QueueForExport(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mode         licensing.TelemetryEventsExportMode
		events       []*telemetrygatewayv1.Event
		expectQueued autogold.Value
	}{{
		name: "disabled",
		mode: licensing.TelemetryEventsExportDisabled,
		events: []*telemetrygatewayv1.Event{{
			Id:      "1",
			Feature: "foobar",
			Action:  "baz",
		}},
		expectQueued: autogold.Expect([]string{}),
	}, {
		name: "enabled: export all",
		mode: licensing.TelemetryEventsExportAll,
		events: []*telemetrygatewayv1.Event{{
			Id:      "1",
			Feature: "foobar",
			Action:  "baz",
		}},
		expectQueued: autogold.Expect([]string{"1"}),
	}, {
		name: "cody-only: drop some",
		mode: licensing.TelemetryEventsExportCodyOnly,
		events: []*telemetrygatewayv1.Event{{
			Id:      "1",
			Feature: "foobar",
			Action:  "baz",
		}, {
			Id:      "2",
			Feature: "cody.foobar",
			Action:  "baz",
		}, {
			Id:      "3",
			Feature: "cody",
			Action:  "bar",
		}},
		expectQueued: autogold.Expect([]string{"2", "3"}),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			logger := logtest.Scoped(t)
			db := NewDB(logger, dbtest.NewDB(logger, t))

			store := TelemetryEventsExportQueueWith(logger, db)

			// Set a mock mode to test enabled exports
			store.(MockExportModeSetterTelemetryEventsExportQueueStore).
				SetMockExportMode(tc.mode)

			require.NoError(t, store.QueueForExport(context.Background(), tc.events))
			queued, err := store.ListForExport(ctx, 999)
			require.NoError(t, err)
			var ids []string
			for _, e := range queued {
				ids = append(ids, e.Id)
			}
			tc.expectQueued.Equal(t, ids)
		})
	}
}

func TestTelemetryEventsExportQueueLifecycle(t *testing.T) {
	ctx := context.Background()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	store := TelemetryEventsExportQueueWith(logger, db)

	// Set a mock mode to test enabled exports
	store.(MockExportModeSetterTelemetryEventsExportQueueStore).
		SetMockExportMode(licensing.TelemetryEventsExportAll)

	events := []*telemetrygatewayv1.Event{{
		Id:        "1",
		Feature:   "Feature",
		Action:    "View",
		Timestamp: timestamppb.New(time.Date(2022, 11, 3, 1, 0, 0, 0, time.UTC)),
		Parameters: &telemetrygatewayv1.EventParameters{
			Metadata: map[string]float64{"public": 1},
		},
	}, {
		Id:        "2",
		Feature:   "Feature",
		Action:    "Click",
		Timestamp: timestamppb.New(time.Date(2022, 11, 3, 2, 0, 0, 0, time.UTC)),
		Parameters: &telemetrygatewayv1.EventParameters{
			PrivateMetadata: &structpb.Struct{
				Fields: map[string]*structpb.Value{"sensitive": structpb.NewStringValue("sensitive")},
			},
		},
	}, {
		Id:        "3",
		Feature:   "Feature",
		Action:    "Show",
		Timestamp: timestamppb.New(time.Date(2022, 11, 3, 3, 0, 0, 0, time.UTC)),
	}}
	eventsToExport := []string{"1", "2"}

	t.Run("feature flag off", func(t *testing.T) {
		// Context with FF disabled.
		ff := featureflag.NewMemoryStore(
			nil, nil, map[string]bool{FeatureFlagTelemetryExport: false})
		ctx := featureflag.WithFlags(context.Background(), ff)

		require.NoError(t, store.QueueForExport(ctx, events))
		export, err := store.ListForExport(ctx, 100)
		require.NoError(t, err)
		assert.Len(t, export, 0)
	})

	t.Run("QueueForExport", func(t *testing.T) {
		require.NoError(t, store.QueueForExport(ctx, events))
	})

	t.Run("CountUnexported", func(t *testing.T) {
		count, err := store.CountUnexported(ctx)
		require.NoError(t, err)
		require.Equal(t, count, int64(3))
	})

	t.Run("ListForExport", func(t *testing.T) {
		limit := len(events) - 1
		export, err := store.ListForExport(ctx, limit)
		require.NoError(t, err)
		assert.Len(t, export, limit)

		// Check we got the exact event IDs we want to export
		var gotIDs []string
		for _, e := range export {
			gotIDs = append(gotIDs, e.GetId())
		}
		assert.Equal(t, eventsToExport, gotIDs)

		// Check integrity of first item
		original, err := proto.Marshal(events[0])
		require.NoError(t, err)
		got, err := proto.Marshal(export[0])
		require.NoError(t, err)
		assert.Equal(t, string(original), string(got))

		// Check second item's private meta is stripped
		assert.NotNil(t, events[1].Parameters.PrivateMetadata) // original
		assert.Nil(t, export[1].Parameters.PrivateMetadata)    // got
	})

	t.Run("before export: DeleteExported", func(t *testing.T) {
		affected, err := store.DeletedExported(ctx, time.Now())
		require.NoError(t, err)
		assert.Zero(t, affected)
	})

	t.Run("MarkAsExported", func(t *testing.T) {
		require.NoError(t, store.MarkAsExported(ctx, eventsToExport))
	})

	t.Run("after export: CountRecentlyExported", func(t *testing.T) {
		export, err := store.CountRecentlyExported(ctx)
		require.NoError(t, err)
		assert.Equal(t, export, int64(2))
	})

	t.Run("after export: ListRecentlyExported", func(t *testing.T) {
		exported, err := store.ListRecentlyExported(ctx, 1, nil)
		require.NoError(t, err)
		require.Len(t, exported, 1)

		// Most recent first
		assert.Equal(t, "2", exported[0].ID)
		assert.Equal(t, "2", exported[0].Payload.GetId())
		assert.NotZero(t, exported[0].ExportedAt)
		assert.NotZero(t, exported[0].Timestamp)

		// Next "page"
		cursor := exported[0].Timestamp
		exported, err = store.ListRecentlyExported(ctx, 1, &cursor)
		require.NoError(t, err)
		require.Len(t, exported, 1)

		assert.Equal(t, "1", exported[0].ID)
		assert.Equal(t, "1", exported[0].Payload.GetId())
		assert.NotZero(t, exported[0].ExportedAt)
		assert.NotZero(t, exported[0].Timestamp)
	})

	t.Run("after export: QueueForExport", func(t *testing.T) {
		export, err := store.ListForExport(ctx, len(events))
		require.NoError(t, err)
		assert.Len(t, export, 1)
		// ID is exactly as expected
		assert.Equal(t, "3", export[0].GetId())
	})

	t.Run("after export: DeleteExported", func(t *testing.T) {
		affected, err := store.DeletedExported(ctx, time.Now())
		require.NoError(t, err)
		assert.Equal(t, int(affected), len(eventsToExport))
	})
}
