// Package okay provides a client to submit custom events to the OkayHQ API.
//
// To ease local development, using a blank token will log a warning and flushing
// the client will result in logging events at the DEBUG level.
package okay

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

// Metric represents a particular metric attached to an event.
type Metric struct {
	// Type is either "count", "durationMs" or "number".
	Type string `json:"type"`
	// Value is the actual value reported for this metric in a given event.
	Value float64 `json:"value"`
}

// customEvent represents a custom event sent to the OkayHQ API.
type customEvent struct {
	// Event is the event type, should be always set to "custom".
	Event string `json:"event"`
	// Name is the custom event name, used to select those events amonst others to build dashboards.
	Name string `json:"customEventName"`
	// Timestamp is the time at which this event occured.
	Timestamp time.Time `json:"timestamp"`
	// Identity ties this specific event to a particular user, enabling to filter events on various group predicates
	// such as teams, organizations, etc ... (optional).
	Identity *eventIdentity `json:"identity,omitempty"`
	// UniqueKey lists the property keys that are used to uniquely identify this event (optional).
	//
	// Sending another event with the same UniqueKey results in overwritting the previous event,
	// enabling to replay events with historical data or to correct incorrect events that were previously sent.
	UniqueKey []string `json:"uniqueKey,omitempty"`
	// Metrics are a map of okayMetric whose keys are the metric name.
	Metrics map[string]Metric `json:"metrics"`
	// Properties are a map of additonal metadata (optional).
	Properties map[string]string `json:"properties,omitempty"`
	// Labels are a list of strings used to tag the event.
	Labels []string `json:"labels,omitempty"`
}

// eventIdentity represents the identity to attach to an event.
type eventIdentity struct {
	// Type represents from where this identity is registered, should always be "sourceControlLogin".
	Type string `json:"type"`
	// User is the unique identifier to reference this identity amongst its Type.
	User string `json:"user"`
}

// Event represents a custom Event to be sent to OkayHQ.
type Event struct {
	// Name is the custom event name, used to select those events amonst others to build dashboards.
	Name string
	// Timestamp is the time at which this event occured.
	Timestamp time.Time
	// OkayURL is used to generate a clickable link in OkayHQ's UI when browsing this event.
	OkayURL string
	// GitHub login this event is attached to (optional).
	GitHubLogin string
	// UniqueKey lists the property keys that are used to uniquely identify this event (optional).
	//
	// Sending another event with the same UniqueKey results in overwritting the previous event,
	// enabling to replay events with historical data or to correct incorrect events that were previously sent.
	UniqueKey []string
	// Properties are a map of additonal metadata (optional).
	Properties map[string]string
	// Metrics are a map of okayMetric whose keys are the metric name.
	Metrics map[string]Metric
	// Labels are a list of strings used to tag the event.
	Labels []string
}

// Client collects and submit metrics to the OkayHQ custom events API and is safe to use
// concurrently.
//
// See https://app.okayhq.com/help/_api/events
type Client struct {
	token    string
	cli      *http.Client
	events   []*customEvent
	endpoint string

	mu sync.Mutex
}

// NewClient returns a new OkayMetricsClient, using the provided http.Client.
func NewClient(client *http.Client, token string) *Client {
	return &Client{
		cli:   client,
		token: token,

		endpoint: okayhqAPIEndpoint,
	}
}

func (o *Client) SetEndpoint(url string) {
	o.endpoint = url
}

// post submits an individual event to the API.
func (o *Client) post(event *customEvent) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	buf := bytes.NewReader(b)
	req, err := http.NewRequest(http.MethodPost, o.endpoint, buf)
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
		return errors.Newf("okayhq: failed to submit custom metric %s to OkayHQ: %q", event.Name, string(body))
	}
	return nil
}

// Push stores a new custom event to be submitted to OkayHQ once Flush is called.
func (o *Client) Push(event *Event) error {
	if event.Name == "" {
		return errors.New("okayhq: event name can't be blank")
	}
	if event.Timestamp.IsZero() {
		return errors.New("okayhq: event timestamp name can't be zero")
	}
	if len(event.Metrics) == 0 {
		return errors.New("okayhq: event must have metrics")
	}
	for _, k := range event.UniqueKey {
		if _, ok := event.Properties[k]; !ok {
			return errors.Newf("okayhq: event proprety %s is marked as unique, but absent from the properties")
		}
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	ce := &customEvent{
		Event:      "custom",
		Name:       event.Name,
		Timestamp:  event.Timestamp,
		UniqueKey:  event.UniqueKey,
		Metrics:    event.Metrics,
		Properties: event.Properties,
		Labels:     event.Labels,
	}
	if event.GitHubLogin != "" {
		ce.Identity = &eventIdentity{
			Type: "sourceControlLogin",
			User: event.GitHubLogin,
		}
	}
	if event.OkayURL != "" {
		ce.Properties["okay.url"] = event.OkayURL
	}
	o.events = append(o.events, ce)

	return nil
}

// Flush empties the list of events accumulated by the client.
func (o *Client) Flush() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	var errs error
	for _, event := range o.events {
		err := o.post(event)
		if err != nil {
			errs = errors.Append(errs, err)
		}
	}
	// Reset the internal events buffer
	o.events = nil

	return errs
}
