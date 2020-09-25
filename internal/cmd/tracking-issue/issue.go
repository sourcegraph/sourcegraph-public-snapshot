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
	LinkedPRs     []*PullRequest `json:"-"` // PRs that resolve this issue
	ChildIssues   []*Issue       `json:"-"` // Tracked issues (only populated for tracking issues)
	ChildPRs      []*PullRequest `json:"-"` // Tracked PRs (only populated for tracking issues)
	Parents       []*Issue       `json:"-"` // Tracking issues watching this issue
}

func (issue *Issue) Closed() bool {
	return strings.EqualFold(issue.State, "closed")
}

func (issue *Issue) Markdown(labelAllowlist []string) string {
	state := " "
	prefixSuffix := ""
	daysSinceClose := ""
	if issue.Closed() {
		state = "x"
		prefixSuffix = "~"
		daysSinceClose = fmt.Sprintf("(🏁 %s) ", formatTimeSince(issue.ClosedAt))
	}

	estimate := Estimate(issue.Labels)
	if estimate == "" {
		est := float64(0)
		for _, child := range issue.ChildIssues {
			est += Days(Estimate(child.Labels))
		}
		if est > 0 {
			estimate = fmt.Sprintf("%.2fd", est)
		}
	}

	if estimate != "" {
		estimate = "__" + estimate + "__ "
	}

	labels := issue.RenderedLabels(labelAllowlist)

	var summaries []string
	for _, pr := range issue.LinkedPRs {
		summaries = append(summaries, pr.Summary())
	}

	pullRequestsPrefix := ""
	if len(summaries) > 0 {
		pullRequestsPrefix = "; PRs: "
	}

	return fmt.Sprintf("- [%s] %s%s ([%s#%d%s](%s)%s%s) %s%s%s\n",
		state,
		daysSinceClose,
		issue.title(),
		prefixSuffix,
		issue.Number,
		prefixSuffix,
		issue.URL,
		pullRequestsPrefix,
		strings.Join(summaries, ", "),
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

func (issue *Issue) TrackedIssues(issues []*Issue) (tracked []*Issue) {
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

func (issue *Issue) TrackedPRs(prs []*PullRequest) (tracked []*PullRequest) {
	if !contains(issue.Labels, "tracking") {
		return nil
	}

outer:
	for _, pr := range prs {
		for _, label := range issue.Labels {
			if label != "tracking" && !contains(pr.Labels, label) {
				continue outer
			}
		}

		tracked = append(tracked, pr)
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
