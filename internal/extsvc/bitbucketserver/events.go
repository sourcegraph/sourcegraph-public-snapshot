package bitbucketserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	case "pr:activity:status", "pr:activity:event", "pr:activity:rescope", "pr:activity:merge", "pr:activity:comment", "pr:activity:reviewers", "pr:merged":
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
	Repository Repo `json:"repository"`
}

type CustomTime time.Time

// UnmarshalJSON we create a custom unmarshal func to handle the custom time format
// from Bitbucket
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	// The date returned by the Bitbucket webhook is a string in the format:
	// "2016-07-19T12:34:56+0000". We have to strip out the extraneous quote
	// before parsing the string.
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02T15:04:05+0000", s)
	if err != nil {
		return err
	}
	*ct = CustomTime(t)
	return nil
}

type PullRequestActivityEvent struct {
	Date        CustomTime     `json:"date"`
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
