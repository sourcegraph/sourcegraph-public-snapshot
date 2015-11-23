package sourcegraph

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
