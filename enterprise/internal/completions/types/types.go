package types

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const HUMAN_MESSAGE_SPEAKER = "human"
const ASISSTANT_MESSAGE_SPEAKER = "assistant"

type Message struct {
	Speaker string `json:"speaker"`
	Text    string `json:"text"`
}

func (m Message) IsValidSpeaker() bool {
	return m.Speaker == HUMAN_MESSAGE_SPEAKER || m.Speaker == ASISSTANT_MESSAGE_SPEAKER
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

type CompletionRequestParameters struct {
	// Prompt exists only for backwards compatibility. Do not use it in new
	// implementations. It will be removed once we are reasonably sure 99%
	// of VSCode extension installations are upgraded to a new Cody version.
	Prompt            string    `json:"prompt"`
	Messages          []Message `json:"messages"`
	MaxTokensToSample int       `json:"maxTokensToSample,omitempty"`
	Temperature       float32   `json:"temperature,omitempty"`
	StopSequences     []string  `json:"stopSequences,omitempty"`
	TopK              int       `json:"topK,omitempty"`
	TopP              float32   `json:"topP,omitempty"`
	Model             string    `json:"model,omitempty"`
}

type CompletionResponse struct {
	Completion string `json:"completion"`
	StopReason string `json:"stopReason"`
}

type SendCompletionEvent func(event CompletionResponse) error

type CompletionsFeature string

const (
	CompletionsFeatureChat CompletionsFeature = "chat_completions"
	CompletionsFeatureCode CompletionsFeature = "code_completions"
)

func (b CompletionsFeature) IsValid() bool {
	switch b {
	case CompletionsFeatureChat,
		CompletionsFeatureCode:
		return true
	}
	return false
}

type CompletionsClient interface {
	Stream(context.Context, CompletionsFeature, CompletionRequestParameters, SendCompletionEvent) error
	Complete(context.Context, CompletionsFeature, CompletionRequestParameters) (*CompletionResponse, error)
}

type RateLimitExceededError struct {
	Feature    CompletionsFeature
	Limit      int
	Used       int
	RetryAfter time.Time
}

func (e RateLimitExceededError) Error() string {
	return fmt.Sprintf("you exceeded the rate limit for %s, only %d requests are allowed per day at the moment to ensure the service stays functional. Current usage: %d. Retry after %s", e.Feature, e.Limit, e.Used, e.RetryAfter.Truncate(time.Second))
}

func (e RateLimitExceededError) WriteHTTPResponse(w http.ResponseWriter) {
	w.Header().Set("x-ratelimit-limit", strconv.Itoa(e.Limit))
	w.Header().Set("x-ratelimit-remaining", strconv.Itoa(max(e.Limit-e.Used, 0)))
	w.Header().Set("retry-after", e.RetryAfter.Format(time.RFC1123))
	http.Error(w, e.Error(), http.StatusTooManyRequests)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
