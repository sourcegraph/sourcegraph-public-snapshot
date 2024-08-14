package goapi

import (
	"encoding/json"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
)

func unmarshalAndMarshalChatCompletionRequest(t *testing.T, jsonStr string) string {
	t.Helper()
	var req CreateChatCompletionRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	serializedJSON, err := json.MarshalIndent(req, "", "    ")
	if err != nil {
		t.Fatalf("Failed to marshal request to JSON: %v", err)
	}
	return string(serializedJSON)
}

func TestDecodeCreateChatCompletionRequest(t *testing.T) {
	autogold.Expect(`{
    "messages": [
        {
            "content": {
                "parts": [
                    {
                        "type": "text",
                        "text": "Respond with \"yes\" and nothing else"
                    }
                ]
            },
            "role": "user"
        }
    ],
    "model": "gpt-4o-mini-2024-07-18",
    "frequency_penalty": 0,
    "max_tokens": 256,
    "presence_penalty": 0,
    "stop": {
        "stop": [
            "\n"
        ]
    },
    "stream": true,
    "stream_options": {
        "include_usage": true
    },
    "temperature": 1,
    "top_p": 1,
    "user": "user-1234567890"
}`).Equal(t,
		unmarshalAndMarshalChatCompletionRequest(t, `{
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
		"stop": ["\n"],
		"stream": true,
		"stream_options": {
			"include_usage": true
		},
        "frequency_penalty": 0,
        "presence_penalty": 0,
		"user": "user-1234567890"
    }`))
}

func TestDecodeCreateChatCompletionRequestStringContent(t *testing.T) {
	autogold.Expect(`{
    "messages": [
        {
            "content": {
                "parts": [
                    {
                        "type": "text",
                        "text": "string message"
                    }
                ]
            },
            "role": "user"
        }
    ],
    "model": "gpt-4o-mini-2024-07-18",
    "stop": {}
}`).Equal(t,
		unmarshalAndMarshalChatCompletionRequest(t, `{
        "model": "gpt-4o-mini-2024-07-18",
        "messages": [
            {
                "role": "user",
                "content": "string message"
            }
        ]
    }`))

}
func TestDecodeCreateChatCompletionRequestStopStringContent(t *testing.T) {
	autogold.Expect(`{
    "messages": [],
    "model": "gpt-4o-mini-2024-07-18",
    "stop": {
        "stop": [
            "foobar"
        ]
    }
}`).Equal(t,
		unmarshalAndMarshalChatCompletionRequest(t, `{
        "model": "gpt-4o-mini-2024-07-18",
        "messages": [],
		"stop": "foobar"
    }`))
}

func TestResponseFormatUnmarshal(t *testing.T) {
	var actual ResponseFormat
	err := json.Unmarshal([]byte(`{"type": "json"}`), &actual)
	assert.Equal(t, "invalid ResponseFormatType: json", err.Error())
	err = json.Unmarshal([]byte(`{"type": "json_object"}`), &actual)
	assert.NoError(t, err)
	assert.Equal(t, actual, ResponseFormat{Type: ResponseFormatTypeJSONObject})
	err = json.Unmarshal([]byte(`{"type": "text"}`), &actual)
	assert.NoError(t, err)
	assert.Equal(t, actual, ResponseFormat{Type: ResponseFormatTypeText})
}
