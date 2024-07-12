package fireworks

import "github.com/sourcegraph/sourcegraph/internal/completions/types"

// fireworksRequest captures fields from https://readme.fireworks.ai/reference/createcompletion
type fireworksRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	MaxTokens   int32    `json:"max_tokens,omitempty"`
	Temperature float32  `json:"temperature,omitempty"`
	TopP        float32  `json:"top_p,omitempty"`
	N           int32    `json:"n,omitempty"`
	Stream      bool     `json:"stream,omitempty"`
	Echo        bool     `json:"echo,omitempty"`
	Stop        []string `json:"stop,omitempty"`
	Logprobs    *uint8   `json:"logprobs,omitempty"`
}

// fireworksChatRequest captures fields from https://readme.fireworks.ai/reference/createchatcompletion
type fireworksChatRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	MaxTokens   int32     `json:"max_tokens,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	TopP        float32   `json:"top_p,omitempty"`
	N           int32     `json:"n,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Stop        []string  `json:"stop,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response for a non-streaming request.
// It differs from the streaming response in the Choices list.
// This response uses the Message field whereas the streaming response uses the Delta field.
// https://readme.fireworks.ai/reference/createchatcompletion
type fireworksResponse struct {
	Choices []struct {
		Text    string `json:"text"`
		Message *struct {
			Content string `json:"content"`
		} `json:"message"`
		Index        int             `json:"index"`
		FinishReason string          `json:"finish_reason"`
		Logprobs     *types.Logprobs `json:"logprobs"`
	} `json:"choices"`
	Usage fireworksUsage `json:"usage"`
}

type fireworksStreamingResponse struct {
	Choices []struct {
		Text  string `json:"text"`
		Delta *struct {
			Content string `json:"content"`
		} `json:"delta"`
		Index        int             `json:"index"`
		FinishReason string          `json:"finish_reason"`
		Logprobs     *types.Logprobs `json:"logprobs"`
	} `json:"choices"`
	Usage fireworksUsage `json:"usage"`
}

type fireworksUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}
