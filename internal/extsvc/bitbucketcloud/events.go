package bitbucketcloud

import (
	"encoding/json"
	"strconv"
	"time"
)

func ParseWebhookEvent(eventKey string, payload []byte) (any, error) {
	var target any
	switch eventKey {
	case "pullrequest:approved":
		target = &PullRequestApprovedEvent{}
	case "pullrequest:changes_request_created":
		target = &PullRequestChangesRequestCreatedEvent{}
	case "pullrequest:changes_request_removed":
		target = &PullRequestChangesRequestRemovedEvent{}
	case "pullrequest:comment_created":
		target = &PullRequestCommentCreatedEvent{}
	case "pullrequest:comment_deleted":
		target = &PullRequestCommentDeletedEvent{}
	case "pullrequest:comment_updated":
		target = &PullRequestCommentUpdatedEvent{}
	case "pullrequest:fulfilled":
		target = &PullRequestFulfilledEvent{}
	case "pullrequest:rejected":
		target = &PullRequestRejectedEvent{}
	case "pullrequest:unapproved":
		target = &PullRequestUnapprovedEvent{}
	case "pullrequest:updated":
		target = &PullRequestUpdatedEvent{}
	case "repo:commit_status_created":
		target = &RepoCommitStatusCreatedEvent{}
	case "repo:commit_status_updated":
		target = &RepoCommitStatusUpdatedEvent{}
	case "repo:push":
		target = &PushEvent{}
	default:
		return nil, UnknownWebhookEventKey(eventKey)
	}

	if err := json.Unmarshal(payload, target); err != nil {
		return nil, err
	}
	return target, nil
}

// Types (and subtypes) that we can unmarshal from a webhook payload.
//
// This is (intentionally) most, but not all, of the payload types as of December
// 2022. Some repo events are unlikely to ever be useful to us, so we don't even
// attempt to unmarshal them.

type PushEvent struct {
	RepoEvent
}

type PullRequestEvent struct {
	RepoEvent
	PullRequest PullRequest `json:"pullrequest"`
}

type PullRequestApprovalEvent struct {
	PullRequestEvent
	Approval DateUserTuple `json:"approval"`
}

type PullRequestApprovedEvent struct {
	PullRequestApprovalEvent
}

type PullRequestUnapprovedEvent struct {
	PullRequestApprovalEvent
}

type PullRequestChangesRequestEvent struct {
	PullRequestEvent
	ChangesRequest DateUserTuple `json:"changes_request"`
}

type PullRequestChangesRequestCreatedEvent struct {
	PullRequestChangesRequestEvent
}

type PullRequestChangesRequestRemovedEvent struct {
	PullRequestChangesRequestEvent
}

type PullRequestCommentEvent struct {
	PullRequestEvent
	Comment Comment `json:"comment"`
}

type PullRequestCommentCreatedEvent struct {
	PullRequestCommentEvent
}

type PullRequestCommentDeletedEvent struct {
	PullRequestCommentEvent
}

type PullRequestCommentUpdatedEvent struct {
	PullRequestCommentEvent
}

type PullRequestFulfilledEvent struct {
	PullRequestEvent
}

type PullRequestRejectedEvent struct {
	PullRequestEvent
}

type DateUserTuple struct {
	Date time.Time `json:"date"`
	User User      `json:"user"`
}

type PullRequestUpdatedEvent struct {
	PullRequestEvent
}

type RepoEvent struct {
	Actor      User `json:"actor"`
	Repository Repo `json:"repository"`
}

type RepoCommitStatusEvent struct {
	RepoEvent
	CommitStatus CommitStatus `json:"commit_status"`
}

type RepoCommitStatusCreatedEvent struct {
	RepoCommitStatusEvent
}

type RepoCommitStatusUpdatedEvent struct {
	RepoCommitStatusEvent
}

type CommitStatus struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	State       PullRequestStatusState `json:"state"`
	Key         string                 `json:"key"`
	URL         string                 `json:"url"`
	Type        CommitStatusType       `json:"type"`
	CreatedOn   time.Time              `json:"created_on"`
	UpdatedOn   time.Time              `json:"updated_on"`
	Commit      Commit                 `json:"commit"`
	Links       Links                  `json:"links"`
}

// The single typed string type in the webhook specific types.

type CommitStatusType string

const (
	CommitStatusTypeBuild CommitStatusType = "build"
)

// Error types.

type UnknownWebhookEventKey string

var _ error = UnknownWebhookEventKey("")

func (e UnknownWebhookEventKey) Error() string {
	return "unknown webhook event key: " + string(e)
}

// Widgetry to ensure all events are keyers.
//
// Annoyingly, most of the pull request events don't have UUIDs associated with
// anything we get, so we just have to do the best we can with what we have.

type keyer interface {
	Key() string
}

var (
	_ keyer = &PullRequestApprovedEvent{}
	_ keyer = &PullRequestChangesRequestCreatedEvent{}
	_ keyer = &PullRequestChangesRequestRemovedEvent{}
	_ keyer = &PullRequestCommentCreatedEvent{}
	_ keyer = &PullRequestCommentDeletedEvent{}
	_ keyer = &PullRequestCommentUpdatedEvent{}
	_ keyer = &PullRequestFulfilledEvent{}
	_ keyer = &PullRequestRejectedEvent{}
	_ keyer = &PullRequestUnapprovedEvent{}
	_ keyer = &PullRequestUpdatedEvent{}
	_ keyer = &RepoCommitStatusCreatedEvent{}
	_ keyer = &RepoCommitStatusUpdatedEvent{}
)

func (e *PullRequestApprovedEvent) Key() string {
	return e.PullRequestApprovalEvent.key() + ":approved"
}

func (e *PullRequestChangesRequestCreatedEvent) Key() string {
	return e.PullRequestChangesRequestEvent.key() + ":created"
}

func (e *PullRequestChangesRequestRemovedEvent) Key() string {
	return e.PullRequestChangesRequestEvent.key() + ":removed"
}

func (e *PullRequestCommentCreatedEvent) Key() string {
	return e.PullRequestCommentEvent.key() + ":created"
}

func (e *PullRequestCommentDeletedEvent) Key() string {
	return e.PullRequestCommentEvent.key() + ":deleted"
}

func (e *PullRequestCommentUpdatedEvent) Key() string {
	return e.PullRequestCommentEvent.key() + ":updated"
}

func (e *PullRequestFulfilledEvent) Key() string {
	return e.PullRequestEvent.key() + ":fulfilled"
}

func (e *PullRequestRejectedEvent) Key() string {
	return e.PullRequestEvent.key() + ":rejected"
}

func (e *PullRequestUnapprovedEvent) Key() string {
	return e.PullRequestApprovalEvent.key() + ":unapproved"
}

func (e *PullRequestUpdatedEvent) Key() string {
	return e.PullRequestEvent.key() + ":" + e.PullRequest.UpdatedOn.String()
}

func (e *RepoCommitStatusCreatedEvent) Key() string {
	return e.RepoCommitStatusEvent.key() + ":created"
}

func (e *RepoCommitStatusUpdatedEvent) Key() string {
	return e.RepoCommitStatusEvent.key() + ":updated"
}

func (e *PullRequestApprovalEvent) key() string {
	return e.PullRequestEvent.key() + ":" +
		e.Approval.User.UUID + ":" +
		e.Approval.Date.String()
}

func (e *PullRequestChangesRequestEvent) key() string {
	return e.PullRequestEvent.key() + ":" +
		e.ChangesRequest.User.UUID + ":" +
		e.ChangesRequest.Date.String()
}

func (e *PullRequestCommentEvent) key() string {
	return e.PullRequestEvent.key() + ":" + strconv.FormatInt(e.Comment.ID, 10)
}

func (e *PullRequestEvent) key() string {
	return e.RepoEvent.key() + ":" + strconv.FormatInt(e.PullRequest.ID, 10)
}

func (e *RepoCommitStatusEvent) key() string {
	return e.RepoEvent.key() + ":" +
		e.CommitStatus.Commit.Hash + ":" +
		e.CommitStatus.CreatedOn.String()
}

func (e *RepoEvent) key() string {
	return e.Repository.UUID
}
