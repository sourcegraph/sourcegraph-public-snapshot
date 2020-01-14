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

func WebHookType(r *http.Request) string {
	return r.Header.Get(eventTypeHeader)
}

func ParseWebHook(event string, payload []byte) (e interface{}, err error) {
	e = &PullRequestEvent{}
	return e, json.Unmarshal(payload, e)
}

type PullRequestEvent struct {
	Date        time.Time   `json:"date"`
	Actor       User        `json:"actor"`
	PullRequest PullRequest `json:"pullRequest"`
	Activity    *Activity   `json:"activity"`
}

type Webhook struct {
	Name     string   `json:"name"`
	Scope    string   `json:"scope"`
	Events   []string `json:"events"`
	Endpoint string   `json:"endpoint"`
	Secret   string   `json:"secret"`
}

// EnsureWebhook upserts a Webhook on a BBS instance.
func (c *Client) EnsureWebhook(ctx context.Context, w Webhook) error {
	raw, err := json.Marshal(w)
	if err != nil {
		return err
	}
	u := "rest/sourcegraph-admin/1.0/webhook"
	req, err := http.NewRequest("PUT", u, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	return c.do(ctx, req, nil)
}
