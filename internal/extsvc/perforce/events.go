package perforce

// TODO: review. BCC'd from ADO implementation to make it "work" initially

import (
	"time"
)

type PerforceEvent string

// BaseEvent is used to parse Azure DevOps events into the correct event struct.
type BaseEvent struct {
	EventType PerforceEvent `json:"eventType"`
}

type ChangelistEvent struct {
	ID          string                 `json:"id"`
	EventType   PerforceEvent          `json:"eventType"`
	Changelist  Changelist             `json:"resource"`
	Message     ChangelistEventMessage `json:"message"`
	CreatedDate time.Time              `json:"createdDate"`
}

type ChangelistSubmittedEvent ChangelistEvent
type ChangelistShelvedEvent ChangelistEvent
type ChangelistUpdatedEvent ChangelistEvent
type ChangelistApprovedEvent ChangelistEvent
type ChangelistApprovedWithSuggestionsEvent ChangelistEvent
type ChangelistRejectedEvent ChangelistEvent
type ChangelistWaitingForAuthorEvent ChangelistEvent

type ChangelistEventMessage struct {
	Text string `json:"text"`
}
