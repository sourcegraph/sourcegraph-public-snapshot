package main

import (
	"testing"

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

		wantAssignees    []string
		wantLabels       []string
		wantBodyContains []string
	}{{
		name: "not reviewed, not planned",
		result: checkResult{
			Reviewed: false,
		},
		wantAssignees:    []string{"robert"},
		wantLabels:       []string{"exception/review", "exception/test-plan", "bobheadxi/robert"},
		wantBodyContains: []string{"has no test plan", "was not reviewed"},
	}, {
		name: "not reviewed, planned",
		result: checkResult{
			Reviewed: false,
			TestPlan: "A plan!",
		},
		wantAssignees:    []string{"robert"},
		wantLabels:       []string{"exception/review", "bobheadxi/robert"},
		wantBodyContains: []string{"has a test plan", "was not reviewed"},
	}, {
		name: "not planned, reviewed",
		result: checkResult{
			Reviewed: true,
		},
		wantAssignees:    []string{"robert"},
		wantLabels:       []string{"exception/test-plan", "bobheadxi/robert"},
		wantBodyContains: []string{"has no test plan"},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateExceptionIssue(&payload, &tt.result)
			t.Log(got.GetTitle(), "\n", got.GetBody())
			assert.Equal(t, tt.wantAssignees, got.GetAssignees())
			assert.Equal(t, tt.wantLabels, got.GetLabels())
			for _, content := range tt.wantBodyContains {
				assert.Contains(t, *got.Body, content)
			}
		})
	}
}
