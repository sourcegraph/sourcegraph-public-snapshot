package telemetrytest

import (
	"context"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

func TestFakeTelemetryEventsExportQueueStore(t *testing.T) {
	s := NewMockEventsExportQueueStore()
	err := s.QueueForExport(
		context.Background(),
		[]*telemetrygatewayv1.Event{
			{
				Id:      "asdfasdf",
				Feature: "Feature",
				Action:  "Action",
			},
		})
	require.NoError(t, err)
	require.Len(t, s.events, 1)
	require.Equal(t, "asdfasdf", s.events[0].Id)
	autogold.Expect([]string{"Feature - Action"}).Equal(t, s.GetMockQueuedEvents().Summary())
}
