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
	// threadsBucket is the bucket used for storing issues by thread ID.
	threadsBucket = "threads"

	// commentsBucket is the bucket name prefix used for storing comments. Actual
	// comments for a thread are stored in "comments-<thread ID>".
	commentsBucket = "comments"

	// eventsBucket is the bucket name prefix used for storing events. Actual
	// events for a thread are stored in "events-<thread ID>".
	eventsBucket = "events"
)

func threadCommentsBucket(threadID uint64) string {
	return commentsBucket + "-" + formatUint64(threadID)
}

func threadEventsBucket(threadID uint64) string {
	return eventsBucket + "-" + formatUint64(threadID)
}
