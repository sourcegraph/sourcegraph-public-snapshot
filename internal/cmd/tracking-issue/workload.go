package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Workload struct {
	Assignee      string
	Days          float64
	CompletedDays float64
	Issues        []*Issue
	PullRequests  []*PullRequest
	Labels        []string
}

func (wl *Workload) AddIssue(newIssue *Issue) {
	for _, issue := range wl.Issues {
		if issue.URL == newIssue.URL {
			return
		}
	}

	wl.Issues = append(wl.Issues, newIssue)
}

func (wl *Workload) Markdown(labelAllowlist []string) string {
	var days string
	if wl.Days > 0 {
		days = fmt.Sprintf(": __%.2fd__", wl.Days)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "\n"+beginAssigneeMarkerFmt+"\n", wl.Assignee)
	fmt.Fprintf(&b, "@%s%s\n\n", wl.Assignee, days)

	// First list all of the incomplete issues and pull requests. This may
	// include an incomplete tracking issue with both complete and incomplete
	// subtasks.

	hasCompletedIssueOrPullRequest := false
	for _, issue := range wl.Issues {
		if issue.Closed() {
			hasCompletedIssueOrPullRequest = true
			continue
		}

		// Render any issue that does not belong to a single sub-tracking
		// issue. We skip these issues on the top level as they will be
		// nested under their parent and we don't want to double-list.
		if len(issue.Parents) != 1 {
			renderIssue(&b, labelAllowlist, issue, 0)
		}
	}

	// Put all PRs that aren't linked to issues or nested under a tracking
	// issue at the end of the top-level.
	for _, pr := range wl.PullRequests {
		if pr.Done() {
			hasCompletedIssueOrPullRequest = true
			continue
		}

		if len(pr.LinkedIssues) == 0 && len(pr.Parents) != 1 {
			b.WriteString(pr.Markdown())
		}
	}

	// If we have a renderable issue or pull request that has been completed,
	// then display a header with the sum of complete work estimates then all
	// of the issues and pull request we skipped in the loops above. This will
	// display all finished issues and pull requests as a flattened list.

	if hasCompletedIssueOrPullRequest {
		type renderedCompletedWork struct {
			Markdown string
			ClosedAt time.Time
		}
		var completedWork []renderedCompletedWork

		for _, issue := range wl.Issues {
			// Render any issue that belongs to zero or more than one
			// tracking issue (excluding the team tracking issue).
			if issue.Closed() {
				completedWork = append(completedWork, renderedCompletedWork{
					Markdown: issue.Markdown(labelAllowlist),
					ClosedAt: issue.ClosedAt,
				})
			}
		}

	outer:
		for _, pr := range wl.PullRequests {
			if pr.Done() {
				// Put all closed PRs that have at least one linked issue that
				// has not been completed at the top level of the finished work.
				for _, issue := range pr.LinkedIssues {
					if issue.Closed() {
						continue outer
					}
				}

				completedWork = append(completedWork, renderedCompletedWork{
					Markdown: pr.Markdown(),
					ClosedAt: pr.ClosedAt,
				})
			}
		}

		sort.Slice(completedWork, func(i, j int) bool {
			// Order rendered markdown by time elapsed since close
			return completedWork[i].ClosedAt.Before(completedWork[j].ClosedAt)
		})

		days = ""
		if wl.CompletedDays > 0 {
			days = fmt.Sprintf(": __%.2fd__", wl.CompletedDays)
		}

		fmt.Fprintf(&b, "\nCompleted%s\n", days)

		// Display completed work chronologically
		for _, issueOrPr := range completedWork {
			b.WriteString(issueOrPr.Markdown)
		}
	}

	fmt.Fprintf(&b, "%s\n", endAssigneeMarker)
	return b.String()
}

func renderIssue(b *strings.Builder, labelAllowlist []string, issue *Issue, depth int) {
	b.WriteString(indent(depth))
	b.WriteString(issue.Markdown(labelAllowlist))

	// Render children tracked _only_ by this issue
	// (excluding the tracking issue being updated) as nested elements
	for _, child := range issue.ChildIssues {
		if len(child.Parents) == 1 {
			renderIssue(b, labelAllowlist, child, depth+1)
		}
	}

	for _, child := range issue.ChildPRs {
		// Nest PRs under the tracking issue they most closely belong to
		// _only if_ it doesn't appear in the list of PRs for any issue
		// in this tracking issue(isn't explicitly linked to any issue).
		if len(child.Parents) == 1 && len(child.LinkedIssues) == 0 {
			b.WriteString(indent(depth + 1))
			b.WriteString(child.Markdown())
		}
	}
}

func indent(depth int) string {
	return strings.Repeat(" ", depth*2)
}

var issueURLMatcher = regexp.MustCompile(`https://github\.com/.+/.+/issues/\d+`)

func (wl *Workload) FillExistingIssuesFromTrackingBody(tracking *TrackingIssue) {
	beginAssigneeMarker := fmt.Sprintf(beginAssigneeMarkerFmt, wl.Assignee)

	start, err := findMarker(tracking.Body, beginAssigneeMarker)
	if err != nil {
		return
	}

	end, err := findMarker(tracking.Body[start:], endAssigneeMarker)
	if err != nil {
		return
	}

	lines := strings.Split(tracking.Body[start:start+end], "\n")

	for _, line := range lines {
		parsedIssueURL := issueURLMatcher.FindString(line)
		if parsedIssueURL == "" {
			continue
		}

		for _, issue := range tracking.Issues {
			if parsedIssueURL == issue.URL && Assignee(issue.Assignees) == wl.Assignee {
				wl.AddIssue(issue)
			}
		}
	}
}
