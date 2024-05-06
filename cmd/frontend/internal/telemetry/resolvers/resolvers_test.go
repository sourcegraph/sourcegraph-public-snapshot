package resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/fakedb"
	"github.com/sourcegraph/sourcegraph/internal/types"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
)

func TestExportedEvents(t *testing.T) {
	exportedEvents := []*telemetrygatewayv1.Event{{
		Id:      "1",
		Feature: "feature",
		Action:  "view",
		// Most recent event
		Timestamp: timestamppb.New(time.Date(2022, 11, 3, 1, 0, 0, 0, time.UTC)),
	}, {
		Id:        "2",
		Feature:   "feature",
		Action:    "click",
		Timestamp: timestamppb.New(time.Date(2022, 11, 3, 2, 0, 0, 0, time.UTC)),
	}, {
		Id:        "3",
		Feature:   "feature",
		Action:    "show",
		Timestamp: timestamppb.New(time.Date(2022, 11, 3, 3, 0, 0, 0, time.UTC)),
	}, {
		Id:        "4",
		Feature:   "feature",
		Action:    "dance",
		Timestamp: timestamppb.New(time.Date(2022, 11, 3, 4, 0, 0, 0, time.UTC)),
	}}
	var exportedEventIDs []string
	for _, e := range exportedEvents {
		exportedEventIDs = append(exportedEventIDs, e.GetId())
	}

	// Use a real DB for TelemetryEventsExportQueue for simplicity and E2E testing
	exportQueueStore := database.TelemetryEventsExportQueueWith(logtest.Scoped(t), database.NewDB(logtest.Scoped(t), dbtest.NewDB(t)))
	require.NoError(t, exportQueueStore.QueueForExport(context.Background(), exportedEvents))
	require.NoError(t, exportQueueStore.MarkAsExported(context.Background(), exportedEventIDs))

	for _, tc := range []struct {
		name string

		actor types.User
		args  graphqlbackend.ExportedEventsArgs

		wantExportedEventsError autogold.Value
		wantCursorTime          autogold.Value
		wantNodes               int
	}{{
		name:                    "not authorized",
		wantExportedEventsError: autogold.Expect("must be site admin"),
	}, {
		name:  "list first half",
		actor: types.User{SiteAdmin: true},
		args: graphqlbackend.ExportedEventsArgs{
			First: int32(len(exportedEvents) / 2),
		},
		wantCursorTime: autogold.Expect("2022-11-03 03:00:00 +0000 UTC"),
		wantNodes:      len(exportedEvents) / 2,
	}, {
		name:  "list last half",
		actor: types.User{SiteAdmin: true},
		args: graphqlbackend.ExportedEventsArgs{
			After: encodeExportedEventsCursor(
				exportedEvents[len(exportedEvents)/2].Timestamp.AsTime(),
			).EndCursor(),
		},
		wantNodes: len(exportedEvents) / 2,
	}, {
		name:      "list all",
		actor:     types.User{SiteAdmin: true},
		args:      graphqlbackend.ExportedEventsArgs{},
		wantNodes: 4,
	}, {
		name:  "no more to list",
		actor: types.User{SiteAdmin: true},
		args: graphqlbackend.ExportedEventsArgs{
			// We sort oldest to newest, so after the newest timestamp, there
			// are no more to list.
			After: encodeExportedEventsCursor(exportedEvents[0].Timestamp.AsTime()).EndCursor(),
		},
		wantNodes: 0,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			// Set up user in fake DB
			fdb := fakedb.New()
			db := dbmocks.NewMockDB()
			fdb.Wire(db)
			userID := fdb.AddUser(tc.actor)
			ctx := actor.WithActor(context.Background(), actor.FromUser(userID))

			// Hook the shared test database into the mock DB
			db.TelemetryEventsExportQueueFunc.SetDefaultReturn(exportQueueStore)

			// List exported events
			events, err := (&Resolver{
				logger: logtest.Scoped(t),
				db:     db,
			}).ExportedEvents(ctx, &tc.args)
			if tc.wantExportedEventsError != nil {
				require.Error(t, err)
				tc.wantExportedEventsError.Equal(t, err.Error())
				return
			}

			t.Run("TotalCount", func(t *testing.T) {
				count, err := events.TotalCount()
				require.NoError(t, err)
				assert.Equal(t, len(exportedEvents), int(count))
			})

			t.Run("PageInfo", func(t *testing.T) {
				page := events.PageInfo()
				assert.Equal(t, tc.wantCursorTime != nil, page.HasNextPage(),
					"unexpectedly has next page")
				if tc.wantCursorTime != nil {
					require.NotNil(t, page.EndCursor())
					c, err := decodeExportedEventsCursor(*page.EndCursor())
					require.NoError(t, err)
					tc.wantCursorTime.Equal(t, c.String())
				}
			})

			t.Run("Nodes", func(t *testing.T) {
				nodes := events.Nodes()
				assert.Equal(t, tc.wantNodes, len(nodes))
			})
		})
	}
}
