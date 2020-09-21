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

	LinkedIssues []*Issue `json:"-"`
}

func (pr *PullRequest) Markdown() string {
	state := " "
	if strings.EqualFold(pr.State, "merged") {
		state = "x"
	}

	return fmt.Sprintf("- [%s] %s [#%d](%s) %s\n",
		state,
		pr.title(),
		pr.Number,
		pr.URL,
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

	if strings.EqualFold(pr.State, "closed") {
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
