package main

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/go-github/v41/github"
)

type acceptanceResult struct {
	// Checked indicates the acceptance checklist checkbox was checked.
	Checked bool
	// Reviewed indicates that *any* review has been made on the PR.
	Reviewed bool
	// Explanation is the content provided after the acceptance checklist checkbox.
	Explanation string
	// Error indicating any issue that might have occured during the check.
	Error error
}

func (r acceptanceResult) Explained() bool {
	return r.Explanation != ""
}

var markdownCommentRegexp = regexp.MustCompile("<!--((.|\n)*?)-->(\n)*")

func checkAcceptance(ctx context.Context, ghc *github.Client, payload *EventPayload) acceptanceResult {
	var err error
	pr := payload.PullRequest

	// Reviewed can be inferred from payload, but if not reviewed we double-check through
	// the GitHub API
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
		acceptanceHeader        = "## Acceptance checklist"
		acceptanceChecklistItem = "I have gone through the [acceptance checklist]"
	)
	sections := strings.Split(pr.Body, acceptanceHeader)
	if len(sections) < 2 {
		return acceptanceResult{
			Checked:  false,
			Reviewed: reviewed,
			Error:    err,
		}
	}
	acceptanceSection := sections[1]
	if strings.Contains(acceptanceSection, acceptanceChecklistItem) && strings.Contains(acceptanceSection, "- [x] ") {
		return acceptanceResult{
			Checked:  true,
			Reviewed: reviewed,
			Error:    err,
		}
	}
	acceptanceLines := strings.Split(acceptanceSection, "\n")
	var explanation []string
	for _, l := range acceptanceLines {
		line := strings.TrimSpace(l)
		if !strings.Contains(line, acceptanceChecklistItem) {
			explanation = append(explanation, line)
		}
	}

	// Merge into single string
	fullExplanation := strings.Join(explanation, "\n")
	// Remove comments
	fullExplanation = markdownCommentRegexp.ReplaceAllString(fullExplanation, "")
	// Remove whitespace
	fullExplanation = strings.TrimSpace(fullExplanation)
	return acceptanceResult{
		Checked:     false,
		Reviewed:    reviewed,
		Explanation: fullExplanation,
		Error:       err,
	}
}
