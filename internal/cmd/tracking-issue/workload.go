package main

import (
	"fmt"
	"regexp"
	"strings"
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
		// Render any issue that belongs to zero or more than one
		// tracking issue (excluding the team tracking issue).
		if len(issue.Parents) != 1 {
			if !strings.EqualFold(issue.State, "closed") {
				renderIssue(&b, labelAllowlist, issue, 0)
			} else {
				hasCompletedIssueOrPullRequest = true
			}
		}
	}

	// Put all PRs that aren't linked to issues top-level
	for _, pr := range wl.PullRequests {
		if len(pr.LinkedIssues) == 0 {
			if !strings.EqualFold(pr.State, "merged") {
				b.WriteString(pr.Markdown())
			} else {
				hasCompletedIssueOrPullRequest = true
			}
		}
	}

	// If we have a renderable issue or pull request that has been completed,
	// then display a header with the sum of complete work estimates then all
	// of the issues and pull request we skipped in the loops above. This will
	// display all finished issues and pull requests as a flattened list.

	if hasCompletedIssueOrPullRequest {
		days = ""
		if wl.CompletedDays > 0 {
			days = fmt.Sprintf(": __%.2fd__", wl.CompletedDays)
		}

		fmt.Fprintf(&b, "\nCompleted%s\n", days)

		for _, issue := range wl.Issues {
			// Render any issue that belongs to zero or more than one
			// tracking issue (excluding the team tracking issue).
			if strings.EqualFold(issue.State, "closed") {
				b.WriteString(indent(0))
				b.WriteString(issue.Markdown(labelAllowlist))
			}
		}

		// Put all PRs that aren't linked to issues top-level
		for _, pr := range wl.PullRequests {
			if strings.EqualFold(pr.State, "merged") {
				b.WriteString(pr.Markdown())
			}
		}
	}

	fmt.Fprintf(&b, "%s\n", endAssigneeMarker)
	return b.String()
}

func renderIssue(b *strings.Builder, labelAllowlist []string, issue *Issue, depth int) {
	b.WriteString(indent(depth))
	b.WriteString(issue.Markdown(labelAllowlist))

	// Render children tracked _only_ by this issue
	// (excluding the team tracking issue) as nested elements
	for _, child := range issue.Children {
		if len(child.Parents) == 1 {
			renderIssue(b, labelAllowlist, child, depth+1)
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
