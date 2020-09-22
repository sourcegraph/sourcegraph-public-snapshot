package main

import (
	"fmt"
	"regexp"
	"strings"
)

type Workload struct {
	Assignee     string
	Days         float64
	Issues       []*Issue
	PullRequests []*PullRequest
	Labels       []string
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

	indent := func(depth int) string {
		return strings.Repeat(" ", depth*2)
	}

	var renderIssue func(issue *Issue, depth int)
	renderIssue = func(issue *Issue, depth int) {
		b.WriteString(indent(depth))
		b.WriteString(issue.Markdown(labelAllowlist))

		// Render children tracked _only_ by this issue
		// (excluding the team tracking issue) as nested elements
		for _, child := range issue.Children {
			if len(child.Parents) == 1 {
				renderIssue(child, depth+1)
			}
		}
	}

	for _, issue := range wl.Issues {
		// Render any issue that belongs to zero or more than one
		// tracking issue (excluding the team tracking issue).
		if len(issue.Parents) != 1 {
			renderIssue(issue, 0)
		}
	}

	// Put all PRs that aren't linked to issues top-level
	for _, pr := range wl.PullRequests {
		if len(pr.LinkedIssues) == 0 {
			b.WriteString(pr.Markdown())
		}
	}

	fmt.Fprintf(&b, "%s\n", endAssigneeMarker)
	return b.String()
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
