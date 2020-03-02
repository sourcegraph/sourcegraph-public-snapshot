package eventlogger

import (
	"encoding/json"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/internal/api"

	"golang.org/x/net/context"
)

// TelemetryRequest represents a request to log telemetry.
type TelemetryRequest struct {
	UserID    int32
	EventName string
	Argument  json.RawMessage
}

// LogEvent sends a payload representing an event to the api/telemetry endpoint.
//
// This method should be invoked after the frontend service has started. It is
// safe to not do so (it will just log an error), but logging the actual event
// will fail otherwise. Consider using e.g. api.InternalClient.RetryPingUntilAvailable
// to wait for the frontend to start.
//
// Note: This does not block since it creates a new goroutine.
func LogEvent(userID int32, name string, argument json.RawMessage) {
	go func() {
		err := logEvent(userID, name, argument)
		if err != nil {
			log15.Warn("eventlogger.LogEvent failed", "event", name, "error", err)
		}
	}()
}

// logEvent sends a payload representing some user event to the InternalClient telemetry API
func logEvent(userID int32, name string, argument json.RawMessage) error {
	reqBody := &TelemetryRequest{
		UserID:    userID,
		EventName: name,
		Argument:  argument,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	return api.InternalClient.LogTelemetry(ctx, reqBody)
}
