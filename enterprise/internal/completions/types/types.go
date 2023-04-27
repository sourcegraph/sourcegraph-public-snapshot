package types

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const HUMAN_MESSAGE_SPEAKER = "human"
const ASISSTANT_MESSAGE_SPEAKER = "assistant"

type Message struct {
	Speaker string `json:"speaker"`
	Text    string `json:"text"`
}

type CodeCompletionRequestParameters struct {
	Prompt            string            `json:"prompt"`
	Temperature       float64           `json:"temperature,omitempty"`
	MaxTokensToSample int               `json:"maxTokensToSample"`
	StopSequences     []string          `json:"stopSequences"`
	TopK              int               `json:"topK,omitempty"`
	TopP              float64           `json:"topP,omitempty"`
	Model             string            `json:"model"`
	Tags              map[string]string `json:"tags,omitempty"`
}

type CodeCompletionResponse struct {
	Completion string  `json:"completion"`
	Stop       *string `json:"stop"`
	StopReason string  `json:"stopReason"`
	Truncated  bool    `json:"truncated"`
	Exception  *string `json:"exception"`
	LogID      string  `json:"logID"`
}

type ChatCompletionRequestParameters struct {
	Messages          []Message `json:"messages"`
	Temperature       float32   `json:"temperature"`
	MaxTokensToSample int       `json:"maxTokensToSample"`
	TopK              int       `json:"topK"`
	TopP              float32   `json:"topP"`
}

type ChatCompletionEvent struct {
	Completion string `json:"completion"`
}

func (m Message) GetPrompt(humanPromptPrefix, assistantPromptPrefix string) (string, error) {
	var prefix string
	switch m.Speaker {
	case HUMAN_MESSAGE_SPEAKER:
		prefix = humanPromptPrefix
	case ASISSTANT_MESSAGE_SPEAKER:
		prefix = assistantPromptPrefix
	default:
		return "", errors.Newf("expected message speaker to be 'human' or 'assistant', got %s", m.Speaker)
	}

	if len(m.Text) == 0 {
		// Important: no trailing space (affects output quality)
		return prefix, nil
	}
	return fmt.Sprintf("%s %s", prefix, m.Text), nil
}

type SendCompletionEvent func(event ChatCompletionEvent) error

type CompletionsClient interface {
	Stream(ctx context.Context, requestParams ChatCompletionRequestParameters, sendEvent SendCompletionEvent) error
	Complete(ctx context.Context, requestParams CodeCompletionRequestParameters) (*CodeCompletionResponse, error)
}
