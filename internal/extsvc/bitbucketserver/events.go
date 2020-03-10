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

// Webhook defines the JSON schema from the BBS Sourcegraph plugin.
// This is not the native BBS webhook.
type Webhook struct {
	Name     string   `json:"name"`
	Scope    string   `json:"scope"`
	Events   []string `json:"events"`
	Endpoint string   `json:"endpoint"`
	Secret   string   `json:"secret"`
}

// UpsertWebhook upserts a Webhook on a BBS instance.
func (c *Client) UpsertWebhook(ctx context.Context, w Webhook) error {
	raw, err := json.Marshal(w)
	if err != nil {
		return err
	}
	u := "rest/sourcegraph-admin/1.0/webhook"
	req, err := http.NewRequest("POST", u, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	return c.do(ctx, req, nil)
}
