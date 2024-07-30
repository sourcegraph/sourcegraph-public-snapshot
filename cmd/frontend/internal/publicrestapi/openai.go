package publicrestapi

import (
	"net/http"

	sglog "github.com/sourcegraph/log"
)

type ChatCompletionRequest struct {
	Model            string                  `json:"model"`
	Messages         []ChatCompletionMessage `json:"messages"`
	Temperature      *float32                `json:"temperature,omitempty"`
	TopP             *float32                `json:"top_p,omitempty"`
	N                *int                    `json:"n,omitempty"`
	Stream           *bool                   `json:"stream,omitempty"`
	Stop             []string                `json:"stop,omitempty"`
	MaxTokens        *int                    `json:"max_tokens,omitempty"`
	PresencePenalty  *float32                `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float32                `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]int          `json:"logit_bias,omitempty"`
	User             string                  `json:"user,omitempty"`
}

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   ChatCompletionUsage    `json:"usage"`
}

type ChatCompletionChoice struct {
	Index        int                   `json:"index"`
	Message      ChatCompletionMessage `json:"message"`
	FinishReason string                `json:"finish_reason"`
}

type ChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// serveOpenAIChatCompletions is a handler for the OpenAI /v1/chat/completions endpoint.
func serveOpenAIChatCompletions(logger sglog.Logger, apiHandler http.Handler) func(w http.ResponseWriter, r *http.Request) (err error) {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		panic("unimplemented")
	}
}
