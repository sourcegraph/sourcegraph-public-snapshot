package modelconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

func TestRedactServerSideConfig(t *testing.T) {
	testCfg := types.ModelConfiguration{
		Providers: []types.Provider{
			{
				ID: "test-provider-1",
				ServerSideConfig: &types.ServerSideProviderConfig{
					AWSBedrock: &types.AWSBedrockProviderConfig{
						AccessToken: "top-secret",
					},
				},
			},
			{
				ID: "test-provider-2",
				ServerSideConfig: &types.ServerSideProviderConfig{
					AzureOpenAI: &types.AzureOpenAIProviderConfig{
						AccessToken: "top-secret",
					},
				},
			},
		},
		Models: []types.Model{
			{
				ServerSideConfig: &types.ServerSideModelConfig{
					AWSBedrockProvisionedThroughput: &types.AWSBedrockProvisionedThroughput{
						ARN: "secret-arn",
					},
				},
			},
		},
	}

	RedactServerSideConfig(&testCfg)

	for _, provider := range testCfg.Providers {
		assert.Nil(t, provider.ServerSideConfig)
	}
	for _, model := range testCfg.Models {
		assert.Nil(t, model.ServerSideConfig)
	}
}

func TestSanitizeResourceName(t *testing.T) {
	tests := []struct {
		Input string
		Want  string
	}{
		{"something with spaces", "something_with_spaces"},
		{"{1234567890-abcdef}", "_1234567890-abcdef_"},
		{"{1234567890-abcdef}", "_1234567890-abcdef_"},

		{
			Input: "anthropic.claude-3-haiku-20240307-v1:0-100k",
			Want:  "anthropic.claude-3-haiku-20240307-v1_0-100k",
		},
		{
			// Note that here we convert AGENT to be lower case.
			Input: "arn:aws:bedrock:aws-region:111122223333:agent/AGENT12345",
			Want:  "arn_aws_bedrock_aws-region_111122223333_agent_agent12345",
		},
	}

	for _, test := range tests {
		got := SanitizeResourceName(test.Input)
		assert.Equal(t, test.Want, got)
	}
}
