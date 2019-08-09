package eventlogger

import (
	"encoding/json"
	"fmt"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/version"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

var backendEventsTrackingSiteID = "SourcegraphBackend"

// defaultLogger is a singleton for event logging from the backend
var defaultLogger = new()

// LogEvent sends a payload representing an event to the api/telemetry endpoint. This
// endpoint only functions on Sourcegraph.com, not on self-hosted instances.
//
// This method should be invoked after the frontend service has started. It is
// safe to not do so (it will just log an error), but logging the actual event
// will fail otherwise. Consider using e.g. api.InternalClient.RetryPingUntilAvailable
// to wait for the frontend to start.
//
// Note: This does not block since it creates a new goroutine.
func LogEvent(userID int32, userEmail string, eventLabel string, eventProperties json.RawMessage) {
	go func() {
		err := defaultLogger.logEvent(userID, userEmail, eventLabel, eventProperties)
		if err != nil {
			log15.Warn("eventlogger.LogEvent failed", "event", eventLabel, "error", err)
		}
	}()
}

// eventLogger represents a connection to a remote URL for sending
// event logs, with environment and user context
type eventLogger struct {
	env string
}

// new returns a new EventLogger client
func new() *eventLogger {
	environment := "production"
	if version.Version() == "dev" {
		environment = "development"
	}
	return &eventLogger{
		env: environment,
	}
}

// newPayload generates a new Payload struct for a provided event
// in the context of the EventLogger client
func (logger *eventLogger) newPayload(userEmail string, event *Event) *Payload {
	userInfo := &UserInfo{
		DomainUserID: "sourcegraph-backend-anonymous",
	}
	if userEmail != "" {
		userInfo = &UserInfo{
			DomainUserID: uuid.New().String(),
			Email:        userEmail,
		}
	}
	return &Payload{
		DeviceInfo: &DeviceInfo{
			Platform:         "Web",
			TrackerNamespace: "sg",
		},
		Events: []*Event{
			event,
		},
		Header: &Header{
			SiteID: backendEventsTrackingSiteID,
			Env:    logger.env,
		},
		BatchInfo: &BatchInfo{
			BatchID:     uuid.New().String(),
			TotalEvents: 1,
			ServerTime:  fmt.Sprintf("%d", time.Now().UTC().Unix()*1000),
		},
		UserInfo: userInfo,
	}
}

// logEvent sends a payload representing some user event to the InternalClient telemetry API
func (logger *eventLogger) logEvent(userID int32, userEmail string, eventLabel string, eventProperties json.RawMessage) error {
	event := &Event{
		Type:            eventLabel,
		EventID:         uuid.New().String(),
		ClientTimestamp: time.Now().UTC().Unix() * 1000,
		Context: &Context{
			EventLabel: eventLabel,
			Backend:    eventProperties,
		},
	}
	payload := logger.newPayload(userEmail, event)
	reqBody := &TelemetryRequest{
		UserID:     userID,
		EventLabel: eventLabel,
		Payload:    payload,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	return api.InternalClient.LogTelemetry(ctx, logger.env, reqBody)
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_777(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
