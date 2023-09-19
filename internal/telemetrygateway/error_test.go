package telemetrygateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

func TestGetSucceededEventsInError(t *testing.T) {
	t.Run("no details", func(t *testing.T) {
		assert.Empty(t, getSucceededEventsInError(nil,
			status.Error(codes.Internal, "oh no!")))
	})

	t.Run("with details, all failed", func(t *testing.T) {
		status, err := status.New(codes.Internal, "oh no!").
			WithDetails(&telemetrygatewayv1.RecordEventsErrorDetails{
				FailedEvents: []*telemetrygatewayv1.RecordEventsErrorDetails_EventError{
					{EventId: "event1", Error: "D:"},
					{EventId: "event2", Error: ":("},
				},
			})
		require.NoError(t, err)
		assert.Empty(t, getSucceededEventsInError(
			[]*telemetrygatewayv1.Event{{
				Id: "event1",
			}, {
				Id: "event2",
			}},
			status.Err()))
	})

	t.Run("with details, partially failed", func(t *testing.T) {
		status, err := status.New(codes.Internal, "oh no!").
			WithDetails(&telemetrygatewayv1.RecordEventsErrorDetails{
				FailedEvents: []*telemetrygatewayv1.RecordEventsErrorDetails_EventError{
					{EventId: "event1", Error: "D:"},
				},
			})
		require.NoError(t, err)
		succeeded := getSucceededEventsInError(
			[]*telemetrygatewayv1.Event{{
				Id: "event1",
			}, {
				Id: "event2",
			}},
			status.Err())
		require.Len(t, succeeded, 1)
		assert.Equal(t, succeeded[0], "event2")
	})
}
