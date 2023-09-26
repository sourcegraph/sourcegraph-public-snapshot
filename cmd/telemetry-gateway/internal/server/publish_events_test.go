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

		summary := summarizePublishEventsResults(results)
		autogold.Expect("all events in batch failed to submit").Equal(t, summary.message)
		autogold.Expect("complete_failure").Equal(t, summary.result)
		assert.Len(t, summary.errorFields, len(results))
		assert.Len(t, summary.succeededEvents, 0)
		assert.Len(t, summary.failedEvents, len(results))
	})

	t.Run("some failed", func(t *testing.T) {
		results := []events.PublishEventResult{{
			EventID:      "asdf",
			PublishError: errors.New("oh no"),
		}, {
			EventID: "asdfasdf",
		}}

		summary := summarizePublishEventsResults(results)
		autogold.Expect("some events in batch failed to submit").Equal(t, summary.message)
		autogold.Expect("partial_failure").Equal(t, summary.result)
		assert.Len(t, summary.errorFields, 1)
		assert.Len(t, summary.succeededEvents, 1)
		assert.Len(t, summary.failedEvents, 1)
	})

	t.Run("all succeeded", func(t *testing.T) {
		results := []events.PublishEventResult{{
			EventID: "asdf",
		}, {
			EventID: "asdfasdf",
		}}

		summary := summarizePublishEventsResults(results)
		autogold.Expect("all events in batch submitted successfully").Equal(t, summary.message)
		autogold.Expect("success").Equal(t, summary.result)
		assert.Len(t, summary.errorFields, 0)
		assert.Len(t, summary.succeededEvents, 2)
		assert.Len(t, summary.failedEvents, 0)
	})
}
