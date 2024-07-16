package completions

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func testAPIProviderAWSBedrock(t *testing.T, infra *apiProviderTestInfra) {

	// getValidTestData returns a valid test scenario for the AWS Bedrock LLM provider.
	getValidTestData := func() completionsRequestTestData {
		initialSiteConfig := schema.SiteConfiguration{
			CodyEnabled:                  pointers.Ptr(true),
			CodyPermissions:              pointers.Ptr(false),
			CodyRestrictUsersFeatureFlag: pointers.Ptr(false),

			// Use a config WITHOUT Provisioned Throughput.
			Completions: &schema.Completions{
				Provider: string(conftypes.CompletionsProviderNameAWSBedrock),

				AccessToken: "<ACCESS_KEY_ID>:<SECRET_ACCESS_KEY>:<SESSION_TOKEN>",
				Endpoint:    "us-west-2",

				ChatModel:       "anthropic.claude-3-opus-20240229-v1:0",
				CompletionModel: "anthropic.claude-instant-v1",
				FastChatModel:   "anthropic.claude-3-opus-20240229-v1_FAST:0",
			},
		}

		// Request from the end user.
		userCompletionRequest := types.CodyCompletionRequestParameters{
			CompletionRequestParameters: types.CompletionRequestParameters{
				Messages: []types.Message{
					{
						Speaker: "human",
						Text:    "please make this code better",
					},
				},
				Stream: pointers.Ptr(false),
			},
		}

		// The request made to AWS Bedrock. The LLM model to use isn't included in the payload,
		// instead it is embedded in the request URL.
		wantRequestToBedrock := map[string]any{
			"anthropic_version": "bedrock-2023-05-31",
			// This is hard-coded in our Bedrock client, as the default if
			// requestParams.MaxTokensToSample is unset.
			"max_tokens": 300,
			"messages": []map[string]any{
				{
					"role": "user",
					"content": []map[string]any{
						{
							"type": "text",
							"text": "please make this code better",
						},
					},
				},
			},
		}
		const wantRequestToBedrockPath = "/model/anthropic.claude-3-opus-20240229-v1:0/invoke"

		responseFromBedrock := map[string]any{
			"content": []map[string]string{
				{
					"speak": "user",
					"text":  "A rewrite in Rust is the only option.",
				},
			},
			"usage": map[string]int{
				"input_token":   100,
				"output_tokens": 200,
			},
			"stop_reason": "max_tokens",
		}

		wantCompletionsResponse := types.CompletionResponse{
			Completion: "A rewrite in Rust is the only option.",
			StopReason: "max_tokens",
			Logprobs:   nil,
		}

		return completionsRequestTestData{
			SiteConfig:                   initialSiteConfig,
			UserCompletionRequest:        userCompletionRequest,
			WantRequestToLLMProvider:     wantRequestToBedrock,
			WantRequestToLLMProviderPath: wantRequestToBedrockPath,
			ResponseFromLLMProvider:      responseFromBedrock,
			WantCompletionResponse:       wantCompletionsResponse,
		}
	}

	t.Run("TestDataIsValid", func(t *testing.T) {
		// Just confirm that the stock test data works as expected,
		// without any test-specific modifications.
		data := getValidTestData()
		runCompletionsTest(t, infra, data)
	})

	t.Run("ProvisionedThroughput", func(t *testing.T) {
		const (
			priovisionedThroughputARN = "arn:aws:bedrock:us-west-2:012345678901:provisioned-model/abcdefghijkl"
			// The chat model encodes the model name and provisioned throughput ARN.
			chatModelInConfig = "anthropic.claude-3-haiku-20240307-v1:0-100k/" + priovisionedThroughputARN
			// Just a regular model name.
			fastChatModelInConfig = "anthropic.claude-v2-fastchat"

			accessTokenInConfig = "<ACCESS_KEY_ID>:<SECRET_ACCESS_KEY>:<SESSION_TOKEN>"
			endpointInConfig    = "https://vpce-0a10b2345cd67e89f-abc0defg.bedrock-runtime.us-west-2.vpce.amazonaws.com"
		)

		getProvisionedThroughputSiteConfig := func() *schema.Completions {
			return &schema.Completions{
				Provider:    string(conftypes.CompletionsProviderNameAWSBedrock),
				AccessToken: accessTokenInConfig,
				Endpoint:    endpointInConfig,

				ChatModel:       chatModelInConfig,
				CompletionModel: "unused",
				FastChatModel:   fastChatModelInConfig,
			}
		}

		t.Run("Chat", func(t *testing.T) {
			data := getValidTestData()
			data.SiteConfig.Completions = getProvisionedThroughputSiteConfig()

			// The chat model is using provisioned throughput, so the
			// URLs are different.
			data.WantRequestToLLMProviderPath = "/model/arn:aws:bedrock:us-west-2:012345678901:provisioned-model/abcdefghijkl/invoke"

			runCompletionsTest(t, infra, data)
		})

		t.Run("FastChat", func(t *testing.T) {
			data := getValidTestData()
			data.SiteConfig.Completions = getProvisionedThroughputSiteConfig()

			data.UserCompletionRequest.Fast = true

			// The fast chat model does not have provisioned throughput, and
			// so the request path to bedrock just has the model's name. (No ARN.)
			data.WantRequestToLLMProviderPath = "/model/anthropic.claude-v2-fastchat/invoke"

			runCompletionsTest(t, infra, data)
		})
	})
}
