package issues

import (
	"time"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// Issue is a problem or story attached to a repository.
type Issue struct {
	issueInternal
	Author *sourcegraph.User
}

type issueInternal struct {
	// Title is the issue's display title.
	Title string

	// Status is the status of an issue, which can be open or closed.
	Status string

	// Events are events associated with an issue, currently any
	// replies made to the initial post.
	Events []Event

	// UID is a unique integer identifier for the issue within its
	// repository. It also is the rank of the issue in terms of post
	// date.
	UID int64

	// AuthorUID is the user ID of the issue author.
	AuthorUID int32

	// Updated is the time at which the issue was last updated.
	Updated time.Time
}

func (i *Issue) GetEvent(n int) *Event {
	if len(i.Events) >= n {
		return &i.Events[n-1]
	}
	return &Event{}
}

type Event struct {
	Created   time.Time
	UID       int
	AuthorUID int32
	Body      string
	Type      eventType
}

type eventType int

const (
	Reply eventType = iota
)

var cl *sourcegraph.Client

type IssueList struct {
	Issues []*Issue
}

func (i IssueList) Open() IssueList {
	return i.filterStatus("open")
}

func (i IssueList) Closed() IssueList {
	return i.filterStatus("closed")
}

func (issues IssueList) filterStatus(s string) IssueList {
	var filtered []*Issue
	for _, issue := range issues.Issues {
		if issue.Status == s {
			filtered = append(filtered, issue)
		}
	}
	return IssueList{filtered}
}
