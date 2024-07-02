package completions

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func testAPIProviderAnthropic(t *testing.T, infra *apiProviderTestInfra) {
	// validAnthropicRequestData bundles the messages sent between major
	// components of the Completions API.
	type validAnthropicRequestData struct {
		// InitialCompletionRequest is the request sent by the user to the
		// Sourcegraph instance.
		InitialCompletionRequest types.CodyCompletionRequestParameters

		// OutboundAnthropicRequest is the data sent from this Sourcegraph
		// instance to the API Provider (e.g. Anthropic, Cody Gateway, AWS
		// Bedrock, etc.)
		OutboundAnthropicRequest map[string]any

		// InboundAnthropicRequest is the response from the API Provider
		// to the Sourcegraph instance.
		InboundAnthropicRequest map[string]any
	}

	// getValidTestData returns a valid set of request data.
	getValidTestData := func() validAnthropicRequestData {
		initialCompletionRequest := types.CodyCompletionRequestParameters{
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

		// Anthropic-specific request object we expect to see sent to Cody Gateway.
		// See `anthropicRequestParameters`.
		outboundAnthropicRequest := map[string]any{
			"model": "claude-2.0",
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
		inboundAnthropicRequest := map[string]any{
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

		return validAnthropicRequestData{
			InitialCompletionRequest: initialCompletionRequest,
			OutboundAnthropicRequest: outboundAnthropicRequest,
			InboundAnthropicRequest:  inboundAnthropicRequest,
		}
	}

	t.Run("WithDefaultConfig", func(t *testing.T) {
		infra.SetSiteConfig(schema.SiteConfiguration{
			CodyEnabled:                  pointers.Ptr(true),
			CodyPermissions:              pointers.Ptr(false),
			CodyRestrictUsersFeatureFlag: pointers.Ptr(false),

			// LicenseKey is required in order to use Cody, but other than
			// that we don't provide any "completions" configuration.
			// This will default to Anthropic models.
			LicenseKey:  "license-key",
			Completions: nil,
		})

		// Confirm that the default configuration `Completions: nil` will use
		// Cody Gateway as the LLM API Provider for the Anthropic models.
		t.Run("ViaCodyGateway", func(t *testing.T) {
			// The Model isn't included in the CompletionRequestParameters, so we have the getModelFn callback
			// return claude-2. The Site Configuration will then route this to Cody Gateway (and not BYOK Anthropic),
			// and we sanity check the request to Cody Gateway matches what is expected, and we serve a valid response.
			infra.PushGetModelResult("anthropic/claude-2", nil)

			// Generate some basic test data and confirm that the completions handler
			// code works as expected.
			testData := getValidTestData()

			// Register our hook to verify Cody Gateway got called with
			// the requested data.
			infra.AssertCodyGatewayReceivesRequestWithResponse(
				t, assertLLMRequestOptions{
					WantRequestPath: "/v1/completions/anthropic-messages",
					WantRequestObj:  &testData.OutboundAnthropicRequest,
					OutResponseObj:  &testData.InboundAnthropicRequest,
				})

			status, responseBody := infra.CallChatCompletionAPI(t, testData.InitialCompletionRequest)

			assert.Equal(t, http.StatusOK, status)
			infra.AssertCompletionsResponse(t, responseBody, types.CompletionResponse{
				Completion: "you should totally rewrite it in Rust!",
				StopReason: "max_tokens",
				Logprobs:   nil,
			})
		})
	})

	t.Run("ViaBYOK", func(t *testing.T) {
		const (
			anthropicAPIKeyInConfig      = "secret-api-key"
			anthropicAPIEndpointInConfig = "https://byok.anthropic.com/path/from/config"
			chatModelInConfig            = "anthropic/claude-3-opus"
			codeModelInConfig            = "anthropic/claude-3-haiku"
		)

		infra.SetSiteConfig(schema.SiteConfiguration{
			CodyEnabled:                  pointers.Ptr(true),
			CodyPermissions:              pointers.Ptr(false),
			CodyRestrictUsersFeatureFlag: pointers.Ptr(false),

			// LicenseKey is required in order to use Cody.
			LicenseKey: "license-key",
			Completions: &schema.Completions{
				Provider:    "anthropic",
				AccessToken: anthropicAPIKeyInConfig,
				Endpoint:    anthropicAPIEndpointInConfig,

				ChatModel:       chatModelInConfig,
				CompletionModel: codeModelInConfig,
			},
		})

		t.Run("ChatModel", func(t *testing.T) {
			// Start with the stock test data, but customize it to reflect
			// what we expect to see based on the site configuration.
			testData := getValidTestData()
			testData.OutboundAnthropicRequest["model"] = "anthropic/claude-3-opus"

			// Register our hook to verify Cody Gateway got called with
			// the requested data.
			infra.AssertGenericExternalAPIRequestWithResponse(
				t, assertLLMRequestOptions{
					WantRequestPath: "/path/from/config",
					WantRequestObj:  &testData.OutboundAnthropicRequest,
					OutResponseObj:  &testData.InboundAnthropicRequest,
					WantHeaders: map[string]string{
						// Yes, Anthropic's API uses "X-Api-Key" rather than the "Authorization" header. 🤷
						"X-Api-Key": anthropicAPIKeyInConfig,
					},
				})

			infra.PushGetModelResult(chatModelInConfig, nil)
			status, responseBody := infra.CallChatCompletionAPI(t, testData.InitialCompletionRequest)

			assert.Equal(t, http.StatusOK, status)
			infra.AssertCompletionsResponse(t, responseBody, types.CompletionResponse{
				Completion: "you should totally rewrite it in Rust!",
				StopReason: "max_tokens",
				Logprobs:   nil,
			})
		})
	})
}
