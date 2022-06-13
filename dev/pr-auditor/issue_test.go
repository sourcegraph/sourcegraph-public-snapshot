package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateExceptionIssue(t *testing.T) {
	payload := EventPayload{
		Repository: RepositoryPayload{FullName: "bobheadxi/robert"},
		PullRequest: PullRequestPayload{
			Title:    "some pull request",
			URL:      "https://bobheadxi.dev",
			MergedBy: UserPayload{Login: "robert"},
		},
	}
	privatePayload := payload
	privatePayload.Repository.Private = true

	protectedPayload := payload
	protectedPayload.PullRequest.Base = RefPayload{Ref: "release"}

	tests := []struct {
		name              string
		payload           EventPayload
		result            checkResult
		additionalContext string

		wantAssignees    []string
		wantLabels       []string
		wantBodyContains []string
		wantBodyExcludes []string
	}{{
		name:    "not reviewed, not planned",
		payload: payload,
		result: checkResult{
			Reviewed: false,
		},
		wantAssignees:    []string{"robert"},
		wantLabels:       []string{"exception/review", "exception/test-plan", "bobheadxi/robert"},
		wantBodyContains: []string{"some pull request", "has no test plan", "was not reviewed"},
		wantBodyExcludes: []string{"protected"},
	}, {
		name:    "not reviewed, planned",
		payload: payload,
		result: checkResult{
			Reviewed: false,
			TestPlan: "A plan!",
		},
		wantAssignees:    []string{"robert"},
		wantLabels:       []string{"exception/review", "bobheadxi/robert"},
		wantBodyContains: []string{"some pull request", "has a test plan", "was not reviewed"},
		wantBodyExcludes: []string{"protected"},
	}, {
		name:    "not planned, reviewed",
		payload: payload,
		result: checkResult{
			Reviewed: true,
		},
		wantAssignees:    []string{"robert"},
		wantLabels:       []string{"exception/test-plan", "bobheadxi/robert"},
		wantBodyContains: []string{"some pull request", "has no test plan"},
		wantBodyExcludes: []string{"protected"},
	}, {
		name:             "private repo, not planned, reviewed",
		payload:          privatePayload,
		result:           checkResult{},
		wantAssignees:    []string{"robert"},
		wantLabels:       []string{"exception/review", "exception/test-plan", "bobheadxi/robert"},
		wantBodyExcludes: []string{"some pull request", "protected"},
	}, {
		name:             "reviewed, planned but protected",
		payload:          protectedPayload,
		result:           checkResult{ProtectedBranch: true},
		wantAssignees:    []string{"robert"},
		wantLabels:       []string{"exception/review", "exception/test-plan", "exception/protected-branch", "bobheadxi/robert"},
		wantBodyContains: []string{"\"release\" is protected"},
	}, {
		name:              "reviewed, planned but protected",
		payload:           protectedPayload,
		additionalContext: "please use preprod branch",
		result:            checkResult{ProtectedBranch: true},
		wantAssignees:     []string{"robert"},
		wantLabels:        []string{"exception/review", "exception/test-plan", "exception/protected-branch", "bobheadxi/robert"},
		wantBodyContains:  []string{"please use preprod branch"},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateExceptionIssue(&tt.payload, &tt.result, tt.additionalContext)
			t.Log(got.GetTitle(), "\n", got.GetBody())
			assert.Equal(t, tt.wantAssignees, got.GetAssignees())
			assert.Equal(t, tt.wantLabels, got.GetLabels())
			for _, content := range tt.wantBodyContains {
				assert.Contains(t, *got.Body, content, "body does not contain expected strings")
			}
			for _, content := range tt.wantBodyExcludes {
				assert.NotContains(t, *got.Body, content, "body contains unexpected strings")
			}
		})
	}
}
