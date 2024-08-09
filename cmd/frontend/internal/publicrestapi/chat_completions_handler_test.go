package publicrestapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/modelconfig"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	types "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func SetSiteConfig(t *testing.T, siteConfig schema.SiteConfiguration) {
	conf.Mock(&conf.Unified{SiteConfiguration: siteConfig})
	if err := modelconfig.ResetMock(); err != nil {
		require.NoError(t, err)
	}
}

func TestChatCompletionsHandler(t *testing.T) {
	c := newTest(t)
	chatModels := c.getChatModels()
	assert.NoError(t, modelconfig.InitMock())
	assert.NoError(t, modelconfig.ResetMockWithStaticData(&types.ModelConfiguration{
		Models: chatModels,
	}))

	t.Run("/api/v1/chat/completions (400 stream=true)", func(t *testing.T) {
		rr := c.chatCompletions(t, `{
			    "model": "anthropic::unknown::claude-3-sonnet-20240229",
			    "messages": [{"role": "user", "content": "Hello"}],
				"stream": true
			}`)
		// For now, we only support non-streaming requests.
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("/api/v1/chat/completions (400 N != 1)", func(t *testing.T) {
		rr := c.chatCompletions(t, `{
			    "model": "anthropic::unknown::claude-3-sonnet-20240229",
			    "messages": [{"role": "user", "content": "Hello"}],
				"n": 2
			}`)
		// For now, we only support non-streaming requests.
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("/api/v1/chat/completions (400 role=system)", func(t *testing.T) {
		rr := c.chatCompletions(t, `{
			    "model": "anthropic::unknown::claude-3-sonnet-20240229",
			    "messages": [{"role": "system", "content": "You are a helpful assistant."}]
			}`)
		// For now, we don't support overriding the system role.
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("/api/v1/chat/completions (400 model is not modelref)", func(t *testing.T) {
		rr := c.chatCompletions(t, `{
			    "model": "anthropic/claude-3-haiku-20240307",
			    "messages": [{"role": "user", "content": "Hello"}]
			}`)
		// For now, we reject requests when the model is not using the new ModelRef format.
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		// Assert that we give a helpful error message nudging the user to use modelref instead of the old syntax.
		assert.Equal(t, "model anthropic/claude-3-haiku-20240307 is not supported (similar to anthropic::unknown::claude-3-haiku-20240307)\n", rr.Body.String())
	})

	t.Run("/api/v1/chat/completions (200 OK)", func(t *testing.T) {
		rr := c.chatCompletions(t, `{
			    "model": "anthropic::unknown::claude-3-sonnet-20240229",
			    "messages": [
			        {
			            "role": "user",
			            "content": "respond with 'yes' in all-lowercase and nothing else"
			        }
			    ],
				"stream": false,
				"max_tokens": 16,
				"logprobs": null,
				"stop": [],
				"temperature": 0.5,
				"top_p": 1,
				"top_k": 0,
				"n": 1,
				"frequency_penalty": 0,
				"presence_penalty": 0
			}`)

		if rr.Code != http.StatusOK {
			extraMessage := ""
			if rr.Code == http.StatusUnauthorized {
				extraMessage = " (to fix authorization issues, you may want to run the command 'source dev/export-http-recording-tokens.sh')"
			}
			t.Fatalf("Expected status code %d, got %d. Body: %s%s", http.StatusOK, rr.Code, rr.Body.String(), extraMessage)
		}

		var resp CreateChatCompletionResponse
		responseBytes := rr.Body.Bytes()
		err := json.Unmarshal(responseBytes, &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v. Body: %s", err, string(responseBytes))
		}

		// Default "Changed" value is time.Now().Unix(), we make it 0 for determinism here.
		resp.Created = 0

		jsonData, err := json.MarshalIndent(resp, "", "    ")
		if err != nil {
			t.Fatalf("Failed to marshal response: %v", err)
		}
		body := string(jsonData)

		// Important assertion. The old /.api/completions/stream endpoint returned text/plain.
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		autogold.Expect(`{
    "id": "chat-mocked-publicrestapi-uuid",
    "choices": [
        {
            "finish_reason": "end_turn",
            "index": 0,
            "message": {
                "content": "yes",
                "role": "assistant"
            }
        }
    ],
    "created": 0,
    "model": "anthropic::unknown::claude-3-sonnet-20240229",
    "system_fingerprint": "",
    "object": "chat.completion",
    "usage": {
        "completion_tokens": 0,
        "prompt_tokens": 0,
        "total_tokens": 0
    }
}`).Equal(t, body)
	})

	for _, model := range chatModels {
		if model.DisplayName == "starcoder" {
			// Skip starcoder because it's not a chat model even if it has the "chat" capability
			// per the /.api/modelconfig/supported-models.json endpoint. Context:
			// https://sourcegraph.slack.com/archives/C04MSD3DP5L/p1723041114247759
			continue
		}
		t.Run(fmt.Sprintf("/api/v1/chat/completions (200 OK, %s)", model.DisplayName), func(t *testing.T) {
			rr := c.chatCompletions(t, fmt.Sprintf(`{
			"model": "%s",
			"messages": [{"role": "user", "content": "respond with 'yes' in all-lowercase and nothing else"}]
		}`, model.ModelRef))

			// Only assert the status code is OK. We assert the response body in the test above.
			assert.Equal(t, http.StatusOK, rr.Code, "Failed to get chat completion for %s", model.ModelRef)
		})
	}

}
