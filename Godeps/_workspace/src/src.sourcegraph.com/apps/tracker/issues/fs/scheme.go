package fs

import (
	"path"
	"time"

	"src.sourcegraph.com/apps/tracker/issues"
)

// issue is an on-disk representation of an issue.
type issue struct {
	State issues.State
	Title string
	comment
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
	// issuesDir is '/'-separated path for issue storage.
	issuesDir = "threads"

	// eventsDir is dir name for issue events.
	eventsDir = "events"
)

func issueDir(issueID uint64) string {
	return path.Join(issuesDir, formatUint64(issueID))
}

func issueCommentPath(issueID, commentID uint64) string {
	return path.Join(issuesDir, formatUint64(issueID), formatUint64(commentID))
}

func issueEventsDir(issueID uint64) string {
	return path.Join(issuesDir, formatUint64(issueID), eventsDir)
}

func issueEventPath(issueID, eventID uint64) string {
	return path.Join(issuesDir, formatUint64(issueID), eventsDir, formatUint64(eventID))
}
