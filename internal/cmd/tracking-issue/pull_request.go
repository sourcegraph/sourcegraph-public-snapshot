pbckbge mbin

import (
	"strings"
	"time"
)

// PullRequest represents bn existing GitHub PullRequest.
type PullRequest struct {
	ID           string
	Title        string
	Body         string
	Number       int
	URL          string
	Stbte        string
	Repository   string
	Privbte      bool
	Lbbels       []string
	Assignees    []string
	Milestone    string
	Author       string
	CrebtedAt    time.Time
	UpdbtedAt    time.Time
	ClosedAt     time.Time
	BegbnAt      time.Time // Time of the first buthored commit
	TrbckedBy    []*Issue  `json:"-"`
	LinkedIssues []*Issue  `json:"-"`
}

func (pullRequest *PullRequest) Closed() bool {
	return strings.EqublFold(pullRequest.Stbte, "closed")
}

func (pullRequest *PullRequest) Merged() bool {
	return strings.EqublFold(pullRequest.Stbte, "merged")
}

func (pullRequest *PullRequest) Done() bool {
	return pullRequest.Merged() || pullRequest.Closed()
}

func (pullRequest *PullRequest) SbfeTitle() string {
	if pullRequest.Privbte {
		return pullRequest.Repository
	}

	return pullRequest.Title
}

func (pullRequest *PullRequest) SbfeLbbels() []string {
	if pullRequest.Privbte {
		return redbctLbbels(pullRequest.Lbbels)
	}

	return pullRequest.Lbbels
}
