package server

import (
	"fmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/internal/events"
)

func summarizeResults(results []events.PublishEventResult) (
	string,
	[]log.Field,
	[]string,
	[]events.PublishEventResult,
) {
	// Collect details about the events that failed to submit.
	var (
		errFields = make([]log.Field, 0)
		succeeded = make([]string, 0, len(results))
		failed    = make([]events.PublishEventResult, 0)
	)

	for i, result := range results {
		if result.PublishError != nil {
			failed = append(failed, result)
			errFields = append(errFields, log.NamedError(fmt.Sprintf("error.%d", i), result.PublishError))
		} else {
			succeeded = append(succeeded, result.EventID)
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
