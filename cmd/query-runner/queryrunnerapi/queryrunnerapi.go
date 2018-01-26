package queryrunnerapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	queryRunnerURL = env.Get("QUERY_RUNNER_URL", "http://query-runner", "URL at which the query-runner service can be reached")

	Client = &client{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
)

type SubjectAndConfig struct {
	Subject api.ConfigurationSubject
	Config  api.PartialConfigSavedQueries
}

type ErrorResponse struct {
	Message string
}

const (
	PathSavedQueryWasCreatedOrUpdated = "/saved-query-was-created-or-updated"
	PathSavedQueryWasDeleted          = "/saved-query-was-deleted"
)

type client struct {
	client *http.Client
}

// SavedQueryWasCreated should be called whenever a saved query was created
// or updated after the server has started.
func (c *client) SavedQueryWasCreatedOrUpdated(ctx context.Context, subject api.ConfigurationSubject, config api.PartialConfigSavedQueries) error {
	return c.post(PathSavedQueryWasCreatedOrUpdated, &SubjectAndConfig{
		Subject: subject,
		Config:  config,
	})
}

// SavedQueryWasDeleted should be called whenever a saved query was deleted
// after the server has started.
func (c *client) SavedQueryWasDeleted(ctx context.Context, spec api.SavedQueryIDSpec) error {
	return c.post(PathSavedQueryWasDeleted, spec)
}

func (c *client) post(path string, data interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		return errors.Wrap(err, "Encoding request")
	}
	u, err := url.Parse(queryRunnerURL)
	if err != nil {
		return errors.Wrap(err, "Parse QUERY_RUNNER_URL")
	}
	u.Path = path
	resp, err := c.client.Post(u.String(), "application/json", &buf)
	if err != nil {
		return errors.Wrap(err, "Post "+u.String())
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	var errResp *ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return errors.Wrap(err, "Decoding response")
	}
	return fmt.Errorf("Error from %s: %s", u.String(), errResp.Message)
}
