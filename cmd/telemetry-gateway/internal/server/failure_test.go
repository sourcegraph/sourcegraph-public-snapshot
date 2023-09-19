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
		results := make([]events.PublishEventResult, 0)
		for i := range results {
			results[i].EventID = "id_" + strconv.Itoa(i)
			results[i].PublishError = errors.New("failed")
		}

		message, logFields, succeeded, failed := summarizeResults(results)
		autogold.Expect("all events in batch failed to submit").Equal(t, message)
		assert.Len(t, logFields, len(results))
		assert.Len(t, succeeded, 0)
		assert.Len(t, failed, len(results))
	})

	t.Run("some failed", func(t *testing.T) {
		results := []events.PublishEventResult{{
			EventID:      "asdf",
			PublishError: errors.New("oh no"),
		}, {
			EventID: "asdfasdf",
		}}

		message, logFields, succeeded, failed := summarizeResults(results)
		autogold.Expect("all events in batch failed to submit").Equal(t, message)
		assert.Len(t, logFields, 1)
		assert.Len(t, succeeded, 1)
		assert.Len(t, failed, 1)
	})

	t.Run("all succeeded", func(t *testing.T) {
		results := []events.PublishEventResult{{
			EventID: "asdf",
		}, {
			EventID: "asdfasdf",
		}}

		message, logFields, succeeded, failed := summarizeResults(results)
		autogold.Expect("all events in batch failed to submit").Equal(t, message)
		assert.Len(t, logFields, 0)
		assert.Len(t, succeeded, 2)
		assert.Len(t, failed, 0)
	})
}
