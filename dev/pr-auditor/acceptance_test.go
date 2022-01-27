package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_checkAcceptance(t *testing.T) {
	tests := []struct {
		name     string
		bodyFile string
		want     acceptanceResult
	}{
		{
			name:     "accepted",
			bodyFile: "testdata/pull_request_body/accepted.md",
			want: acceptanceResult{
				Checked: true,
			},
		},
		{
			name:     "no explanation",
			bodyFile: "testdata/pull_request_body/no-explanation.md",
			want: acceptanceResult{
				Checked: false,
			},
		},
		{
			name:     "has explanation",
			bodyFile: "testdata/pull_request_body/explanation.md",
			want: acceptanceResult{
				Checked:     false,
				Explanation: "This is an exception explanation!",
			},
		},
		{
			name:     "complicated explanation",
			bodyFile: "testdata/pull_request_body/complicated-explanation.md",
			want: acceptanceResult{
				Checked: false,
				Explanation: `This is an exception explanation!
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

			got := checkAcceptance(string(body))
			assert.Equal(t, tt.want.Checked, got.Checked)
			t.Log("got.Explanation: ", got.Explanation)
			if tt.want.Explanation == "" {
				assert.Empty(t, got.Explanation)
			} else {
				assert.True(t, strings.Contains(got.Explanation, tt.want.Explanation),
					"Expected explanation:\n\n%s\n\nTo include:\n\n%s\n", got.Explanation, tt.want.Explanation)
			}
		})
	}
}
