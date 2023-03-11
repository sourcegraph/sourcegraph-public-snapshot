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

type CompletionRequestParameters struct {
	Messages          []Message `json:"messages"`
	Temperature       float32   `json:"temperature"`
	MaxTokensToSample int       `json:"maxTokensToSample"`
	TopK              int       `json:"topK"`
	TopP              float32   `json:"topP"`
}

type CompletionEvent struct {
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

type SendCompletionEvent func(event CompletionEvent) error

type CompletionStreamClient interface {
	Stream(ctx context.Context, requestParams CompletionRequestParameters, sendEvent SendCompletionEvent) error
}
