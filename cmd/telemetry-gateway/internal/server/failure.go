package server

import (
	"fmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/internal/events"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

func summarizeResults(results []events.PublishEventResult) (
	string,
	[]log.Field,
	[]*telemetrygatewayv1.RecordEventsResponse_RecordingSuccess,
	[]*telemetrygatewayv1.RecordEventsResponse_RecordingError,
) {
	// Collect details about the events that failed to submit.
	var (
		errFields = make([]log.Field, 0)
		succeeded = make([]*telemetrygatewayv1.RecordEventsResponse_RecordingSuccess, 0, len(results))
		failed    = make([]*telemetrygatewayv1.RecordEventsResponse_RecordingError, 0)
	)

	for i, result := range results {
		if result.PublishError != nil {
			failed = append(failed, &telemetrygatewayv1.RecordEventsResponse_RecordingError{
				EventId: result.EventID,
				Error:   result.PublishError.Error(),
			})
			errFields[i] = log.NamedError(fmt.Sprintf("error.%d", i), result.PublishError)
		} else {
			succeeded = append(succeeded, &telemetrygatewayv1.RecordEventsResponse_RecordingSuccess{
				EventId: result.EventID,
			})
		}
	}

	var message string
	switch {
	case len(failed) == len(results):
		message = "all events in batch failed to submit"
	case len(failed) > 0 && len(failed) < len(results):
		message = "some events in batch failed to submit"
	case len(failed) == 0:
		message = "all events in batch submitted successfully"
	}

	return message, errFields, succeeded, failed
}
