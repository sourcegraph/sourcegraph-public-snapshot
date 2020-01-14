package bitbucketserver

import (
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
