package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var okayhqAPIEndpoint = "https://app.okayhq.com/api/events/v1"

// OkayMetric represents a particular metric attached to an event.
type OkayMetric struct {
	// Type is either "count", "durationMs" or "number".
	Type string `json:"type"`
	// Value is the actual value reported for this metric in a given event.
	Value float64 `json:"value"`
}

// okayEvent represents a custom event sent to the OkayHQ API.
type okayEvent struct {
	// Event is the event type, should be always set to "custom".
	Event string `json:"event"`
	// Name is the custom event name, used to select those events amonst others to build dashboards.
	Name string `json:"customEventName"`
	// Timestamp is the time at which this event occured.
	Timestamp time.Time `json:"timestamp"`
	// Identity ties this specific event to a particular user, enabling to filter events on various group predicates
	// such as teams, organizations, etc ...
	Identity okayEventIdentity `json:"identity"`
	// UniqueKey lists the property keys that are used to uniquely identify this event (optional).
	//
	// Sending another event with the same UniqueKey results in overwritting the previous event,
	// enabling to replay events with historical data or to correct incorrect events that were previously sent.
	UniqueKey []string `json:"uniqueKey,omitempty"`
	// Metrics are a map of okayMetric whose keys are the metric name.
	Metrics map[string]OkayMetric `json:"metrics"`
	// Properties are a map of additonal metadata (optional).
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type okayEventIdentity struct {
	// Type represents from where this identity is registered, should always be "sourceControlLogin".
	Type string `json:"type"`
	// User is the unique identifier to reference this identity amongst its Type.
	User string `json:"user"`
}

type OkayEvent struct {
	// Name is the custom event name, used to select those events amonst others to build dashboards.
	Name string
	// Timestamp is the time at which this event occured.
	Timestamp time.Time
	// GitHub login this event is attached to
	GitHubLogin string
	// UniqueKey lists the property keys that are used to uniquely identify this event (optional).
	//
	// Sending another event with the same UniqueKey results in overwritting the previous event,
	// enabling to replay events with historical data or to correct incorrect events that were previously sent.
	UniqueKey []string
	// Properties are a map of additonal metadata (optional).
	Properties map[string]interface{}
	// Metrics are a map of okayMetric whose keys are the metric name.
	Metrics map[string]OkayMetric
}

// OkayMetricsClient collects and submit metrics to the OkayHQ custom events API.
//
// See https://app.okayhq.com/help/_api/events
//
// TODO: If we were to extract this into a package:
//       - Flush should take a context.Context
//       - Flush should check if the context is cancelled in between making requests.
//       - Add some tests.
type OkayMetricsClient struct {
	token  string
	cli    *http.Client
	events []*okayEvent
	mu     sync.Mutex
}

// NewOkayMetricsClient returns a new OkayMetricsClient, using the provided http.Client.
func NewOkayMetricsClient(client *http.Client, token string) *OkayMetricsClient {
	return &OkayMetricsClient{
		cli:   client,
		token: token,
	}
}

// post submits an individual event to the API.
func (o *OkayMetricsClient) post(event *okayEvent) error {
	fmt.Println(okayhqAPIEndpoint)
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	buf := bytes.NewReader(b)
	req, err := http.NewRequest(http.MethodPost, okayhqAPIEndpoint, buf)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", o.token))
	req.Header.Add("Content-Type", "application/json")
	resp, err := o.cli.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("can't read response body")
		}
		defer resp.Body.Close()
		return errors.Newf("failed to submit custom metric to OkayHQ: %q", string(body))
	}
	fmt.Println(resp.StatusCode)
	return nil
}

// Push stores a new custom event to be submitted to OkayHQ once Flush is called.
func (o *OkayMetricsClient) Push(name string, event *OkayEvent) error {
	if name == "" {
		return errors.New("Okay event name can't be blank")
	}
	if event.Timestamp.IsZero() {
		return errors.New("Okay event timestamp name can't be zero")
	}
	if len(event.Metrics) == 0 {
		return errors.New("Okay event must have metrics")
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	o.events = append(o.events, &okayEvent{
		Event:     "custom",
		Name:      name,
		Timestamp: event.Timestamp,
		UniqueKey: event.UniqueKey,
		Identity: okayEventIdentity{
			Type: "sourceControlLogin",
			User: event.GitHubLogin,
		},
		Metrics:    event.Metrics,
		Properties: event.Properties,
	})

	return nil
}

// Flush empties the list of events accumulated by the client.
func (o *OkayMetricsClient) Flush() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	var errs error
	for _, event := range o.events {
		err := o.post(event)
		if err != nil {
			errs = errors.Append(err)
		}
	}
	// Reset the internal events buffer
	o.events = nil
	return errs
}
