package graphqlbackend

import "context"

type CompletionsResolver interface {
	Completions(ctx context.Context, args CompletionsArgs) (string, error)
}

type CompletionsArgs struct {
	Input CompletionsInput
	Fast  bool
}

type Message struct {
	Speaker string `json:"speaker"`
	Text    string `json:"text"`
}

type CompletionsInput struct {
	Messages          []Message `json:"messages"`
	Temperature       float64   `json:"temperature"`
	MaxTokensToSample int32     `json:"maxTokensToSample"`
	TopK              int32     `json:"topK"`
	TopP              int32     `json:"topP"`
}
