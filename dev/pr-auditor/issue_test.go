package main

import (
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/stretchr/testify/assert"
)

func TestGenerateExceptionIssue(t *testing.T) {
	payload := EventPayload{
		Repository:  RepositoryPayload{FullName: "bobheadxi/robert"},
		PullRequest: PullRequestPayload{URL: "https://bobheadxi.dev", MergedBy: UserPayload{Login: "robert"}},
	}
	tests := []struct {
		name    string
		payload EventPayload
		result  checkResult
		want    *github.IssueRequest
	}{{
		name: "not reviewed, not planned",
		result: checkResult{
			Reviewed: false,
		},
		want: &github.IssueRequest{
			Assignees: &[]string{"robert"},
			Labels:    &[]string{"exception/review", "exception/test-plan", "bobheadxi/robert"},
		},
	}, {
		name: "not reviewed, planned",
		result: checkResult{
			Reviewed: false,
			TestPlan: "A plan!",
		},
		want: &github.IssueRequest{
			Assignees: &[]string{"robert"},
			Labels:    &[]string{"exception/review", "bobheadxi/robert"},
		},
	}, {
		name: "not planned, reviewed",
		result: checkResult{
			Reviewed: true,
		},
		want: &github.IssueRequest{
			Assignees: &[]string{"robert"},
			Labels:    &[]string{"exception/test-plan", "bobheadxi/robert"},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateExceptionIssue(&payload, &tt.result)
			t.Log(got.GetTitle(), "\n", got.GetBody())
			assert.Equal(t, tt.want.GetAssignees(), got.GetAssignees())
			assert.Equal(t, tt.want.GetLabels(), got.GetLabels())
		})
	}
}
