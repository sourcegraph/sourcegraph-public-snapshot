package server

import (
	"errors"
	"strconv"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/internal/events"
)

func TestSummarizeFailedEvents(t *testing.T) {
	t.Run("all failed", func(t *testing.T) {
		submitted := 5
		failed := make([]events.PublishEventResult, submitted)
		for i := range failed {
			failed[i].EventID = "id_" + strconv.Itoa(i)
			failed[i].PublishError = errors.New("failed")
		}

		message, logFields, details := summarizeFailedEvents(submitted, failed)
		assert.Len(t, logFields, submitted)
		assert.Len(t, details.FailedEvents, submitted)
		autogold.Expect("all events in batch failed to submit").Equal(t, message)
	})

	t.Run("some failed", func(t *testing.T) {
		submitted := 5
		failed := []events.PublishEventResult{{
			EventID:      "asdf",
			PublishError: errors.New("oh no"),
		}}

		message, logFields, details := summarizeFailedEvents(submitted, failed)
		assert.Len(t, logFields, 1)
		assert.Len(t, details.FailedEvents, 1)
		autogold.Expect("some events in batch failed to submit").Equal(t, message)
	})
}
