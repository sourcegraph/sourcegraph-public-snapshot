package main

import (
	"fmt"
	"strings"
	"time"
)

type PullRequest struct {
	ID         string
	Title      string
	Body       string
	Number     int
	URL        string
	State      string
	Repository string
	Private    bool
	Labels     []string
	Assignees  []string
	Milestone  string
	Author     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ClosedAt   time.Time
	BeganAt    time.Time // Time of the first authored commit

	LinkedIssues []*Issue `json:"-"` // Issues this PR resolves
	Parents      []*Issue `json:"-"` // Tracking issues watching this PR
}

func (pr *PullRequest) Closed() bool {
	return strings.EqualFold(pr.State, "closed")
}

func (pr *PullRequest) Merged() bool {
	return strings.EqualFold(pr.State, "merged")
}

func (pr *PullRequest) Done() bool {
	return pr.Merged() || pr.Closed()
}

func (pr *PullRequest) Summary() string {
	prefixSuffix := ""
	if pr.Done() {
		prefixSuffix = "~"
	}

	return fmt.Sprintf("%s[#%d](%s)%s", prefixSuffix, pr.Number, pr.URL, prefixSuffix)
}

func (pr *PullRequest) Markdown() string {
	state := " "
	prefixSuffix := ""
	daysSinceClose := ""
	if pr.Done() {
		state = "x"
		prefixSuffix = "~"
		daysSinceClose = fmt.Sprintf("(🏁 %s) ", formatTimeSince(pr.ClosedAt))
	}

	return fmt.Sprintf("- [%s] %s%s (%s[#%d](%s)%s) %s\n",
		state,
		daysSinceClose,
		pr.title(),
		prefixSuffix,
		pr.Number,
		pr.URL,
		prefixSuffix,
		pr.Emojis(),
	)
}

func (pr *PullRequest) Emojis() string {
	categories := Categories(pr.Labels, pr.Repository, pr.Body)
	categories["pull-request"] = ":shipit:"
	return Emojis(categories)
}

func (pr *PullRequest) title() string {
	var title string

	if pr.Private {
		title = pr.Repository
	} else {
		title = pr.Title
	}

	if pr.Closed() {
		title = "~" + strings.TrimSpace(title) + "~"
	}

	return title
}

func (pr *PullRequest) Redact() {
	if pr.Private {
		pr.Title = "REDACTED"
		pr.Labels = RedactLabels(pr.Labels)
	}
}
