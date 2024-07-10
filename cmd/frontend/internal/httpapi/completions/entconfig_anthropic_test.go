package completions

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func testAPIProviderAnthropic(t *testing.T, infra *apiProviderTestInfra) {
	getValidTestData := func() completionsRequestTestData {
		// Use the default completions configuration, which will use Claude for chat.
		siteConfig := schema.SiteConfiguration{
			CodyEnabled:                  pointers.Ptr(true),
			CodyPermissions:              pointers.Ptr(false),
			CodyRestrictUsersFeatureFlag: pointers.Ptr(false),

			// LicenseKey is required in order to use Cody, but other than
			// that we don't provide any "completions" configuration.
			// This will default to Anthropic models.
			LicenseKey:  "license-key",
			Completions: nil,
		}

		initialCompletionRequest := types.CodyCompletionRequestParameters{
			CompletionRequestParameters: types.CompletionRequestParameters{
				// Model is unset. We expect it to default to the site config's ChatModel.
				// Which based on the `Completions: nil` line above, will be "claude-3-sonnet-20240229".
				Messages: []types.Message{
					{
						Speaker: "human",
						Text:    "please make this code better",
					},
				},
				Stream: pointers.Ptr(false),
			},
		}

		// Anthropic-specific request object we expect to see sent to Cody Gateway.
		// See `anthropicRequestParameters`.
		outboundAnthropicRequest := map[string]any{
			"model": "claude-3-sonnet-20240229",
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

		// Stock response we would receive from Anthropic.
		//
		// The expected output is found also defined in the Anthropic completion provider codebase,
		// as `anthropicNonStreamingResponse`.` But it's easier to keep those types unexported.
		inboundAnthropicResponse := map[string]any{
			"content": []map[string]string{
				{
					"speak": "user",
					"text":  "you should totally rewrite it in Rust!",
				},
			},
			"usage": map[string]int{
				"input_token":   100,
				"output_tokens": 200,
			},
			"stop_reason": "max_tokens",
		}

		wantCompletionsResponse := types.CompletionResponse{
			Completion: "you should totally rewrite it in Rust!",
			StopReason: "max_tokens",
			Logprobs:   nil,
		}

		return completionsRequestTestData{
			SiteConfig:                   siteConfig,
			UserCompletionRequest:        initialCompletionRequest,
			WantRequestToLLMProvider:     outboundAnthropicRequest,
			WantRequestToLLMProviderPath: "/v1/completions/anthropic-messages",
			ResponseFromLLMProvider:      inboundAnthropicResponse,
			WantCompletionResponse:       wantCompletionsResponse,
		}
	}

	t.Run("TestDataIsValid", func(t *testing.T) {
		// Just confirm that the stock test data works as expected,
		// without any test-specific modifications.
		data := getValidTestData()

		// Confirm that the test data is using the "default completions config".
		require.Nil(t, data.SiteConfig.Completions)

		runCompletionsTest(t, infra, data)
	})

	t.Run("BYOK", func(t *testing.T) {
		const (
			anthropicAPIKeyInConfig      = "secret-api-key"
			anthropicAPIEndpointInConfig = "https://byok.anthropic.com/path/from/config"

			chatModelInConfig     = "anthropic/claude-3-opus"
			codeModelInConfig     = "anthropic/claude-3-haiku"
			fastChatModelInConfig = "anthropic/fast-chat-model"
		)
		getBYOKSiteConfig := func() *schema.Completions {
			return &schema.Completions{
				Provider:    "anthropic",
				AccessToken: anthropicAPIKeyInConfig,
				Endpoint:    anthropicAPIEndpointInConfig,

				ChatModel:       chatModelInConfig,
				CompletionModel: codeModelInConfig,
				FastChatModel:   fastChatModelInConfig,
			}
		}

		t.Run("Chat", func(t *testing.T) {
			testData := getValidTestData()
			testData.SiteConfig.Completions = getBYOKSiteConfig()

			// No set all of the other things that we expect to be impacted by that.
			testData.WantRequestToLLMProvider["model"] = chatModelInConfig
			testData.WantRequestToLLMProviderPath = "/path/from/config"

			runCompletionsTest(t, infra, testData)
		})

		t.Run("CustomModel", func(t *testing.T) {
			testData := getValidTestData()
			testData.SiteConfig.Completions = getBYOKSiteConfig()

			// BUG: Cody Enterprise doesn't support using any user-provided models.
			// This confirms the current behavior which we want to change soon.
			testData.UserCompletionRequest.Model = "anthropic/latest-and-greatest"

			// Confirm the user-supplied model is ignored.
			testData.WantRequestToLLMProvider["model"] = chatModelInConfig
			testData.WantRequestToLLMProviderPath = "/path/from/config"

			runCompletionsTest(t, infra, testData)
		})

		t.Run("FastChat", func(t *testing.T) {
			testData := getValidTestData()
			testData.SiteConfig.Completions = getBYOKSiteConfig()

			// Flag the request as using the "FastChat" model.
			testData.UserCompletionRequest.Fast = true

			// No set all of the other things that we expect to be impacted by that.
			testData.WantRequestToLLMProvider["model"] = fastChatModelInConfig
			testData.WantRequestToLLMProviderPath = "/path/from/config"

			runCompletionsTest(t, infra, testData)
		})
	})
}
