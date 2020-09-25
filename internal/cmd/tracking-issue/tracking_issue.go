package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type TrackingIssue struct {
	*Issue
	Issues         []*Issue
	PRs            []*PullRequest
	LabelAllowlist []string
}

func NewTrackingIssue(issue *Issue) *TrackingIssue {
	t := &TrackingIssue{Issue: issue}
	t.FillLabelAllowlist()
	return t
}

var labelMatcher = regexp.MustCompile(labelMarkerRegexp)

// NOTE: labels specified inside the WORK section will be silently discarded
func (t *TrackingIssue) FillLabelAllowlist() {
	lines := strings.Split(t.Body, "\n")
	for _, line := range lines {
		matches := labelMatcher.FindStringSubmatch(line)
		if matches != nil {
			t.LabelAllowlist = append(t.LabelAllowlist, matches[1])
		}
	}
}

func (t *TrackingIssue) UpdateWork(work string) (updated bool, err error) {
	before := t.Body

	after, err := patch(t.Body, work)
	if err != nil {
		return false, err
	}

	t.Body = after
	return before != after, nil
}

func (t *TrackingIssue) Workloads() Workloads {
	workloads := map[string]*Workload{}

	workload := func(assignee string) *Workload {
		w := workloads[assignee]
		if w == nil {
			w = &Workload{
				TrackingIssue: t,
				Assignee:      assignee,
			}

			workloads[assignee] = w
			w.FillExistingIssuesFromTrackingBody(t)
		}
		return w
	}

	for _, issue := range t.Issues {
		// Exclude listing the tracking issue in the tracking issue.
		if issue.URL == t.Issue.URL {
			continue
		}

		linked := issue.LinkedPullRequests(t.PRs)
		for _, pr := range linked {
			issue.LinkedPRs = append(issue.LinkedPRs, pr)
			pr.LinkedIssues = append(pr.LinkedIssues, issue)
		}

		trackedIssues := issue.TrackedIssues(t.Issues)
		for _, child := range trackedIssues {
			issue.ChildIssues = append(issue.ChildIssues, child)
			child.Parents = append(child.Parents, issue)
		}

		trackedPRs := issue.TrackedPRs(t.PRs)
		for _, child := range trackedPRs {
			issue.ChildPRs = append(issue.ChildPRs, child)
			child.Parents = append(child.Parents, issue)
		}

		if t.Milestone != "" && issue.Milestone != t.Milestone {
			// Check to see if it's planned. Otherwise this will mark nested
			// tracking issues as deprioritized, even if they never assigned
			// a milestons.
			if contains(issue.Labels, fmt.Sprintf("planned/%s", t.Milestone)) {
				issue.Deprioritised = true
			}
		}
	}

	for _, issue := range t.Issues {
		// This influences render order, so make it deterministic
		sort.Slice(issue.ChildIssues, func(i, j int) bool { return issue.ChildIssues[i].Number < issue.ChildIssues[j].Number })
		sort.Slice(issue.ChildPRs, func(i, j int) bool { return issue.ChildPRs[i].Number < issue.ChildPRs[j].Number })
	}

	for _, issue := range t.Issues {
		// Exclude listing the tracking issue in the tracking issue.
		if issue.URL == t.Issue.URL {
			continue
		}

		issueAssignees := ListOfAssignees(issue.Assignees)
		for _, assignee := range issueAssignees {
			w := workload(assignee)
			w.AddIssue(issue)

			if t.Milestone == "" || issue.Milestone == t.Milestone {
				estimate := Estimate(issue.Labels)
				w.Days += Days(estimate)

				if issue.Closed() {
					w.CompletedDays += Days(estimate)
				}
			}
		}
	}

	for _, pr := range t.PRs {
		w := workload(pr.Author)
		w.PullRequests = append(w.PullRequests, pr)
	}

	// Do not tag someone unless they've done work
	for k, wl := range workloads {
		if !wl.Visible() {
			delete(workloads, k)
		}
	}

	return workloads
}
