package publicrestapi

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
)

func TestAPI(t *testing.T) {
	c := newTest(t, "chat_completions")

	t.Run("/api/v1/chat/completions (400 stream=true)", func(t *testing.T) {
		rr := c.chatCompletions(t, `{
			    "model": "anthropic/claude-3-haiku-20240307",
			    "messages": [{"role": "user", "content": "Hello"}],
				"stream": true
			}`)
		// For now, we only support non-streaming requests.
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("/api/v1/chat/completions (400 role=system)", func(t *testing.T) {
		rr := c.chatCompletions(t, `{
			    "model": "anthropic/claude-3-haiku-20240307",
			    "messages": [{"role": "system", "content": "You are a helpful assistant."}]
			}`)
		// For now, we don't support overriding the system role.
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("/api/v1/chat/completions (200 OK)", func(t *testing.T) {
		rr := c.chatCompletions(t, `{
			    "model": "anthropic/claude-3-haiku-20240307",
			    "messages": [{"role": "user", "content": "respond with 'yes' in all-lowercase and nothing else"}]
			}`)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status code %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
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
    "model": "anthropic/claude-3-haiku-20240307",
    "system_fingerprint": "",
    "object": "chat.completion",
    "usage": {
        "completion_tokens": 0,
        "prompt_tokens": 0,
        "total_tokens": 0
    }
}`).Equal(t, body)
	})
}
