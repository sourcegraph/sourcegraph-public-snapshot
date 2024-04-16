package server

import (
	"testing"

	"github.com/stretchr/testify/assert"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

func TestMigrateEvents(t *testing.T) {
	t.Run("legacy metadata format", func(t *testing.T) {
		events := []*telemetrygatewayv1.Event{
			{
				Parameters: &telemetrygatewayv1.EventParameters{
					LegacyMetadata: map[string]int64{
						"foo": 123,
					},
				},
			},
			{
				Parameters: &telemetrygatewayv1.EventParameters{
					LegacyMetadata: map[string]int64{
						"foo": 123,
					},
					Metadata: map[string]float64{
						"bar": 456,
					},
				},
			},
		}
		migrateEvents(events)

		// first event
		assert.Equal(t, float64(123), events[0].Parameters.Metadata["foo"])
		assert.Nil(t, events[0].Parameters.LegacyMetadata)

		// second event
		assert.Equal(t, float64(123), events[1].Parameters.Metadata["foo"])
		assert.Equal(t, float64(456), events[1].Parameters.Metadata["bar"])
		assert.Nil(t, events[1].Parameters.LegacyMetadata)
	})
}
