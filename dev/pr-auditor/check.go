package main

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/go-github/v41/github"
)

type checkResult struct {
	// Reviewed indicates that *any* review has been made on the PR.
	Reviewed bool
	// TestPlan is the content provided after the acceptance checklist checkbox.
	TestPlan string
	// Error indicating any issue that might have occured during the check.
	Error error
}

func (r checkResult) HasTestPlan() bool {
	return r.TestPlan != ""
}

var markdownCommentRegexp = regexp.MustCompile("<!--((.|\n)*?)-->(\n)*")

func checkTestPlan(ctx context.Context, ghc *github.Client, payload *EventPayload) checkResult {
	pr := payload.PullRequest

	// Reviewed can be inferred from payload, but if not reviewed we double-check through
	// the GitHub API
	var err error
	reviewed := pr.ReviewComments > 0
	if !reviewed {
		repoParts := strings.Split(payload.Repository.FullName, "/")
		var reviews []*github.PullRequestReview
		// Continue, but return err later
		reviews, _, err = ghc.PullRequests.ListReviews(ctx, repoParts[0], repoParts[1], payload.PullRequest.Number, &github.ListOptions{})
		reviewed = len(reviews) > 0
	}

	// Parse aceptance data from body
	const (
		acceptanceHeader = "## Test plan"
	)
	sections := strings.Split(pr.Body, acceptanceHeader)
	if len(sections) < 2 {
		return checkResult{
			Reviewed: reviewed,
			Error:    err,
		}
	}
	acceptanceSection := sections[1]
	acceptanceLines := strings.Split(acceptanceSection, "\n")
	var explanation []string
	for _, l := range acceptanceLines {
		line := strings.TrimSpace(l)
		explanation = append(explanation, line)
	}

	// Merge into single string
	fullExplanation := strings.Join(explanation, "\n")
	// Remove comments
	fullExplanation = markdownCommentRegexp.ReplaceAllString(fullExplanation, "")
	// Remove whitespace
	fullExplanation = strings.TrimSpace(fullExplanation)
	return checkResult{
		Reviewed: reviewed,
		TestPlan: fullExplanation,
		Error:    err,
	}
}
