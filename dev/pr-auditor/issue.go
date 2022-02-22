package main

import (
	"fmt"

	"github.com/google/go-github/v41/github"
)

const (
	testPlanDocs = "https://docs.sourcegraph.com/dev/background-information/testing_principles#test-plans"
)

func generateExceptionIssue(payload *EventPayload, result *checkResult) *github.IssueRequest {
	var (
		issueTitle      = fmt.Sprintf("%s#%d: %q", payload.Repository.FullName, payload.PullRequest.Number, payload.PullRequest.Title)
		issueBody       string
		exceptionLabels = []string{}
		issueAssignees  = []string{}
	)

	if !result.Reviewed {
		exceptionLabels = append(exceptionLabels, "exception/review")
	}
	if !result.HasTestPlan() {
		exceptionLabels = append(exceptionLabels, "exception/test-plan")
	}

	if !result.Reviewed {
		if result.HasTestPlan() {
			issueBody = fmt.Sprintf("%s %q **has a test plan** but **was not reviewed**.", payload.PullRequest.URL, payload.PullRequest.Title)
		} else {
			issueBody = fmt.Sprintf("%s %q **has no test plan** and **was not reviewed**.", payload.PullRequest.URL, payload.PullRequest.Title)
		}
	} else if !result.HasTestPlan() {
		issueBody = fmt.Sprintf("%s %q **has no test plan**.", payload.PullRequest.URL, payload.PullRequest.Title)
	}

	if !result.HasTestPlan() {
		issueBody += fmt.Sprintf("\n\nLearn more about test plans in our [testing guidelines](%s).", testPlanDocs)
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
