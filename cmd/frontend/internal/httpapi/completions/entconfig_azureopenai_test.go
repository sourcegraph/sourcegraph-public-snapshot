package completions

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func testAPIProviderAzureOpenAI(t *testing.T, infra *apiProviderTestInfra) {

	// getValidTestData returns a valid test scenario for the AzureOpenAI LLM provider.
	getValidTestData := func() completionsRequestTestData {
		initialSiteConfig := schema.SiteConfiguration{
			CodyEnabled:                  pointers.Ptr(true),
			CodyPermissions:              pointers.Ptr(false),
			CodyRestrictUsersFeatureFlag: pointers.Ptr(false),

			Completions: &schema.Completions{
				Provider: string(conftypes.CompletionsProviderNameAzureOpenAI),

				AccessToken: "horse-battery-staple",
				Endpoint:    "https://endpoint-from-config.example.com",

				ChatModel:       "deployment-id-1_chat",
				CompletionModel: "deployment-id-2_completion",
				FastChatModel:   "deployment-id-3_fastchat",

				// Configure hooks to provide a friendly name for the opaque model names.
				AzureChatModel:       "gpt-35-turbo_chat",
				AzureCompletionModel: "gpt-35-turbo_completion",
				// BUG: There is no way to express the FastChat model used.
				User: "azure-user-string",
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

		// azopenai.ChatCompletionsOptions. Although the actual JSON format is burried
		// deep in the Azure SDK code.
		wantRequestToAzureOpenAI := map[string]any{
			"max_tokens": 0,
			"messages": []map[string]any{
				map[string]any{
					"content": "please make this code better",
					"role":    "user",
				},
			},
			"model":       "deployment-id-1_chat",
			"n":           1,
			"temperature": 0,
			"top_p":       0,
			"user":        "azure-user-string",
		}
		const wantRequestToAzureOpenAIPath = "/openai/deployments/deployment-id-1_chat/chat/completions"

		// azopenai.GetChatCompletionsResponse. Note that the way the SDK
		// parses the HTTP response, the outermost "ChatCompletions" field
		// is omitted. And all the field names are remapped to snake_case.
		responseFromAzureOpenAI := map[string]any{
			"id": "azure-id",
			"choices": []map[string]any{
				map[string]any{
					"finish_reason": "stop", // CompletionsFinishReasonStop
					"index":         100,
					"log_probs":     nil,
					"delta": map[string]any{
						"content": "A rewrite in Rust is the only option.",
					},
				},
			},
			"usage": map[string]any{
				"completion_tokens": 1,
				"prompt_tokens":     2,
				"total_tokens":      3,
			},
		}

		wantCompletionsResponse := types.CompletionResponse{
			Completion: "A rewrite in Rust is the only option.",
			StopReason: "stop",
			Logprobs:   nil,
		}

		return completionsRequestTestData{
			SiteConfig:                   initialSiteConfig,
			UserCompletionRequest:        userCompletionRequest,
			WantRequestToLLMProvider:     wantRequestToAzureOpenAI,
			WantRequestToLLMProviderPath: wantRequestToAzureOpenAIPath,
			ResponseFromLLMProvider:      responseFromAzureOpenAI,
			WantCompletionResponse:       wantCompletionsResponse,
		}
	}

	t.Run("TestDataIsValid", func(t *testing.T) {
		// Just confirm that the stock test data works as expected,
		// without any test-specific modifications.
		data := getValidTestData()
		runCompletionsTest(t, infra, data)
	})

	t.Run("AzureSpecificConfigFlags", func(t *testing.T) {
		// Customize the site configuration data and verify
		// that the Azure-specific values are actually used.
		data := getValidTestData()

		compConfig := data.SiteConfig.Completions
		compConfig.User = "some-other-azure-user"

		data.WantRequestToLLMProvider["user"] = "some-other-azure-user"

		runCompletionsTest(t, infra, data)
	})
}
