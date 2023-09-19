package server

import (
	"fmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/internal/events"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

func summarizeFailedEvents(submittedEvents int, failedEvents []events.PublishEventResult) (string, []log.Field, *telemetrygatewayv1.RecordEventsErrorDetails) {
	// Generate failure message
	var message string
	if len(failedEvents) == submittedEvents {
		message = "all events in batch failed to submit"
	} else {
		message = "some events in batch failed to submit"
	}

	// Collect details about the events that failed to submit.
	failedEventsDetails := make([]*telemetrygatewayv1.RecordEventsErrorDetails_EventError, len(failedEvents))
	errFields := make([]log.Field, len(failedEvents))
	for i, failure := range failedEvents {
		failedEventsDetails[i] = &telemetrygatewayv1.RecordEventsErrorDetails_EventError{
			EventId: failure.EventID,
			Error:   failure.PublishError.Error(),
		}
		errFields[i] = log.NamedError(fmt.Sprintf("error.%d", i), failure.PublishError)
	}

	return message, errFields, &telemetrygatewayv1.RecordEventsErrorDetails{
		FailedEvents: failedEventsDetails,
	}
}
