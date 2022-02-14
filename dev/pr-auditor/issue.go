package main

import (
	"fmt"

	"github.com/google/go-github/v41/github"
)

func generateExceptionIssue(payload *EventPayload, result *checkResult) *github.IssueRequest {
	var (
		exceptionLabels = []string{}
		issueTitle      string
		issueBody       string
		issueAssignees  = []string{}
	)

	if !result.Reviewed {
		exceptionLabels = append(exceptionLabels, "exception/review")
	}
	if !result.HasTestPlan() {
		exceptionLabels = append(exceptionLabels, "exception/test-plan")
	}

	if !result.Reviewed {
		issueTitle = fmt.Sprintf("exception/review: PR %s#%d", payload.Repository.FullName, payload.PullRequest.Number)
		if !result.HasTestPlan() {
			issueBody = fmt.Sprintf("%s has a test plan but was not reviewed.", payload.PullRequest.URL)
		} else {
			issueBody = fmt.Sprintf("%s has no test plan and was not reviewed.", payload.PullRequest.URL)
		}
	} else if !result.HasTestPlan() {
		issueTitle = fmt.Sprintf("exception/test-plan: PR %s#%d", payload.Repository.FullName, payload.PullRequest.Number)
		issueBody = fmt.Sprintf("%s did not provide a test plan.", payload.PullRequest.URL)
	}

	user := payload.PullRequest.MergedBy.Login
	issueAssignees = append(issueAssignees, user)
	issueBody += fmt.Sprintf("\n\n@%s please comment in this issue with an explanation for this exception and close this issue.", user)

	if result.Error != nil {
		// Log the error in the issue
		issueBody += fmt.Sprintf("\n\nEncountered error when checking PR: %s", result.Error)
	}

	labels := append(exceptionLabels, payload.Repository.FullName)
	return &github.IssueRequest{
		Title:     github.String(issueTitle),
		Body:      github.String(issueBody),
		Assignees: &issueAssignees,
		Labels:    &labels,
	}
}
