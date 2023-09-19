package telemetrygateway

import (
	"google.golang.org/grpc/status"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

func extractErrorDetails(err error) *telemetrygatewayv1.RecordEventsErrorDetails {
	st, ok := status.FromError(err)
	if ok {
		for _, detail := range st.Details() {
			switch d := detail.(type) {
			case *telemetrygatewayv1.RecordEventsErrorDetails:
				return d
			}
		}
	}
	return nil
}

func getSucceededEventsInError(submitted []*telemetrygatewayv1.Event, err error) []string {
	// If we can't get these error details, we assume the entire batch failed -
	// better to over-submit than under-submit. We can deduplicate on our end
	// using each event's unique persistent ID.
	details := extractErrorDetails(err)
	if details == nil {
		return nil // assume all failed
	}

	if len(submitted) == len(details.FailedEvents) {
		return nil // all failed
	}

	failedEvents := make(map[string]struct{})
	for _, ev := range details.FailedEvents {
		failedEvents[ev.EventId] = struct{}{}
	}
	succeededEvents := make([]string, 0, len(submitted))
	for _, ev := range submitted {
		if _, failed := failedEvents[ev.GetId()]; !failed {
			succeededEvents = append(succeededEvents, ev.Id)
		}
	}
	return succeededEvents
}
