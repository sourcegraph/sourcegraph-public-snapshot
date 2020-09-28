package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Workload struct {
	TrackingIssue *TrackingIssue
	Assignee      string
	Days          float64
	CompletedDays float64
	Issues        []*Issue
	PullRequests  []*PullRequest
	Labels        []string
}

func (wl *Workload) Visible() bool {
	for _, issue := range wl.Issues {
		if wl.issueVisible(issue) {
			return true
		}
	}

	for _, pr := range wl.PullRequests {
		if wl.pullRequestVisible(pr) {
			return true
		}
	}

	return false
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
		if !wl.issueVisible(issue) {
			continue
		}
		if issue.Closed() {
			hasCompletedIssueOrPullRequest = true
			continue
		}

		// Render any issue that does not belong to a single sub-tracking
		// issue. We skip these issues on the top level as they will be
		// nested under their parent and we don't want to double-list.
		if len(wl.issuesVisibleAncestors(issue)) != 1 {
			wl.renderIssue(&b, labelAllowlist, issue, 0)
		}
	}

	// Put all PRs that aren't linked to issues or nested under a tracking
	// issue at the end of the top-level.
	for _, pr := range wl.PullRequests {
		if !wl.pullRequestVisible(pr) {
			continue
		}
		if pr.Done() {
			hasCompletedIssueOrPullRequest = true
			continue
		}

		if len(pr.LinkedIssues) == 0 && len(wl.pullRequestsVisibleAncestors(pr)) != 1 {
			b.WriteString(pr.Markdown())
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

		for _, issueOrPr := range wl.gatherCompletedWork(labelAllowlist) {
			b.WriteString(issueOrPr.Markdown)
		}
	}

	fmt.Fprintf(&b, "%s\n", endAssigneeMarker)
	return b.String()
}

type CompletedWork struct {
	Title    string
	Markdown string
	ClosedAt time.Time
}

func (wl *Workload) gatherCompletedWork(labelAllowList []string) []CompletedWork {
	completedWork := append(
		wl.gatherCompletedIssues(labelAllowList),
		wl.gatherCompletedPullRequests()...,
	)

	sort.Slice(completedWork, func(i, j int) bool {
		// Order rendered markdown by time elapsed since close
		if completedWork[i].ClosedAt.Before(completedWork[j].ClosedAt) {
			return true
		}
		if completedWork[i].ClosedAt != completedWork[j].ClosedAt {
			return false
		}

		// Break ties alphabetically
		return strings.Compare(completedWork[j].Title, completedWork[i].Title) < 0
	})

	return completedWork
}

// gatherCompletedIssues returns all completed issues whose sole parent is not
// also complete. This will ensure that we display the roots of large completed
// work instead of every element in that subtree.
func (wl *Workload) gatherCompletedIssues(labelAllowList []string) (completedWork []CompletedWork) {
	for _, issue := range wl.Issues {
		if !wl.issueVisible(issue) {
			continue
		}
		if !issue.Closed() {
			continue
		}

		// This would be displayed nested in a tracking issue. Because the
		// parent is also complete we don't want to show this duplicate data.
		visibleAncestors := wl.issuesVisibleAncestors(issue)
		if len(visibleAncestors) == 1 && visibleAncestors[0].Closed() {
			continue
		}

		completedWork = append(completedWork, CompletedWork{
			Title:    issue.Title,
			Markdown: issue.Markdown(labelAllowList),
			ClosedAt: issue.ClosedAt,
		})
	}

	return completedWork
}

// gatherCompletedPullRequests returns all completed pull requests that has an
// open parent or linked issue.
func (wl *Workload) gatherCompletedPullRequests() (completedWork []CompletedWork) {
	for _, pr := range wl.PullRequests {
		if !wl.pullRequestVisible(pr) {
			continue
		}
		if !pr.Done() {
			continue
		}

		// This would be displayed nested in a tracking issue but isn't explicitly
		// linked to a particular issue. We don't want to show these if that parent
		// is also complete.
		if len(pr.LinkedIssues) == 0 {
			visibleAncestors := wl.pullRequestsVisibleAncestors(pr)
			if len(visibleAncestors) == 1 && visibleAncestors[0].Closed() {
				continue
			}
		}

		closedIssues := 0
		for _, issue := range pr.LinkedIssues {
			if issue.Closed() {
				closedIssues++
			}
		}
		// If all linked issues are closed we can skip this. If there is at least one
		// open linked issue then the work related to this PR is not complete and we
		// want to show this in the progress.
		if len(pr.LinkedIssues) > 0 && len(pr.LinkedIssues) == closedIssues {
			continue
		}

		completedWork = append(completedWork, CompletedWork{
			Title:    pr.Title,
			Markdown: pr.Markdown(),
			ClosedAt: pr.ClosedAt,
		})
	}

	return completedWork
}

func (wl *Workload) renderIssue(b *strings.Builder, labelAllowlist []string, issue *Issue, depth int) {
	b.WriteString(indent(depth))
	b.WriteString(issue.Markdown(labelAllowlist))

	for _, child := range issue.ChildIssues {
		if !wl.issueVisible(child) {
			continue
		}

		// Render children tracked _only_ by this issue
		if len(wl.issuesVisibleAncestors(child)) == 1 {
			wl.renderIssue(b, labelAllowlist, child, depth+1)
		}
	}

	for _, child := range issue.ChildPRs {
		if !wl.pullRequestVisible(child) {
			continue
		}

		// Nest PRs under the tracking issue they most closely belong to
		// _only if_ it doesn't appear in the list of PRs for any issue
		// in this tracking issue(isn't explicitly linked to any issue).
		if len(child.LinkedIssues) == 0 && len(wl.pullRequestsVisibleAncestors(child)) == 1 {
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

// issuesVisibleAncestors gets a slice of ancestor issues which are visible also to this workload.
func (wl *Workload) issuesVisibleAncestors(issue *Issue) (ancestors []*Issue) {
	var frontier []*Issue
	frontier = append(frontier, issue.Parents...)

	visited := map[*Issue]struct{}{}
	for len(frontier) > 0 {
		var top *Issue
		top, frontier = frontier[0], frontier[1:]

		if _, ok := visited[top]; ok {
			continue
		}
		visited[top] = struct{}{}

		if wl.issueVisible(top) {
			// stop traversal on this branch
			ancestors = append(ancestors, top)
		} else {
			frontier = append(frontier, top.Parents...)
		}
	}

	return ancestors
}

// pullRequestsVisibleAncestors gets a slice of linked issues which are also visible to this workload.
func (wl *Workload) pullRequestsVisibleAncestors(pr *PullRequest) (ancestors []*Issue) {
	ancestorMap := map[*Issue]struct{}{}
	for _, parent := range pr.Parents {
		if wl.issueVisible(parent) {
			ancestorMap[parent] = struct{}{}
		} else {
			// linked issue isn't itself visible, but we're still
			// interested in the ancestors of that issue.
			for _, issue := range wl.issuesVisibleAncestors(parent) {
				ancestorMap[issue] = struct{}{}
			}
		}
	}

	// deduplicate
	for issue := range ancestorMap {
		ancestors = append(ancestors, issue)
	}
	return ancestors
}

// issueVisible determines if this issue should be rendered. An issue should be
// rendered if it is tracked by the tracking issue, or tracks another issue or
// pull request that is.
func (wl *Workload) issueVisible(issue *Issue) bool {
	// Cycle breaker: We wrap the recursive term O in the function above with
	// a check of its parameter, returning zero if it's already been seen as
	// part of the recursion. There are meta and cyclicly referential tracking
	// issues that become each other's parent AND child.
	visited := map[*Issue]struct{}{}

	f := func(rec issueRec, issue *Issue) bool {
		if _, ok := visited[issue]; ok {
			return false
		}
		visited[issue] = struct{}{}

		return wl.issueVisibleRec(rec, issue)
	}

	return f(f, issue)
}

type issueRec func(rec issueRec, issue *Issue) bool

func (wl *Workload) issueVisibleRec(rec issueRec, issue *Issue) bool {
	if contains(issue.Labels, "tracking") {
		// Tracking issues are visible if something they track is. Tracking issues
		// may be visible if the milestones match as well (this check falls through).
		for _, child := range issue.ChildIssues {
			if rec(rec, child) {
				return true
			}
		}
	} else {
		// Mismatched leaf issue authors
		if !contains(issue.Assignees, wl.Assignee) {
			return false
		}
	}

	if wl.TrackingIssue.Milestone != "" {
		// Mismatched milestones
		if issue.Milestone != "" && wl.TrackingIssue.Milestone != issue.Milestone {
			return false
		}
	}

	// Tags already match due to search query
	return true
}

// pullRequestVisible determines if this pull request should be rendered. A pull
// request should be rendered if it's attached to an issue that's also rendered.
func (wl *Workload) pullRequestVisible(pr *PullRequest) bool {
	if pr.Author != wl.Assignee {
		// Mismatched authors
		return false
	}

	if wl.TrackingIssue.Milestone != "" {
		if pr.Milestone != "" && wl.TrackingIssue.Milestone != pr.Milestone {
			return false
		}
	}

	return true
}
