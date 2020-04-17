package bitbucketserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const (
	eventTypeHeader = "X-Event-Key"
)

func WebhookEventType(r *http.Request) string {
	return r.Header.Get(eventTypeHeader)
}

func ParseWebhookEvent(eventType string, payload []byte) (e interface{}, err error) {
	switch eventType {
	case "ping":
		return PingEvent{}, nil
	case "repo:build_status":
		e = &BuildStatusEvent{}
		return e, json.Unmarshal(payload, e)
	default:
		e = &PullRequestEvent{}
		return e, json.Unmarshal(payload, e)
	}
}

type PingEvent struct{}

type PullRequestEvent struct {
	Date        time.Time   `json:"date"`
	Actor       User        `json:"actor"`
	PullRequest PullRequest `json:"pullRequest"`
	Activity    *Activity   `json:"activity"`
}

type BuildStatusEvent struct {
	Commit       string        `json:"commit"`
	Status       BuildStatus   `json:"status"`
	PullRequests []PullRequest `json:"pullRequests"`
}

// Webhook defines the JSON schema from the BBS Sourcegraph plugin.
// This is not the native BBS webhook.
type Webhook struct {
	Name     string   `json:"name"`
	Scope    string   `json:"scope"`
	Events   []string `json:"events"`
	Endpoint string   `json:"endpoint"`
	Secret   string   `json:"secret"`
}

const webhookURL = "rest/sourcegraph-admin/1.0/webhook"

// UpsertWebhook upserts a Webhook on a BBS instance.
func (c *Client) UpsertWebhook(ctx context.Context, w Webhook) error {
	raw, err := json.Marshal(w)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	return c.do(ctx, req, nil)
}

// DeleteWebhook deletes the webhook with the given name
func (c *Client) DeleteWebhook(ctx context.Context, name string) error {
	u := webhookURL + "?name=" + name
	req, err := http.NewRequestWithContext(ctx, "DELETE", u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf8")
	return c.do(ctx, req, nil)
}
