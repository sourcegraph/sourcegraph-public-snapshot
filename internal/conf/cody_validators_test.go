package conf

import (
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func Test_completionsConfigValidator(t *testing.T) {
	tests := []struct {
		name             string
		config           Unified
		expectedProblems Problems
	}{
		{
			name: "bedrock arn's are not valid",
			config: Unified{
				SiteConfiguration: schema.SiteConfiguration{
					Completions: &schema.Completions{
						Provider:  "aws-bedrock",
						ChatModel: "arn:aws:bedrock:us-west-2:012345678901:provisioned-model/abcdefghijkl",
					},
				},
			},
			expectedProblems: NewSiteProblems(
				fmt.Sprintf(bedrockArnMessageTemplate, "chatModel", "arn:aws:bedrock:us-west-2:012345678901:provisioned-model/abcdefghijkl"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := completionsConfigValidator(tt.config)
			if len(problems) != len(tt.expectedProblems) {
				t.Errorf("got %d problems, expected %d", len(problems), len(tt.expectedProblems))
				return
			}

			// Check each expected problem is present in actual problems
			problemsMap := make(map[string]bool)
			for _, p := range problems {
				problemsMap[p.String()] = true
			}
			for _, expected := range tt.expectedProblems {
				if _, ok := problemsMap[expected.String()]; !ok {
					t.Errorf("expected problem not present: %s", expected)
				}
			}
		})
	}
}
