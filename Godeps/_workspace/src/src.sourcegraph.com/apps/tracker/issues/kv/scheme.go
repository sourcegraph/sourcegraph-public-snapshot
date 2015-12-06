package kv

import (
	"time"

	"src.sourcegraph.com/apps/tracker/issues"
)

// issue is an on-disk representation of an issue.
type issue struct {
	State     issues.State
	Title     string
	AuthorUID int32
	CreatedAt time.Time
	Reference *reference `json:",omitempty"`
}

// reference is an on-disk representation of a reference.
type reference struct {
	Repo      issues.RepoSpec
	Path      string
	CommitID  string
	StartLine uint32
	EndLine   uint32
}

// comment is an on-disk representation of a comment.
type comment struct {
	AuthorUID int32
	CreatedAt time.Time
	Body      string
}

// event is an on-disk representation of an event.
type event struct {
	ActorUID  int32
	CreatedAt time.Time
	Type      issues.EventType
	Rename    *issues.Rename `json:",omitempty"`
}

const (
	// issuesBucket is the bucket used for storing issues by issue ID.
	issuesBucket = "threads"

	// commentsBucket is the bucket name prefix used for storing comments. Actual
	// comments for an issue are stored in "comments-{{.IssueID}}".
	commentsBucket = "comments"

	// eventsBucket is the bucket name prefix used for storing events. Actual
	// events for an issue are stored in "events-{{.IssueID}}".
	eventsBucket = "events"
)

func issueCommentsBucket(issueID uint64) string {
	return commentsBucket + "-" + formatUint64(issueID)
}

func issueEventsBucket(issueID uint64) string {
	return eventsBucket + "-" + formatUint64(issueID)
}
