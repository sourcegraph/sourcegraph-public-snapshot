package publicrestapi

import (
	"encoding/json"
	"testing"
)

func TestDecodeCreateChatCompletionRequest(t *testing.T) {
	jsonStr := `{
        "model": "gpt-4o-mini-2024-07-18",
        "messages": [
            {
                "role": "user",
                "content": [
                    {
                        "type": "text",
                        "text": "Respond with \"yes\" and nothing else"
                    }
                ]
            }
        ],
        "temperature": 1,
        "max_tokens": 256,
        "top_p": 1,
        "frequency_penalty": 0,
        "presence_penalty": 0
    }`

	var req CreateChatCompletionRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	// Assert the decoded values
	if req.Model != "gpt-4o-mini-2024-07-18" {
		t.Errorf("Expected model to be 'gpt-4o-mini-2024-07-18', got '%s'", req.Model)
	}

	if len(req.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(req.Messages))
	}

	msg := req.Messages[0]
	if msg.Role != "user" {
		t.Errorf("Expected role to be 'user', got '%s'", msg.Role)
	}

	content, ok := msg.Content.([]interface{})
	if !ok {
		t.Fatalf("Expected content to be a slice of interfaces")
	}

	if len(content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(content))
	}

	contentMap, ok := content[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected content item to be a map")
	}

	if contentMap["type"] != "text" {
		t.Errorf("Expected content type to be 'text', got '%v'", contentMap["type"])
	}

	if contentMap["text"] != "Respond with \"yes\" and nothing else" {
		t.Errorf("Expected content text to be 'Respond with \"yes\" and nothing else', got '%v'", contentMap["text"])
	}

	if *req.Temperature != 1 {
		t.Errorf("Expected temperature to be 1, got %f", *req.Temperature)
	}

	if *req.MaxTokens != 256 {
		t.Errorf("Expected max_tokens to be 256, got %d", req.MaxTokens)
	}

	if *req.TopP != 1 {
		t.Errorf("Expected top_p to be 1, got %f", *req.TopP)
	}

	if *req.FrequencyPenalty != 0 {
		t.Errorf("Expected frequency_penalty to be 0, got %f", *req.FrequencyPenalty)
	}

	if *req.PresencePenalty != 0 {
		t.Errorf("Expected presence_penalty to be 0, got %f", *req.PresencePenalty)
	}
}

func TestDecodeCreateChatCompletionRequestStringContent(t *testing.T) {
	jsonStr := `{
        "model": "gpt-4o-mini-2024-07-18",
        "messages": [
            {
                "role": "user",
                "content": "string message"
            }
        ]
    }`

	var req CreateChatCompletionRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(req.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(req.Messages))
	}

	message := req.Messages[0]
	if message.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", message.Role)
	}

	content, ok := message.Content.(string)
	if !ok {
		t.Fatalf("Expected content to be string, got %T", message.Content)
	}

	if content != "string message" {
		t.Errorf("Expected content 'string message', got '%s'", content)
	}
}
