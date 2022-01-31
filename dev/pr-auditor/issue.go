package main

import (
	"fmt"

	"github.com/google/go-github/v41/github"
)

func generateAcceptanceAuditTrailIssue(payload *EventPayload, acceptance *acceptanceResult) (*github.IssueRequest, bool) {
	var (
		issueTitle     = fmt.Sprintf("acceptance checklist exception: PR %s#%d", payload.Repository.FullName, payload.PullRequest.Number)
		issueBody      string
		issueAssignees = []string{}
		closeIssue     bool
	)

	if acceptance.Checked && !acceptance.Reviewed {
		issueBody = fmt.Sprintf("%s went through the acceptance checklist but was not actually reviewed.", payload.PullRequest.URL)
	} else {
		issueBody = fmt.Sprintf("%s did not go through the acceptance checklist.", payload.PullRequest.URL)
	}

	if acceptance.Error != nil {
		// Log the error in the issue
		issueBody += fmt.Sprintf("\n\nEncountered error when checking acceptance: %s", acceptance.Error)
	}
	if !acceptance.Explained() {
		user := payload.PullRequest.MergedBy.Login
		issueAssignees = append(issueAssignees, user)
		issueBody += fmt.Sprintf("\n\nNo explanation was provided - @%s please comment in this issue with an explanation for this exception and close this issue.", user)
	} else {
		closeIssue = true
		issueBody += fmt.Sprintf("\n\nProvided explanation:\n\n%s", acceptance.Explanation)
	}

	return &github.IssueRequest{
		Title:     github.String(issueTitle),
		Body:      github.String(issueBody),
		Assignees: &issueAssignees,
	}, closeIssue
}
