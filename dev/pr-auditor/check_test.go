package main

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckTestPlan(t *testing.T) {
	tests := []struct {
		name            string
		bodyFile        string
		labels          []string
		baseBranch      string
		protectedBranch string
		want            checkResult
	}{
		{
			name:     "has test plan",
			bodyFile: "testdata/pull_request_body/has-plan.md",
			want: checkResult{
				Reviewed: false,
				TestPlan: "I have a plan!",
			},
		},
		{
			name:            "protected branch",
			bodyFile:        "testdata/pull_request_body/has-plan.md",
			baseBranch:      "release",
			protectedBranch: "release",
			want: checkResult{
				Reviewed:        false,
				TestPlan:        "I have a plan!",
				ProtectedBranch: true,
			},
		},
		{
			name:            "non protected branch",
			bodyFile:        "testdata/pull_request_body/has-plan.md",
			baseBranch:      "preprod",
			protectedBranch: "release",
			want: checkResult{
				Reviewed:        false,
				TestPlan:        "I have a plan!",
				ProtectedBranch: false,
			},
		},
		{
			name:     "no test plan",
			bodyFile: "testdata/pull_request_body/no-plan.md",
			want: checkResult{
				Reviewed: false,
			},
		},
		{
			name:     "complicated test plan",
			bodyFile: "testdata/pull_request_body/has-plan-fancy.md",
			want: checkResult{
				Reviewed: false,
				TestPlan: `This is a plan!
Quite lengthy

And a little complicated; there's also the following reasons:

1. A
2. B
3. C`,
			},
		},
		{
			name:     "inline test plan",
			bodyFile: "testdata/pull_request_body/inline-plan.md",
			want: checkResult{
				Reviewed: false,
				TestPlan: `This is a plan!
Quite lengthy

And a little complicated; there's also the following reasons:

1. A
2. B
3. C`,
			},
		},
		{
			name:     "no review required",
			bodyFile: "testdata/pull_request_body/no-review-required.md",
			want: checkResult{
				Reviewed: true,
				TestPlan: "I have a plan! No review required: this is a bot PR",
			},
		},
		{
			name:     "bad markdown still passes",
			bodyFile: "testdata/pull_request_body/bad-markdown.md",
			want: checkResult{
				Reviewed: true,
				TestPlan: "This is still a plan! No review required: just trust me",
			},
		},
		{
			name:     "no review required via automerge label",
			bodyFile: "testdata/pull_request_body/has-plan.md",
			labels:   []string{"automerge"},
			want: checkResult{
				Reviewed: true,
				TestPlan: "I have a plan!",
			},
		},
		{
			name:     "no review required via no-review-required label",
			bodyFile: "testdata/pull_request_body/has-plan.md",
			labels:   []string{"no-review-required"},
			want: checkResult{
				Reviewed: true,
				TestPlan: "I have a plan!",
			},
		},
		{
			name:     "no review required via automerge label but no plan",
			bodyFile: "testdata/pull_request_body/no-plan.md",
			labels:   []string{"automerge"},
			want: checkResult{
				Reviewed: false,
			},
		},
		{
			name:     "no review required via no-review-required label but no plan",
			bodyFile: "testdata/pull_request_body/no-plan.md",
			labels:   []string{"no-review-required"},
			want: checkResult{
				Reviewed: false,
			},
		},
		{
			name:     "no review required but with the wrong label",
			bodyFile: "testdata/pull_request_body/has-plan.md",
			labels:   []string{"random-label"},
			want: checkResult{
				Reviewed: false,
				TestPlan: "I have a plan!",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := os.ReadFile(tt.bodyFile)
			require.NoError(t, err)

			payload := &EventPayload{
				PullRequest: PullRequestPayload{
					Body: string(body),
				},
			}
			if tt.labels != nil {
				payload.PullRequest.Labels = make([]Label, len(tt.labels))
				for i, label := range tt.labels {
					payload.PullRequest.Labels[i] = Label{Name: label}
				}
			}
			checkOpts := checkOpts{
				ValidateReviews: false,
			}

			if tt.baseBranch != "" && tt.protectedBranch != "" {
				payload.PullRequest.Base = RefPayload{Ref: tt.baseBranch}
				checkOpts.ProtectedBranch = tt.protectedBranch
			}

			got := checkPR(context.Background(), nil, payload, checkOpts)
			assert.Equal(t, tt.want.HasTestPlan(), got.HasTestPlan())
			t.Log("got.TestPlan: ", got.TestPlan)
			if tt.want.TestPlan == "" {
				assert.Empty(t, got.TestPlan)
			} else {
				assert.True(t, strings.Contains(got.TestPlan, tt.want.TestPlan),
					cmp.Diff(got.TestPlan, tt.want.TestPlan))
			}
			assert.Equal(t, tt.want.ProtectedBranch, got.ProtectedBranch)
			assert.Equal(t, tt.want.Reviewed, got.Reviewed)
		})
	}
}
