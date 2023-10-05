package main

import (
	"fmt"

	"github.com/google/go-github/v55/github"
)

const (
	testPlanDocs = "https://docs.sourcegraph.com/dev/background-information/testing_principles#test-plans"
)

func generateExceptionIssue(payload *EventPayload, result *checkResult, additionalContext string) *github.IssueRequest {
	// 🚨 SECURITY: Do not reference other potentially sensitive fields of pull requests
	prTitle := payload.PullRequest.Title
	if payload.Repository.Private {
		prTitle = "<redacted>"
	}

	var (
		issueTitle      = fmt.Sprintf("%s#%d: %q", payload.Repository.FullName, payload.PullRequest.Number, prTitle)
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
	if result.ProtectedBranch {
		exceptionLabels = append(exceptionLabels, "exception/protected-branch")
	}

	if !result.Reviewed {
		if result.HasTestPlan() {
			issueBody = fmt.Sprintf("%s %q **has a test plan** but **was not reviewed**.", payload.PullRequest.URL, prTitle)
		} else {
			issueBody = fmt.Sprintf("%s %q **has no test plan** and **was not reviewed**.", payload.PullRequest.URL, prTitle)
		}
	} else if !result.HasTestPlan() {
		issueBody = fmt.Sprintf("%s %q **has no test plan**.", payload.PullRequest.URL, prTitle)
	}

	if !result.HasTestPlan() {
		issueBody += fmt.Sprintf("\n\nLearn more about test plans in our [testing guidelines](%s).", testPlanDocs)
	}

	if result.ProtectedBranch {
		issueBody += fmt.Sprintf("\n\nThe base branch %q is protected and should not have direct pull requests to it.", payload.PullRequest.Base.Ref)
	}

	if additionalContext != "" {
		issueBody += fmt.Sprintf("\n\n%s", additionalContext)
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
