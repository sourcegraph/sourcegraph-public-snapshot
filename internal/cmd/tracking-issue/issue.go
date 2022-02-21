package main

import (
	"strings"
	"sync"
	"time"

	"github.com/grafana/regexp"
)

// Issue represents an existing GitHub Issue.
type Issue struct {
	ID                  string
	Title               string
	Body                string
	Number              int
	URL                 string
	State               string
	Repository          string
	Private             bool
	Labels              []string
	Assignees           []string
	Milestone           string
	MilestoneNumber     int
	Author              string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	ClosedAt            time.Time
	TrackedBy           []*Issue       `json:"-"`
	TrackedIssues       []*Issue       `json:"-"`
	TrackedPullRequests []*PullRequest `json:"-"`
	LinkedPullRequests  []*PullRequest `json:"-"`

	// Populate and get with .IdentifyingLabels()
	identifyingLabels     []string
	identifyingLabelsOnce sync.Once
}

func (issue *Issue) Closed() bool {
	return strings.EqualFold(issue.State, "closed")
}

var optionalLabelMatcher = regexp.MustCompile(optionalLabelMarkerRegexp)

func (issue *Issue) IdentifyingLabels() []string {
	issue.identifyingLabelsOnce.Do(func() {
		issue.identifyingLabels = nil

		// Parse out optional labels
		optionalLabels := map[string]struct{}{}
		lines := strings.Split(issue.Body, "\n")
		for _, line := range lines {
			matches := optionalLabelMatcher.FindStringSubmatch(line)
			if matches != nil {
				optionalLabels[matches[1]] = struct{}{}
			}
		}

		// Get non-optional and non-tracking labels
		for _, label := range issue.Labels {
			if _, optional := optionalLabels[label]; !optional && label != "tracking" {
				issue.identifyingLabels = append(issue.identifyingLabels, label)
			}
		}
	})

	return issue.identifyingLabels
}

func (issue *Issue) SafeTitle() string {
	if issue.Private {
		return issue.Repository
	}

	return issue.Title
}

func (issue *Issue) SafeLabels() []string {
	if issue.Private {
		return redactLabels(issue.Labels)
	}

	return issue.Labels
}

func (issue *Issue) UpdateBody(markdown string) (updated bool, ok bool) {
	prefix, _, suffix, ok := partition(issue.Body, beginWorkMarker, endWorkMarker)
	if !ok {
		return false, false
	}

	newBody := prefix + "\n" + markdown + suffix
	if newBody == issue.Body {
		return false, true
	}

	issue.Body = newBody
	return true, true
}
