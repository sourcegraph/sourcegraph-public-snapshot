package azuredevops

import (
	"encoding/json"
	"strconv"
	"time"
)

var (
	PullRequestCreatedEventType AzureDevOpsEvent = "git.pullrequest.created"
	PullRequestMergedEventType  AzureDevOpsEvent = "git.pullrequest.merged"
	PullRequestUpdatedEventType AzureDevOpsEvent = "git.pullrequest.updated"
)

func ParseWebhookEvent(eventKey AzureDevOpsEvent, payload []byte) (any, error) {
	var target any
	switch eventKey {
	case PullRequestCreatedEventType:
		target = &PullRequestCreatedEvent{}
	case PullRequestMergedEventType:
		target = &PullRequestMergedEvent{}
	case PullRequestUpdatedEventType:
		target = &PullRequestUpdatedEvent{}
	default:
		return nil, webhookNotFoundErr{}
	}

	if err := json.Unmarshal(payload, target); err != nil {
		return nil, err
	}
	return target, nil
}

type AzureDevOpsEvent string

// BaseEvent is used to parse Azure DevOps events into the correct event struct.
type BaseEvent struct {
	EventType AzureDevOpsEvent `json:"eventType"`
}

type PullRequestEvent struct {
	ID          string                  `json:"id"`
	EventType   AzureDevOpsEvent        `json:"eventType"`
	PullRequest PullRequest             `json:"resource"`
	Message     PullRequestEventMessage `json:"message"`
	CreatedDate time.Time               `json:"createdDate"`
}

type PullRequestCreatedEvent PullRequestEvent
type PullRequestMergedEvent PullRequestEvent
type PullRequestUpdatedEvent PullRequestEvent

type PullRequestEventMessage struct {
	Text string `json:"text"`
}

// Widgetry to ensure all events are keyers.
//
// Annoyingly, most of the pull request events don't have UUIDs associated with
// anything we get, so we just have to do the best we can with what we have.

type keyer interface {
	Key() string
}

var (
	_ keyer = &PullRequestUpdatedEvent{}
	_ keyer = &PullRequestMergedEvent{}
	_ keyer = &PullRequestCreatedEvent{}
)

func (e *PullRequestUpdatedEvent) Key() string {
	return strconv.Itoa(e.PullRequest.ID) + ":updated"
}

func (e *PullRequestMergedEvent) Key() string {
	return strconv.Itoa(e.PullRequest.ID) + ":merged"
}

func (e *PullRequestCreatedEvent) Key() string {
	return strconv.Itoa(e.PullRequest.ID) + ":created"
}

type webhookNotFoundErr struct{}

func (w webhookNotFoundErr) Error() string {
	return "webhook not found"
}

func (w webhookNotFoundErr) NotFound() bool {
	return true
}
