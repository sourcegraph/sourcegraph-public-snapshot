package eventlogger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var backendEventsTrackingAppID = "SourcegraphBackend"
var defaultRemoteURL = "https://sourcegraph-logging.telligentdata.com/log/v1/"

// EventLogger is a singleton for event logging from the backend
// TODO(Dan): build this to handle custom end-points for on-prem deployments
var EventLogger = new(nil)

// eventLogger represents a connection to a remote URL for sending
// event logs, with environment and user context
type eventLogger struct {
	env, url string
	ctx      context.Context
}

type eventLoggerOptions struct {
	remoteURL string
}

// new returns a new EventLogger client
func new(opt *eventLoggerOptions) *eventLogger {
	environment := "production"
	if env.Version == "dev" {
		environment = "development"
	}
	url := defaultRemoteURL + environment
	if opt != nil && opt.remoteURL != "" {
		url = opt.remoteURL
	}
	return &eventLogger{
		env: environment,
		ctx: context.Background(),
		url: url,
	}
}

// post sends payload to the remote analytics endpoint
func (logger *eventLogger) post(payload *Payload) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "eventLogger: marshal json")
	}
	req, err := http.NewRequest("POST", logger.url, bytes.NewReader(payloadJSON))
	if err != nil {
		return errors.Wrap(err, "eventLogger: create post request")
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "eventLogger: http request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("eventLogger: %s failed with %d %s", payloadJSON, resp.StatusCode, string(body))
	}
	return nil
}

// newPayload generates a new Payload struct for a provided event
// in the context of the EventLogger client
func (logger *eventLogger) newPayload(userEmail *string, event *Event) *Payload {
	userInfo := &UserInfo{
		DomainUserID: "sourcegraph-backend-anonymous",
	}
	if userEmail != nil {
		userInfo = &UserInfo{
			DomainUserID: uuid.New().String(),
			Email:        *userEmail,
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
			AppID: backendEventsTrackingAppID,
			Env:   logger.env,
		},
		BatchInfo: &BatchInfo{
			BatchID:     uuid.New().String(),
			TotalEvents: 1,
			ServerTime:  fmt.Sprintf("%d", time.Now().UTC().Unix()*1000),
		},
		UserInfo: userInfo,
	}
}

// LogEvent sends a payload representing some user event to the remote analytics endpoint
func (logger *eventLogger) LogEvent(userEmail *string, eventLabel string, eventProperties map[string]string) error {
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
	return logger.post(payload)
}
