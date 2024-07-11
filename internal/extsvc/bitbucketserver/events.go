package bitbucketserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	eventTypeHeader = "X-Event-Key"
)

func WebhookEventType(r *http.Request) string {
	return r.Header.Get(eventTypeHeader)
}

func ParseWebhookEvent(eventType string, payload []byte) (e any, err error) {
	switch eventType {
	case "ping", "diagnostics:ping":
		return PingEvent{}, nil
	case "repo:refs_changed":
		e = &PushEvent{}
		return e, json.Unmarshal(payload, e)
	case "repo:build_status":
		e = &BuildStatusEvent{}
		return e, json.Unmarshal(payload, e)
	case "pr:activity:status", "pr:activity:event", "pr:activity:rescope", "pr:activity:merge", "pr:activity:comment", "pr:activity:reviewers":
		e = &PullRequestActivityEvent{}
		return e, json.Unmarshal(payload, e)
	case "pr:participant:status":
		e = &PullRequestParticipantStatusEvent{}
		return e, json.Unmarshal(payload, e)
	default:
		return nil, errors.Errorf("unknown webhook event type: %q", eventType)
	}
}

type PingEvent struct{}

type PushEvent struct {
	Repository  Repo        `json:"repository"`
	CreatedDate int64       `json:"createdDate"`
	Changes     []RefChange `json:"changes"`
}

type RefChange struct {
	Ref RefChangeRef `json:"ref"`
	// "refId": "refs/heads/master",
	RefID string `json:"refId"`
	// "fromHash": "e54a46d4e4ddb7bb370070aa8a9b68e0ed959e5b",
	FromHash string `json:"fromHash"`
	// "toHash": "b14f55bd2b206c1128676131f9d66c56ec19e388",
	ToHash string `json:"toHash"`
	// "type": "UPDATE"
	Type string `json:"type"`
}

type RefChangeRef struct {
	// "id": "refs/heads/master",
	ID string `json:"id"`
	// "displayId": "master",
	DisplayID string `json:"displayId"`
	// "type": "BRANCH"
	Type string `json:"type"`
}

type PullRequestActivityEvent struct {
	Date        time.Time      `json:"date"`
	Actor       User           `json:"actor"`
	PullRequest PullRequest    `json:"pullRequest"`
	Action      ActivityAction `json:"action"`
	Activity    *Activity      `json:"activity"`
}

type PullRequestParticipantStatusEvent struct {
	*ParticipantStatusEvent
	PullRequest PullRequest `json:"pullRequest"`
}

type ParticipantStatusEvent struct {
	CreatedDate int            `json:"createdDate"`
	User        User           `json:"user"`
	Action      ActivityAction `json:"action"`
}

func (a *ParticipantStatusEvent) Key() string {
	return fmt.Sprintf("%s:%d:%d", a.Action, a.User.ID, a.CreatedDate)
}

type BuildStatusEvent struct {
	Commit       string        `json:"commit"`
	Status       BuildStatus   `json:"status"`
	PullRequests []PullRequest `json:"pullRequests"`
}
