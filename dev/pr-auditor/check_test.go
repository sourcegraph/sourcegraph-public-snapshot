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
		name     string
		bodyFile string
		want     checkResult
	}{
		{
			name:     "has test plan",
			bodyFile: "testdata/pull_request_body/has-plan.md",
			want: checkResult{
				TestPlan: "I have a plan!",
			},
		},
		{
			name:     "no test plan",
			bodyFile: "testdata/pull_request_body/no-plan.md",
			want:     checkResult{},
		},
		{
			name:     "complicated test plan",
			bodyFile: "testdata/pull_request_body/has-plan-fancy.md",
			want: checkResult{
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
				TestPlan: `This is a plan!
Quite lengthy

And a little complicated; there's also the following reasons:

1. A
2. B
3. C`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := os.ReadFile(tt.bodyFile)
			require.NoError(t, err)

			got := checkPR(context.Background(), nil, &EventPayload{
				PullRequest: PullRequestPayload{
					Body:           string(body),
					ReviewComments: 1, // Happy path
				},
			}, checkOpts{
				ValidateReviews: false,
			})
			assert.Equal(t, tt.want.HasTestPlan(), got.HasTestPlan())
			t.Log("got.Explanation: ", got.TestPlan)
			if tt.want.TestPlan == "" {
				assert.Empty(t, got.TestPlan)
			} else {
				assert.True(t, strings.Contains(got.TestPlan, tt.want.TestPlan),
					cmp.Diff(got.TestPlan, tt.want.TestPlan))
			}
			assert.True(t, got.Reviewed) // Check this is always set
		})
	}
}
