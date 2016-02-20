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
	Reactions []reaction `json:",omitempty"`
}

// reaction is an on-disk representation of a reaction.
type reaction struct {
	EmojiID    issues.EmojiID
	AuthorUIDs []int32 // Order does not matter; this would be better represented as a set like map[int32]struct{}, but we're using JSON and it doesn't support that.
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
