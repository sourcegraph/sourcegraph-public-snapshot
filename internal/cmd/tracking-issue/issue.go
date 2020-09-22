package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Issue struct {
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

	Deprioritised bool           `json:"-"`
	LinkedPRs     []*PullRequest `json:"-"`
	Children      []*Issue       `json:"-"`
	Parents       []*Issue       `json:"-"`
}

func (issue *Issue) Markdown(labelAllowlist []string) string {
	state := " "
	if strings.EqualFold(issue.State, "closed") {
		state = "x"
	}

	estimate := Estimate(issue.Labels)

	if estimate != "" {
		estimate = "__" + estimate + "__ "
	}

	labels := issue.RenderedLabels(labelAllowlist)

	return fmt.Sprintf("- [%s] %s [#%d](%s) %s%s%s\n",
		state,
		issue.title(),
		issue.Number,
		issue.URL,
		labels,
		estimate,
		issue.Emojis(),
	)
}

func (issue *Issue) RenderedLabels(labelAllowlist []string) string {
	var b strings.Builder
	for _, label := range issue.Labels {
		for _, allowedLabel := range labelAllowlist {
			if allowedLabel == label {
				b.WriteString(fmt.Sprintf("`%s` ", label))
				break
			}
		}
	}
	return b.String()
}

func (issue *Issue) Emojis() string {
	categories := Categories(issue.Labels, issue.Repository, issue.Body)
	return Emojis(categories)
}

func (issue *Issue) title() string {
	var title string

	if issue.Private {
		title = issue.Repository
	} else {
		title = issue.Title
	}

	// Cross off issues that were originally planned
	// for the milestone but are no longer in it.
	if issue.Deprioritised {
		title = "~" + strings.TrimSpace(title) + "~"
	}

	return title
}

func (issue *Issue) Tracked(issues []*Issue) (tracked []*Issue) {
	if !contains(issue.Labels, "tracking") {
		return nil
	}

outer:
	for _, other := range issues {
		if other == issue {
			continue
		}

		for _, label := range issue.Labels {
			if label != "tracking" && !contains(other.Labels, label) {
				continue outer
			}
		}

		tracked = append(tracked, other)
	}

	return tracked
}

func (issue *Issue) LinkedPullRequests(prs []*PullRequest) (linked []*PullRequest) {
	for _, pr := range prs {
		hasMatch, err := regexp.MatchString(fmt.Sprintf(`#%d([^\d]|$)`, issue.Number), pr.Body)
		if err != nil {
			panic(err)
		}
		if hasMatch {
			linked = append(linked, pr)
		}

		hasMatch, err = regexp.MatchString(fmt.Sprintf(`https://github.com/[^/]+/[^/]+/issues/%d([^\d]|$)`, issue.Number), pr.Body)
		if err != nil {
			panic(err)
		}
		if hasMatch {
			linked = append(linked, pr)
		}
	}

	return linked
}

func (issue *Issue) Redact() {
	if issue.Private {
		issue.Title = "REDACTED"
		issue.Labels = RedactLabels(issue.Labels)
	}
}
