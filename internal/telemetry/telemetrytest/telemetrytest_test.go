package telemetrytest

import (
	"context"
	"errors"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
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

func TestBestEffort(t *testing.T) {
	t.Run("no panic on nil parameters", func(t *testing.T) {
		erroringStore := NewMockEventsStore()
		erroringStore.StoreEventsFunc.SetDefaultReturn(errors.New("BOOM"))

		r := telemetry.NewBestEffortEventRecorder(
			logtest.Scoped(t),
			telemetry.NewEventRecorder(erroringStore),
		)

		r.Record(context.Background(), telemetry.FeatureExample, telemetry.ActionAttempted, nil)
	})
}
