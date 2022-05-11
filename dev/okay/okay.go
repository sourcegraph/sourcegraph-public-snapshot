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

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

var okayhqAPIEndpoint = "https://app.okayhq.com/api/events/v1"

// Client collects and submit metrics to the OkayHQ custom events API and is safe to use
// concurrently.
//
// See https://app.okayhq.com/help/_api/events
type Client struct {
	token    string
	cli      *http.Client
	events   []*customEvent
	endpoint string
	logger   log.Logger

	mu sync.Mutex
}

// NewClient returns a new OkayMetricsClient, using the provided http.Client.
func NewClient(client *http.Client, token string) *Client {
	logger := log.Scoped("okayhq", "OkayHQ Metrics")
	if token == "" {
		logger.Warn("empty token, will log events at DEBUG level instead of submitting them to the API")
	}
	return &Client{
		cli:    client,
		token:  token,
		logger: logger,

		endpoint: okayhqAPIEndpoint,
	}
}

// SetEndpoint replace the default endpoint to OkayHQ API.
func (c *Client) SetEndpoint(url string) {
	c.endpoint = url
}

// post submits an individual event to the API.
func (c *Client) post(event *customEvent) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if c.token == "" {
		// If the token is empty, just log the events
		c.logger.Debug("pretending to send event", log.String("event", string(b)))
		return nil
	}

	buf := bytes.NewReader(b)
	req, err := http.NewRequest(http.MethodPost, c.endpoint, buf)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.cli.Do(req)
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
func (c *Client) Push(event *Event) error {
	if event.Name == "" {
		return errors.New("okayhq: event name can't be blank")
	}
	if event.Timestamp.IsZero() {
		return errors.New("okayhq: event timestamp name can't be zero")
	}
	if len(event.Metrics) == 0 {
		return errors.New("okayhq: event must have metrics")
	}
	if len(event.UniqueKey) == 0 {
		return errors.New("okayhq: event must have unqiue property keys")
	}
	for _, k := range event.UniqueKey {
		if _, ok := event.Properties[k]; !ok {
			return errors.Newf("okayhq: event proprety %s is marked as unique, but absent from the properties")
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

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
	c.events = append(c.events, ce)

	return nil
}

// Flush empties the list of events accumulated by the client.
func (c *Client) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs error
	for _, event := range c.events {
		err := c.post(event)
		if err != nil {
			errs = errors.Append(errs, err)
		}
	}
	// Reset the internal events buffer
	c.events = nil

	return errs
}
