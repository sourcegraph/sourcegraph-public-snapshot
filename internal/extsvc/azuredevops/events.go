package azuredevops

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

var (
	PULL_REQUEST_APPROVED_TEXT                  = "approved pull request"
	PULL_REQUEST_APPROVED_WITH_SUGGESTIONS_TEXT = "has approved and left suggestions in pull request"
	PULL_REQUEST_REJECTED_TEXT                  = "rejected pull request"
	PULL_REQUEST_WAITING_FOR_AUTHOR_TEXT        = "is waiting for the author in pull request"

	PullRequestMergedEventType                  AzureDevOpsEvent = "git.pullrequest.merged"
	PullRequestUpdatedEventType                 AzureDevOpsEvent = "git.pullrequest.updated"
	PullRequestApprovedEventType                AzureDevOpsEvent = "git.pullrequest.approved"
	PullRequestApprovedWithSuggestionsEventType AzureDevOpsEvent = "git.pullrequest.approved_with_suggestions"
	PullRequestRejectedEventType                AzureDevOpsEvent = "git.pullrequest.rejected"
	PullRequestWaitingForAuthorEventType        AzureDevOpsEvent = "git.pullrequest.waiting_for_author"
)

func ParseWebhookEvent(eventKey AzureDevOpsEvent, payload []byte) (any, error) {
	var target any
	switch eventKey {
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

	// Azure DevOps doesn't give us much in the way of differentiating webhook events, so we are going
	// to try to parse the event message so that we can ideally simulate the different event types.
	if eventKey == PullRequestUpdatedEventType {
		var returnTarget any
		newTarget := target.(*PullRequestUpdatedEvent)
		text := newTarget.Message.Text

		switch {
		case strings.Contains(text, PULL_REQUEST_APPROVED_TEXT):
			newTarget.EventType = PullRequestApprovedEventType
			returnTarget = PullRequestApprovedEvent(*newTarget)
		case strings.Contains(text, PULL_REQUEST_REJECTED_TEXT):
			newTarget.EventType = PullRequestRejectedEventType
			returnTarget = PullRequestRejectedEvent(*newTarget)
		case strings.Contains(text, PULL_REQUEST_WAITING_FOR_AUTHOR_TEXT):
			newTarget.EventType = PullRequestWaitingForAuthorEventType
			returnTarget = PullRequestWaitingForAuthorEvent(*newTarget)
		case strings.Contains(text, PULL_REQUEST_APPROVED_WITH_SUGGESTIONS_TEXT):
			newTarget.EventType = PullRequestApprovedWithSuggestionsEventType
			returnTarget = PullRequestApprovedWithSuggestionsEvent(*newTarget)
		default:
			return target, nil
		}

		return returnTarget, nil
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

type PullRequestMergedEvent PullRequestEvent
type PullRequestUpdatedEvent PullRequestEvent
type PullRequestApprovedEvent PullRequestEvent
type PullRequestApprovedWithSuggestionsEvent PullRequestEvent
type PullRequestRejectedEvent PullRequestEvent
type PullRequestWaitingForAuthorEvent PullRequestEvent

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
	_ keyer = &PullRequestApprovedEvent{}
	_ keyer = &PullRequestApprovedWithSuggestionsEvent{}
	_ keyer = &PullRequestRejectedEvent{}
	_ keyer = &PullRequestWaitingForAuthorEvent{}
)

func (e *PullRequestUpdatedEvent) Key() string {
	return strconv.Itoa(e.PullRequest.ID) + ":updated:" + e.CreatedDate.String()
}

func (e *PullRequestMergedEvent) Key() string {
	return strconv.Itoa(e.PullRequest.ID) + ":merged:" + e.CreatedDate.String()
}

func (e *PullRequestApprovedEvent) Key() string {
	return strconv.Itoa(e.PullRequest.ID) + ":approved:" + e.CreatedDate.String()
}

func (e *PullRequestApprovedWithSuggestionsEvent) Key() string {
	return strconv.Itoa(e.PullRequest.ID) + ":approved_with_suggestions:" + e.CreatedDate.String()
}

func (e *PullRequestRejectedEvent) Key() string {
	return strconv.Itoa(e.PullRequest.ID) + ":rejected:" + e.CreatedDate.String()
}

func (e *PullRequestWaitingForAuthorEvent) Key() string {
	return strconv.Itoa(e.PullRequest.ID) + ":waiting_for_author:" + e.CreatedDate.String()
}

type webhookNotFoundErr struct{}

func (w webhookNotFoundErr) Error() string {
	return "webhook not found"
}

func (w webhookNotFoundErr) NotFound() bool {
	return true
}
