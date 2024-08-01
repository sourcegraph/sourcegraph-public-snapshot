package publicrestapi

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertRountripJson(t *testing.T, jsonStr, expectedJSON string) {
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
	assert.Equal(t, expectedJSON, string(serializedJSON))
}

func TestDecodeCreateChatCompletionRequest(t *testing.T) {
	assertRountripJson(t, `{
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
    }`,
		`{
    "messages": [
        {
            "content": {
                "Parts": [
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
        "Stop": [
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
}`)
}

func TestDecodeCreateChatCompletionRequestStringContent(t *testing.T) {
	assertRountripJson(t, `{
        "model": "gpt-4o-mini-2024-07-18",
        "messages": [
            {
                "role": "user",
                "content": "string message"
            }
        ]
    }`,
		`{
    "messages": [
        {
            "content": {
                "Parts": [
                    {
                        "type": "text",
                        "text": "string message"
                    }
                ]
            },
            "role": "user"
        }
    ],
    "model": "gpt-4o-mini-2024-07-18"
}`)

}
func TestDecodeCreateChatCompletionRequestStopStringContent(t *testing.T) {
	assertRountripJson(t, `{
        "model": "gpt-4o-mini-2024-07-18",
        "messages": [],
		"stop": "foobar"
    }`,
		`{
    "messages": [],
    "model": "gpt-4o-mini-2024-07-18",
    "stop": {
        "Stop": [
            "foobar"
        ]
    }
}`)
}

func TestResponseFormatUnmarshal(t *testing.T) {
	var actual ResponseFormat
	err := json.Unmarshal([]byte(`{"type": "json"}`), &actual)
	assert.Equal(t, err, fmt.Errorf("invalid ResponseFormatType: json"))
	err = json.Unmarshal([]byte(`{"type": "json_object"}`), &actual)
	assert.NoError(t, err)
	assert.Equal(t, actual, ResponseFormat{Type: ResponseFormatTypeJSONObject})
	err = json.Unmarshal([]byte(`{"type": "text"}`), &actual)
	assert.NoError(t, err)
	assert.Equal(t, actual, ResponseFormat{Type: ResponseFormatTypeText})
}
